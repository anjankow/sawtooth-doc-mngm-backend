package http

import (
	"context"
	"doc-management/internal/config"
	"doc-management/internal/model"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

const (
	notQueryPrefix = "!"
)

type retrivedProposal struct {
	ProposalID     string   `json:"proposalID"`
	Name           string   `json:"name"`
	Category       string   `json:"category"`
	Content        string   `json:"content"`
	Author         string   `json:"author"`
	Signers        []string `json:"signers"`
	ProposedStatus string   `json:"proposedStatus"`
}

func (ser server) signProposal(w http.ResponseWriter, r *http.Request) {

	proposalID, signer, err := ser.readSignProposalParams(r)
	if err != nil {
		ser.badRequest(w, err.Error())
		return
	}

	if err := ser.app.SignProposal(r.Context(), proposalID, signer); err != nil {
		ser.serverError(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)

}

func (ser server) getAllProposals(w http.ResponseWriter, r *http.Request) {
	userID := normalize(r.URL.Query().Get("userID"))
	category := normalize(r.URL.Query().Get("category"))

	ser.logger.Info("getting all the proposals", zap.String("userID", userID), zap.String("category", category))

	if userID == "" && category == "" {
		ser.badRequest(w, ErrSearchTooBroad.Error())
		return
	}

	var proposals []model.Proposal
	var err error

	if strings.HasPrefix(userID, notQueryPrefix) {
		userID = strings.TrimPrefix(userID, notQueryPrefix)
		proposals, err = ser.app.GetToSignProposals(r.Context(), userID)

	} else {
		proposals, err = ser.app.GetUserProposals(r.Context(), userID)
	}

	if err != nil {
		ser.serverError(w, "getting the proposals failed: "+err.Error())
		return
	}

	proposToReturn := make([]retrivedProposal, len(proposals))
	for i, proposal := range proposals {
		proposToReturn[i] = retrivedProposal{
			ProposalID:     proposal.ProposalID,
			Name:           proposal.DocumentName,
			Category:       proposal.Category,
			Author:         proposal.ModificationAuthor,
			Signers:        proposal.Signers,
			ProposedStatus: proposal.ProposedStatus.String(),
		}
		// limit the content length to display
		if len(proposal.Content) > 80 {
			proposToReturn[i].Content = string(proposal.Content[:80])
		} else {
			proposToReturn[i].Content = string(proposal.Content)
		}
	}

	response, err := json.Marshal(proposToReturn)
	if err != nil {
		ser.serverError(w, "marshalling the response failed: "+err.Error())
		return
	}

	if _, err := w.Write(response); err != nil {
		ser.serverError(w, "failed to write the response: "+err.Error())
		return
	}
}

func (ser server) putProposal(w http.ResponseWriter, r *http.Request) {

	proposal, err := ser.readAddProposalParams(r)
	if err != nil {
		ser.badRequest(w, err.Error())
		return
	}

	// TODO: fix context to come from the client
	ctx, cancel := context.WithTimeout(context.Background(), config.GetRequestTimeout())
	defer cancel()

	if err := ser.app.AddProposal(ctx, proposal); err != nil {
		ser.serverError(w, "saving the proposal failed: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (ser server) readAddProposalParams(r *http.Request) (model.Proposal, error) {
	// max file size is 10MB
	if err := r.ParseMultipartForm(10e7); err != nil {
		return model.Proposal{}, errors.New("failed to parse the form: " + err.Error())
	}

	var err error
	params := mux.Vars(r)

	docName := normalize(params["docName"])
	if docName == "" {
		err = multierr.Append(err, errors.New("docName is missing"))
	}

	userID := normalize(r.FormValue("userID"))
	if userID == "" {
		err = multierr.Append(err, errors.New("userID is missing"))
	}

	category := normalize(r.FormValue("category"))
	docStatus := normalize(r.FormValue("docStatus"))

	var content []byte
	// if the doc is to be removed, there is no file content
	if docStatus != model.DocStatusRemoved.String() {

		file, handler, err := r.FormFile("docFile")
		if err != nil {
			err = multierr.Append(err, errors.New("failed to get the proposal file from form: "+err.Error()))
		}
		defer file.Close()

		if err != nil {
			return model.Proposal{}, err
		}

		bytes, err := ioutil.ReadAll(file)
		if err != nil {
			return model.Proposal{}, errors.New("failed to read the proposal file: " + err.Error())
		}

		if len(bytes) != int(handler.Size) {
			return model.Proposal{}, errors.New(fmt.Sprintf("upload error: size of received file: %v, size declared in the header: %v", len(bytes), handler.Size))
		}

		ser.logger.Info(fmt.Sprintf("received file: %s, size %v", handler.Filename, handler.Size))
		content = bytes

	} else {
		// just check previous validation errors
		if err != nil {
			return model.Proposal{}, err
		}
	}

	return model.Proposal{
		DocumentName:       docName,
		Category:           category,
		ModificationAuthor: userID,
		Content:            content,
		ProposedStatus:     model.DocStatus(docStatus),
	}, nil
}

func (ser server) readSignProposalParams(r *http.Request) (proposalID, signer string, err error) {
	params := mux.Vars(r)

	proposalID = normalize(params["proposalID"])
	if proposalID == "" {
		err = multierr.Append(err, errors.New("proposalID is missing"))
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		err = errors.New("can't read the request body: " + err.Error())
		return
	}

	var body struct {
		Signer string `json:"signer"`
	}

	if err = json.Unmarshal(bodyBytes, &body); err != nil {
		err = errors.New("invalid body: " + err.Error())
		return
	}

	return proposalID, body.Signer, nil
}

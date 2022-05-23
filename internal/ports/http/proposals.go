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

	"github.com/gorilla/mux"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

type retrivedProposal struct {
	ProposalID string   `json:"proposalID"`
	Name       string   `json:"name"`
	Category   string   `json:"category"`
	Content    string   `json:"content"`
	Author     string   `json:"author"`
	Signers    []string `json:"signers"`
}

func (ser server) getDocProposals(w http.ResponseWriter, r *http.Request) {
}

func (ser server) getAllProposals(w http.ResponseWriter, r *http.Request) {
	userID := normalize(r.URL.Query().Get("userID"))
	category := normalize(r.URL.Query().Get("category"))

	ser.logger.Info("getting all the proposals", zap.String("userID", userID), zap.String("category", category))

	proposals, err := ser.app.GetAllProposals(r.Context(), category, userID)
	if err != nil {
		ser.serverError(w, "getting the proposals failed: "+err.Error())
		return
	}

	proposToReturn := make([]retrivedProposal, len(proposals))
	for i, proposal := range proposals {
		proposToReturn[i] = retrivedProposal{
			ProposalID: proposal.TransactionID,
			Name:       proposal.DocumentName,
			Category:   proposal.Category,
			Author:     proposal.ModificationAuthor,
			Signers:    proposal.Signers,
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

	proposal, err := ser.readProposalParams(r)
	if err != nil {
		ser.badRequest(w, err.Error())
		return
	}

	// TODO: fix context to come from the client
	ctx, cancel := context.WithTimeout(context.Background(), config.GetRequestTimeout())
	defer cancel()

	if err := ser.app.SaveProposal(ctx, proposal); err != nil {
		ser.serverError(w, "saving the proposal failed: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (ser server) readProposalParams(r *http.Request) (model.Proposal, error) {
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

	return model.Proposal{
		DocumentID: model.DocumentID{
			DocumentName: docName,
			Category:     category,
		},
		ProposalContent: model.ProposalContent{
			ModificationAuthor: userID,
			Content:            bytes,
			ProposedStatus:     docStatus,
		},
	}, nil
}

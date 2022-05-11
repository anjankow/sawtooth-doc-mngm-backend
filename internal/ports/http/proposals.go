package http

import (
	"context"
	"doc-management/internal/config"
	"doc-management/internal/model"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/multierr"
)

func (ser server) getProposals(w http.ResponseWriter, r *http.Request) {

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

	if err := ser.app.SaveDocumentProposal(ctx, proposal); err != nil {
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
	userID := params["userID"]

	if userID == "" {
		err = multierr.Append(err, errors.New("userID is missing"))
	}

	docName := r.FormValue("docID")
	if docName == "" {
		err = multierr.Append(err, errors.New("docID is missing"))
	}
	category := r.FormValue("category")
	docStatus := r.FormValue("docStatus")

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

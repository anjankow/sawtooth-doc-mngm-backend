package http

import (
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

	fmt.Println(r.Header.Get("Content-Type"))
	// max file size is 10MB
	if err := r.ParseMultipartForm(10e7); err != nil {
		ser.badRequest(w, "failed to parse the form: "+err.Error())
		return
	}

	var validationErr error
	params := mux.Vars(r)
	userID := params["userID"]

	if userID == "" {
		validationErr = multierr.Append(validationErr, errors.New("userID is missing"))
	}

	docID := r.FormValue("docID")
	if docID == "" {
		validationErr = multierr.Append(validationErr, errors.New("docID is missing"))
	}
	file, handler, err := r.FormFile("docFile")
	if err != nil {
		err = multierr.Append(validationErr, errors.New("failed to get the proposal file from form: "+err.Error()))
	}
	defer file.Close()

	if validationErr != nil {
		ser.badRequest(w, validationErr.Error())
		return
	}

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		ser.badRequest(w, "failed to read the proposal file: "+err.Error())
		return
	}

	if len(bytes) != int(handler.Size) {
		ser.badRequest(w, fmt.Sprintf("upload error: size of received file: %v, size declared in the header: %v", len(bytes), handler.Size))
		return
	}

	ser.logger.Info(fmt.Sprintf("received file: %s, size %v", handler.Filename, handler.Size))

	document := model.Document{
		Author:     userID,
		DocumentID: docID,
		DocBytes:   bytes,
	}

	if err := ser.app.SaveDocumentProposal(r.Context(), document); err != nil {
		ser.serverError(w, "saving the proposal failed: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
}
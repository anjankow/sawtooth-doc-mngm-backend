package http

import (
	"doc-management/internal/model"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
)

func (ser server) getProposals(w http.ResponseWriter, r *http.Request) {

}

func (ser server) putProposal(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	userID := params["userID"]
	docID := params["docID"]

	if userID == "" {
		ser.badRequest(w, "invalid putProposal: userID is missing")
		return
	}
	if docID == "" {
		ser.badRequest(w, "invalid putProposal: docID is missing")
		return
	}

	// max file size is 10MB
	r.ParseMultipartForm(10e7)
	file, handler, err := r.FormFile("docFile")
	if err != nil {
		ser.badRequest(w, "failed to get the proposal file from form: "+err.Error())
		return
	}
	defer file.Close()

	ser.logger.Info(fmt.Sprintf("received file: %s, size %v", handler.Filename, handler.Size))

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		ser.badRequest(w, "failed to read the proposal file: "+err.Error())
		return
	}
	if len(bytes) != int(handler.Size) {
		ser.badRequest(w, fmt.Sprintf("upload error: size of received file: %v, size declared in the header: %v", len(bytes), handler.Size))
		return
	}

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

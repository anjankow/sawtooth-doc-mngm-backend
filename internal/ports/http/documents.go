package http

import (
	"doc-management/internal/model"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type retrivedDocVersion struct {
	ProposalID string `json:"proposalID"`
	Name       string `json:"name"`
	Category   string `json:"category"`
	Version    int    `json:"version"`
	Content    string `json:"content"`
	Author     string `json:"author"`
	Status     string `json:"status"`
}

func (r *retrivedDocVersion) assign(doc model.Document) {
	r.ProposalID = doc.ProposalID
	r.Name = doc.DocumentName
	r.Category = doc.Category
	r.Version = doc.Version
	r.Author = doc.Author
	r.Status = string(doc.Status)

	// limit the content length to display
	if len(doc.Content) > 80 {
		r.Content = string(doc.Content[:80])
	} else {
		r.Content = string(doc.Content)
	}
}

func (ser server) getDocVersions(w http.ResponseWriter, r *http.Request) {
	docName, category := ser.readGetDocVersionParams(r)
	ser.logger.Debug("getting doc " + docName + ", category " + category)

	if docName == "" || category == "" {
		ser.badRequest(w, "both docName and category need to be given")
		return
	}

	docs, err := ser.app.GetDocumentVersions(r.Context(), docName, category)
	if err != nil {
		ser.serverError(w, err.Error())
		return
	}

	ser.respondDocRequest(w, docs)
}

func (ser server) getDocuments(w http.ResponseWriter, r *http.Request) {

	author, signer := ser.readGetDocsParams(r)
	ser.logger.Debug("getting docs, author {" + author + "}, signer {" + signer + "}")

	if author == "" && signer == "" {
		ser.badRequest(w, ErrSearchTooBroad.Error())
		return
	}

	docs, err := ser.app.GetDocuments(r.Context(), author, signer)
	if err != nil {
		ser.serverError(w, err.Error())
		return
	}

	ser.respondDocRequest(w, docs)
}

func (ser server) readGetDocsParams(r *http.Request) (author, signer string) {
	queryParams := r.URL.Query()

	author = normalize(queryParams.Get("author"))
	signer = normalize(queryParams.Get("signer"))

	return
}

func (ser server) readGetDocVersionParams(r *http.Request) (docName, category string) {
	params := mux.Vars(r)

	docName = normalize(params["docName"])
	category = normalize(params["category"])

	return
}

func (ser server) respondDocRequest(w http.ResponseWriter, docs []model.Document) {
	retDocs := make([]retrivedDocVersion, len(docs))
	for i, doc := range docs {
		retDocs[i].assign(doc)
	}

	response, err := json.Marshal(retDocs)
	if err != nil {
		ser.serverError(w, "marshalling the response failed: "+err.Error())
		return
	}

	if _, err := w.Write(response); err != nil {
		ser.serverError(w, "failed to write the response: "+err.Error())
		return
	}
}

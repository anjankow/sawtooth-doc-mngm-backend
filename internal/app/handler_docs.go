package app

import (
	"context"
	"doc-management/internal/hashing"
	"doc-management/internal/model"
	"errors"
	"fmt"

	"go.uber.org/zap"
)

var (
	invalidContent = []byte("INVALID")
)

func (a App) GetDocumentVersions(ctx context.Context, docName, category string) ([]model.Document, error) {
	docs, err := a.blkchnClient.GetDocumentVersions(ctx, category, docName)
	if err != nil {
		return []model.Document{}, err
	}

	return a.fillAndVerifyDocContent(ctx, docs)
}

func (a App) GetDocuments(ctx context.Context, author, signer string) (docs []model.Document, err error) {
	if author == "" && signer == "" {
		err = errors.New("at least one of params author and signer needs to be given")
		return
	}

	if author != "" {
		docs, err = a.blkchnClient.GetDocumentsOfAuthor(ctx, author)

	} else {
		docs, err = a.blkchnClient.GetDocumentsSignedBy(ctx, signer)
	}

	return a.fillAndVerifyDocContent(ctx, docs)
}

func (a App) fillAndVerifyDocContent(ctx context.Context, docs []model.Document) (verified []model.Document, err error) {

	// TODO: parallelize
	for i, doc := range docs {

		if doc.Status == model.DocStatusRemoved {
			a.logger.Debug("skipping a doc, status: "+doc.Status.String(), zap.String("docName", doc.DocumentName), zap.String("category", doc.Category), zap.Int("version", doc.Version))
			continue
		}

		if doc.Status == model.DocStatusActive {
			docWithContent, err := a.db.FillDocumentContent(ctx, doc)
			if err != nil {
				a.logger.Error("error when getting the document content: "+err.Error(), zap.String("docName", doc.DocumentName), zap.String("category", doc.Category), zap.Int("version", doc.Version))
				docs[i].Content = []byte("ERROR")
				continue
			}

			dbContentHash := hashing.CalculateSHA512(string(docWithContent.Content))
			if dbContentHash == doc.ContentHash {
				docs[i] = docWithContent
				continue
			} else {
				a.invalidateDoc(doc)
			}
		}

		// for cases when the status is already invalid or the content hash doesn't match
		docs[i].Status = model.DocStatusInvalid
	}

	a.logger.Info(fmt.Sprint("content hash checked, returning ", len(verified), "/", len(docs), " documents"))

	return docs, nil
}

func (a App) invalidateDoc(doc model.Document) {
	// keep the invalid content in the db
	key, err := a.keyManager.GetAppKey()
	if err != nil {
		a.logger.Error("can't remove from blockchain, getting app key failed: " + err.Error())
		return
	}
	if _, err := a.blkchnClient.InvalidateDocumentVersion(context.Background(), doc, key.GetSigner()); err != nil {
		a.logger.Error("can't invalidate the doc: " + err.Error())
	}
}

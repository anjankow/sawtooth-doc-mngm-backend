package app

import (
	"context"
	"doc-management/internal/hashing"
	"doc-management/internal/model"
	"errors"
	"fmt"

	"go.uber.org/zap"
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
	for _, doc := range docs {
		docWithContent, err := a.db.FillDocContent(ctx, doc)
		if err != nil {
			a.logger.Error("error when getting the document content: "+err.Error(), zap.String("docName", doc.DocumentName), zap.String("category", doc.Category), zap.Int("version", doc.Version))
			continue
		}

		dbContentHash := hashing.CalculateSHA512(string(docWithContent.Content))
		if dbContentHash != doc.ContentHash {
			a.logger.Error("proposal content hash not matched! invalidating doc...", zap.String("docName", doc.DocumentName), zap.String("category", doc.Category), zap.Int("version", doc.Version), zap.String("dbHash", dbContentHash), zap.String("expectedHash", doc.ContentHash))

			// keep the invalid content in the db
			key, err := a.keyManager.GetAppKey()
			if err != nil {
				a.logger.Error("can't remove from blockchain, getting app key failed: " + err.Error())
				continue
			}
			if _, err := a.blkchnClient.InvalidateDocumentVersion(context.Background(), doc, key.GetSigner()); err != nil {
				a.logger.Error("can't invalidate the doc: " + err.Error())
			}

			continue
		}

		verified = append(verified, docWithContent)
	}

	a.logger.Info(fmt.Sprint("content hash checked, returning ", len(verified), "/", len(docs), " documents"))

	return verified, nil
}

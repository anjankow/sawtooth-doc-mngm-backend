package blockchain

import (
	"context"
	"doc-management/internal/model"

	"github.com/hyperledger/sawtooth-sdk-go/signing"
)

func (c Client) GetDocumentVersions(ctx context.Context, category string, docName string) (docs []model.Document, err error) {
	return
}

func (c Client) GetDocumentsOfAuthor(ctx context.Context, author string) (docs []model.Document, err error) {
	return
}

func (c Client) GetDocumentsSignedBy(ctx context.Context, signer string) (docs []model.Document, err error) {
	return
}

func (c Client) InvalidateDocumentVersion(ctx context.Context, doc model.Document, signer *signing.Signer) (transactionID string, err error) {
	return
}

func (c Client) SubmitDocumentVersion(ctx context.Context, doc model.Document, signer *signing.Signer) (transactionID string, err error) {
	return
}

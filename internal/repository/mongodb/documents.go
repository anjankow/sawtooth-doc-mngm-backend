package mongodb

import (
	"context"
	"doc-management/internal/model"
)

func (b Repository) InsertDocumentVersion(ctx context.Context, doc model.Document) error {
	return nil
}

func (b Repository) FillDocContent(ctx context.Context, doc model.Document) (model.Document, error) {
	return doc, nil
}

func (b Repository) RemoveDocumentVersion(ctx context.Context, doc model.Document) error {
	return nil
}

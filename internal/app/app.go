package app

import (
	"context"
	"doc-management/internal/model"
)

type App struct {
}

func (a App) SaveDocumentProposal(ctx context.Context, doc model.Document) error {
	return nil
}

package app

import (
	"context"
	"doc-management/internal/blockchain"
	"doc-management/internal/hashing"
	"doc-management/internal/keymanager"
	"doc-management/internal/model"
	"errors"
)

type App struct {
	client blockchain.Client
}

func (a App) SaveDocumentProposal(ctx context.Context, proposal model.Proposal) error {

	if proposal.Category == "" {
		proposal.Category = model.DefaultCategory
	}
	if proposal.ProposedStatus == "" {
		proposal.ProposedStatus = string(model.DocStatusAccepted)
	}
	if status := (model.DocStatus)(proposal.ProposedStatus); !status.IsValid() {
		return errors.New("invalid document status: " + proposal.ProposedStatus)
	}

	proposal.ContentHash = hashing.Calculate(proposal.Content)

	keys, err := keymanager.GenerateKeys()
	if err != nil {
		return err
	}

	return a.client.SubmitProposal(ctx, proposal, keys.GetSigner())
}

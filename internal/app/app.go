package app

import (
	"context"
	"doc-management/internal/blockchain"
	"doc-management/internal/config"
	"doc-management/internal/hashing"
	"doc-management/internal/keymanager"
	"doc-management/internal/model"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type App struct {
	client *blockchain.Client
	logger *zap.Logger
}

func NewApp(logger *zap.Logger) App {
	return App{
		client: blockchain.NewClient(logger, config.GetValidatorRestApiAddr()),
		logger: logger,
	}
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
	proposal.ProposalID = uuid.NewString()
	proposal.ContentHash = hashing.Calculate(proposal.Content)

	keys, err := keymanager.GenerateKeys()
	if err != nil {
		return err
	}

	a.logger.Info("submitting proposal", zap.String("docName", proposal.DocumentName), zap.String("author", proposal.ModificationAuthor), zap.String("proposalID", proposal.ProposalID))

	return a.client.SubmitProposal(ctx, proposal, keys.GetSigner())
}

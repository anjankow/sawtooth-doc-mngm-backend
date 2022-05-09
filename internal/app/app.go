package app

import (
	"context"
	"doc-management/internal/blockchain"
	"doc-management/internal/config"
	"doc-management/internal/hashing"
	"doc-management/internal/keymanager"
	"doc-management/internal/model"
	"doc-management/internal/repository/mongodb"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type App struct {
	blkchnClient *blockchain.Client
	logger       *zap.Logger
	db           mongodb.Repository
}

func NewApp(logger *zap.Logger, db mongodb.Repository) App {
	return App{
		blkchnClient: blockchain.NewClient(logger, config.GetValidatorRestApiAddr()),
		logger:       logger,
		db:           db,
	}
}

func (a App) SaveDocumentProposal(ctx context.Context, proposal model.Proposal) error {

	proposal = completeProposalData(proposal)
	if err := validateProposal(proposal); err != nil {
		return err
	}

	a.logger.Info("submitting proposal", zap.String("docName", proposal.DocumentName), zap.String("author", proposal.ModificationAuthor), zap.String("proposalID", proposal.ProposalID))

	keys, err := keymanager.GenerateKeys()
	if err != nil {
		return err
	}

	transactionID, err := a.blkchnClient.SubmitProposal(context.Background(), proposal, keys.GetSigner())
	if err != nil {
		a.logger.Error(err.Error())
		// return err
	}

	return a.db.InsertProposal(context.Background(), proposal, transactionID)
}

func validateProposal(proposal model.Proposal) error {
	if status := (model.DocStatus)(proposal.ProposedStatus); !status.IsValid() {
		return errors.New("invalid document status: " + proposal.ProposedStatus)
	}

	return nil
}

func completeProposalData(proposal model.Proposal) model.Proposal {
	if proposal.Category == "" {
		proposal.Category = model.DefaultCategory
	}
	if proposal.ProposedStatus == "" {
		proposal.ProposedStatus = string(model.DocStatusAccepted)
	}
	proposal.ProposalID = uuid.NewString()
	proposal.ContentHash = hashing.Calculate(proposal.Content)

	return proposal
}

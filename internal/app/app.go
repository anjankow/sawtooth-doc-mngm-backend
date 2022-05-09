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
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	submitTimeout = 10 * time.Second
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

	submitCtx, _ := context.WithTimeout(context.Background(), submitTimeout)
	transactionID, err := a.blkchnClient.SubmitProposal(submitCtx, proposal, keys.GetSigner())
	if err != nil {
		a.logger.Error(err.Error())
		// return err
	}
	a.logger.Info("proposal submitted, transaction ID: "+transactionID, zap.String("docName", proposal.DocumentName), zap.String("author", proposal.ModificationAuthor), zap.String("proposalID", proposal.ProposalID))

	dbCtx, cancel := context.WithTimeout(context.Background(), submitTimeout)
	defer cancel()

	return a.db.InsertProposal(dbCtx, proposal, transactionID)
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

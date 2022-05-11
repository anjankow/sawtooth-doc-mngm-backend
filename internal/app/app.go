package app

import (
	"context"
	"doc-management/internal/blockchain"
	"doc-management/internal/config"
	"doc-management/internal/keymanager"
	"doc-management/internal/model"
	"doc-management/internal/repository/mongodb"
	"time"

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

	proposal.Complete()
	if err := proposal.Validate(); err != nil {
		return err
	}

	keys, err := keymanager.GenerateKeys()
	if err != nil {
		return err
	}

	transaction, err := blockchain.NewProposalTransaction(proposal, keys.GetSigner())
	if err != nil {
		return err
	}

	a.logger.Info("submitting proposal", zap.String("docName", proposal.DocumentName), zap.String("author", proposal.ModificationAuthor), zap.String("proposalID", proposal.ProposalID))

	transactionID, err := a.blkchnClient.Submit(ctx, transaction)
	if err != nil {
		a.logger.Error(err.Error())
		// return err
	}
	a.logger.Info("proposal submitted, transaction ID: "+transactionID, zap.String("docName", proposal.DocumentName), zap.String("author", proposal.ModificationAuthor), zap.String("proposalID", proposal.ProposalID))

	return a.db.InsertProposal(ctx, proposal, transactionID)
}

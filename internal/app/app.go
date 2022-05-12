package app

import (
	"context"
	"doc-management/internal/blockchain"
	"doc-management/internal/config"
	"doc-management/internal/keymanager"
	"doc-management/internal/model"
	"doc-management/internal/repository/mongodb"
	"errors"
	"time"

	"go.uber.org/zap"
)

const (
	submitTimeout = 10 * time.Second
)

var ErrSearchTooBroad = errors.New("missing params to GET query")

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

func (a App) GetAllProposals(ctx context.Context, category string, userID string) ([]model.Proposal, error) {

	if userID != "" {
		// if the user is defined, get all no matter the category
		return a.db.GetUserProposals(ctx, userID)
	}

	if category == "" {
		// if the user is not given and category is not given too, return error
		return nil, ErrSearchTooBroad
	}

	return a.db.GetCategoryProposals(ctx, category)
}

func (a App) SaveProposal(ctx context.Context, proposal model.Proposal) error {

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
	proposal.TransactionID = transaction.GetTransactionID()

	a.logger.Info("submitting proposal", zap.String("docName", proposal.DocumentName), zap.String("author", proposal.ModificationAuthor), zap.String("transactionID", proposal.TransactionID))

	if err := a.db.InsertProposal(ctx, proposal, transaction.GetTransactionID()); err != nil {
		return err
	}

	if _, err = a.blkchnClient.Submit(ctx, transaction); err != nil {
		// remove the doc from the database
		a.logger.Debug("removing the proposal content from the database on error", zap.String("transactionID", proposal.TransactionID))
		_ = a.db.RemoveProposal(context.Background(), proposal)
		return err
	}

	a.logger.Info("proposal submitted, transaction ID: "+transaction.GetTransactionID(), zap.String("docName", proposal.DocumentName), zap.String("author", proposal.ModificationAuthor))

	return nil
}

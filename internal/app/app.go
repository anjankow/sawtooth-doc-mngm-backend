package app

import (
	"context"
	"doc-management/internal/blockchain"
	"doc-management/internal/config"
	"doc-management/internal/keymanager"
	"doc-management/internal/model"
	"doc-management/internal/repository/mongodb"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
)

const (
	submitTimeout = 10 * time.Second
)

var ErrSearchTooBroad = errors.New("missing params to GET query")

type App struct {
	blkchnClient *blockchain.Client
	keyManager   keymanager.KeyManager
	logger       *zap.Logger
	db           mongodb.Repository
}

func NewApp(logger *zap.Logger, db mongodb.Repository) App {
	return App{
		blkchnClient: blockchain.NewClient(logger, config.GetValidatorRestApiAddr()),
		keyManager:   keymanager.NewKeyManager(logger),
		logger:       logger,
		db:           db,
	}
}

func (a App) GetAllProposals(ctx context.Context, category string, userID string) (propos []model.Proposal, err error) {

	if userID == "" && category == "" {
		// if the user is not given and category is not given too, return error
		return nil, ErrSearchTooBroad
	}

	if userID != "" {
		// if the user is defined, get all no matter the category
		propos, err = a.db.GetUserProposals(ctx, userID)
	} else {
		propos, err = a.db.GetCategoryProposals(ctx, category)
	}

	if err != nil {
		return nil, err
	}

	a.logger.Info(fmt.Sprint("read ", len(propos), " proposals, checking the content hash..."))

	// verify the content hash against the blockchain
	verified := []model.Proposal{}
	for _, prop := range propos {
		// TODO: parallelize
		if err := a.blkchnClient.VerifyContentHash(ctx, prop); err != nil {

			if err == blockchain.ErrInvalidContentHash {
				a.logger.Error("proposal content hash not matched! removing...", zap.String("proposalID", prop.TransactionID), zap.String("dbHash", prop.ContentHash))

				_ = a.db.RemoveProposal(context.Background(), prop)
				_ = a.blkchnClient.RemoveProposal(context.Background(), prop)
				continue
			}

			a.logger.Warn("error when getting content hash from blockchain, skipping the check", zap.String("proposalID", prop.TransactionID), zap.String("dbHash", prop.ContentHash))
		}

		verified = append(verified, prop)
	}

	a.logger.Info(fmt.Sprint("content hash checked, returning ", len(verified), "/", len(propos), " proposals"))
	return verified, nil
}

func (a App) SaveProposal(ctx context.Context, proposal model.Proposal) error {

	proposal.Complete()
	if err := proposal.Validate(); err != nil {
		return err
	}

	keys, err := a.keyManager.GenerateKeys()
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

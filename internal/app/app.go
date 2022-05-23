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
	"strings"
	"time"

	"go.uber.org/zap"
)

const (
	submitTimeout  = 10 * time.Second
	notQueryPrefix = "!"
)

var ErrSearchTooBroad = errors.New("missing params to GET query")
var ErrProposalExists = errors.New("proposal already exists")

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
		if strings.HasPrefix(userID, notQueryPrefix) {
			propos, err = a.db.GetToSignProposals(ctx, strings.TrimPrefix(userID, notQueryPrefix))
		} else {
			propos, err = a.db.GetUserProposals(ctx, userID)
		}

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
				a.logger.Error("proposal content hash not matched! removing...", zap.String("proposalID", prop.ProposalID), zap.String("dbHash", prop.ContentHash))

				_ = a.db.RemoveProposal(context.Background(), prop)
				_ = a.blkchnClient.RemoveProposal(context.Background(), prop)
				continue
			}

			a.logger.Warn("error when getting content hash from blockchain, skipping the check", zap.String("proposalID", prop.ProposalID), zap.String("dbHash", prop.ContentHash))
		}

		verified = append(verified, prop)
	}

	a.logger.Info(fmt.Sprint("content hash checked, returning ", len(verified), "/", len(propos), " proposals"))
	return verified, nil
}

func (a App) SaveProposal(ctx context.Context, proposal model.Proposal) error {

	// fill in the missing fields with defaults and validate
	proposal.Complete()
	if err := proposal.Validate(); err != nil {
		return err
	}

	// // check if this proposal already exists
	// existingPropos, err := a.blkchnClient.GetProposals(ctx, proposal)
	// if err != nil {
	// 	a.logger.Error(fmt.Sprint("failed to get the existing proposals for the document ", proposal.DocumentName,
	// 		", category ", proposal.Category, "; proceeding with submitting the new proposal"))
	// }

	// for _, existing := range existingPropos {
	// 	if existing.ContentHash == proposal.ContentHash {
	// 		a.logger.Debug("proposal already exists", zap.String("category", proposal.Category), zap.String("docName", proposal.DocumentName), zap.String("contentHash", proposal.ContentHash))
	// 	}
	// 	return ErrProposalExists
	// }

	// TODO: use the user's keys obtained from the key manager
	keys, err := a.keyManager.GenerateKeys()
	if err != nil {
		return err
	}

	// craete a new blockchain transaction to get the transaction ID == proposal ID
	transaction, err := blockchain.NewProposalTransaction(proposal, keys.GetSigner())
	if err != nil {
		return err
	}

	a.logger.Info("submitting proposal", zap.String("docName", proposal.DocumentName), zap.String("author", proposal.ModificationAuthor), zap.String("proposalID", proposal.ProposalID))

	// first insert the transaction to the DB
	if err := a.db.InsertProposal(ctx, proposal, transaction.GetTransactionID()); err != nil {
		return err
	}

	// submit to blockchain only if all the previous operations succeeded, as this action is irreversible
	if _, err = a.blkchnClient.Submit(ctx, transaction); err != nil {
		// remove the doc from the database
		a.logger.Debug("removing the proposal content from the database on error", zap.String("proposalID", proposal.ProposalID))
		_ = a.db.RemoveProposal(context.Background(), proposal)
		return err
	}

	a.logger.Info("proposal submitted, transaction ID: "+transaction.GetTransactionID(), zap.String("docName", proposal.DocumentName), zap.String("author", proposal.ModificationAuthor))

	return nil
}

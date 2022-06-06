package app

import (
	"context"
	"doc-management/internal/blockchain"
	"doc-management/internal/blockchain/events"
	"doc-management/internal/config"
	"doc-management/internal/hashing"
	"doc-management/internal/keymanager"
	"doc-management/internal/model"
	"doc-management/internal/repository/mongodb"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
)

const (
	submitTimeout           = 10 * time.Second
	acceptingProcessTimeout = 10 * time.Second
)

var ErrProposalExists = errors.New("proposal already exists")

type App struct {
	blkchnClient *blockchain.Client
	keyManager   keymanager.KeyManager
	logger       *zap.Logger
	db           mongodb.Repository
	listener     *events.EventListener
}

func NewApp(logger *zap.Logger, db mongodb.Repository) App {
	return App{
		blkchnClient: blockchain.NewClient(logger, config.GetValidatorRestApiAddr()),
		listener:     events.NewEventListener(logger, config.GetValidatorHostname()),
		keyManager:   keymanager.NewKeyManager(logger),
		logger:       logger,
		db:           db,
	}
}

func (a App) Start() error {
	if err := a.listener.SetHandler("proposal_accepted", a.handleProposalAccepted); err != nil {
		return errors.New("failed to set the handler for 'proposal_accepted' event: " + err.Error())
	}

	if err := a.listener.Start(); err != nil {
		return errors.New("failed to start the listener: " + err.Error())
	}

	return nil
}

func (a App) Stop() {
	if err := a.listener.Stop(); err != nil {
		a.logger.Warn("error when stopping the listener: " + err.Error())
	}
}

func (a App) handleProposalAccepted(data []byte) error {
	// proposalID := string(data)
	// if proposalID == "" {
	// 	return errors.New("can't process accepted proposal, proposal ID is missing")
	// }

	// ctx, cancel := context.WithTimeout(context.Background(), acceptingProcessTimeout)
	// defer cancel()

	// // TODO: use the user's keys obtained from the key manager
	// keys, err := a.keyManager.GenerateKeys()
	// if err != nil {
	// 	return err
	// }

	// a.logger.Info("submitting accepted doc version", zap.String("proposalID", proposalID))

	// proposal, err := a.getProposalData(ctx, proposalID)
	// if err != nil {
	// 	return err
	// }

	// doc, err := a.blkchnClient.GetLatestDocVersion(ctx, proposal.Category, proposal.DocumentName)
	// if err != nil {
	// 	return errors.New("failed to get the latest doc version: " + err.Error())
	// }

	// newDoc := model.ConvertProposalToDoc(proposal, doc.Version+1)

	// if err := a.db.InsertDocumentVersion(ctx, newDoc); err != nil {
	// 	return errors.New("failed to insert the accepted doc into db: " + err.Error())
	// }

	// // submit to blockchain only if all the previous operations succeeded, as this action is irreversible
	// transactionID, err := a.blkchnClient.SubmitDocumentVersion(ctx, newDoc, keys.GetSigner())
	// if err != nil {
	// 	// remove the doc from the database
	// 	a.logger.Debug("removing the doc content from the database on error", zap.String("proposalID", proposal.ProposalID), zap.Error(err))
	// 	_ = a.db.RemoveDocumentVersion(context.Background(), newDoc.Category, newDoc.DocumentName, newDoc.Version)
	// 	return err
	// }

	// a.logger.Info("new doc version saved, transaction ID: "+transactionID, zap.String("docName", newDoc.DocumentName), zap.String("author", newDoc.Author))

	return nil
}

func (a App) GetDocumentVersions(ctx context.Context, docName, category string) ([]model.Document, error) {
	return []model.Document{
		{
			DocumentName: docName,
			Category:     category,
			Author:       "sdkfm",
			Content:      []byte("strfdsgdf"),
			Version:      1,
			Status:       "safs",
			ProposalID:   "21321",
		},
		{
			DocumentName: docName,
			Category:     category,
			Author:       "sdkfmaa",
			Content:      []byte("asdasdadaaaaaa"),
			Version:      3,
			Status:       "safs",
			ProposalID:   "21321s",
		},
		{
			DocumentName: docName,
			Category:     category,
			Author:       "sdkfmaa11",
			Content:      []byte("as2dasdadaaaaaa"),
			Version:      2,
			Status:       "safs",
			ProposalID:   "321321s",
		},
	}, nil
}

func (a App) GetDocuments(ctx context.Context, docName, category string) ([]model.Document, error) {
	return []model.Document{
		{
			DocumentName: docName,
			Category:     category,
			Author:       "sdkfm",
			Content:      []byte("strfdsgdf"),
			Version:      1,
			Status:       "safs",
			ProposalID:   "21321",
		},
		{
			DocumentName: "aaa",
			Category:     category,
			Author:       "sdkfmaa",
			Content:      []byte("asdasdadaaaaaa"),
			Version:      3,
			Status:       "safs",
			ProposalID:   "21321s",
		},
		{
			DocumentName: docName,
			Category:     "aaa",
			Author:       "sdkfmaa11",
			Content:      []byte("as2dasdadaaaaaa"),
			Version:      2,
			Status:       "safs",
			ProposalID:   "321321s",
		},
	}, nil
}

func (a App) getProposalData(ctx context.Context, proposalID string) (model.Proposal, error) {
	proposal, err := a.blkchnClient.GetProposal(ctx, proposalID)
	if err != nil {
		return model.Proposal{}, err
	}

	filled, err := a.fillAndVerifyContent(ctx, []model.Proposal{proposal})
	if err != nil {
		return model.Proposal{}, err
	}

	if len(filled) < 1 {
		return model.Proposal{}, errors.New("can't accept the proposal, content verification failed, proposalID: " + proposalID)
	}

	return filled[0], nil

}

// SignProposal user ID refers to a user who signs the proposal
func (a App) SignProposal(ctx context.Context, proposalID string, userID string) error {
	// TODO: use the user's keys obtained from the key manager
	keys, err := a.keyManager.GenerateKeys()
	if err != nil {
		return err
	}

	transactionID, err := a.blkchnClient.SignProposal(ctx, proposalID, userID, keys.GetSigner())
	if err != nil {
		return err
	}

	a.logger.Debug("proposal signed, transaction ID: " + transactionID)
	return nil
}

func (a App) GetToSignProposals(ctx context.Context, userID string) (propos []model.Proposal, err error) {
	propos, err = a.blkchnClient.GetActiveProposals(ctx)
	if err != nil {
		return
	}

	var proposByOthers []model.Proposal
	for _, p := range propos {
		if p.ModificationAuthor != userID {
			proposByOthers = append(proposByOthers, p)
		}
	}
	a.logger.Info(fmt.Sprint("to sign: ", len(proposByOthers), "/", len(propos)), zap.String("userID", userID))

	return a.fillAndVerifyContent(ctx, proposByOthers)
}

func (a App) fillAndVerifyContent(ctx context.Context, propos []model.Proposal) ([]model.Proposal, error) {
	var verified []model.Proposal
	// TODO: parallelize
	for _, p := range propos {
		pWithContent, err := a.db.FillProposalContent(ctx, p)
		if err != nil {
			a.logger.Error("error when getting the proposal content: "+err.Error(), zap.String("proposalID", p.ProposalID))
			continue
		}

		dbContentHash := hashing.CalculateSHA512(string(pWithContent.Content))
		if dbContentHash != p.ContentHash {
			a.logger.Error("proposal content hash not matched! removing...", zap.String("proposalID", p.ProposalID), zap.String("dbHash", dbContentHash), zap.String("expectedHash", p.ContentHash))

			if err := a.db.RemoveProposal(context.Background(), p); err != nil {
				a.logger.Error("failed to remove the proposal from db: " + err.Error())
			}

			key, err := a.keyManager.GetAppKey()
			if err != nil {
				a.logger.Error("can't remove from blockchain, getting app key failed: " + err.Error())
				continue
			}
			if _, err := a.blkchnClient.RemoveProposal(context.Background(), p.ProposalID, key.GetSigner()); err != nil {
				a.logger.Error("can't remove the proposal from blockchain: " + err.Error())
			}

			continue
		}

		verified = append(verified, pWithContent)
	}

	a.logger.Info(fmt.Sprint("content hash checked, returning ", len(verified), "/", len(propos), " proposals"))

	return verified, nil
}

func (a App) GetUserProposals(ctx context.Context, userID string) (propos []model.Proposal, err error) {
	propos, err = a.blkchnClient.GetUserProposals(ctx, userID)
	if err != nil {
		// if this user doesn't yet exist on the blockchain
		// he simply hasn't created anything = no error
		if err == blockchain.ErrNotFound {
			return propos, nil
		}

		return propos, err
	}

	return a.fillAndVerifyContent(ctx, propos)
}

func (a App) AddProposal(ctx context.Context, proposal model.Proposal) error {

	// fill in the missing fields with defaults and validate
	proposal.Complete()
	if err := proposal.Validate(); err != nil {
		return err
	}

	// check if this proposal already exists
	existingPropos, err := a.blkchnClient.GetDocProposals(ctx, proposal.Category, proposal.DocumentName)
	if err != nil && err != blockchain.ErrNotFound {
		a.logger.Error(fmt.Sprint("failed to get the existing proposals for the document ", proposal.DocumentName,
			", category ", proposal.Category, "; proceeding with submitting the new proposal"))
	}

	for _, existing := range existingPropos {
		if existing.ContentHash == proposal.ContentHash {
			a.logger.Debug("proposal already exists", zap.String("category", proposal.Category), zap.String("docName", proposal.DocumentName), zap.String("existingProposalID", existing.ProposalID))
			return ErrProposalExists
		}
	}

	// TODO: use the user's keys obtained from the key manager
	keys, err := a.keyManager.GenerateKeys()
	if err != nil {
		return err
	}

	a.logger.Info("submitting proposal", zap.String("docName", proposal.DocumentName), zap.String("author", proposal.ModificationAuthor), zap.String("proposalID", proposal.ProposalID))

	// first insert the transaction to the DB
	if err := a.db.InsertProposal(ctx, proposal); err != nil {
		return err
	}

	// submit to blockchain only if all the previous operations succeeded, as this action is irreversible
	transactionID, err := a.blkchnClient.SubmitProposal(ctx, proposal, keys.GetSigner())
	if err != nil {
		// remove the doc from the database
		a.logger.Debug("removing the proposal content from the database on error", zap.String("proposalID", proposal.ProposalID))
		_ = a.db.RemoveProposal(context.Background(), proposal)
		return err
	}

	a.logger.Info("proposal submitted, transaction ID: "+transactionID, zap.String("docName", proposal.DocumentName), zap.String("author", proposal.ModificationAuthor))

	return nil
}

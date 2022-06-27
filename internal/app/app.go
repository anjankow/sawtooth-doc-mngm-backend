package app

import (
	"context"
	"doc-management/internal/blockchain"
	"doc-management/internal/blockchain/events"
	"doc-management/internal/config"
	"doc-management/internal/model"
	"doc-management/internal/repository/mongodb"
	"doc-management/internal/signkeys"
	"doc-management/internal/usermanager"
	"errors"
	"time"

	"go.uber.org/zap"
)

const (
	submitTimeout           = 10 * time.Second
	acceptingProcessTimeout = 10 * time.Second
)

type App struct {
	blkchnClient *blockchain.Client
	userManager  usermanager.UserManager
	logger       *zap.Logger
	db           mongodb.Repository
	listener     *events.EventListener

	appKeys signkeys.UserKeys
}

func NewApp(logger *zap.Logger, db mongodb.Repository) App {

	return App{
		blkchnClient: blockchain.NewClient(logger, config.GetValidatorRestAPIAddr()),
		listener:     events.NewEventListener(logger, config.GetValidatorAddr()),
		logger:       logger,
		db:           db,
		// initialize when starting the app
		userManager: usermanager.UserManager{},
		appKeys:     signkeys.UserKeys{},
	}
}

func (a *App) Start() error {
	userManag, err := usermanager.NewUserManager(config.GetTenantID(), config.GetClientID(), config.GetMsExtensionID(), config.GetAppSecret())
	if err != nil {
		return errors.New("failed to initialize user manager: " + err.Error())
	}
	a.userManager = userManag
	ctx, cancel := context.WithTimeout(context.Background(), submitTimeout)
	defer cancel()
	appKeys, err := userManag.InitAndReadAppKeys(ctx, config.GetAppUserID())
	a.appKeys = appKeys
	a.logger.Info("app keys initialized", zap.String("publicKeyShort", appKeys.PublicKey.AsHex()[:20]))

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
	proposalID := string(data)
	if proposalID == "" {
		return errors.New("can't process accepted proposal, proposal ID is missing")
	}

	ctx, cancel := context.WithTimeout(context.Background(), acceptingProcessTimeout)
	defer cancel()

	a.logger.Info("submitting accepted doc version", zap.String("proposalID", proposalID))

	proposal, err := a.getProposalData(ctx, proposalID)
	if err != nil {
		return err
	}

	docs, err := a.blkchnClient.GetDocumentVersions(ctx, proposal.Category, proposal.DocumentName)
	if err != nil {
		return errors.New("failed to get the doc versions: " + err.Error())
	}

	newVersion := model.GetNextDocVersion(docs)

	newDoc := model.NewDocumentFromProposal(proposal, newVersion)

	if err := a.db.InsertDocumentVersion(ctx, newDoc); err != nil {
		return errors.New("failed to insert the accepted doc into db: " + err.Error())
	}

	// submit to blockchain only if all the previous operations succeeded, as this action is irreversible
	transactionID, err := a.blkchnClient.SubmitDocumentVersion(ctx, newDoc, a.appKeys.GetSigner())
	if err != nil {
		// remove the doc from the database
		a.logger.Debug("removing the doc content from the database on error", zap.String("proposalID", proposal.ProposalID), zap.Error(err))
		_ = a.db.RemoveDocumentVersion(context.Background(), newDoc)
		return err
	}

	a.logger.Info("new doc version saved, transaction ID: "+transactionID, zap.String("docName", newDoc.DocumentName), zap.String("author", newDoc.Author))

	return nil
}

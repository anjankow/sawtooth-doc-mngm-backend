package app

import (
	"context"
	"doc-management/internal/blockchain"
	"doc-management/internal/hashing"
	"doc-management/internal/model"
	"errors"
	"fmt"

	"go.uber.org/zap"
)

var ErrProposalExists = errors.New("proposal already exists")

func (a App) getProposalData(ctx context.Context, proposalID string) (model.Proposal, error) {
	proposal, err := a.blkchnClient.GetProposal(ctx, proposalID)
	if err != nil {
		return model.Proposal{}, err
	}

	filled, err := a.fillAndVerifyProposalContent(ctx, []model.Proposal{proposal})
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

	user, err := a.userManager.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	transactionID, err := a.blkchnClient.SignProposal(ctx, proposalID, userID, user.Keys.GetSigner())
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

	return a.fillAndVerifyProposalContent(ctx, proposByOthers)
}

func (a App) fillAndVerifyProposalContent(ctx context.Context, propos []model.Proposal) ([]model.Proposal, error) {
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

			if _, err := a.blkchnClient.RemoveProposal(context.Background(), p.ProposalID, a.appKeys.GetSigner()); err != nil {
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

	return a.fillAndVerifyProposalContent(ctx, propos)
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

	user, err := a.userManager.GetUserByID(ctx, proposal.ModificationAuthor)
	if err != nil {
		return err
	}

	a.logger.Info("submitting proposal", zap.String("docName", proposal.DocumentName), zap.String("author", proposal.ModificationAuthor), zap.String("proposalID", proposal.ProposalID))

	// first insert the transaction to the DB
	if err := a.db.InsertProposal(ctx, proposal); err != nil {
		return err
	}

	// submit to blockchain only if all the previous operations succeeded, as this action is irreversible
	transactionID, err := a.blkchnClient.SubmitProposal(ctx, proposal, user.Keys.GetSigner())
	if err != nil {
		// remove the doc from the database
		a.logger.Debug("removing the proposal content from the database on error", zap.String("proposalID", proposal.ProposalID))
		_ = a.db.RemoveProposal(context.Background(), proposal)
		return err
	}

	a.logger.Info("proposal submitted, transaction ID: "+transactionID, zap.String("docName", proposal.DocumentName), zap.String("author", proposal.ModificationAuthor))

	return nil
}

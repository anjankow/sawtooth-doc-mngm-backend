package mongodb

import (
	"context"
	"doc-management/internal/config"
	"doc-management/internal/hashing"
	"doc-management/internal/model"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

const (
	proposalsCollection = "proposals"
)

func (b Repository) InsertProposal(ctx context.Context, proposal model.Proposal, transactionID string) error {

	// category name is the collection name
	coll := b.client.Database(config.GetDatabaseName()).Collection(proposalsCollection)

	storedPropos := Proposal{
		TransactionID:  proposal.TransactionID,
		Category:       proposal.Category,
		Name:           proposal.DocumentName,
		Author:         proposal.ModificationAuthor,
		Content:        proposal.Content,
		Signers:        []string{},
		ProposedStatus: proposal.ProposedStatus,
	}

	data, err := bson.Marshal(storedPropos)
	if err != nil {
		return errors.New("failed to marshal the proposal: " + err.Error())
	}

	result, err := coll.InsertOne(ctx, data)
	if err != nil {
		return errors.New("failed to insert a new proposal: " + err.Error())
	}
	if result.InsertedID != proposal.TransactionID {
		return errors.New(fmt.Sprint("inserted a proposal with unexpected ID: ", result.InsertedID, "; expected: ", proposal.TransactionID))
	}

	return nil

}

func (b Repository) RemoveProposal(ctx context.Context, proposal model.Proposal) error {

	// category name is the collection name
	coll := b.client.Database(config.GetDatabaseName()).Collection(proposalsCollection)

	filter := bson.M{
		"_id": proposal.TransactionID,
	}
	result, err := coll.DeleteOne(ctx, filter)

	if err != nil {
		b.logger.Debug("failed to remove the proposal: "+err.Error(), zap.String("docName", proposal.DocumentName), zap.String("transactionID", proposal.TransactionID))
		return err
	}

	if result.DeletedCount == 0 {
		b.logger.Debug("trying to remove non existing proposal", zap.String("docName", proposal.DocumentName), zap.String("transactionID", proposal.TransactionID))
	}

	return nil

}

func (b Repository) GetToSignProposals(ctx context.Context, userID string) ([]model.Proposal, error) {

	filter := bson.M{
		"author": bson.M{
			"$ne": userID,
		},
	}

	return b.getProposals(ctx, filter)
}

func (b Repository) getProposals(ctx context.Context, filter bson.M) ([]model.Proposal, error) {
	coll := b.client.Database(config.GetDatabaseName()).Collection(proposalsCollection)

	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		return nil, errors.New("failed to find the user proposals: " + err.Error())
	}

	var storedPropos []Proposal
	if err := cursor.All(ctx, &storedPropos); err != nil {
		return nil, errors.New("failed to get all proposals from the cursor: " + err.Error())
	}

	var modelPropos = make([]model.Proposal, len(storedPropos))
	for i, stored := range storedPropos {
		// content hash is always recalculated everytime the data is retrieved
		contentHash := hashing.Calculate(stored.Content)

		modelPropos[i] = model.Proposal{
			DocumentID: model.DocumentID{
				DocumentName: stored.Name,
				Category:     stored.Category,
			},
			ProposalContent: model.ProposalContent{
				TransactionID:      stored.TransactionID,
				ModificationAuthor: stored.Author,
				Content:            stored.Content,
				ContentHash:        contentHash,
				ProposedStatus:     stored.ProposedStatus,
			},
			Signers: stored.Signers,
		}
	}

	return modelPropos, nil
}

func (b Repository) GetUserProposals(ctx context.Context, userID string) ([]model.Proposal, error) {

	filter := bson.M{
		"author": userID,
	}

	return b.getProposals(ctx, filter)
}

func (b Repository) GetCategoryProposals(ctx context.Context, category string) ([]model.Proposal, error) {
	filter := bson.M{
		"category": category,
	}

	return b.getProposals(ctx, filter)
}

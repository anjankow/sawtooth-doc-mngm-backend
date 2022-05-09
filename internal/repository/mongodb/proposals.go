package mongodb

import (
	"context"
	"doc-management/internal/config"
	"doc-management/internal/model"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (b Repository) insertNewDocWithProposal(ctx context.Context, proposal model.Proposal, transactionID string, collection *mongo.Collection) error {
	doc := StoredDocument{
		DocumentName: proposal.DocumentName,
		Proposals: []Proposal{
			{
				ProposalID:    proposal.ProposalID,
				TransactionID: transactionID,
				Content:       proposal.Content,
				Author:        proposal.ModificationAuthor,
			},
		},
		Versions: []DocVersion{},
	}

	data, err := bson.Marshal(doc)
	if err != nil {
		return errors.New("failed to marshal the document: " + err.Error())
	}

	result, err := collection.InsertOne(ctx, data)
	if err != nil {
		return errors.New("failed to insert a new doc proposal: " + err.Error())
	}
	if result.InsertedID != proposal.DocumentName {
		return errors.New(fmt.Sprint("inserted a document with unexpected ID: ", result.InsertedID, "; expected: ", proposal.DocumentName))
	}

	return nil
}

func (b Repository) updateDocWithProposal(ctx context.Context, proposal model.Proposal, transactionID string, collection *mongo.Collection, queryResult *mongo.SingleResult) error {
	var doc StoredDocument
	if err := queryResult.Decode(&doc); err != nil {
		return errors.New("failed to decode the document: " + err.Error())
	}

	doc.Proposals = append(doc.Proposals, Proposal{
		ProposalID:    proposal.ProposalID,
		TransactionID: transactionID,
		Content:       proposal.Content,
		Author:        proposal.ModificationAuthor,
	})

	data, err := bson.Marshal(doc)
	if err != nil {
		return errors.New("failed to marshal back the document: " + err.Error())
	}

	filter := bson.M{
		"_id": proposal.DocumentName,
	}
	result, err := collection.ReplaceOne(ctx, filter, data)
	if err != nil {
		return errors.New("failed to update the document with a new proposal: " + err.Error())
	}

	if result.ModifiedCount != 1 {
		return errors.New(fmt.Sprint("failed to update the document with a new proposal, modified count = ", result.ModifiedCount))
	}

	return nil
}

func (b Repository) InsertProposal(ctx context.Context, proposal model.Proposal, transactionID string) error {

	// category name is the collection name
	coll := b.client.Database(config.GetDatabaseName()).Collection(proposal.Category)

	filter := bson.M{
		"_id": proposal.DocumentName,
	}
	result := coll.FindOne(ctx, filter)
	if result.Err() == nil {
		return b.updateDocWithProposal(ctx, proposal, transactionID, coll, result)
	}
	if result.Err() == mongo.ErrNoDocuments {
		return b.insertNewDocWithProposal(ctx, proposal, transactionID, coll)
	}

	// else there was an error of the find operation
	return result.Err()
}

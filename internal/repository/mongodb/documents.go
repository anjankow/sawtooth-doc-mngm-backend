package mongodb

import (
	"context"
	"doc-management/internal/config"
	"doc-management/internal/model"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
)

const (
	docsCollection = "documents"
)

func getDocID(doc model.Document) string {
	return doc.Category + ";" + doc.DocumentName + ";" + fmt.Sprint(doc.Version)
}

type storedDoc struct {
	DocID   string `bson:"_id" json:"id"`
	Content []byte
}

func (b Repository) InsertDocumentVersion(ctx context.Context, doc model.Document) error {
	coll := b.client.Database(config.GetDatabaseName()).Collection(docsCollection)

	toInsert := storedDoc{
		DocID:   getDocID(doc),
		Content: doc.Content,
	}

	data, err := bson.Marshal(toInsert)
	if err != nil {
		return errors.New("failed to marshal the doc: " + err.Error())
	}

	_, err = coll.InsertOne(ctx, data)
	if err != nil {
		return errors.New("failed to insert a new doc: " + err.Error())
	}

	return nil
}

func (b Repository) FillDocumentContent(ctx context.Context, doc model.Document) (model.Document, error) {
	coll := b.client.Database(config.GetDatabaseName()).Collection(docsCollection)

	filter := bson.M{
		"_id": getDocID(doc),
	}

	result := coll.FindOne(ctx, filter)
	if result.Err() != nil {
		return model.Document{}, errors.New("failed to find the doc: " + result.Err().Error())
	}

	var fromDB storedDoc
	if err := result.Decode(&fromDB); err != nil {
		return model.Document{}, errors.New("failed to decode the doc: " + err.Error())
	}

	doc.Content = fromDB.Content
	return doc, nil
}

func (b Repository) RemoveDocumentVersion(ctx context.Context, doc model.Document) error {
	coll := b.client.Database(config.GetDatabaseName()).Collection(docsCollection)

	filter := bson.M{
		"_id": getDocID(doc),
	}

	result, err := coll.DeleteOne(ctx, filter)
	if err != nil {
		return errors.New("failed to remove the doc: " + err.Error())
	}

	if result.DeletedCount == 0 {
		b.logger.Info("document not found, can't be deleted: " + getDocID(doc))
	}

	return nil
}

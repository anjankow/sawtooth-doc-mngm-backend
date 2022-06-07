package blockchain

import (
	"context"
	"doc-management/internal/blockchain/doctrackerfamily"
	"doc-management/internal/model"
	"errors"
	"fmt"
	"net/url"

	"github.com/hyperledger/sawtooth-sdk-go/signing"
	"go.uber.org/zap"
)

func (c Client) GetDocumentVersions(ctx context.Context, category string, docName string) (data []model.Document, err error) {

	addr := doctrackerfamily.GetDocAddress(category, docName)
	filter := url.Values{}
	filter.Set("address", addr)

	url := fmt.Sprintf("%s?%s", stateAPI, filter.Encode())

	response, err := c.sendRequest(ctx, url, nil, "")
	if err != nil {
		return
	}

	unmarshalled, err := unmarshalDataList(response)
	if err != nil {
		return
	}

	data = make([]model.Document, len(unmarshalled))
	for i, payload := range unmarshalled {
		if err := unmarshalStatePayload(&data[i], string(payload)); err != nil {
			c.logger.Error("get doc state: failed to unmarshal the payload: " + err.Error())
			continue
		}

	}

	return data, nil
}

func (c Client) GetDocumentsOfAuthor(ctx context.Context, author string) (docs []model.Document, err error) {
	user, err := c.getUserData(ctx, author)
	if err != nil {
		return
	}

	if len(user.Authored) == 0 {
		c.logger.Debug("user " + author + " has no authored documents")
		return
	}

	docs, err = c.getDocsData(ctx, user.Authored)
	if err != nil {
		return
	}
	if len(docs) == 0 {
		return docs, errors.New("failed to get any document authored by " + author)
	}

	return docs, nil
}

func (c Client) getDocsData(ctx context.Context, addresses []string) ([]model.Document, error) {

	data := make([]model.Document, len(addresses))
	for i, addr := range addresses {
		// TODO: parallelize
		url := fmt.Sprintf("%s/%s", stateAPI, addr)
		response, err := c.sendRequest(ctx, url, nil, "")
		if err != nil {
			c.logger.Error("failed to get the state of doc: "+err.Error(), zap.String("address", addr))
			continue
		}

		if err := unmarshalStatePayload(&data[i], response); err != nil {
			c.logger.Error("failed to unmarshal the state of doc: "+err.Error(), zap.String("address", addr))
			continue
		}
	}

	return data, nil
}

func (c Client) getUserData(ctx context.Context, user string) (doctrackerfamily.UserData, error) {
	addr := doctrackerfamily.GetUserAddress(user)

	url := fmt.Sprintf("%s/%s", stateAPI, addr)
	response, err := c.sendRequest(ctx, url, nil, "")
	if err != nil {
		return doctrackerfamily.UserData{}, err
	}

	var payload doctrackerfamily.UserData
	if err := unmarshalStatePayload(&payload, response); err != nil {
		return doctrackerfamily.UserData{}, errors.New("get docs of user: " + err.Error())
	}

	return payload, nil
}

func (c Client) GetDocumentsSignedBy(ctx context.Context, signer string) (docs []model.Document, err error) {
	user, err := c.getUserData(ctx, signer)
	if err != nil {
		return
	}

	if len(user.Signed) == 0 {
		c.logger.Debug("user " + signer + " hasn't signed any documents")
		return
	}

	docs, err = c.getDocsData(ctx, user.Signed)
	if err != nil {
		return
	}
	if len(docs) == 0 {
		return docs, errors.New("failed to get any document signed by " + signer)
	}

	return docs, nil
}

func (c Client) InvalidateDocumentVersion(ctx context.Context, doc model.Document, signer *signing.Signer) (transactionID string, err error) {
	docDataAddress := doctrackerfamily.GetDocVersionAddress(doc)

	payload := make(map[interface{}]interface{})
	payload["action"] = doctrackerfamily.ActionInvalidate
	payload["address"] = docDataAddress

	transaction, err := NewTransaction(payload, signer, []string{docDataAddress}, doctrackerfamily.FamilyName, doctrackerfamily.FamilyVersion)
	if err != nil {
		return "", errors.New("failed to invalidate a document version: " + err.Error())
	}

	return c.submitTransaction(ctx, transaction, signer)

}

func (c Client) SubmitDocumentVersion(ctx context.Context, doc model.Document, signer *signing.Signer) (transactionID string, err error) {
	docDataAddress := doctrackerfamily.GetDocVersionAddress(doc)
	authorAddress := doctrackerfamily.GetUserAddress(doc.Author)
	signerAddresses := make([]string, len(doc.Signers))
	for i, signer := range doc.Signers {
		signerAddresses[i] = doctrackerfamily.GetUserAddress(signer)
	}

	payload := make(map[interface{}]interface{})
	payload["action"] = doctrackerfamily.ActionInsert
	payload["proposalID"] = doc.ProposalID
	payload["category"] = doc.Category
	payload["documentName"] = doc.DocumentName
	payload["contentHash"] = doc.ContentHash
	payload["status"] = doc.Status
	payload["author"] = doc.Author
	payload["version"] = doc.Version
	payload["signers"] = doc.Signers

	transaction, err := NewTransaction(payload, signer, append(signerAddresses, []string{authorAddress, docDataAddress}...), doctrackerfamily.FamilyName, doctrackerfamily.FamilyVersion)
	if err != nil {
		return "", errors.New("failed to create a new document version transaction: " + err.Error())
	}

	return c.submitTransaction(ctx, transaction, signer)
}

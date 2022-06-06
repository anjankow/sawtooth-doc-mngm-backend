package blockchain

import (
	"context"
	"net/url"

	"doc-management/internal/blockchain/proposalfamily"
	propfamily "doc-management/internal/blockchain/proposalfamily"
	"doc-management/internal/blockchain/settingsfamily"
	"doc-management/internal/model"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/fxamacker/cbor"
	"github.com/hyperledger/sawtooth-sdk-go/signing"
	"go.uber.org/zap"
)

type action string

const (
	actionInsert action = "insert"
	actionVote   action = "vote"
	actionDelete action = "delete"
)

func (c Client) RemoveProposal(ctx context.Context, proposalID string, signer *signing.Signer) (transactionID string, err error) {
	proposalAddr := propfamily.GetProposalAddressFromID(proposalID)

	docAddr, authorAddr, err := c.getAddrByProposalID(ctx, proposalID)
	if err != nil {
		c.logger.Warn("failed to get the doc addr from proposal ID by getting the state: " + err.Error())
	}

	payload := make(map[interface{}]interface{})
	payload["action"] = actionDelete
	payload["proposalID"] = proposalID

	transaction, err := NewTransaction(payload, signer, []string{proposalAddr, authorAddr, docAddr}, propfamily.FamilyName, propfamily.FamilyVersion)
	if err != nil {
		return "", errors.New("failed to create a proposal sign transaction: " + err.Error())
	}

	return c.submitTransaction(ctx, transaction, signer)
}

func (c Client) getAddrByProposalID(ctx context.Context, proposalID string) (docAddr string, authorAddr string, err error) {
	prop, err := c.getProposalState(ctx, proposalID)
	if err != nil {
		return "", "", err
	}

	docAddr = propfamily.GetDocAddress(prop.Category, prop.DocName)
	authorAddr = propfamily.GetUserAddress(prop.Author)

	c.logger.Debug("received proposal info: " + prop.Category + ", " + prop.DocName + ", address: " + docAddr)
	return
}

func (c Client) SignProposal(ctx context.Context, proposalID string, userID string, signer *signing.Signer) (transactionID string, err error) {
	proposalAddr := propfamily.GetProposalAddressFromID(proposalID)
	voterAddr := propfamily.GetUserAddress(userID)
	settingAddr := settingsfamily.GetAddress("proposal.vote.threshold")

	docAddr, authorAddr, err := c.getAddrByProposalID(ctx, proposalID)
	if err != nil {
		c.logger.Warn("failed to get the doc addr from proposal ID by getting the state: " + err.Error())
	}

	payload := make(map[interface{}]interface{})
	payload["action"] = actionVote
	payload["proposalID"] = proposalID
	payload["voter"] = userID

	transaction, err := NewTransaction(payload, signer, []string{proposalAddr, voterAddr, authorAddr, docAddr, settingAddr}, propfamily.FamilyName, propfamily.FamilyVersion)
	if err != nil {
		return "", errors.New("failed to create a proposal sign transaction: " + err.Error())
	}

	return c.submitTransaction(ctx, transaction, signer)
}

// GetActiveProposals returns all active proposals
func (c Client) GetActiveProposals(ctx context.Context) (proposals []model.Proposal, err error) {
	// fetch all the proposals = pass only the part of address corresponding to proposals
	addr := propfamily.GetProposalAddressFromID("")
	filter := url.Values{}
	filter.Set("address", addr)

	url := fmt.Sprintf("%s?%s", stateAPI, filter.Encode())
	// TODO: paging, limits
	response, err := c.sendRequest(ctx, url, nil, "")
	if err != nil {
		return
	}

	var unmarshalled struct {
		Data []json.RawMessage
	}
	if err := json.Unmarshal([]byte(response), &unmarshalled); err != nil {
		return proposals, errors.New("get active proposals: failed to unmarshal the response: " + err.Error())
	}
	c.logger.Info(fmt.Sprint("fetched ", len(unmarshalled.Data), " proposals from the current state"))

	for _, payload := range unmarshalled.Data {
		var proposal propfamily.ProposalData
		if err := c.unmarshalStatePayload(&proposal, string(payload)); err != nil {
			c.logger.Error("get active proposals: failed to unmarshal the proposal payload: " + err.Error())
			continue
		}

		if proposal.CurrentStatus != string(model.DocStatusActive) {
			// return only active proposals
			// c.logger.Debug("get active proposals: found inactive proposal: " + proposal.ProposalID)
			continue
		}

		m := convertToModelProposal(proposal)
		proposals = append(proposals, m)

	}

	return proposals, nil
}

func convertToModelProposal(propData propfamily.ProposalData) model.Proposal {
	return model.Proposal{
		ProposalID:         propData.ProposalID,
		DocumentName:       propData.DocName,
		Category:           propData.Category,
		ModificationAuthor: propData.Author,
		Content:            []byte{},
		ContentHash:        propData.ContentHash,
		ProposedStatus:     propData.ProposedDocStatus,
		CurrentStatus:      propData.CurrentStatus,
		Signers:            propData.Signers,
	}
}

// GetDocProposals fills in only proposal ID and content hash
func (c Client) GetDocProposals(ctx context.Context, category string, documentName string) (proposals []model.Proposal, err error) {
	payload, err := c.getDocState(ctx, category, documentName)
	if err != nil {
		return
	}

	for proposalID, contentHash := range payload.Proposals {
		proposals = append(proposals, model.Proposal{
			DocumentName: documentName,
			Category:     category,
			ProposalID:   proposalID,
			ContentHash:  contentHash,
		})
	}

	return proposals, nil
}
func (c Client) GetProposal(ctx context.Context, proposalID string) (model.Proposal, error) {
	propData, err := c.getProposalState(ctx, proposalID)
	if err != nil {
		return model.Proposal{}, errors.New("failed to get the proposal from blockchain: " + err.Error())
	}

	return convertToModelProposal(propData), nil
}

// GetUserProposals returns only active proposals created by the user
func (c Client) GetUserProposals(ctx context.Context, user string) (proposals []model.Proposal, err error) {
	userPropos, err := c.getUserState(ctx, user)
	if err != nil {
		return
	}

	// TODO: parallelize
	for _, id := range userPropos.Active {
		propData, err := c.getProposalState(ctx, id)
		if err != nil {
			c.logger.Error("getting user proposal error, skipping... error: "+err.Error(), zap.String("proposalID", id))
			continue
		}

		proposals = append(proposals, convertToModelProposal(propData))
	}

	if len(proposals) != len(userPropos.Active) {
		c.logger.Warn(fmt.Sprint("returning ", len(proposals), "/", len(userPropos.Active), " proposals due to get proposal state errors"))
	}

	return proposals, nil
}

func (c Client) unmarshalStatePayload(out interface{}, response string) error {
	var unmarshalled struct {
		Data string
	}
	if err := json.Unmarshal([]byte(response), &unmarshalled); err != nil {
		return errors.New("failed to unmarshal the response: " + err.Error())
	}
	decoded, err := base64.StdEncoding.DecodeString(unmarshalled.Data)
	if err != nil {
		return errors.New("failed to decode the payload: " + err.Error())
	}

	if err := cbor.Unmarshal([]byte(decoded), out); err != nil {
		return errors.New("failed to unmarshal the payload: " + err.Error())
	}

	return nil
}

func (c Client) getDocState(ctx context.Context, category string, docName string) (data propfamily.DocData, err error) {
	addr := propfamily.GetDocAddress(category, docName)
	url := fmt.Sprintf("%s/%s", stateAPI, addr)
	response, err := c.sendRequest(ctx, url, nil, "")
	if err != nil {
		return data, err
	}

	var payload propfamily.DocData
	if err := c.unmarshalStatePayload(&payload, response); err != nil {
		return data, errors.New("get proposal state error: " + err.Error())
	}

	return payload, nil
}

func (c Client) getProposalState(ctx context.Context, proposalID string) (data propfamily.ProposalData, err error) {
	addr := propfamily.GetProposalAddressFromID(proposalID)
	url := fmt.Sprintf("%s/%s", stateAPI, addr)
	response, err := c.sendRequest(ctx, url, nil, "")
	if err != nil {
		return data, err
	}

	var payload propfamily.ProposalData
	if err := c.unmarshalStatePayload(&payload, response); err != nil {
		return data, errors.New("get proposal state error: " + err.Error())
	}

	return payload, nil
}

func (c Client) getUserState(ctx context.Context, user string) (data propfamily.UserData, err error) {
	addr := propfamily.GetUserAddress(user)
	url := fmt.Sprintf("%s/%s", stateAPI, addr)
	response, err := c.sendRequest(ctx, url, nil, "")
	if err != nil {
		return data, err
	}

	var payload propfamily.UserData
	if err := c.unmarshalStatePayload(&payload, response); err != nil {
		return data, errors.New("get user state error: " + err.Error())
	}

	return payload, nil
}

func (c Client) SubmitProposal(ctx context.Context, proposal model.Proposal, signer *signing.Signer) (transactionID string, err error) {

	proposalDataAddress := proposalfamily.GetProposalAddress(proposal)
	authorAddress := proposalfamily.GetUserAddress(proposal.ModificationAuthor)
	docAddress := proposalfamily.GetDocAddress(proposal.Category, proposal.DocumentName)

	payload := make(map[interface{}]interface{})
	payload["action"] = actionInsert
	payload["proposalID"] = proposal.ProposalID
	payload["category"] = proposal.Category
	payload["docName"] = proposal.DocumentName
	payload["contentHash"] = proposal.ContentHash
	payload["proposedStatus"] = proposal.ProposedStatus
	payload["author"] = proposal.ModificationAuthor

	transaction, err := NewTransaction(payload, signer, []string{proposalDataAddress, authorAddress, docAddress}, propfamily.FamilyName, propfamily.FamilyVersion)
	if err != nil {
		return "", errors.New("failed to create a new proposal transaction: " + err.Error())
	}

	return c.submitTransaction(ctx, transaction, signer)

}

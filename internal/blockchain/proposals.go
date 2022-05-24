/**
 * Copyright 2018 Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * ------------------------------------------------------------------------------
 */

// based on https://github.com/hyperledger/sawtooth-sdk-go/blob/21f3d02d2446b6a91a945c93a8b94b1ddf616841/examples/intkey_go/src/sawtooth_intkey_client/intkey_client.go
package blockchain

import (
	"context"
	"net/url"
	"time"

	"doc-management/internal/model"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/fxamacker/cbor"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/transaction_pb2"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type action string

const (
	actionInsert action = "insert"
	actionVote   action = "vote"
	actionDelete action = "delete"
)

const (
	proposalFamilyVersion  string = "1.0"
	batchAPI               string = "batches"
	batchStatusAPI         string = "batch_statuses"
	stateAPI               string = "state"
	contentTypeOctetStream string = "application/octet-stream"

	wait uint = 10
)

func (c Client) RemoveProposal(ctx context.Context, proposal model.Proposal) error {
	// TODO
	return nil
}

// GetActiveProposals returns all active proposals
func (c Client) GetActiveProposals(ctx context.Context) (proposals []model.Proposal, err error) {
	// fetch all the proposals = pass only the part of address corresponding to proposals
	addr := getProposalAddressFromID("")
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
		var proposal proposalData
		if err := c.unmarshalStatePayload(&proposal, string(payload)); err != nil {
			c.logger.Error("get active proposals: failed to unmarshal the proposal payload: " + err.Error())
			continue
		}

		if proposal.CurrentStatus != string(model.DocStatusActive) {
			// return only active proposals
			c.logger.Debug("get active proposals: found inactive proposal: ")
			continue
		}

		m := convertToModelProposal(proposal)
		proposals = append(proposals, m)

	}

	return proposals, nil
}

func convertToModelProposal(propData proposalData) model.Proposal {
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
	c.logger.Debug("unmarshalled payload", zap.Any("payload", out))

	return nil
}

func (c Client) getDocState(ctx context.Context, category string, docName string) (data docData, err error) {
	addr := getDocAddress(category, docName)
	url := fmt.Sprintf("%s/%s", stateAPI, addr)
	response, err := c.sendRequest(ctx, url, nil, "")
	if err != nil {
		return data, err
	}

	var payload docData
	if err := c.unmarshalStatePayload(&payload, response); err != nil {
		return data, errors.New("get proposal state error: " + err.Error())
	}

	return payload, nil
}

func (c Client) getProposalState(ctx context.Context, proposalID string) (data proposalData, err error) {
	addr := getProposalAddressFromID(proposalID)
	url := fmt.Sprintf("%s/%s", stateAPI, addr)
	response, err := c.sendRequest(ctx, url, nil, "")
	if err != nil {
		return data, err
	}

	var payload proposalData
	if err := c.unmarshalStatePayload(&payload, response); err != nil {
		return data, errors.New("get proposal state error: " + err.Error())
	}

	return payload, nil
}

func (c Client) getUserState(ctx context.Context, user string) (data userData, err error) {
	addr := getUserAddress(user)
	url := fmt.Sprintf("%s/%s", stateAPI, addr)
	response, err := c.sendRequest(ctx, url, nil, "")
	if err != nil {
		return data, err
	}

	var payload userData
	if err := c.unmarshalStatePayload(&payload, response); err != nil {
		return data, errors.New("get user state error: " + err.Error())
	}

	return payload, nil
}

func (c Client) Submit(ctx context.Context, proposalTxn ProposalTransaction) (transactionID string, err error) {

	c.logger.Debug("submitting a proposal to address " + proposalTxn.proposalAddress)

	// Get BatchList
	rawBatchList, err := createBatchList(
		[]*transaction_pb2.Transaction{&proposalTxn.transaction}, proposalTxn.signer)
	if err != nil {
		return "", errors.New(
			fmt.Sprintf("unable to construct batch list: %v", err))
	}
	batchId := rawBatchList.Batches[0].HeaderSignature
	batchList, err := proto.Marshal(&rawBatchList)
	if err != nil {
		return "", errors.New(
			fmt.Sprintf("unable to serialize batch list: %v", err))
	}

	waitTime := uint(0)
	startTime := time.Now()
	response, err := c.sendRequest(
		ctx, batchAPI, batchList, contentTypeOctetStream)
	if err != nil {
		return "", err
	}
	for waitTime < wait {
		status, err := c.getStatus(context.Background(), batchId, wait-waitTime)
		if err != nil {
			return "", err
		}
		waitTime = uint(time.Now().Sub(startTime))
		if status != "PENDING" {
			c.logger.Info("getStatus response: " + response)
			return proposalTxn.transaction.HeaderSignature, nil
		}
	}

	c.logger.Info("getStatus response: " + response)
	return proposalTxn.transaction.HeaderSignature, nil

}

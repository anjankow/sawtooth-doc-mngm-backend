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
	"time"

	"doc-management/internal/hashing"
	"doc-management/internal/model"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/fxamacker/cbor"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/transaction_pb2"
	"github.com/hyperledger/sawtooth-sdk-go/signing"
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
	transactionsAPI        string = "transactions"
	stateAPI               string = "state"
	contentTypeOctetStream string = "application/octet-stream"

	wait uint = 10
)

var ErrInvalidContentHash = errors.New("content hash is invalid")

type ProposalTransaction struct {
	proposalAddress string
	transaction     transaction_pb2.Transaction
	signer          *signing.Signer
}

func (t ProposalTransaction) GetTransactionID() string {
	return t.transaction.HeaderSignature
}

func (t ProposalTransaction) GetProposalAddress() string {
	return t.proposalAddress
}

func NewProposalTransaction(proposal model.Proposal, signer *signing.Signer) (ProposalTransaction, error) {

	proposalDataAddress := getProposalAddress(proposal)
	authorAddress := getUserAddress(proposal.ModificationAuthor)
	docAddress := getDocAddress(proposal.Category, proposal.DocumentName)

	payload := make(map[interface{}]interface{})
	payload["action"] = actionInsert
	payload["proposalID"] = proposal.ProposalID
	payload["category"] = proposal.Category
	payload["docName"] = proposal.DocumentName
	payload["contentHash"] = proposal.ContentHash
	payload["proposedStatus"] = proposal.ProposedStatus
	payload["author"] = proposal.ModificationAuthor

	payloadDump, err := cbor.Marshal(payload, cbor.CanonicalEncOptions())
	if err != nil {
		return ProposalTransaction{}, errors.New("failed to dump the payload: " + err.Error())
	}

	// Construct TransactionHeader
	rawTransactionHeader := transaction_pb2.TransactionHeader{
		SignerPublicKey:  signer.GetPublicKey().AsHex(),
		FamilyName:       proposalFamily,
		FamilyVersion:    proposalFamilyVersion,
		Nonce:            strconv.Itoa(rand.Int()),
		BatcherPublicKey: signer.GetPublicKey().AsHex(),
		Inputs:           []string{proposalDataAddress, authorAddress, docAddress},
		Outputs:          []string{proposalDataAddress, authorAddress, docAddress},
		PayloadSha512:    hashing.Calculate(payloadDump),
	}

	transactionHeader, err := proto.Marshal(&rawTransactionHeader)
	if err != nil {
		return ProposalTransaction{}, errors.New(
			fmt.Sprintf("unable to serialize transaction header: %v", err))
	}

	// Signature of TransactionHeader
	transactionHeaderSignature := hex.EncodeToString(
		signer.Sign(transactionHeader))

	// Construct Transaction
	transaction := transaction_pb2.Transaction{
		Header:          transactionHeader,
		HeaderSignature: transactionHeaderSignature,
		Payload:         payloadDump,
	}

	return ProposalTransaction{
		proposalAddress: proposalDataAddress,
		transaction:     transaction,
		signer:          signer,
	}, nil
}

func (c Client) RemoveProposal(ctx context.Context, proposal model.Proposal) error {
	// TODO
	return nil
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

		proposals = append(proposals, model.Proposal{
			ProposalID:         id,
			DocumentName:       propData.DocName,
			Category:           propData.Category,
			ModificationAuthor: propData.Author,
			Content:            []byte{},
			ContentHash:        propData.ContentHash,
			ProposedStatus:     propData.ProposedDocStatus,
			CurrentStatus:      propData.CurrentStatus,
			Signers:            propData.Signers,
		})
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

func (c Client) GetProposals(ctx context.Context, proposal model.Proposal) ([]model.Proposal, error) {
	url := fmt.Sprintf("%s/%s", stateAPI, getProposalAddress(proposal))
	response, err := c.sendRequest(ctx, url, nil, "")
	if err != nil {
		return []model.Proposal{}, err
	}

	return c.readExistingProposals(response)
}

func (c Client) VerifyContentHash(ctx context.Context, proposal model.Proposal) error {
	url := fmt.Sprintf("%s/%s", stateAPI, getProposalAddress(proposal))
	response, err := c.sendRequest(ctx, url, nil, "")
	if err != nil {
		return err
	}
	contentHash, err := c.readContentHash(response)

	if contentHash != proposal.ContentHash {
		return ErrInvalidContentHash
	}

	return nil
}

func (c Client) readExistingProposals(response string) ([]model.Proposal, error) {
	var proposals []model.Proposal
	var unmarshalled struct {
		Data string
	}
	if err := json.Unmarshal([]byte(response), &unmarshalled); err != nil {
		return proposals, errors.New("failed to unmarshal /state GET response: " + err.Error())
	}
	decoded, err := base64.StdEncoding.DecodeString(unmarshalled.Data)
	if err != nil {
		return proposals, errors.New("failed to decode /state GET payload: " + err.Error())
	}

	c.logger.Debug("decoded payload", zap.String("payload", string(decoded)))

	var payload struct {
		Category  string
		DocName   string
		Proposals []storedProposal `cbor:"proposals"`
	}
	if err := cbor.Unmarshal([]byte(decoded), &payload); err != nil {
		return proposals, errors.New("failed to unmarshal /state GET payload: " + err.Error())
	}
	c.logger.Debug("unmarshalled payload", zap.Any("payload", payload))

	proposals = make([]model.Proposal, len(payload.Proposals))
	for i, existing := range payload.Proposals {
		proposals[i] = model.Proposal{
			DocumentName:       payload.DocName,
			Category:           payload.Category,
			ProposalID:         existing.ProposalID,
			ModificationAuthor: existing.Author,
			Content:            []byte{},
			ContentHash:        existing.ContentHash,
			ProposedStatus:     existing.ProposedDocStatus,
		}
	}

	return proposals, nil
}

func (c Client) readContentHash(response string) (string, error) {
	var unmarshalled struct {
		Data struct {
			Payload string
		}
	}
	if err := json.Unmarshal([]byte(response), &unmarshalled); err != nil {
		return "", errors.New("failed to unmarshal /transactions GET response: " + err.Error())
	}
	decoded, err := base64.StdEncoding.DecodeString(unmarshalled.Data.Payload)
	if err != nil {
		return "", errors.New("failed to decode /transactions GET payload: " + err.Error())
	}

	c.logger.Debug("unmarshalled payload", zap.String("payload", string(decoded)))

	var payload struct {
		ContentHash string `cbor:"contentHash"`
	}
	if err := cbor.Unmarshal([]byte(decoded), &payload); err != nil {
		return "", errors.New("failed to unmarshal /transactions GET payload: " + err.Error())
	}

	contentHash := payload.ContentHash
	c.logger.Debug("unmarshalled content hash info", zap.String("contentHash", contentHash))

	return contentHash, nil

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

type storedProposal struct {
	_ struct{} `cbor:",toarray"`

	ProposalID        string   `cbor:"proposalID"`
	DocName           string   `cbor:"docName"`
	Category          string   `cbor:"category"`
	Author            string   `cbor:"author"`
	Signers           []string `cbor:"signers"`
	ProposedDocStatus string   `cbor:"proposedDocStatus"`
	CurrentStatus     string   `cbor:"currentStatus"`
	ContentHash       string   `cbor:"contentHash"`
}

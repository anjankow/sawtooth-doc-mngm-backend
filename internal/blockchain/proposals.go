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
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"strconv"

	cbor "github.com/brianolson/cbor_go"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/transaction_pb2"
	"github.com/hyperledger/sawtooth-sdk-go/signing"
	"google.golang.org/protobuf/proto"
)

const (
	proposalFamily         string = "proposals"
	proposalFamilyVersion  string = "1.0"
	batchSubmitAPI         string = "batches"
	batchStatusAPI         string = "batch_statuses"
	stateAPI               string = "state"
	contentTypeOctetStream string = "application/octet-stream"

	wait uint = 10
)

type ProposalTransaction struct {
	address     string
	transaction transaction_pb2.Transaction
	signer      *signing.Signer
}

func (t ProposalTransaction) GetTransactionID() string {
	return t.transaction.HeaderSignature
}

func NewProposalTransaction(proposal model.Proposal, signer *signing.Signer) (ProposalTransaction, error) {

	address := getProposalAddress(proposal)

	payload := make(map[string]interface{})
	payload["category"] = proposal.Category
	payload["docName"] = proposal.DocumentName
	payload["contentHash"] = proposal.ContentHash
	payload["proposedStatus"] = proposal.ProposedStatus
	payload["author"] = proposal.ModificationAuthor
	payloadDump, err := cbor.Dumps(payload)
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
		Inputs:           []string{address},
		Outputs:          []string{address},
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
		address:     address,
		transaction: transaction,
		signer:      signer,
	}, nil
}

func getProposalAddress(proposal model.Proposal) (address string) {
	proposalFamilyHash := hashing.CalculateFromStr(proposalFamily)
	categoryHash := hashing.CalculateFromStr(proposal.Category)
	docNameHash := hashing.CalculateFromStr(proposal.DocumentName)

	return proposalFamilyHash[0:6] + categoryHash[0:6] + docNameHash[0:58]
}

func (c Client) Submit(ctx context.Context, proposalTxn ProposalTransaction) (transactionID string, err error) {

	c.logger.Debug("submitting a proposal to address " + proposalTxn.address)

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
		ctx, batchSubmitAPI, batchList, contentTypeOctetStream)
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

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
	bytes2 "bytes"
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"doc-management/internal/config"
	"doc-management/internal/hashing"
	"doc-management/internal/model"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"strconv"

	cbor "github.com/brianolson/cbor_go"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/batch_pb2"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/transaction_pb2"
	"github.com/hyperledger/sawtooth-sdk-go/signing"
	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v3"
)

const (
	proposalFamily         string = "proposals"
	proposalFamilyVersion  string = "1.0"
	batchSubmitAPI         string = "batches"
	batchStatusAPI         string = "batch_statuses"
	stateAPI               string = "state"
	contentTypeOctetStream string = "application/octet-stream"

	wait uint = 5
)

func (c Client) SubmitProposal(ctx context.Context, proposal model.Proposal, signer *signing.Signer) error {

	proposalFamilyHash := hashing.CalculateFromStr(proposalFamily)
	categoryHash := hashing.CalculateFromStr(proposal.Category)
	docNameHash := hashing.CalculateFromStr(proposal.DocumentName)

	address := proposalFamilyHash[0:6] + categoryHash[0:6] + docNameHash[0:58]

	payload := make(map[string]interface{})
	payload["docName"] = proposal.DocumentName
	payload["contentHash"] = proposal.ContentHash
	payload["proposedStatus"] = proposal.ProposedStatus
	payload["author"] = proposal.ModificationAuthor
	payloadDump, err := cbor.Dumps(payload)
	if err != nil {
		return errors.New("failed to dump the payload: " + err.Error())
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
		return errors.New(
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

	// Get BatchList
	rawBatchList, err := createBatchList(
		[]*transaction_pb2.Transaction{&transaction}, signer)
	if err != nil {
		return errors.New(
			fmt.Sprintf("unable to construct batch list: %v", err))
	}
	batchId := rawBatchList.Batches[0].HeaderSignature
	batchList, err := proto.Marshal(&rawBatchList)
	if err != nil {
		return errors.New(
			fmt.Sprintf("unable to serialize batch list: %v", err))
	}

	waitTime := uint(0)
	startTime := time.Now()
	response, err := c.sendRequest(
		batchSubmitAPI, batchList, contentTypeOctetStream)
	if err != nil {
		return err
	}
	for waitTime < wait {
		status, err := c.getStatus(batchId, wait-waitTime)
		if err != nil {
			return err
		}
		waitTime = uint(time.Now().Sub(startTime))
		if status != "PENDING" {
			c.logger.Info("getStatus response: " + response)
			return nil
		}
	}

	c.logger.Info("getStatus response: " + response)
	return nil

}

func (c Client) getStatus(
	batchId string, wait uint) (string, error) {

	// API to call
	apiSuffix := fmt.Sprintf("%s?id=%s&wait=%d",
		batchStatusAPI, batchId, wait)
	response, err := c.sendRequest(apiSuffix, []byte{}, "")
	if err != nil {
		return "", err
	}

	responseMap := make(map[interface{}]interface{})
	err = yaml.Unmarshal([]byte(response), &responseMap)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error reading response: %v", err))
	}
	entry :=
		responseMap["data"].([]interface{})[0].(map[interface{}]interface{})
	return fmt.Sprint(entry["status"]), nil
}

func (c Client) sendRequest(
	apiSuffix string,
	data []byte,
	contentType string) (string, error) {

	// Construct URL
	var url string
	if strings.HasPrefix(config.GetValidatorRestApiAddr(), "http://") {
		url = fmt.Sprintf("%s/%s", config.GetValidatorRestApiAddr(), apiSuffix)
	} else {
		url = fmt.Sprintf("http://%s/%s", config.GetValidatorRestApiAddr(), apiSuffix)
	}

	// Send request to validator URL
	var response *http.Response
	var err error
	if len(data) > 0 {
		response, err = http.Post(url, contentType, bytes2.NewBuffer(data))
	} else {
		response, err = http.Get(url)
	}
	if err != nil {
		return "", errors.New(
			fmt.Sprintf("Failed to connect to REST API: %v", err))
	}
	if response.StatusCode == 404 {
		c.logger.Debug(fmt.Sprintf("%v", response))
		return "", errors.New("responded with status 404")
	} else if response.StatusCode >= 400 {
		return "", errors.New(
			fmt.Sprintf("Error %d: %s", response.StatusCode, response.Status))
	}
	defer response.Body.Close()
	reponseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error reading response: %v", err))
	}
	return string(reponseBody), nil
}

func createBatchList(
	transactions []*transaction_pb2.Transaction, signer *signing.Signer) (batch_pb2.BatchList, error) {

	// Get list of TransactionHeader signatures
	transactionSignatures := []string{}
	for _, transaction := range transactions {
		transactionSignatures =
			append(transactionSignatures, transaction.HeaderSignature)
	}

	// Construct BatchHeader
	rawBatchHeader := batch_pb2.BatchHeader{
		SignerPublicKey: signer.GetPublicKey().AsHex(),
		TransactionIds:  transactionSignatures,
	}
	batchHeader, err := proto.Marshal(&rawBatchHeader)
	if err != nil {
		return batch_pb2.BatchList{}, errors.New(
			fmt.Sprintf("unable to serialize batch header: %v", err))
	}

	// Signature of BatchHeader
	batchHeaderSignature := hex.EncodeToString(
		signer.Sign(batchHeader))

	// Construct Batch
	batch := batch_pb2.Batch{
		Header:          batchHeader,
		Transactions:    transactions,
		HeaderSignature: batchHeaderSignature,
	}

	// Construct BatchList
	return batch_pb2.BatchList{
		Batches: []*batch_pb2.Batch{&batch},
	}, nil
}

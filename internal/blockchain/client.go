package blockchain

import (
	bytes2 "bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/hyperledger/sawtooth-sdk-go/protobuf/batch_pb2"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/transaction_pb2"
	"github.com/hyperledger/sawtooth-sdk-go/signing"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v3"
)

type Client struct {
	logger *zap.Logger
	url    string
}

func NewClient(logger *zap.Logger, validatorRestAPIUrl string) *Client {
	url := validatorRestAPIUrl
	if !strings.HasPrefix(validatorRestAPIUrl, "http://") {
		url = "http://" + validatorRestAPIUrl
	}

	return &Client{logger: logger, url: url}
}

func (c Client) getStatus(ctx context.Context,
	batchId string, wait uint) (string, error) {

	// API to call
	apiSuffix := fmt.Sprintf("%s?id=%s&wait=%d",
		batchStatusAPI, batchId, wait)
	response, err := c.sendRequest(ctx, apiSuffix, []byte{}, "")
	if err != nil {
		return "", err
	}

	responseMap := make(map[string]interface{})
	err = yaml.Unmarshal([]byte(response), &responseMap)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error reading response: %v", err))
	}
	entry :=
		responseMap["data"].([]interface{})[0].(map[string]interface{})
	return fmt.Sprint(entry["status"]), nil
}

func (c Client) sendRequest(
	ctx context.Context,
	apiSuffix string,
	data []byte,
	contentType string) (string, error) {

	// Construct URL
	url := fmt.Sprintf("%s/%s", c.url, apiSuffix)

	var req *http.Request
	var err error
	// Send request to validator URL
	if len(data) > 0 {
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, url, bytes2.NewBuffer(data))
	} else {
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	}

	if err != nil {
		return "", errors.New("failed to create a new request: " + err.Error())
	}
	req.Header.Add("Content-Type", contentType)

	c.logger.Debug("sending " + req.Method + " request to " + url)
	response, err := http.DefaultClient.Do(req)

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

	c.logger.Debug("reading the response body")
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

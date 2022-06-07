package blockchain

import (
	bytes2 "bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/fxamacker/cbor"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/batch_pb2"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/transaction_pb2"
	"github.com/hyperledger/sawtooth-sdk-go/signing"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v3"
)

var (
	ErrNotFound = errors.New("responded with status 404")
)

const (
	batchAPI               string = "batches"
	batchStatusAPI         string = "batch_statuses"
	stateAPI               string = "state"
	contentTypeOctetStream string = "application/octet-stream"

	wait uint = 10
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

func (c Client) submitTransaction(ctx context.Context, transaction transaction_pb2.Transaction, signer *signing.Signer) (string, error) {

	batchId, batchList, err := createBatchList(
		[]*transaction_pb2.Transaction{&transaction}, signer)

	startTime := time.Now()
	response, err := c.sendRequest(
		ctx, batchAPI, batchList, contentTypeOctetStream)
	if err != nil {
		return "", err
	}
	status, err := c.getStatus(batchId, startTime)
	if err != nil {
		return "", err
	}

	c.logger.Info("request response: " + response + ", status: " + status)
	return transaction.HeaderSignature, nil
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
		return "", ErrNotFound
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
	transactions []*transaction_pb2.Transaction, signer *signing.Signer) (batchId string, batchList []byte, err error) {

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
		err = errors.New(
			fmt.Sprintf("unable to serialize batch header: %v", err))
		return
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
	rawBatchList := batch_pb2.BatchList{
		Batches: []*batch_pb2.Batch{&batch},
	}

	batchId = rawBatchList.Batches[0].HeaderSignature
	batchList, err = proto.Marshal(&rawBatchList)
	if err != nil {
		err = errors.New(
			fmt.Sprintf("unable to serialize batch list: %v", err))
		return
	}

	return batchId, batchList, nil
}

func (c Client) getStatus(batchId string, startTime time.Time) (string, error) {

	waitTime := uint(0)
	statusPending := "PENDING"
	for waitTime < wait {
		status, err := c.getStatusRequest(context.Background(), batchId, wait-waitTime)
		if err != nil {
			return "", err
		}
		waitTime = uint(time.Now().Sub(startTime))
		if status != statusPending {
			return status, nil
		}
	}

	return statusPending, nil

}

func (c Client) getStatusRequest(ctx context.Context,
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

func unmarshalDataList(response string) ([]json.RawMessage, error) {

	var unmarshalled struct {
		Data []json.RawMessage
	}
	if err := json.Unmarshal([]byte(response), &unmarshalled); err != nil {
		return []json.RawMessage{}, errors.New("failed to unmarshal data list: " + err.Error())
	}

	return unmarshalled.Data, nil
}

func unmarshalStatePayload(out interface{}, response string) error {
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

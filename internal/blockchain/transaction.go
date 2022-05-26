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
	"doc-management/internal/hashing"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/fxamacker/cbor"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/transaction_pb2"
	"github.com/hyperledger/sawtooth-sdk-go/signing"
	"google.golang.org/protobuf/proto"
)

func NewTransaction(payload map[interface{}]interface{}, signer *signing.Signer, addresses []string, familyName, familyVersion string) (transaction_pb2.Transaction, error) {

	payloadDump, err := cbor.Marshal(payload, cbor.CanonicalEncOptions())
	if err != nil {
		return transaction_pb2.Transaction{}, errors.New("failed to dump the payload: " + err.Error())
	}

	// Construct TransactionHeader
	rawTransactionHeader := transaction_pb2.TransactionHeader{
		SignerPublicKey:  signer.GetPublicKey().AsHex(),
		FamilyName:       familyName,
		FamilyVersion:    familyVersion,
		Nonce:            strconv.Itoa(rand.Int()),
		BatcherPublicKey: signer.GetPublicKey().AsHex(),
		Inputs:           addresses,
		Outputs:          addresses,
		PayloadSha512:    hashing.Calculate(payloadDump),
	}

	transactionHeader, err := proto.Marshal(&rawTransactionHeader)
	if err != nil {
		return transaction_pb2.Transaction{}, errors.New(
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

	return transaction, nil
}

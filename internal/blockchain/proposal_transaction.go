package blockchain

import (
	"doc-management/internal/blockchain/proposalfamily"
	"doc-management/internal/hashing"
	"doc-management/internal/model"
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

	payloadDump, err := cbor.Marshal(payload, cbor.CanonicalEncOptions())
	if err != nil {
		return ProposalTransaction{}, errors.New("failed to dump the payload: " + err.Error())
	}

	// Construct TransactionHeader
	rawTransactionHeader := transaction_pb2.TransactionHeader{
		SignerPublicKey:  signer.GetPublicKey().AsHex(),
		FamilyName:       proposalfamily.FamilyName,
		FamilyVersion:    proposalfamily.FamilyVersion,
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

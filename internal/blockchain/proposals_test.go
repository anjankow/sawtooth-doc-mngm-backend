package blockchain_test

import (
	"context"
	"doc-management/internal/blockchain"
	"doc-management/internal/config"
	"doc-management/internal/hashing"
	"doc-management/internal/keymanager"
	"doc-management/internal/model"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.uber.org/zap"
)

func TestVerifyContentHash(t *testing.T) {
	logger := zap.NewExample()
	client := blockchain.NewClient(logger, config.GetValidatorRestApiAddr())
	hashing.Initialize(logger)

	content := "hulajnogi sa ze stonogi"
	proposal := model.Proposal{
		DocumentID: model.DocumentID{
			DocumentName: "tralala",
			Category:     "general",
		},
		ProposalContent: model.ProposalContent{
			TransactionID:      "",
			ModificationAuthor: "ja",
			Content:            []byte(content),
			ContentHash:        hashing.CalculateFromStr(content),
			ProposedStatus:     "accepted",
		},
	}
	keys, err := keymanager.GenerateKeys()
	require.NoError(t, err)

	txn, err := blockchain.NewProposalTransaction(proposal, keys.GetSigner())
	require.NoError(t, err)
	proposal.TransactionID = txn.GetTransactionID()

	_, err = client.Submit(context.TODO(), txn)
	require.NoError(t, err)

	assert.NoError(t, client.VerifyContentHash(context.TODO(), proposal))

	// proposal.ContentHash += "dsada"
	// assert.NoError(t, client.VerifyContentHash(context.TODO(), proposal))
}

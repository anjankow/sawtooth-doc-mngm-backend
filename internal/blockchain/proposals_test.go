package blockchain

import (
	"context"
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
	client := NewClient(logger, config.GetValidatorRestApiAddr())
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
	keyMan := keymanager.NewKeyManager(logger)
	keys, err := keyMan.GenerateKeys()
	require.NoError(t, err)

	txn, err := NewProposalTransaction(proposal, keys.GetSigner())
	require.NoError(t, err)
	proposal.TransactionID = txn.GetTransactionID()

	_, err = client.Submit(context.TODO(), txn)
	require.NoError(t, err)

	assert.NoError(t, client.VerifyContentHash(context.TODO(), proposal))

	proposal.ContentHash += "dsada"
	assert.ErrorIs(t, client.VerifyContentHash(context.TODO(), proposal), ErrInvalidContentHash)
}

func TestReadContentHash(t *testing.T) {
	logger := zap.NewExample()
	client := NewClient(logger, config.GetValidatorRestApiAddr())
	hashing.Initialize(logger)

	responsePayload := `
	{
		"data": {
		  "header": {
			"batcher_public_key": "0244c35f53451679ffe90bb80ec5e6c6a344c3a73de18aa75548e198192b75f2a1",
			"dependencies": [],
			"family_name": "proposals",
			"family_version": "1.0",
			"inputs": [
			  "8ed94c1d995bf5ca37360740d1e77f6d0c346e699331b2a50116dcc9925aa0f3c3d3a2"
			],
			"nonce": "5577006791947779410",
			"outputs": [
			  "8ed94c1d995bf5ca37360740d1e77f6d0c346e699331b2a50116dcc9925aa0f3c3d3a2"
			],
			"payload_sha512": "39fa2fe4ecbfbff9d4605d75dc778e129bd46722423294d7914bd037a38b0b5070971d6529fc47d3699ffb00f9730b05c71647022ca54ba2d2bc737704a48e92",
			"signer_public_key": "0244c35f53451679ffe90bb80ec5e6c6a344c3a73de18aa75548e198192b75f2a1"
		  },
		  "header_signature": "d648a5cd38a68181a65a650d78b20c66ddd40a320afeaac2467c62c84fb57db5751f19e6caafbdfb2bf6c4e925f25c019b7ac93d73ac6c346306376e436e59e6",
		  "payload": "pWZhdXRob3JkZmFlZmdkb2NOYW1lY2Rhc2hjYXRlZ29yeWdnZW5lcmFsa2NvbnRlbnRIYXNoeIBlNTFkZTdlNzFkZTFkMTBmMmE2YWI4MzJjMjUwZjJkOTExZmRkNmFhYmY5MGZlNjBmNzZkMDMzMTE4YTA5ODc5MWJjNDM4MWM0MTI5Mjc1YjgzZmVlNjIwMjNhNmNmODc1NGQyZTE1NTY5MzkyNTY2ZTJkNTU5YmI1ZmNlZWQyM25wcm9wb3NlZFN0YXR1c2hhY2NlcHRlZA=="
		},
		"link": "http://localhost:8008/transactions/d648a5cd38a68181a65a650d78b20c66ddd40a320afeaac2467c62c84fb57db5751f19e6caafbdfb2bf6c4e925f25c019b7ac93d73ac6c346306376e436e59e6"
	}
	`
	contentHash, err := client.readContentHash(responsePayload)

	require.NoError(t, err)
	assert.Equal(t, `e51de7e71de1d10f2a6ab832c250f2d911fdd6aabf90fe60f76d033118a098791bc4381c4129275b83fee62023a6cf8754d2e15569392566e2d559bb5fceed23`,
		contentHash)
}

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
		DocumentName:       "tralala",
		Category:           "general",
		ModificationAuthor: "ja",
		Content:            []byte(content),
		ContentHash:        hashing.CalculateFromStr(content),
		ProposedStatus:     "accepted",
	}
	proposal.Complete()

	keyMan := keymanager.NewKeyManager(logger)
	keys, err := keyMan.GenerateKeys()
	require.NoError(t, err)

	txn, err := NewProposalTransaction(proposal, keys.GetSigner())
	require.NoError(t, err)

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

func TestGetPayload(t *testing.T) {
	logger := zap.NewExample()
	client := NewClient(logger, config.GetValidatorRestApiAddr())
	hashing.Initialize(logger)

	responsePayload := `
	{
		"data": "oWlwcm9wb3NhbHOjeCQ4ZDZkMzg5MS01MGRkLTQxOGUtYjNhMS01YzZkMmQ4NTBmYzV4gDYzZmNiYzhmYzk0ZDBmZGVkMDM5ODMwMmZhMDdmYTRlNDQ5YTg2NjBkNmE1NWY2MTUwNGI4YzQ2Mzc0YTE2ZjlkMjgxZWM2ZTBhMzZkNWI5MGY0OGE1NDU0MzMxZWJmNzNkZTQ4MjQwNTQxYWU0MzQ1NTlhMTk0OWJiNjZjZTljeCRkMjVkMmY5Ni00ZWQ0LTQ3NDctYTA3YS05MDQ4NGI1ZjgyNmR4gDA5NTU2ZWM1NDRhYWIzYzczYmUxOGJjNDQ4NjIzMjY4OTg0ZTkxMzk2YjgzNjNhZDkzMTVlNjA5YjlmN2NiZDA4ZTFlZGQ1MTkyYWExMjdiMmMwOGFjZGVhYjZkM2MxOTBmMDUxYjNhYmFlNzI4N2E4NDIzNzgwMmZjOTg1NjNmeCQ3MGNmOTRlMy00ZjIyLTQ4ZTEtYmRlOS1jNGVmY2M2NTcyZWZ4gDk3YjQ3OTg0ZjA4YWMzOTJjMGY3MWE0MmRhYWZkNmQ3OTVkNDJlYzlkYWZjNzg2Mzk2MzAzYjgwZDM4MDQ0MGZhMjhkYTdhYzBhYjMyYjk3MTA2ZmVmZjEwMzc1NzczYzVhYWRiOGViYmM4MWNjNjYyOTVkNWU1ODZmMmQ5MTQz",
		"head": "8060272602d50c0394710a1db6bc5621d83f908682762d0d22739a65d8f2de6715fbc7f492dc01688da1f5e79c139aad7634b96d71212956a4fb998341b6e4c3",
		"link": "http://localhost:8008/state/8ed94c3c52af1d995bde32c32b4fae693f807da208dc86dcf20fa2f620e20d9edcc17d?head=8060272602d50c0394710a1db6bc5621d83f908682762d0d22739a65d8f2de6715fbc7f492dc01688da1f5e79c139aad7634b96d71212956a4fb998341b6e4c3"
	  }
	`
	var decoded struct {
		Proposals map[string]string `cbor:"proposals"`
	}
	err := client.unmarshalStatePayload(&decoded, responsePayload)

	require.NoError(t, err)
	assert.Equal(t, `63fcbc8fc94d0fded0398302fa07fa4e449a8660d6a55f61504b8c46374a16f9d281ec6e0a36d5b90f48a5454331ebf73de48240541ae434559a1949bb66ce9c`,
		decoded.Proposals["8d6d3891-50dd-418e-b3a1-5c6d2d850fc5"])
}

func TestReadExistingProposals(t *testing.T) {
	logger := zap.NewExample()
	client := NewClient(logger, config.GetValidatorRestApiAddr())
	hashing.Initialize(logger)

	responsePayload := `{
		"data": "o2hjYXRlZ29yeWdnZW5lcmFsZ2RvY05hbWVhd2lwcm9wb3NhbHOChHiAMGEyYmUyYmE1MDdjMzZhMjZhOTNjODg2MzcwNDZhYmYwNTIwNDg4YTZiZjNmOGUzNjg3YjU3NzQzYWNmYzk2NTIwZWQ2ZTg4ZWE1ODUxNDU3MDk4MWIxMzczNDA5OGE3NmFhZWViYWEyOWFhODZkYTRkN2RkNjAyOWNmNzRlNjZqdXNlcnNmYWRzZmhhY2NlcHRlZHiANTJhZGQ0ZmJjOTQ2MTMwOWNhODIyNWY5ZmQ0ZDgwMjAyYmIzZTU3NmUzZDhkYzkwNDk0M2E0YjU5NmRlMWY1ZjdiM2M4OWZjOGI4MjNmMzMyYWVhNmE2ZTZkNWE5MDQ1ODQ2ODExN2I1YzU5NDVhNDY3M2QxM2E2ODY0MjM0ZDGEeIA0MWJkYTE0ZjliZjA0MDQ3NTY4NWUxNmQ4MGY1OGU5YmJhMGY1NDE3YWExMWQ2ZGUwMDRmZWY5ZWFkYjgwNzZhNmM0NDJjYTYwMTA2MmQxMTJlMzg3NzBhYTFiZWM2ZGE5MzVhYWQ0NzE0MDgxZDgwODllNTc2MmU0M2M0ZjlmZmp1c2Vyc2ZhZHNmaGFjY2VwdGVkeIA1MmFkZDRmYmM5NDYxMzA5Y2E4MjI1ZjlmZDRkODAyMDJiYjNlNTc2ZTNkOGRjOTA0OTQzYTRiNTk2ZGUxZjVmN2IzYzg5ZmM4YjgyM2YzMzJhZWE2YTZlNmQ1YTkwNDU4NDY4MTE3YjVjNTk0NWE0NjczZDEzYTY4NjQyMzRkMQ==",
		"head": "33f53e2648059773251c0f08d44b04262c3a743cc65b7da41488baeef1e5f5b976b7b0d1ddb61822b9589c5cbb30b3d5fdcd2f8645e10976bbb2ed177fc26a92",
		"link": "http://localhost:8008/state/8ed94c1d995baa66509891ad28030349ba9581e8c92528faab6a34349061a44b6f8fcd?head=33f53e2648059773251c0f08d44b04262c3a743cc65b7da41488baeef1e5f5b976b7b0d1ddb61822b9589c5cbb30b3d5fdcd2f8645e10976bbb2ed177fc26a92"
	  }`
	proposals, err := client.readExistingProposals(responsePayload)

	require.NoError(t, err)
	assert.Equal(t, 2, len(proposals))
	assert.Equal(t, "general", proposals[0].Category)
	assert.Equal(t, "52add4fbc9461309ca8225f9fd4d80202bb3e576e3d8dc904943a4b596de1f5f7b3c89fc8b823f332aea6a6e6d5a90458468117b5c5945a4673d13a6864234d1", proposals[1].ContentHash)

}

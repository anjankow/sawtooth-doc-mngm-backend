package blockchain

import (
	"doc-management/internal/hashing"
	"doc-management/internal/model"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestAddress(t *testing.T) {
	hashing.Initialize(zap.NewNop())
	proposal := model.Proposal{
		TransactionID:      "8ed94cb143611f6fd9a184c21965ba642251f0792b10b000fbc9878b69179b1e7bcce9",
		DocumentName:       "docname",
		Category:           "general2",
		ModificationAuthor: "alabaster",
	}

	expectedDocAddr := `8ed94c3c52af3884d10c98b34cbf1a8c2b39f03406d32d96116e1d427755a4d2ad6195`
	expectedProposalAddr := `8ed94c5290e9373cdf2f5cb0a7b8561d6cd6bce82f96230e9d0e37bac8aaff85f86154`
	expectedUserAddr := `8ed94cb143611f6fd9a184c21965ba642251f0792b10b000fbc9878b69179b1e7bcce9`

	assert.Equal(t, expectedDocAddr, getDocAddress(proposal))
	assert.Equal(t, expectedProposalAddr, getProposalDataAddress(proposal))
	assert.Equal(t, expectedUserAddr, getUserAddress(proposal.ModificationAuthor))

}

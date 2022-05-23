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
		ProposalID:         "60a9e27b2ca2d845d7304a0955a1b358ec6e66d952bfc199b862d05ad365588d4f2272a0d570117518bb781667b6012b0f89206e89baabfe1bc8792c009bfcff",
		DocumentName:       "docname",
		Category:           "general2",
		ModificationAuthor: "alabaster",
	}

	expectedDocAddr := `8ed94c3c52af3884d10c98b34cbf1a8c2b39f03406d32d96116e1d427755a4d2ad6195`
	expectedProposalAddr := `8ed94c5290e964cc4cc0674fbd42665971d649f83ded200beda9732a984eb6e5d69f6b`
	expectedUserAddr := `8ed94cb143611f6fd9a184c21965ba642251f0792b10b000fbc9878b69179b1e7bcce9`

	assert.Equal(t, expectedDocAddr, getDocAddress(proposal))
	assert.Equal(t, expectedProposalAddr, getProposalAddress(proposal))
	assert.Equal(t, expectedUserAddr, getUserAddress(proposal.ModificationAuthor))

}

package doctrackerfamily_test

import (
	"doc-management/internal/blockchain/doctrackerfamily"
	"doc-management/internal/hashing"
	"doc-management/internal/model"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestGetDocVersionAddress(t *testing.T) {
	hashing.Initialize(zap.NewNop())
	doc := model.Document{
		Category:     "aaa",
		DocumentName: "name",
		Version:      55,
	}
	addr := doctrackerfamily.GetDocVersionAddress(doc)
	assert.Equal(t, len(addr), 70)
	t.Log(addr)
}

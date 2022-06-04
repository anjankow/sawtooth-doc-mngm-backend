package settingsfamily_test

import (
	"doc-management/internal/blockchain/settingsfamily"
	"doc-management/internal/hashing"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestGetAddress(t *testing.T) {
	hashing.Initialize(zap.NewNop())
	name := "proposal.vote.threshold"
	expectedAddr := "000000ecd1378bc9dc1300ab274474a6aa82c1497e22fe854a24bce3b0c44298fc1c14"
	assert.Equal(t, expectedAddr, settingsfamily.GetAddress(name))
}

package keymanager_test

import (
	"doc-management/internal/keymanager"
	"testing"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/stretchr/testify/assert"
)

func TestGenerateKey(t *testing.T) {
	keys, err := keymanager.GenerateKeys()
	assert.NoError(t, err)
	assert.NotEmpty(t, keys.PrivateKey)
	assert.NotEmpty(t, keys.PublicKey)

	priv := secp256k1.PrivKeyFromBytes(keys.PrivateKey.AsBytes())

	assert.Equal(t, priv.PubKey().SerializeUncompressed(), keys.PublicKey.AsBytes())
}

package keymanager

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"

	"github.com/btcsuite/btcd/btcec/v2"
)

type UserKeys struct {
	PublicKey  []byte
	PrivateKey []byte
}

// source: https://github.com/ethereum/go-ethereum/blob/86d547707965685cef732aa28c15e6811ea98408/crypto/secp256k1/secp256_test.go#L19
func GenerateKeys() (UserKeys, error) {
	key, err := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
	if err != nil {
		return UserKeys{}, errors.New("failed to generate the keys: " + err.Error())
	}
	pubkey := elliptic.Marshal(btcec.S256(), key.X, key.Y)

	privkey := make([]byte, 32)
	blob := key.D.Bytes()
	copy(privkey[32-len(blob):], blob)

	return UserKeys{
		PublicKey:  pubkey,
		PrivateKey: privkey,
	}, nil
}

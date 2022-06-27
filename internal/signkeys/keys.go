package signkeys

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"errors"

	"github.com/btcsuite/btcd/btcec"
	"github.com/hyperledger/sawtooth-sdk-go/signing"
	"go.uber.org/multierr"
)

type UserKeys struct {
	PrivateKey signing.PrivateKey
	PublicKey  signing.PublicKey
}

func (u UserKeys) GetSigner() *signing.Signer {
	cryptoFactory := signing.NewCryptoFactory(signing.NewSecp256k1Context())
	return cryptoFactory.NewSigner(u.PrivateKey)
}

func NewUserKeys(privateKeyHex, publicKeyHex string) (UserKeys, error) {
	priv, errPriv := hex.DecodeString(privateKeyHex)
	pub, errPub := hex.DecodeString(publicKeyHex)
	if err := multierr.Combine(errPriv, errPub); err != nil {
		return UserKeys{}, err
	}

	return UserKeys{
		PublicKey:  signing.NewSecp256k1PublicKey(pub),
		PrivateKey: signing.NewSecp256k1PrivateKey(priv),
	}, nil
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

	keys := UserKeys{
		PublicKey:  signing.NewSecp256k1PublicKey(pubkey),
		PrivateKey: signing.NewSecp256k1PrivateKey(privkey),
	}

	return keys, nil
}

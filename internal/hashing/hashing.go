package hashing

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"

	"hash"

	"go.uber.org/zap"
)

type hashWrapper struct {
	logger *zap.Logger
	sha256 hash.Hash
	sha512 hash.Hash
}

var hashInstance hashWrapper

func Initialize(logger *zap.Logger) {
	hashInstance.sha512 = sha512.New()
	hashInstance.sha256 = sha256.New()
	hashInstance.logger = logger
}

func CalculateSHA512(data string) string {
	hashInstance.sha512.Reset()

	if _, err := hashInstance.sha512.Write([]byte(data)); err != nil {
		hashInstance.logger.Error("failed to initialize hash function: " + err.Error())
		return ""
	}

	h := hashInstance.sha512.Sum(nil)

	return hex.EncodeToString(h)
}

func CalculateSHA256(data string) string {
	hashInstance.sha256.Reset()
	_, _ = hashInstance.sha256.Write([]byte(data))
	h := hashInstance.sha256.Sum(nil)

	return hex.EncodeToString(h)
}

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
	hash   hash.Hash
}

var hashInstance hashWrapper

func Initialize(logger *zap.Logger) {
	hashInstance.hash = sha512.New()
	hashInstance.logger = logger
}

func Calculate(data []byte) string {
	hashInstance.hash.Reset()

	if _, err := hashInstance.hash.Write(data); err != nil {
		hashInstance.logger.Error("failed to initialize hash function: " + err.Error())
		return ""
	}

	h := hashInstance.hash.Sum(nil)

	return hex.EncodeToString(h)
}

func CalculateFromStr(data string) string {
	return Calculate([]byte(data))
}

func CalculateSHA256(data string) string {
	hash := sha256.New()
	_, _ = hash.Write([]byte(data))
	h := hash.Sum(nil)

	return hex.EncodeToString(h)
}

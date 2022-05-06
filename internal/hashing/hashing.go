package hashing

import (
	"crypto/sha512"
	"encoding/hex"
	"errors"
)

type Hash struct {
	logger *zap.logger
}

func Hash(data []byte) string {
	hash := sha512.New()
	if _, err := hash.Write(data); err != nil {
		return "", errors.New("failed to initialize hash function: " + err.Error())
	}

	h := hash.Sum(nil)

	return hex.EncodeToString(h), nil
}

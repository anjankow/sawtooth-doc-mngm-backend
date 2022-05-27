package hashing_test

import (
	"crypto/sha256"
	"crypto/sha512"
	"doc-management/internal/hashing"
	"encoding/hex"
	"hash"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// python script for obtaining hash, the hash output need to match
// // // // // // // // // // // // // // // // // // //
// import hashlib

// def hash(data):
//     return hashlib.sha512(data.encode()).hexdigest()
// // // // // // // // // // // // // // // // // // //

func TestHashing(t *testing.T) {
	text := "mala agatka"

	hashing.Initialize(zap.NewNop())
	output := hashing.CalculateSHA512(text)
	assert.Equal(t,
		"3768b1bbee7097f5c98f0b2cfc516ae08e0e442ae333b4d7a3648d1e9d798e7e42a734cb48570d379af3c38df5996febc9a0cc1c8c7356ee8926e1b88aeeff15",
		output)
}

func TestHashing2Times(t *testing.T) {
	hashing.Initialize(zap.NewNop())

	text := "mala agatka"
	output := hashing.CalculateSHA512(text)
	assert.Equal(t,
		"3768b1bbee7097f5c98f0b2cfc516ae08e0e442ae333b4d7a3648d1e9d798e7e42a734cb48570d379af3c38df5996febc9a0cc1c8c7356ee8926e1b88aeeff15",
		output)

	text = "mniejsza agatka"
	output = hashing.CalculateSHA512(text)
	assert.Equal(t,
		"f2e0c1ae62cd64d5495b8342d2efd3f9e197bc4225b582885ee34ffa5325e26a23121c745ac347f37f62f699c6997a2c649de5379011d576a576ad05ae8524df",
		output)

}

func TestFamilyHash(t *testing.T) {
	proposalFamily := "proposals"
	hashing.Initialize(zap.NewNop())
	proposalFamilyHash := hashing.CalculateSHA512(proposalFamily)
	assert.Equal(t, proposalFamilyHash[0:6], "8ed94c")
}

func BenchmarkStaticHashInstance(b *testing.B) {
	var input = []string{
		"aaaa", "bbb", "", "sdfsadfas",
	}
	loops := 1000000
	hashing.Initialize(zap.NewNop())

	b.Run("static hash instance", func(b *testing.B) {
		for i := 0; i < loops; i++ {
			hashing.CalculateSHA256(input[i%4])
			hashing.CalculateSHA512(input[i%4])
		}
	})

	b.Run("on demand instance", func(b *testing.B) {
		for i := 0; i < loops; i++ {
			hashOnDemand(input[i%4], sha512.New())
			hashOnDemand(input[i%4], sha256.New())
		}
	})
}

func hashOnDemand(data string, hash hash.Hash) error {

	if _, err := hash.Write([]byte(data)); err != nil {
		return err
	}

	h := hash.Sum(nil)

	hex.EncodeToString(h)
	return nil
}

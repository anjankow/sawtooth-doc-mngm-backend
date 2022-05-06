package hashing_test

import (
	"doc-management/internal/hashing"
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
	data := []byte(text)

	hashing.Initialize(zap.NewNop())
	output := hashing.Calculate(data)
	assert.Equal(t,
		"3768b1bbee7097f5c98f0b2cfc516ae08e0e442ae333b4d7a3648d1e9d798e7e42a734cb48570d379af3c38df5996febc9a0cc1c8c7356ee8926e1b88aeeff15",
		output)
}

func TestHashing2Times(t *testing.T) {
	hashing.Initialize(zap.NewNop())

	text := "mala agatka"
	output := hashing.Calculate([]byte(text))
	assert.Equal(t,
		"3768b1bbee7097f5c98f0b2cfc516ae08e0e442ae333b4d7a3648d1e9d798e7e42a734cb48570d379af3c38df5996febc9a0cc1c8c7356ee8926e1b88aeeff15",
		output)

	text = "mniejsza agatka"
	output = hashing.Calculate([]byte(text))
	assert.Equal(t,
		"f2e0c1ae62cd64d5495b8342d2efd3f9e197bc4225b582885ee34ffa5325e26a23121c745ac347f37f62f699c6997a2c649de5379011d576a576ad05ae8524df",
		output)

}

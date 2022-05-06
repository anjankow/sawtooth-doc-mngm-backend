package hashing_test

import (
	"doc-management/internal/model"
	"testing"

	"github.com/stretchr/testify/assert"
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

	output, err := model.Hash(data)
	assert.NoError(t, err)
	assert.Equal(t,
		"3768b1bbee7097f5c98f0b2cfc516ae08e0e442ae333b4d7a3648d1e9d798e7e42a734cb48570d379af3c38df5996febc9a0cc1c8c7356ee8926e1b88aeeff15",
		output)
}

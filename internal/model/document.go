package model

type Document struct {
	Author     string
	DocumentID string

	DocBytes []byte
	Hash     []byte
}

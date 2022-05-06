package model

type DocStatus string

const (
	DocStatusAccepted DocStatus = "accepted"
	DocStatusDeleted  DocStatus = "deleted"
)

const DefaultCategory = "general"

// DocumentID to uniquely identify a document
type DocumentID struct {
	DocumentName string
	Category     string
}

// Document existing on the blockchain
type Document struct {
	DocumentID

	ModificationAuthor string
	Content            []byte
	ContentHash        []byte

	Version int
	Status  DocStatus
}

func (status DocStatus) IsValid() bool {
	return status == DocStatusAccepted || status == DocStatusDeleted
}

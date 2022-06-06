package model

type DocStatus string

const (
	DocStatusActive  DocStatus = "active"
	DocStatusRemoved DocStatus = "removed"
)

const DefaultCategory = "general"

// Document existing on the blockchain
type Document struct {
	DocumentName string
	Category     string

	Author      string
	Content     []byte
	ContentHash []byte

	Version int
	Status  DocStatus

	ProposalID string
}

func (status DocStatus) IsValid() bool {
	return status == DocStatusRemoved || status == DocStatusActive
}

func (status DocStatus) String() string {
	return string(status)
}

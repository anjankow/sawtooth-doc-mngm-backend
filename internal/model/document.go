package model

type DocStatus string

const (
	DocStatusActive   DocStatus = "active"
	DocStatusAccepted DocStatus = "accepted"
	DocStatusDeleted  DocStatus = "deleted"
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
	return status == DocStatusAccepted || status == DocStatusDeleted || status == DocStatusActive
}

package model

type DocStatus string

const (
	DocStatusActive  DocStatus = "active"
	DocStatusRemoved DocStatus = "removed"
	DocStatusInvalid DocStatus = "invalid"
)

const DefaultCategory = "general"

// Document existing on the blockchain
type Document struct {
	DocumentName string
	Category     string

	Author      string
	Content     []byte
	ContentHash string

	Version int
	Status  DocStatus

	ProposalID string

	Signers []string
}

func (status DocStatus) IsValid() bool {
	return status == DocStatusRemoved || status == DocStatusActive
}

func (status DocStatus) String() string {
	return string(status)
}

func NewDocumentFromProposal(proposal Proposal, version int) Document {
	return Document{
		DocumentName: proposal.DocumentName,
		Category:     proposal.Category,
		Author:       proposal.ModificationAuthor,
		Content:      proposal.Content,
		ContentHash:  proposal.ContentHash,
		Version:      version,
		Status:       proposal.ProposedStatus,
		ProposalID:   proposal.ProposalID,
		Signers:      proposal.Signers,
	}
}

func GetNextDocVersion(docs []Document) int {
	latestVersion := 0

	for _, doc := range docs {
		if doc.Version > latestVersion {
			latestVersion = doc.Version
		}
	}

	return latestVersion + 1
}

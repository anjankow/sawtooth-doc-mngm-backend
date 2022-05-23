package model

import (
	"doc-management/internal/hashing"
	"errors"

	"github.com/google/uuid"
)

type Proposal struct {
	ProposalID   string
	DocumentName string
	Category     string

	ModificationAuthor string
	Content            []byte
	ContentHash        string
	ProposedStatus     string
	CurrentStatus      string

	Signers []string
}

func (proposal Proposal) Validate() error {
	if status := (DocStatus)(proposal.ProposedStatus); !status.IsValid() {
		return errors.New("invalid document status: " + proposal.ProposedStatus)
	}

	return nil
}

func (proposal *Proposal) Complete() {

	proposal.ProposalID = uuid.NewString()

	if proposal.Category == "" {
		proposal.Category = DefaultCategory
	}
	if proposal.ProposedStatus == "" {
		proposal.ProposedStatus = string(DocStatusActive)
	}
	proposal.ContentHash = hashing.Calculate(proposal.Content)
}

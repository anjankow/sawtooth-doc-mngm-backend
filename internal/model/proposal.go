package model

import (
	"doc-management/internal/hashing"
	"errors"

	"github.com/google/uuid"
)

type Proposal struct {
	DocumentID
	ProposalContent
}

type ProposalContent struct {
	ProposalID         string
	ModificationAuthor string
	Content            []byte
	ContentHash        string

	ProposedStatus string
}

func (proposal Proposal) Validate() error {
	if status := (DocStatus)(proposal.ProposedStatus); !status.IsValid() {
		return errors.New("invalid document status: " + proposal.ProposedStatus)
	}

	return nil
}

func (proposal *Proposal) Complete() {
	if proposal.Category == "" {
		proposal.Category = DefaultCategory
	}
	if proposal.ProposedStatus == "" {
		proposal.ProposedStatus = string(DocStatusAccepted)
	}
	proposal.ProposalID = uuid.NewString()
	proposal.ContentHash = hashing.Calculate(proposal.Content)
}

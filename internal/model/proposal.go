package model

import (
	"doc-management/internal/hashing"
	"errors"
)

type Proposal struct {
	DocumentID
	ProposalContent
	Signers []string
}

type ProposalContent struct {
	// transaction ID is unique for each proposal and is used as a proposal ID
	TransactionID string

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
	proposal.ContentHash = hashing.Calculate(proposal.Content)
}

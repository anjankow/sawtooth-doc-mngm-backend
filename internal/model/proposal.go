package model

import (
	"doc-management/internal/hashing"
	"errors"

	"github.com/google/uuid"
)

type ProposalStatus string

const (
	ProposalStatusActive   ProposalStatus = "active"
	ProposalStatusAccepted ProposalStatus = "accepted"
	ProposalStatusRemoved  ProposalStatus = "removed"
)

type Proposal struct {
	ProposalID   string
	DocumentName string
	Category     string

	ModificationAuthor string
	Content            []byte
	ContentHash        string
	ProposedStatus     DocStatus
	CurrentStatus      ProposalStatus

	Signers []string
}

func (proposal Proposal) Validate() error {
	if proposal.ProposedStatus.IsValid() {
		return errors.New("invalid document status: " + proposal.ProposedStatus.String())
	}

	return nil
}

func (proposal *Proposal) Complete() {

	proposal.ProposalID = uuid.NewString()

	if proposal.Category == "" {
		proposal.Category = DefaultCategory
	}
	if proposal.ProposedStatus == "" {
		proposal.ProposedStatus = DocStatusActive
	}
	proposal.ContentHash = hashing.CalculateSHA512(string(proposal.Content))
}

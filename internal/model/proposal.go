package model

type Proposal struct {
	DocumentID
	ProposalContent
}

type ProposalContent struct {
	ProposalID         string
	ModificationAuthor string
	Content            []byte
	ContentHash        []byte

	ProposedStatus string
}

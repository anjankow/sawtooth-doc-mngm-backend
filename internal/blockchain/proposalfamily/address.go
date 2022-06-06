package proposalfamily

import (
	"doc-management/internal/hashing"
	"doc-management/internal/model"
	"sync"
)

var (
	familyHash             = ""
	proposalDataPrefixHash = ""
	userPrefixHash         = ""
	docPrefixHash          = ""

	calcOnce sync.Once
)

// hashing lib needs to be initialized first
func initHashVars() {
	calcOnce.Do(func() {
		familyHash = hashing.CalculateSHA512(FamilyName)
		proposalDataPrefixHash = hashing.CalculateSHA512(proposalDataPrefix)
		userPrefixHash = hashing.CalculateSHA512(userPrefix)
		docPrefixHash = hashing.CalculateSHA512(docPrefix)

	})
}

func GetDocAddress(category string, docName string) (address string) {
	initHashVars()

	categoryHash := hashing.CalculateSHA512(category)
	docNameHash := hashing.CalculateSHA512(docName)

	return familyHash[0:6] + docPrefixHash[0:6] + categoryHash[0:6] + docNameHash[0:52]
}

// GetProposalAddressFromID calculates the proposal address; if the proposal ID is empty,
// it returns the address of all the proposals
func GetProposalAddressFromID(proposalID string) (address string) {
	initHashVars()

	address = familyHash[0:6] + proposalDataPrefixHash[0:6]

	if proposalID != "" {
		proposalIDHash := hashing.CalculateSHA512(proposalID)
		address += proposalIDHash[0:58]
	}

	return address
}

func GetProposalAddress(proposal model.Proposal) (address string) {
	return GetProposalAddressFromID(proposal.ProposalID)
}

func GetUserAddress(user string) (address string) {
	initHashVars()

	userHash := hashing.CalculateSHA512(user)

	return familyHash[0:6] + userPrefixHash[0:6] + userHash[0:58]
}

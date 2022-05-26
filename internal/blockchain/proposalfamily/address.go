package proposalfamily

import (
	"doc-management/internal/hashing"
	"doc-management/internal/model"
	"sync"
)

const (
	FamilyName    string = "proposals"
	FamilyVersion string = "1.0"

	// to hold all proposal related data
	proposalDataPrefix = "proposaldata"
	// to hold proposalIDs of active and accepted proposals created by the user
	// and proposalIDs he was voting on
	userPrefix = "user"
	// to hold proposal IDs of active proposals for the doc
	docPrefix = "doc"
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
		familyHash = hashing.CalculateFromStr(FamilyName)
		proposalDataPrefixHash = hashing.CalculateFromStr(proposalDataPrefix)
		userPrefixHash = hashing.CalculateFromStr(userPrefix)
		docPrefixHash = hashing.CalculateFromStr(docPrefix)

	})
}

func GetDocAddress(category string, docName string) (address string) {
	initHashVars()

	categoryHash := hashing.CalculateFromStr(category)
	docNameHash := hashing.CalculateFromStr(docName)

	return familyHash[0:6] + docPrefixHash[0:6] + categoryHash[0:6] + docNameHash[0:52]
}

// GetProposalAddressFromID calculates the proposal address; if the proposal ID is empty,
// it returns the address of all the proposals
func GetProposalAddressFromID(proposalID string) (address string) {
	initHashVars()

	address = familyHash[0:6] + proposalDataPrefixHash[0:6]

	if proposalID != "" {
		proposalIDHash := hashing.CalculateFromStr(proposalID)
		address += proposalIDHash[0:58]
	}

	return address
}

func GetProposalAddress(proposal model.Proposal) (address string) {
	return GetProposalAddressFromID(proposal.ProposalID)
}

func GetUserAddress(user string) (address string) {
	initHashVars()

	userHash := hashing.CalculateFromStr(user)

	return familyHash[0:6] + userPrefixHash[0:6] + userHash[0:58]
}

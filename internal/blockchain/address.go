package blockchain

import (
	"doc-management/internal/hashing"
	"doc-management/internal/model"
	"sync"
)

const (
	proposalFamily string = "proposals"

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

func initHashVars() {
	calcOnce.Do(func() {
		familyHash = hashing.CalculateFromStr(proposalFamily)
		proposalDataPrefixHash = hashing.CalculateFromStr(proposalDataPrefix)
		userPrefixHash = hashing.CalculateFromStr(userPrefix)
		docPrefixHash = hashing.CalculateFromStr(docPrefix)

	})
}

func getDocAddress(proposal model.Proposal) (address string) {
	initHashVars()

	categoryHash := hashing.CalculateFromStr(proposal.Category)
	docNameHash := hashing.CalculateFromStr(proposal.DocumentName)

	return familyHash[0:6] + docPrefixHash[0:6] + categoryHash[0:6] + docNameHash[0:52]
}

func getProposalDataAddress(proposal model.Proposal) (address string) {
	initHashVars()

	proposalIDHash := hashing.CalculateFromStr(proposal.TransactionID)

	return familyHash[0:6] + proposalDataPrefixHash[0:6] + proposalIDHash[0:58]
}

func getUserAddress(user string) (address string) {
	initHashVars()

	userHash := hashing.CalculateFromStr(user)

	return familyHash[0:6] + userPrefixHash[0:6] + userHash[0:58]
}

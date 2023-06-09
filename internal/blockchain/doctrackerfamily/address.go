package doctrackerfamily

import (
	"doc-management/internal/hashing"
	"doc-management/internal/model"
	"fmt"
	"sync"
)

var (
	familyHash     = ""
	userPrefixHash = ""
	docPrefixHash  = ""

	calcOnce sync.Once
)

// hashing lib needs to be initialized first
func initHashVars() {
	calcOnce.Do(func() {
		familyHash = hashing.CalculateSHA512(FamilyName)
		userPrefixHash = hashing.CalculateSHA512(userPrefix)
		docPrefixHash = hashing.CalculateSHA512(docPrefix)
	})
}

func GetDocAddress(category string, docName string) (address string) {
	initHashVars()

	categoryHash := hashing.CalculateSHA512(category)
	docNameHash := hashing.CalculateSHA512(docName)

	return familyHash[0:6] + docPrefixHash[0:6] + categoryHash[0:6] + docNameHash[0:48]
}

func GetDocVersionAddress(doc model.Document) (address string) {
	initHashVars()

	categoryHash := hashing.CalculateSHA512(doc.Category)
	docNameHash := hashing.CalculateSHA512(doc.DocumentName)
	docVersion := fmt.Sprintf("%04d", doc.Version)

	return familyHash[0:6] + docPrefixHash[0:6] + categoryHash[0:6] + docNameHash[0:48] + docVersion[0:4]
}

func GetUserAddress(user string) (address string) {
	initHashVars()

	userHash := hashing.CalculateSHA512(user)

	return familyHash[0:6] + userPrefixHash[0:6] + userHash[0:58]
}

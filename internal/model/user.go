package model

import "doc-management/internal/signkeys"

type User struct {
	ID         string
	Name       string
	PrivateKey string
	PublicKey  string
}

func (u User) HasValidKeys() bool {
	return u.PrivateKey != "" && u.PublicKey != ""
}

func (u User) SetKeys(userKeys signkeys.UserKeys) User {
	u.PrivateKey = userKeys.PrivateKey.AsHex()
	u.PublicKey = userKeys.PublicKey.AsHex()
	return u
}

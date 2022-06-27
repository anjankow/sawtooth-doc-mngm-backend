package model

import "doc-management/internal/signkeys"

type User struct {
	ID   string
	Name string
	Keys signkeys.UserKeys
}

func (u User) HasValidKeys() bool {
	return u.Keys.Valid()
}

func (u User) SetKeys(userKeys signkeys.UserKeys) User {
	u.Keys = userKeys
	return u
}

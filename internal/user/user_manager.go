package user

type UserManager struct {
	tenantID string
	clientID string
	secret   string
	token    string
}

type User struct {
	ID         string
	Name       string
	PrivateKey string
	PublicKey  string
}

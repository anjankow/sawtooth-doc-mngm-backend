package mongodb

type DocVersion struct {
	Version       string
	TransactionID string
	Content       []byte
	Author        string
	Signers       []string
}

type Proposal struct {
	// transaction ID is used as a proposal ID
	TransactionID string
	Content       []byte
	Author        string
}

type StoredDocument struct {
	DocumentName string              `bson:"_id" json:"id"`
	Proposals    map[string]Proposal `bson:"proposals" json:"proposals"`
	Versions     []DocVersion        `bson:"versions" json:"versions"`
}

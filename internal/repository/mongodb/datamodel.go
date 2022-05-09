package mongodb

type DocVersion struct {
	Version       string
	TransactionID string
	Content       []byte
	Author        string
	Signers       []string
}

type Proposal struct {
	ProposalID    string
	TransactionID string
	Content       []byte
	Author        string
}

type StoredDocument struct {
	DocumentName string       `bson:"_id" json:"id"`
	Proposals    []Proposal   `bson:"proposals" json:"proposals"`
	Versions     []DocVersion `bson:"versions" json:"versions"`
}

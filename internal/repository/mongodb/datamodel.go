package mongodb

type DocVersion struct {
	// transaction ID on creation in doctracker family, used as a doc ID
	TransactionID string `bson:"_id" json:"id"`
	// transaction ID on creation in proposal family
	ProposalTransactionID string `bson:"proposal_id" json:"proposal_id"`

	Category string
	Name     string

	Author  string
	Version string
	Content []byte

	Signers []string
}

type Proposal struct {
	// transaction ID on creation in proposal family, used as a proposal ID
	TransactionID string `bson:"_id" json:"id"`

	Category string `bson:"category" json:"category"`
	Name     string

	Author  string `bson:"author" json:"author"`
	Content []byte

	Signers        []string
	ProposedStatus string
}

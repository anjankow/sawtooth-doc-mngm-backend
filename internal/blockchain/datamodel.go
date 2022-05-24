package blockchain

type userData struct {
	Signed   []string `cbor:"signed"`
	Accepted []string `cbor:"accepted"`
	Active   []string `cbor:"active"`
}

type proposalData struct {
	// _ struct{} `cbor:",toarray"`

	ProposalID        string   `cbor:"proposalID"`
	DocName           string   `cbor:"docName"`
	Category          string   `cbor:"category"`
	Author            string   `cbor:"author"`
	Signers           []string `cbor:"signers"`
	ProposedDocStatus string   `cbor:"proposedDocStatus"`
	CurrentStatus     string   `cbor:"currentStatus"`
	ContentHash       string   `cbor:"contentHash"`
}

type docData struct {
	Proposals map[string]string `cbor:"proposals"`
}

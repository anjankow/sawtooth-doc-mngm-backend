package proposalfamily

type Action string

const (
	ActionInsert Action = "insert"
	ActionVote   Action = "vote"
	ActionDelete Action = "delete"
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

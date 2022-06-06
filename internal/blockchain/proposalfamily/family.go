package proposalfamily

type Action string

const (
	ActionInsert Action = "insert"
	ActionVote   Action = "vote"
	ActionDelete Action = "delete"
)

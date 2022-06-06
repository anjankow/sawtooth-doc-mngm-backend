package doctrackerfamily

type Action string

const (
	ActionInsert     Action = "insert"
	ActionInvalidate Action = "invalidate"
)

const (
	FamilyName    string = "doctracker"
	FamilyVersion string = "1.0"

	// to hold all the versions of a doc
	docPrefix = "doc"
	// to hold ids of documents created by this user
	// and those that the user signed
	userPrefix = "user"
)

package project

type ProjectStatus string

const (
	NOT_INITIALIZED ProjectStatus = "NOT_INITIALIZED"
	IN_PROGRESS     ProjectStatus = "IN_PROGRESS"
	COMPLETED       ProjectStatus = "COMPLETED"
)

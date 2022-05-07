package models

type Alert struct {
	ID         uint64
	Kind       string
	Difference string
	IssuedAt   int64
	Path       string
	Modified   int64
	AgentID    uint64
}

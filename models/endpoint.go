package models

type Endpoint struct {
	ID                uint64
	Name              string
	Kind              string
	HasBaseline       bool
	BaselineIsCurrent bool
	WatchedPaths      []string
}

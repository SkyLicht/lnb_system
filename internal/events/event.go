package events

import "time"

type ChangeType string

const (
	ChangeCreated  ChangeType = "created"
	ChangeModified ChangeType = "modified"
	ChangeDeleted  ChangeType = "deleted"
)

type ChangeEvent struct {
	Watcher   string     `json:"watcher"`
	Type      ChangeType `json:"type"`
	Path      string     `json:"path"`
	Size      int64      `json:"size,omitempty"`
	OldSize   int64      `json:"oldSize,omitempty"`
	Hash      string     `json:"hash,omitempty"`
	OldHash   string     `json:"oldHash,omitempty"`
	Timestamp time.Time  `json:"timestamp"`
}

type Logger interface {
	Write(ChangeEvent) error
}

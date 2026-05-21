package watcher

import "time"

type FileState struct {
	Path    string
	Size    int64
	ModTime time.Time
	Hash    string
}

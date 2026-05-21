package watcher

import (
	"testing"
	"time"

	"lnb/internal/events"
)

func TestDiffDetectsCreatedModifiedAndDeletedFiles(t *testing.T) {
	timestamp := time.Date(2026, 5, 21, 10, 0, 0, 0, time.UTC)
	previous := map[string]FileState{
		"/tmp/deleted.txt":  {Path: "/tmp/deleted.txt", Size: 10, Hash: "old-deleted", ModTime: timestamp},
		"/tmp/modified.txt": {Path: "/tmp/modified.txt", Size: 10, Hash: "old-modified", ModTime: timestamp},
	}
	current := map[string]FileState{
		"/tmp/created.txt":  {Path: "/tmp/created.txt", Size: 20, Hash: "new-created", ModTime: timestamp},
		"/tmp/modified.txt": {Path: "/tmp/modified.txt", Size: 11, Hash: "new-modified", ModTime: timestamp.Add(time.Second)},
	}

	changes := Diff("test", previous, current, timestamp)
	seen := map[events.ChangeType]bool{}
	for _, change := range changes {
		seen[change.Type] = true
	}

	if len(changes) != 3 {
		t.Fatalf("expected 3 changes, got %d", len(changes))
	}
	if !seen[events.ChangeCreated] {
		t.Fatal("expected a created event")
	}
	if !seen[events.ChangeModified] {
		t.Fatal("expected a modified event")
	}
	if !seen[events.ChangeDeleted] {
		t.Fatal("expected a deleted event")
	}
}

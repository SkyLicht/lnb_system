package watcher

import (
	"time"

	"lnb/internal/events"
)

func Diff(watcherName string, previous, current map[string]FileState, timestamp time.Time) []events.ChangeEvent {
	changes := make([]events.ChangeEvent, 0)

	for path, currentState := range current {
		previousState, exists := previous[path]
		if !exists {
			changes = append(changes, events.ChangeEvent{
				Watcher:   watcherName,
				Type:      events.ChangeCreated,
				Path:      path,
				Size:      currentState.Size,
				Hash:      currentState.Hash,
				Timestamp: timestamp,
			})
			continue
		}

		if previousState.Hash != currentState.Hash || !previousState.ModTime.Equal(currentState.ModTime) || previousState.Size != currentState.Size {
			changes = append(changes, events.ChangeEvent{
				Watcher:   watcherName,
				Type:      events.ChangeModified,
				Path:      path,
				Size:      currentState.Size,
				OldSize:   previousState.Size,
				Hash:      currentState.Hash,
				OldHash:   previousState.Hash,
				Timestamp: timestamp,
			})
		}
	}

	for path, previousState := range previous {
		if _, exists := current[path]; !exists {
			changes = append(changes, events.ChangeEvent{
				Watcher:   watcherName,
				Type:      events.ChangeDeleted,
				Path:      path,
				OldSize:   previousState.Size,
				OldHash:   previousState.Hash,
				Timestamp: timestamp,
			})
		}
	}

	return changes
}

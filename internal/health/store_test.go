package health

import "testing"

func TestStoreSnapshotReportsAliveAndErrorWatchers(t *testing.T) {
	store := NewStore()
	store.Register("ok", "watcher_on_file", "watched/input", false)
	store.MarkAlive("ok", true)
	store.MarkScan("ok")

	store.Register("bad", "watcher_on_file_creation", `\\server\share`, true)
	store.MarkError("bad", assertError("share unavailable"))

	snapshot := store.Snapshot()
	if snapshot.Status != StatusError {
		t.Fatalf("expected overall error status, got %s", snapshot.Status)
	}
	if snapshot.WatchersOK != 1 {
		t.Fatalf("expected one ok watcher, got %d", snapshot.WatchersOK)
	}
	if snapshot.Errors != 1 {
		t.Fatalf("expected one error watcher, got %d", snapshot.Errors)
	}
}

func TestStoreSnapshotSortsWatchers(t *testing.T) {
	store := NewStore()
	store.Register("zeta", "watcher_on_file", "watched/z", false)
	store.Register("alpha", "watcher_on_file", "watched/a", false)
	store.Register("creation", "watcher_on_file_creation", "watched/c", false)

	snapshot := store.Snapshot()
	names := []string{
		snapshot.Watchers[0].Name,
		snapshot.Watchers[1].Name,
		snapshot.Watchers[2].Name,
	}

	expected := []string{"alpha", "zeta", "creation"}
	for index := range expected {
		if names[index] != expected[index] {
			t.Fatalf("expected sorted watcher names %#v, got %#v", expected, names)
		}
	}
}

type assertError string

func (e assertError) Error() string {
	return string(e)
}

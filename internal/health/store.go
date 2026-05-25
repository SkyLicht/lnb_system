package health

import (
	"encoding/json"
	"net/http"
	"sort"
	"sync"
	"time"
)

type Status string

const (
	StatusStarting Status = "starting"
	StatusAlive    Status = "alive"
	StatusStopped  Status = "stopped"
	StatusError    Status = "error"
)

type WatcherStatus struct {
	Name            string    `json:"name"`
	Feature         string    `json:"feature"`
	Status          Status    `json:"status"`
	Alive           bool      `json:"alive"`
	Path            string    `json:"path"`
	SambaEnabled    bool      `json:"sambaEnabled"`
	SambaConnected  bool      `json:"sambaConnected"`
	LastScanAt      time.Time `json:"lastScanAt,omitempty"`
	LastHeartbeatAt time.Time `json:"lastHeartbeatAt,omitempty"`
	LastError       string    `json:"lastError,omitempty"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type Snapshot struct {
	Status     Status          `json:"status"`
	CheckedAt  time.Time       `json:"checkedAt"`
	Watchers   []WatcherStatus `json:"watchers"`
	WatchersOK int             `json:"watchersOk"`
	Errors     int             `json:"errors"`
}

type Store struct {
	mutex    sync.RWMutex
	watchers map[string]WatcherStatus
}

func NewStore() *Store {
	return &Store{
		watchers: map[string]WatcherStatus{},
	}
}

func (s *Store) Register(name string, feature string, path string, sambaEnabled bool) {
	now := time.Now().UTC()
	s.update(name, func(status WatcherStatus) WatcherStatus {
		status.Name = name
		status.Feature = feature
		status.Path = path
		status.SambaEnabled = sambaEnabled
		status.Status = StatusStarting
		status.UpdatedAt = now
		return status
	})
}

func (s *Store) MarkAlive(name string, sambaConnected bool) {
	now := time.Now().UTC()
	s.update(name, func(status WatcherStatus) WatcherStatus {
		status.Status = StatusAlive
		status.Alive = true
		status.SambaConnected = sambaConnected
		status.LastHeartbeatAt = now
		status.LastError = ""
		status.UpdatedAt = now
		return status
	})
}

func (s *Store) MarkScan(name string) {
	now := time.Now().UTC()
	s.update(name, func(status WatcherStatus) WatcherStatus {
		status.Status = StatusAlive
		status.Alive = true
		status.LastScanAt = now
		status.LastHeartbeatAt = now
		status.LastError = ""
		status.UpdatedAt = now
		return status
	})
}

func (s *Store) MarkError(name string, err error) {
	now := time.Now().UTC()
	message := ""
	if err != nil {
		message = err.Error()
	}
	s.update(name, func(status WatcherStatus) WatcherStatus {
		status.Status = StatusError
		status.Alive = false
		status.LastError = message
		status.UpdatedAt = now
		return status
	})
}

func (s *Store) MarkStopped(name string) {
	now := time.Now().UTC()
	s.update(name, func(status WatcherStatus) WatcherStatus {
		status.Status = StatusStopped
		status.Alive = false
		status.UpdatedAt = now
		return status
	})
}

func (s *Store) Snapshot() Snapshot {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	snapshot := Snapshot{
		Status:    StatusAlive,
		CheckedAt: time.Now().UTC(),
		Watchers:  make([]WatcherStatus, 0, len(s.watchers)),
	}

	for _, watcher := range s.watchers {
		snapshot.Watchers = append(snapshot.Watchers, watcher)
		if watcher.Status == StatusAlive {
			snapshot.WatchersOK++
			continue
		}
		if watcher.Status == StatusError {
			snapshot.Errors++
			snapshot.Status = StatusError
		}
	}

	if len(snapshot.Watchers) == 0 {
		snapshot.Status = StatusStarting
	}

	sort.Slice(snapshot.Watchers, func(i, j int) bool {
		if snapshot.Watchers[i].Feature == snapshot.Watchers[j].Feature {
			return snapshot.Watchers[i].Name < snapshot.Watchers[j].Name
		}
		return snapshot.Watchers[i].Feature < snapshot.Watchers[j].Feature
	})

	return snapshot
}

func (s *Store) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(writer http.ResponseWriter, request *http.Request) {
		snapshot := s.Snapshot()
		statusCode := http.StatusOK
		if snapshot.Status == StatusError {
			statusCode = http.StatusServiceUnavailable
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(statusCode)
		_ = json.NewEncoder(writer).Encode(snapshot)
	})
	return mux
}

func (s *Store) update(name string, mutate func(WatcherStatus) WatcherStatus) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.watchers[name] = mutate(s.watchers[name])
}

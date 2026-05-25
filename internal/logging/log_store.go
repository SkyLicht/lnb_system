package logging

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type LogEntry struct {
	Time    time.Time         `json:"time"`
	Level   string            `json:"level"`
	Message string            `json:"message"`
	Attrs   map[string]string `json:"attrs,omitempty"`
}

type LogStore struct {
	mutex   sync.RWMutex
	limit   int
	entries []LogEntry
}

func NewLogStore(limit int) *LogStore {
	if limit <= 0 {
		limit = 500
	}
	return &LogStore{limit: limit}
}

func (s *LogStore) Add(entry LogEntry) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.entries = append(s.entries, entry)
	if len(s.entries) > s.limit {
		s.entries = append([]LogEntry{}, s.entries[len(s.entries)-s.limit:]...)
	}
}

func (s *LogStore) Entries() []LogEntry {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	entries := make([]LogEntry, len(s.entries))
	copy(entries, s.entries)
	return entries
}

type MemoryHandler struct {
	store *LogStore
	level slog.Leveler
}

func NewMemoryHandler(store *LogStore, level slog.Leveler) MemoryHandler {
	if level == nil {
		level = slog.LevelInfo
	}
	return MemoryHandler{store: store, level: level}
}

func (h MemoryHandler) Enabled(_ context.Context, level slog.Level) bool {
	return h.store != nil && level >= h.level.Level()
}

func (h MemoryHandler) Handle(_ context.Context, record slog.Record) error {
	attrs := map[string]string{}
	record.Attrs(func(attr slog.Attr) bool {
		attr.Value = attr.Value.Resolve()
		attrs[attr.Key] = formatValue(attr.Value)
		return true
	})

	h.store.Add(LogEntry{
		Time:    record.Time,
		Level:   record.Level.String(),
		Message: record.Message,
		Attrs:   attrs,
	})
	return nil
}

func (h MemoryHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h MemoryHandler) WithGroup(name string) slog.Handler {
	return h
}

type TeeHandler struct {
	handlers []slog.Handler
}

func NewTeeHandler(handlers ...slog.Handler) TeeHandler {
	return TeeHandler{handlers: handlers}
}

func (h TeeHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h TeeHandler) Handle(ctx context.Context, record slog.Record) error {
	for _, handler := range h.handlers {
		if !handler.Enabled(ctx, record.Level) {
			continue
		}
		if err := handler.Handle(ctx, record); err != nil {
			return err
		}
	}
	return nil
}

func (h TeeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, 0, len(h.handlers))
	for _, handler := range h.handlers {
		handlers = append(handlers, handler.WithAttrs(attrs))
	}
	return TeeHandler{handlers: handlers}
}

func (h TeeHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, 0, len(h.handlers))
	for _, handler := range h.handlers {
		handlers = append(handlers, handler.WithGroup(name))
	}
	return TeeHandler{handlers: handlers}
}

package watcher

import (
	"context"
	"log/slog"
	"path/filepath"
	"time"

	"lnb/internal/config"
	"lnb/internal/events"
)

type Watcher struct {
	config   config.PathConfig
	interval time.Duration
	events   events.Logger
	logger   *slog.Logger
}

func New(pathConfig config.PathConfig, interval time.Duration, eventLogger events.Logger, logger *slog.Logger) Watcher {
	return Watcher{
		config:   pathConfig,
		interval: interval,
		events:   eventLogger,
		logger:   logger,
	}
}

func (w Watcher) Run(ctx context.Context) {
	root, err := filepath.Abs(w.config.Path)
	if err != nil {
		w.logger.Error("invalid watcher path", "watcher", w.config.Name, "path", w.config.Path, "error", err)
		return
	}

	scanner := NewScanner(root, w.config)
	previous, err := scanner.Scan()
	if err != nil {
		w.logger.Error("failed to initialize watcher", "watcher", w.config.Name, "path", root, "error", err)
		return
	}

	w.logger.Info("watcher started", "watcher", w.config.Name, "path", root, "recursive", w.config.Recursive)
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("watcher stopped", "watcher", w.config.Name)
			return
		case <-ticker.C:
			current, err := scanner.Scan()
			if err != nil {
				w.logger.Error("failed to scan watcher path", "watcher", w.config.Name, "path", root, "error", err)
				continue
			}

			changes := Diff(w.config.Name, previous, current, time.Now().UTC())
			for _, change := range changes {
				if err := w.events.Write(change); err != nil {
					w.logger.Error("failed to write change event", "watcher", w.config.Name, "path", change.Path, "error", err)
				}
			}

			previous = current
		}
	}
}

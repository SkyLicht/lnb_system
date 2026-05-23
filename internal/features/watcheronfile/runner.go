package watcheronfile

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"lnb/internal/config"
	"lnb/internal/events"
	"lnb/internal/functions"
	"lnb/internal/ioaccess"
	"lnb/internal/watcher"
)

type Runner struct {
	config      config.WatcherOnFileConfig
	interval    time.Duration
	logger      *slog.Logger
	eventLogger events.Logger
	functions   functions.Executor
}

func New(featureConfig config.WatcherOnFileConfig, logger *slog.Logger, eventLogger events.Logger, functions functions.Executor) Runner {
	return Runner{
		config:      featureConfig,
		interval:    time.Duration(featureConfig.PollIntervalMs) * time.Millisecond,
		logger:      logger,
		eventLogger: eventLogger,
		functions:   functions,
	}
}

func (r Runner) Run(ctx context.Context) {
	if err := ioaccess.Ensure(r.config.Input, r.logger); err != nil {
		r.logger.Error("failed to access watcher_on_file input", "watcher", r.config.Name, "path", r.config.Input.Path, "error", err)
		return
	}
	if err := ioaccess.Ensure(r.config.Output, r.logger); err != nil {
		r.logger.Error("failed to access watcher_on_file output", "watcher", r.config.Name, "path", r.config.Output.Path, "error", err)
		return
	}

	root, err := filepath.Abs(r.config.Input.Path)
	if err != nil {
		r.logger.Error("invalid watcher_on_file path", "watcher", r.config.Name, "path", r.config.Input.Path, "error", err)
		return
	}

	targetPath := r.resolveTargetPath(root)
	scanner := watcher.NewScanner(root, watcher.ScanOptions{
		Recursive: r.config.Recursive,
		Ignore:    r.config.Ignore,
	})

	previous, err := scanner.Scan()
	if err != nil {
		r.logger.Error("failed to initialize watcher_on_file", "watcher", r.config.Name, "path", root, "error", err)
		return
	}

	r.logger.Info("watcher_on_file started", "watcher", r.config.Name, "path", root, "file", targetPath, "recursive", r.config.Recursive)

	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			r.logger.Info("watcher_on_file stopped", "watcher", r.config.Name)
			return
		case <-ticker.C:
			current, err := scanner.Scan()
			if err != nil {
				r.logger.Error("failed to scan watcher_on_file", "watcher", r.config.Name, "path", root, "error", err)
				continue
			}

			for path, currentState := range current {
				if filepath.Clean(path) != targetPath {
					continue
				}

				previousState, exists := previous[path]
				if !exists {
					continue
				}

				if previousState.Hash == currentState.Hash &&
					previousState.Size == currentState.Size &&
					previousState.ModTime.Equal(currentState.ModTime) {
					continue
				}

				content, err := os.ReadFile(path)
				if err != nil {
					r.logger.Error("failed to read modified file", "watcher", r.config.Name, "path", path, "error", err)
					continue
				}

				if err := r.functions.Execute(r.config.Function, functions.Payload{
					Ref:        r.config.ID,
					Feature:    "watcher_on_file",
					Watcher:    r.config.Name,
					Path:       path,
					OutputPath: r.config.Output.Path,
					Content:    string(content),
				}); err != nil {
					r.logger.Error("failed to execute function", "watcher", r.config.Name, "function", r.config.Function, "path", path, "error", err)
				}

				if err := r.eventLogger.Write(events.ChangeEvent{
					Watcher:   r.config.Name,
					Type:      events.ChangeModified,
					Path:      path,
					Size:      currentState.Size,
					OldSize:   previousState.Size,
					Hash:      currentState.Hash,
					OldHash:   previousState.Hash,
					Timestamp: time.Now().UTC(),
				}); err != nil {
					r.logger.Error("failed to write modification event", "watcher", r.config.Name, "path", path, "error", err)
				}
			}

			previous = current
		}
	}
}

func (r Runner) resolveTargetPath(root string) string {
	if filepath.IsAbs(r.config.File) {
		return filepath.Clean(r.config.File)
	}

	return filepath.Clean(filepath.Join(root, r.config.File))
}

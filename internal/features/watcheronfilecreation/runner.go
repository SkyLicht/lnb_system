package watcheronfilecreation

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"lnb/internal/config"
	"lnb/internal/functions"
	"lnb/internal/ioaccess"
	"lnb/internal/watcher"
)

type Runner struct {
	config    config.WatcherOnFileCreationConfig
	interval  time.Duration
	logger    *slog.Logger
	functions functions.Executor
}

func New(featureConfig config.WatcherOnFileCreationConfig, logger *slog.Logger, functions functions.Executor) Runner {
	return Runner{
		config:    featureConfig,
		interval:  time.Duration(featureConfig.PollIntervalMs) * time.Millisecond,
		logger:    logger,
		functions: functions,
	}
}

func (r Runner) Run(ctx context.Context) {
	if err := ioaccess.Ensure(r.config.Input, r.logger); err != nil {
		r.logger.Error("failed to access watcher_on_file_creation input", "watcher", r.config.Name, "path", r.config.Input.Path, "error", err)
		return
	}
	if err := ioaccess.Ensure(r.config.Output, r.logger); err != nil {
		r.logger.Error("failed to access watcher_on_file_creation output", "watcher", r.config.Name, "path", r.config.Output.Path, "error", err)
		return
	}

	root, err := filepath.Abs(r.config.Input.Path)
	if err != nil {
		r.logger.Error("invalid watcher_on_file_creation path", "watcher", r.config.Name, "path", r.config.Input.Path, "error", err)
		return
	}

	scanner := watcher.NewScanner(root, watcher.ScanOptions{
		Recursive: r.config.Recursive,
		Ignore:    r.config.Ignore,
	})

	previous, err := scanner.Scan()
	if err != nil {
		r.logger.Error("failed to initialize watcher_on_file_creation", "watcher", r.config.Name, "path", root, "error", err)
		return
	}

	r.logger.Info("watcher_on_file_creation started", "watcher", r.config.Name, "path", root, "recursive", r.config.Recursive)

	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			r.logger.Info("watcher_on_file_creation stopped", "watcher", r.config.Name)
			return
		case <-ticker.C:
			current, err := scanner.Scan()
			if err != nil {
				r.logger.Error("failed to scan watcher_on_file_creation", "watcher", r.config.Name, "path", root, "error", err)
				continue
			}

			for path := range current {
				if _, exists := previous[path]; exists {
					continue
				}

				content, err := os.ReadFile(path)
				if err != nil {
					r.logger.Error("failed to read new file", "watcher", r.config.Name, "path", path, "error", err)
					continue
				}

				if err := r.functions.Execute(r.config.Function, functions.Payload{
					Ref:        r.config.ID,
					Feature:    "watcher_on_file_creation",
					Watcher:    r.config.Name,
					Path:       path,
					OutputPath: r.config.Output.Path,
					Content:    string(content),
				}); err != nil {
					r.logger.Error("failed to execute function", "watcher", r.config.Name, "function", r.config.Function, "path", path, "error", err)
				}
			}

			previous = current
		}
	}
}

package watcheronfilecreation

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"lnb/internal/config"
	"lnb/internal/functions"
	"lnb/internal/health"
	"lnb/internal/ioaccess"
	"lnb/internal/watcher"
)

type Runner struct {
	config    config.WatcherOnFileCreationConfig
	interval  time.Duration
	logger    *slog.Logger
	functions functions.Executor
	health    *health.Store
}

func New(featureConfig config.WatcherOnFileCreationConfig, logger *slog.Logger, functions functions.Executor, healthStore *health.Store) Runner {
	return Runner{
		config:    featureConfig,
		interval:  time.Duration(featureConfig.PollIntervalMs) * time.Millisecond,
		logger:    logger,
		functions: functions,
		health:    healthStore,
	}
}

func (r Runner) Run(ctx context.Context) error {
	r.registerHealth()
	if ok, err := r.waitForInputAccess(ctx); !ok {
		return err
	}

	root, err := filepath.Abs(r.config.Input.Path)
	if err != nil {
		r.markHealthError(err)
		return fmt.Errorf("invalid watcher_on_file_creation path %q at %s: %w", r.config.Name, r.config.Input.Path, err)
	}

	scanner := watcher.NewScanner(root, watcher.ScanOptions{
		Recursive: r.config.Recursive,
		Ignore:    r.config.Ignore,
	})

	previous, err := scanner.Scan()
	if err != nil {
		r.markHealthError(err)
		return fmt.Errorf("failed to initialize watcher_on_file_creation %q at %s: %w", r.config.Name, root, err)
	}
	r.markHealthScan()
	pending := map[string]watcher.FileState{}
	failures := map[string]int{}

	r.logger.Info("watcher_on_file_creation started", "watcher", r.config.Name, "path", root, "recursive", r.config.Recursive, "retry", r.config.Retry, "accepted", r.config.Accepted)

	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			r.logger.Info("watcher_on_file_creation stopped", "watcher", r.config.Name)
			r.markHealthStopped()
			return nil
		case <-ticker.C:
			current, err := scanner.Scan()
			if err != nil {
				r.logger.Error("failed to scan watcher_on_file_creation", "watcher", r.config.Name, "path", root, "error", err)
				r.markHealthError(err)
				continue
			}
			r.markHealthScan()

			for path, currentState := range current {
				if _, exists := previous[path]; exists {
					continue
				}

				if !r.acceptsPath(path) {
					r.logger.Warn("new file skipped because extension is not accepted", "watcher", r.config.Name, "path", path, "accepted", r.config.Accepted)
					delete(pending, path)
					delete(failures, path)
					continue
				}

				pendingState, exists := pending[path]
				if !exists || !sameFileState(pendingState, currentState) {
					pending[path] = currentState
					continue
				}

				if r.processCreatedFile(path) {
					delete(pending, path)
					delete(failures, path)
					continue
				}

				failures[path]++
				if failures[path] >= r.config.Retry {
					r.logger.Error(
						"file marked as bad after retry limit",
						"watcher", r.config.Name,
						"path", path,
						"retry", r.config.Retry,
						"failures", failures[path],
					)
					delete(pending, path)
					delete(failures, path)
				}
			}

			for path := range pending {
				if _, exists := current[path]; !exists {
					delete(pending, path)
					delete(failures, path)
				}
			}

			previous = current
			for path := range pending {
				delete(previous, path)
			}
		}
	}
}

func (r Runner) waitForInputAccess(ctx context.Context) (bool, error) {
	for {
		if err := ioaccess.Ensure(r.config.Input, r.logger); err != nil {
			r.markHealthError(err)
			if !r.config.IgnoreFailedConnections {
				return false, fmt.Errorf("failed to access watcher_on_file_creation input %q at %s: %w", r.config.Name, r.config.Input.Path, err)
			}

			r.logger.Error(
				"failed to access watcher_on_file_creation input; watcher will retry",
				"watcher", r.config.Name,
				"path", r.config.Input.Path,
				"error", err,
				"retry_after", r.interval.String(),
			)
			select {
			case <-ctx.Done():
				r.markHealthStopped()
				return false, nil
			case <-time.After(r.interval):
				continue
			}
		}

		r.markHealthAlive()
		return true, nil
	}
}

func (r Runner) processCreatedFile(path string) bool {
	content, err := os.ReadFile(path)
	if err != nil {
		r.logger.Error("failed to read new file", "watcher", r.config.Name, "path", path, "error", err)
		return false
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
		return false
	}

	return true
}

func sameFileState(left watcher.FileState, right watcher.FileState) bool {
	return left.Size == right.Size &&
		left.Hash == right.Hash &&
		left.ModTime.Equal(right.ModTime)
}

func (r Runner) acceptsPath(path string) bool {
	if len(r.config.Accepted) == 0 {
		return true
	}

	extension := strings.ToLower(filepath.Ext(path))
	for _, accepted := range r.config.Accepted {
		if extension == accepted {
			return true
		}
	}
	return false
}

func (r Runner) registerHealth() {
	if r.health != nil {
		r.health.Register(r.config.Name, "watcher_on_file_creation", r.config.Input.Path, r.config.Input.Samba)
	}
}

func (r Runner) markHealthAlive() {
	if r.health != nil {
		r.health.MarkAlive(r.config.Name, !r.config.Input.Samba || ioaccess.IsAccessible(r.config.Input.Path))
	}
}

func (r Runner) markHealthScan() {
	if r.health != nil {
		r.health.MarkScan(r.config.Name)
	}
}

func (r Runner) markHealthError(err error) {
	if r.health != nil {
		r.health.MarkError(r.config.Name, err)
	}
}

func (r Runner) markHealthStopped() {
	if r.health != nil {
		r.health.MarkStopped(r.config.Name)
	}
}

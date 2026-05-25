package watcheronfile

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"lnb/internal/config"
	"lnb/internal/events"
	"lnb/internal/functions"
	"lnb/internal/health"
	"lnb/internal/ioaccess"
	"lnb/internal/watcher"
)

type Runner struct {
	config      config.WatcherOnFileConfig
	interval    time.Duration
	logger      *slog.Logger
	eventLogger events.Logger
	functions   functions.Executor
	health      *health.Store
}

func New(featureConfig config.WatcherOnFileConfig, logger *slog.Logger, eventLogger events.Logger, functions functions.Executor, healthStore *health.Store) Runner {
	return Runner{
		config:      featureConfig,
		interval:    time.Duration(featureConfig.PollIntervalMs) * time.Millisecond,
		logger:      logger,
		eventLogger: eventLogger,
		functions:   functions,
		health:      healthStore,
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
		return fmt.Errorf("invalid watcher_on_file path %q at %s: %w", r.config.Name, r.config.Input.Path, err)
	}

	targetPath := r.resolveTargetPath(root)
	if !r.acceptsPath(targetPath) {
		err := fmt.Errorf("target %s is not in accepted extensions %v", targetPath, r.config.Accepted)
		r.markHealthError(err)
		return fmt.Errorf("watcher_on_file %q target %s is not in accepted extensions %v", r.config.Name, targetPath, r.config.Accepted)
	}
	if err := ensureRootExists(root); err != nil {
		r.markHealthError(err)
		return fmt.Errorf("failed to initialize watcher_on_file %q at %s: %w", r.config.Name, root, err)
	}

	previous, err := r.scanTarget(targetPath, nil)
	if err != nil {
		r.markHealthError(err)
		return fmt.Errorf("failed to initialize watcher_on_file %q at %s: %w", r.config.Name, targetPath, err)
	}
	r.markHealthScan()

	r.logger.Info("watcher_on_file started", "watcher", r.config.Name, "path", root, "file", targetPath, "recursive", r.config.Recursive)

	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			r.logger.Info("watcher_on_file stopped", "watcher", r.config.Name)
			r.markHealthStopped()
			return nil
		case <-ticker.C:
			current, err := r.scanTarget(targetPath, previous)
			if err != nil {
				r.logger.Error("failed to scan watcher_on_file", "watcher", r.config.Name, "path", targetPath, "error", err)
				r.markHealthError(err)
				continue
			}
			r.markHealthScan()

			changes := watcher.Diff(r.config.Name, previous, current, time.Now().UTC())
			processed := true
			for _, change := range changes {
				if err := r.processChange(change); err != nil {
					r.logger.Error("failed to process target file change", "watcher", r.config.Name, "type", change.Type, "path", change.Path, "error", err)
					processed = false
				}
			}

			if processed {
				previous = current
			}
		}
	}
}

func (r Runner) waitForInputAccess(ctx context.Context) (bool, error) {
	for {
		if err := ioaccess.Ensure(r.config.Input, r.logger); err != nil {
			r.markHealthError(err)
			if !r.config.IgnoreFailedConnections {
				return false, fmt.Errorf("failed to access watcher_on_file input %q at %s: %w", r.config.Name, r.config.Input.Path, err)
			}

			r.logger.Error(
				"failed to access watcher_on_file input; watcher will retry",
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

func (r Runner) resolveTargetPath(root string) string {
	if filepath.IsAbs(r.config.File) {
		return filepath.Clean(r.config.File)
	}

	return filepath.Clean(filepath.Join(root, r.config.File))
}

func (r Runner) scanTarget(targetPath string, previous map[string]watcher.FileState) (map[string]watcher.FileState, error) {
	var previousState *watcher.FileState
	if state, exists := previous[targetPath]; exists {
		previousState = &state
	}

	state, _, err := watcher.SnapshotFile(targetPath, previousState)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]watcher.FileState{}, nil
		}
		return nil, err
	}

	return map[string]watcher.FileState{targetPath: state}, nil
}

func (r Runner) processChange(change events.ChangeEvent) error {
	if change.Type == events.ChangeCreated || change.Type == events.ChangeModified {
		content, err := os.ReadFile(change.Path)
		if err != nil {
			return fmt.Errorf("failed to read changed file: %w", err)
		}

		if err := r.functions.Execute(r.config.Function, functions.Payload{
			Ref:        r.config.ID,
			Feature:    "watcher_on_file",
			Watcher:    r.config.Name,
			Path:       change.Path,
			OutputPath: r.config.Output.Path,
			Content:    string(content),
		}); err != nil {
			return fmt.Errorf("failed to execute function %q: %w", r.config.Function, err)
		}
	}

	if err := r.eventLogger.Write(change); err != nil {
		r.logger.Error("failed to write change event", "watcher", r.config.Name, "type", change.Type, "path", change.Path, "error", err)
	}

	return nil
}

func ensureRootExists(root string) error {
	info, err := os.Stat(root)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("expected input path to be a directory")
	}
	return nil
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
		r.health.Register(r.config.Name, "watcher_on_file", r.config.Input.Path, r.config.Input.Samba)
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

package watcher

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type ScanOptions struct {
	Recursive bool
	Ignore    []string
}

type Scanner struct {
	root    string
	options ScanOptions
}

func NewScanner(root string, options ScanOptions) Scanner {
	return Scanner{
		root:    root,
		options: options,
	}
}

func (s Scanner) Scan() (map[string]FileState, error) {
	info, err := os.Stat(s.root)
	if err != nil {
		return nil, err
	}

	states := make(map[string]FileState)
	if !info.IsDir() {
		state, err := buildFileState(s.root)
		if err != nil {
			return nil, err
		}
		states[s.root] = state
		return states, nil
	}

	err = filepath.WalkDir(s.root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if path != s.root && shouldIgnore(path, s.options.Ignore) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if entry.IsDir() {
			if path != s.root && !s.options.Recursive {
				return filepath.SkipDir
			}
			return nil
		}

		state, err := buildFileState(path)
		if err != nil {
			return err
		}

		states[path] = state
		return nil
	})
	if err != nil {
		return nil, err
	}

	return states, nil
}

func buildFileState(path string) (FileState, error) {
	state, _, err := SnapshotFile(path, nil)
	return state, err
}

func SnapshotFile(path string, previous *FileState) (FileState, bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return FileState{}, false, err
	}
	if info.IsDir() {
		return FileState{}, false, errors.New("expected a file, got a directory")
	}

	if previous != nil && previous.Size == info.Size() && previous.ModTime.Equal(info.ModTime()) {
		return *previous, false, nil
	}

	hash, err := hashFile(path)
	if err != nil {
		return FileState{}, false, err
	}

	return FileState{
		Path:    path,
		Size:    info.Size(),
		ModTime: info.ModTime(),
		Hash:    hash,
	}, true, nil
}

func shouldIgnore(path string, patterns []string) bool {
	base := filepath.Base(path)
	for _, pattern := range patterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}

		matchedBase, _ := filepath.Match(pattern, base)
		matchedPath, _ := filepath.Match(pattern, path)
		if matchedBase || matchedPath || strings.Contains(path, pattern) {
			return true
		}
	}

	return false
}

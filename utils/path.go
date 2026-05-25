package utils

import (
	"errors"
	"os"
)

type PathStatus struct {
	Path  string
	Found bool
	IsDir bool
	Size  int64
}

func TestPathExistence(path string) (PathStatus, error) {
	if path == "" {
		return PathStatus{}, errors.New("path is required")
	}

	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return PathStatus{Path: path, Found: false}, nil
		}
		return PathStatus{Path: path, Found: false}, err
	}

	return PathStatus{
		Path:  path,
		Found: true,
		IsDir: info.IsDir(),
		Size:  info.Size(),
	}, nil
}

func buildPathStatus(path string, info os.FileInfo) PathStatus {
	return PathStatus{
		Path:  path,
		Found: true,
		IsDir: info.IsDir(),
		Size:  info.Size(),
	}
}

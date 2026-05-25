package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPathExistenceReturnsFoundForExistingFile(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "sample.txt")

	if err := os.WriteFile(filePath, []byte("test"), 0o600); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	status, err := TestPathExistence(filePath)
	if err != nil {
		t.Fatalf("test path existence: %v", err)
	}

	if !status.Found {
		t.Fatal("expected existing file to be found")
	}
	if status.IsDir {
		t.Fatal("expected file status, got directory")
	}
	if status.Size == 0 {
		t.Fatal("expected file size to be reported")
	}
}

func TestPathExistenceReturnsNotFoundForMissingPath(t *testing.T) {
	status, err := TestPathExistence(filepath.Join(t.TempDir(), "missing.txt"))
	if err != nil {
		t.Fatalf("test path existence: %v", err)
	}

	if status.Found {
		t.Fatal("expected missing path not to be found")
	}
}

func TestPathExistenceRequiresPath(t *testing.T) {
	if _, err := TestPathExistence(""); err == nil {
		t.Fatal("expected empty path to return an error")
	}
}

func TestNetworkShareRootReturnsUNCShare(t *testing.T) {
	share, ok := NetworkShareRoot(`\\server\share\folder\file.txt`)
	if !ok {
		t.Fatal("expected UNC path to be accepted")
	}
	if share != `\\server\share` {
		t.Fatalf("expected share root, got %q", share)
	}
}

func TestNetworkShareRootAcceptsForwardSlashes(t *testing.T) {
	share, ok := NetworkShareRoot(`//server/share/folder`)
	if !ok {
		t.Fatal("expected network path to be accepted")
	}
	if share != `\\server\share` {
		t.Fatalf("expected share root, got %q", share)
	}
}

func TestNetworkShareRootRejectsLocalPath(t *testing.T) {
	if _, ok := NetworkShareRoot(`C:\data\file.txt`); ok {
		t.Fatal("expected local path to be rejected")
	}
}

package envfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadReadsEnvValues(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	content := "# comment\nWATCHER_CONFIG=configs/watcher.config.json\nQUOTED=\"value with spaces\"\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	values, err := Load(path)
	if err != nil {
		t.Fatalf("load env file: %v", err)
	}

	if values["WATCHER_CONFIG"] != "configs/watcher.config.json" {
		t.Fatalf("expected WATCHER_CONFIG value, got %q", values["WATCHER_CONFIG"])
	}
	if values["QUOTED"] != "value with spaces" {
		t.Fatalf("expected quoted value to be trimmed, got %q", values["QUOTED"])
	}
}

func TestRequiredRejectsMissingValue(t *testing.T) {
	if _, err := Required(map[string]string{}, "WATCHER_CONFIG"); err == nil {
		t.Fatal("expected missing WATCHER_CONFIG to return an error")
	}
}

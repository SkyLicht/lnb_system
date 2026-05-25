package logging

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestConsoleHandlerWritesReadableLogLine(t *testing.T) {
	var output bytes.Buffer
	handler := NewConsoleHandler(&output, slog.LevelInfo, false)
	record := slog.NewRecord(time.Date(2026, 5, 24, 23, 30, 0, 0, time.UTC), slog.LevelInfo, "watcher started", 0)
	record.AddAttrs(slog.String("watcher", "panasonic_top"))

	if err := handler.Handle(context.Background(), record); err != nil {
		t.Fatalf("handle record: %v", err)
	}

	line := output.String()
	for _, expected := range []string{"23:30:00.000", "INFO", "watcher started", "watcher=panasonic_top"} {
		if !strings.Contains(line, expected) {
			t.Fatalf("expected log line to contain %q, got %q", expected, line)
		}
	}
}

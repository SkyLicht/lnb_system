package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"lnb/internal/envfile"
	"lnb/internal/health"
)

func main() {
	url, err := healthURL()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load watcher status URL: %v\n", err)
		os.Exit(1)
	}
	client := http.Client{Timeout: 5 * time.Second}

	response, err := client.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read watcher health from %s: %v\n", url, err)
		os.Exit(1)
	}
	defer response.Body.Close()

	var snapshot health.Snapshot
	if err := json.NewDecoder(response.Body).Decode(&snapshot); err != nil {
		fmt.Fprintf(os.Stderr, "failed to decode watcher health response: %v\n", err)
		os.Exit(1)
	}

	printSnapshot(url, snapshot)
	if response.StatusCode >= http.StatusBadRequest || snapshot.Status == health.StatusError {
		os.Exit(1)
	}
}

func healthURL() (string, error) {
	values, err := envfile.Load(".env")
	if err != nil {
		return "", err
	}

	value, err := envfile.Required(values, "WATCHER_HTTP_URL")
	if err != nil {
		return "", err
	}
	return strings.TrimRight(value, "/") + "/health", nil
}

func printSnapshot(url string, snapshot health.Snapshot) {
	fmt.Println("Watcher Status")
	fmt.Println("==============")
	fmt.Printf("Endpoint:    %s\n", url)
	fmt.Printf("Status:      %s\n", snapshot.Status)
	fmt.Printf("Checked at:  %s\n", formatTime(snapshot.CheckedAt))
	fmt.Printf("Watchers:    %d ok, %d errors, %d total\n", snapshot.WatchersOK, snapshot.Errors, len(snapshot.Watchers))
	fmt.Println()

	for _, watcher := range snapshot.Watchers {
		fmt.Printf("- %s [%s]\n", watcher.Name, watcher.Feature)
		fmt.Printf("  Alive:          %t\n", watcher.Alive)
		fmt.Printf("  Status:         %s\n", watcher.Status)
		fmt.Printf("  Path:           %s\n", watcher.Path)
		fmt.Printf("  Samba enabled:  %t\n", watcher.SambaEnabled)
		fmt.Printf("  Samba connected: %t\n", watcher.SambaConnected)
		fmt.Printf("  Last scan:      %s\n", formatTime(watcher.LastScanAt))
		fmt.Printf("  Heartbeat:      %s\n", formatTime(watcher.LastHeartbeatAt))
		if watcher.LastError != "" {
			fmt.Printf("  Last error:     %s\n", watcher.LastError)
		}
		fmt.Println()
	}
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return "(none)"
	}
	return value.Local().Format(time.RFC3339)
}

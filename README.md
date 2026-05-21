# File Watcher

Production-oriented Go service that watches multiple configured file paths and writes detected file changes as JSON Lines events.

## Project Structure

```text
cmd/file-watcher/      Application entry point
configs/               Runtime configuration examples
deployments/           Deployment artifacts
internal/app/          Application orchestration
internal/config/       Configuration loading and validation
internal/events/       Event models and event writers
internal/watcher/      File scanning, hashing, and change detection
watched/               Local sample folders
```

## Run Locally

```powershell
go run ./cmd/file-watcher -config configs/watcher.config.json
```

The service writes operational logs to stdout and change events to both stdout and `logs/file-changes.jsonl`.

## Configuration

```json
{
  "pollIntervalMs": 1000,
  "logFile": "logs/file-changes.jsonl",
  "paths": [
    {
      "name": "input",
      "path": "watched/input",
      "recursive": true,
      "ignore": ["*.tmp", ".git"]
    }
  ]
}
```

## Event Format

```json
{
  "watcher": "input",
  "type": "modified",
  "path": "C:\\code\\lnb\\watched\\input\\file.txt",
  "size": 42,
  "oldSize": 18,
  "hash": "new-sha256",
  "oldHash": "old-sha256",
  "timestamp": "2026-05-21T16:00:00Z"
}
```

## Test

```powershell
go test ./...
```

## Build

```powershell
go build -o bin/file-watcher.exe ./cmd/file-watcher
```

## Docker

```powershell
docker build -f deployments/Dockerfile -t file-watcher .
docker run --rm -v ${PWD}/watched:/app/watched -v ${PWD}/logs:/app/logs file-watcher
```

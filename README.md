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
internal/functions/    Function handlers and registry
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
  "logFile": "logs/file-changes.jsonl",
  "os": "windows",
  "shutdownTimeoutMs": 15000,
  "watcher_on_file_creation": [
    {
      "pollIntervalMs": 1000,
      "name": "creation_logs",
      "function": "lnb_type1_logs",
      "input": {
        "samba": true,
        "path": "\\\\10.13.104.240\\lws",
        "credentials": true,
        "user": "domain\\user",
        "pass": "secret"
      },
      "output": {
        "samba": false,
        "path": "C:\\data\\logs\\panasonic\\J01",
        "credentials": false,
        "user": "",
        "pass": ""
      },
      "recursive": true,
      "ignore": ["*.tmp", ".git"]
    }
  ],
  "watcher_on_file": [
    {
      "pollIntervalMs": 1000,
      "name": "input",
      "file": "log_ss",
      "function": "log_to_console",
      "input": {
        "samba": false,
        "path": "watched/input",
        "credentials": false,
        "user": "",
        "pass": ""
      },
      "output": {
        "samba": false,
        "path": "",
        "credentials": false,
        "user": "",
        "pass": ""
      },
      "recursive": true,
      "ignore": ["*.tmp", ".git"]
    }
  ]
}
```

- `watcher_on_file_creation`: triggers when a new file is created and logs the file content to stdout.
- `watcher_on_file`: triggers when the configured `file` is modified and logs the updated content to stdout.
- `input`: source location options, including Samba and optional credentials.
- `output`: destination location options, including Samba and optional credentials.
- `function`: selects how the trigger is handled. Implemented now: `log_to_console`, `lnb_type1_logs`.
- `shutdownTimeoutMs`: max time to wait for graceful shutdown before returning an error.

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

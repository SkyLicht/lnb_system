# File Watcher

Production-oriented Go service that watches multiple configured file paths and writes detected file changes as JSON Lines events.

## Project Structure

```text
cmd/file-watcher/      Application entry point
cmd/config-print/      Pretty printer for the watcher configuration
cmd/path-test/         Utility entry point for checking path existence
cmd/status/            Health/status command for the running watcher
configs/               Runtime configuration examples
deployments/           Deployment artifacts
internal/app/          Application orchestration
internal/config/       Configuration loading and validation
internal/events/       Event models and event writers
internal/functions/    Function handlers and registry
internal/watcher/      File scanning, hashing, and change detection
utils/                 Shared utility helpers
watched/               Local sample folders
```

## Run Locally

```powershell
go run ./cmd/file-watcher
```

The config file path is read from `.env`:

```text
WATCHER_CONFIG=configs/watcher.config.json
WATCHER_HTTP_URL=http://127.0.0.1:8080
```

The service writes readable console logs to stdout and change events to both stdout and `logs/file-changes.jsonl`.

For machine-readable operational logs:

```powershell
go run ./cmd/file-watcher -log-format json
```

To disable console colors:

```powershell
go run ./cmd/file-watcher -no-color
```

To print the loaded watcher configuration in a readable format:

```powershell
go run ./cmd/config-print
```

To read watcher alive and Samba connection status from the running service:

```powershell
go run ./cmd/status
```

The service also exposes a packaged dashboard and JSON APIs:

```text
http://127.0.0.1:8080/
http://127.0.0.1:8080/health
http://127.0.0.1:8080/logs
```

## Test Path Existence

```powershell
go run ./cmd/path-test -path watched/input
```

For a network path:

```powershell
go run ./cmd/path-test -network -path "\\server\share\folder"
```

To connect the share before testing it on Windows:

```powershell
go run ./cmd/path-test -network -connect -path "\\server\share\folder" -user "domain\user" -pass "secret"
```

The command exits with status `0` when the path exists, `1` when it does not exist, and `2` when the path argument is invalid or cannot be checked.

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
      "retry": 3,
      "accepted": [".xml"],
      "ignoreFailedConnections": true,
      "input": {
        "samba": true,
        "path": "\\\\10.13.104.240\\lws",
        "credentials": true,
        "user": "domain\\user",
        "pass": "secret"
      },
      "output": {
        "path": "C:\\data\\logs\\panasonic\\J01"
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
        "path": ""
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
- `output`: local destination path.
- `function`: selects how the trigger is handled. Implemented now: `log_to_console`, `lnb_type1_logs`.
- `retry`: number of failed processing attempts before a new file is marked as bad and no longer retried. Defaults to `3`.
- `accepted`: optional list of accepted file extensions. Empty means all extensions are accepted. Values can be written as `"xml"` or `".xml"`.
- `ignoreFailedConnections`: when `true`, input connection failures are logged and retried without stopping the whole app.
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

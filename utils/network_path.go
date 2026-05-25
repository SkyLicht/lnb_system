package utils

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type NetworkPathOptions struct {
	Path           string
	Connect        bool
	UseCredentials bool
	User           string
	Pass           string
}

type NetworkPathStatus struct {
	PathStatus
	Share     string
	Connected bool
}

func TestNetworkPathExistence(options NetworkPathOptions) (NetworkPathStatus, error) {
	path := strings.TrimSpace(options.Path)
	if path == "" {
		return NetworkPathStatus{}, errors.New("path is required")
	}

	share, ok := NetworkShareRoot(path)
	if !ok {
		return NetworkPathStatus{}, fmt.Errorf("network path must be a UNC path like \\\\server\\share")
	}

	status, err := statNetworkPath(path, share)
	if err == nil {
		return status, nil
	}
	if !options.Connect {
		return status, err
	}

	if runtime.GOOS != "windows" {
		return status, fmt.Errorf("auto-connect is only supported on windows; mount the share first: %w", err)
	}

	if err := connectNetworkShare(share, options); err != nil {
		return status, err
	}

	status, err = statNetworkPath(path, share)
	status.Connected = true
	return status, err
}

func NetworkShareRoot(path string) (string, bool) {
	normalized := strings.ReplaceAll(strings.TrimSpace(path), "/", "\\")
	if !strings.HasPrefix(normalized, `\\`) {
		return "", false
	}

	parts := strings.Split(strings.TrimPrefix(normalized, `\\`), `\`)
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", false
	}

	return `\\` + parts[0] + `\` + parts[1], true
}

func statNetworkPath(path string, share string) (NetworkPathStatus, error) {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return NetworkPathStatus{
				PathStatus: PathStatus{Path: path, Found: false},
				Share:      share,
			}, nil
		}
		return NetworkPathStatus{
			PathStatus: PathStatus{Path: path, Found: false},
			Share:      share,
		}, err
	}

	return NetworkPathStatus{
		PathStatus: buildPathStatus(path, info),
		Share:      share,
	}, nil
}

func connectNetworkShare(share string, options NetworkPathOptions) error {
	args := []string{"use", share}
	if options.UseCredentials {
		user := strings.TrimSpace(options.User)
		if user == "" {
			return errors.New("user is required when credentials are enabled")
		}
		args = append(args, "/user:"+user, options.Pass)
	}
	args = append(args, "/persistent:no")

	output, err := exec.Command("net", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to connect network share %s: %w (%s)", share, err, strings.TrimSpace(string(output)))
	}

	return nil
}

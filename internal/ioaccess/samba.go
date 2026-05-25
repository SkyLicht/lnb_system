package ioaccess

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"lnb/internal/config"
)

func Ensure(location config.IOConfig, logger *slog.Logger) error {
	if !location.Samba {
		return nil
	}

	if location.Path == "" {
		return errorsPathRequired()
	}

	if _, err := os.Stat(location.Path); err == nil {
		return nil
	}

	if runtime.GOOS != "windows" {
		return fmt.Errorf("samba auto-connect is only supported on windows runtime; path must be pre-mounted: %s", location.Path)
	}

	share, ok := ShareRoot(location.Path)
	if !ok {
		return fmt.Errorf("samba path must be a UNC path like \\\\server\\share: %s", location.Path)
	}

	commandArgs := []string{"use", share}
	if location.Credentials {
		commandArgs = append(commandArgs, "/user:"+location.User, location.Pass)
	}
	commandArgs = append(commandArgs, "/persistent:no")

	command := exec.Command("net", commandArgs...)
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to connect samba share %s: %w (%s)", share, err, strings.TrimSpace(string(output)))
	}

	if _, err := os.Stat(location.Path); err != nil {
		return fmt.Errorf("samba share %s is still not accessible after connect: %w", location.Path, err)
	}

	return nil
}

func IsAccessible(path string) bool {
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

func errorsPathRequired() error {
	return fmt.Errorf("samba location requires a non-empty path")
}

func ShareRoot(path string) (string, bool) {
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

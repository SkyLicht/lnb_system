package ioaccess

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"runtime"

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
		logger.Warn("samba auto-connect is only supported on windows runtime; path must be pre-mounted", "path", location.Path)
		return nil
	}

	commandArgs := []string{"use", location.Path}
	if location.Credentials {
		commandArgs = append(commandArgs, "/user:"+location.User, location.Pass)
	}
	commandArgs = append(commandArgs, "/persistent:no")

	command := exec.Command("net", commandArgs...)
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to connect samba share %s: %w (%s)", location.Path, err, string(output))
	}

	if _, err := os.Stat(location.Path); err != nil {
		return fmt.Errorf("samba share %s is still not accessible after connect: %w", location.Path, err)
	}

	return nil
}

func errorsPathRequired() error {
	return fmt.Errorf("samba location requires a non-empty path")
}

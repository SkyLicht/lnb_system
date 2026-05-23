package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type Config struct {
	LogFile               string                        `json:"logFile"`
	OS                    string                        `json:"os"`
	ShutdownTimeoutMs     int                           `json:"shutdownTimeoutMs"`
	WatcherOnFileCreation []WatcherOnFileCreationConfig `json:"watcher_on_file_creation"`
	WatcherOnFile         []WatcherOnFileConfig         `json:"watcher_on_file"`
}

type WatchTarget struct {
	ID             string   `json:"id"`
	PollIntervalMs int      `json:"pollIntervalMs"`
	Name           string   `json:"name"`
	Input          IOConfig `json:"input"`
	Output         IOConfig `json:"output"`
	Path           string   `json:"path"`
	OutputPath     string   `json:"output_path"`
	Function       string   `json:"function"`
	Recursive      bool     `json:"recursive"`
	Ignore         []string `json:"ignore"`
}

type IOConfig struct {
	Samba       bool   `json:"samba"`
	Path        string `json:"path"`
	Credentials bool   `json:"credentials"`
	User        string `json:"user"`
	Pass        string `json:"pass"`
}

type WatcherOnFileCreationConfig struct {
	WatchTarget
}

type WatcherOnFileConfig struct {
	WatchTarget
	File string `json:"file"`
}

func Load(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c *Config) ApplyDefaults() {
	c.OS = normalizeOS(c.OS)
	if c.OS == "" {
		c.OS = runtimeOS()
	}

	if c.ShutdownTimeoutMs <= 0 {
		c.ShutdownTimeoutMs = 15000
	}

	for i := range c.WatcherOnFileCreation {
		applyWatchTargetDefaults(&c.WatcherOnFileCreation[i].WatchTarget, c.OS)
	}

	for i := range c.WatcherOnFile {
		applyWatchTargetDefaults(&c.WatcherOnFile[i].WatchTarget, c.OS)
		c.WatcherOnFile[i].File = normalizePathForOS(strings.TrimSpace(c.WatcherOnFile[i].File), c.OS)
	}
}

func (c Config) Validate() error {
	if c.OS != "windows" && c.OS != "linux" {
		return errors.New("os must be one of: windows, linux")
	}

	if c.ShutdownTimeoutMs < 1000 {
		return errors.New("shutdownTimeoutMs must be at least 1000 milliseconds")
	}

	if len(c.WatcherOnFileCreation) == 0 && len(c.WatcherOnFile) == 0 {
		return errors.New("at least one watcher must be configured in watcher_on_file_creation or watcher_on_file")
	}

	for index, watcherConfig := range c.WatcherOnFileCreation {
		if err := validateWatchTarget("watcher_on_file_creation", index, watcherConfig.WatchTarget); err != nil {
			return err
		}
	}

	for index, watcherConfig := range c.WatcherOnFile {
		if err := validateWatchTarget("watcher_on_file", index, watcherConfig.WatchTarget); err != nil {
			return err
		}
		if watcherConfig.File == "" {
			return fmt.Errorf("watcher_on_file[%d].file is required", index)
		}
	}

	return nil
}

func applyWatchTargetDefaults(target *WatchTarget, osName string) {
	target.ID = strings.TrimSpace(target.ID)
	target.Name = strings.TrimSpace(target.Name)
	target.Path = normalizePathForOS(strings.TrimSpace(target.Path), osName)
	target.OutputPath = normalizePathForOS(strings.TrimSpace(target.OutputPath), osName)
	target.Function = strings.TrimSpace(target.Function)
	target.Input.Path = normalizePathForOS(strings.TrimSpace(target.Input.Path), osName)
	target.Input.User = strings.TrimSpace(target.Input.User)
	target.Input.Pass = strings.TrimSpace(target.Input.Pass)
	target.Output.Path = normalizePathForOS(strings.TrimSpace(target.Output.Path), osName)
	target.Output.User = strings.TrimSpace(target.Output.User)
	target.Output.Pass = strings.TrimSpace(target.Output.Pass)

	if target.PollIntervalMs <= 0 {
		target.PollIntervalMs = 1000
	}

	if target.Input.Path == "" && target.Path != "" {
		target.Input.Path = target.Path
	}
	if target.Output.Path == "" && target.OutputPath != "" {
		target.Output.Path = target.OutputPath
	}

	if target.Name == "" {
		target.Name = target.Input.Path
	}

	if target.Function == "" {
		target.Function = "log_to_console"
	}
}

func validateWatchTarget(group string, index int, target WatchTarget) error {
	if target.PollIntervalMs < 100 {
		return fmt.Errorf("%s[%d].pollIntervalMs must be at least 100 milliseconds", group, index)
	}
	if target.Input.Path == "" {
		return fmt.Errorf("%s[%d].input.path is required", group, index)
	}
	if target.Name == "" {
		return fmt.Errorf("%s[%d].name is required", group, index)
	}
	if target.Function == "" {
		return fmt.Errorf("%s[%d].function is required", group, index)
	}
	if target.Input.Credentials {
		if target.Input.User == "" {
			return fmt.Errorf("%s[%d].input.user is required when input.credentials is true", group, index)
		}
		if target.Input.Pass == "" {
			return fmt.Errorf("%s[%d].input.pass is required when input.credentials is true", group, index)
		}
	}
	if target.Output.Credentials {
		if target.Output.User == "" {
			return fmt.Errorf("%s[%d].output.user is required when output.credentials is true", group, index)
		}
		if target.Output.Pass == "" {
			return fmt.Errorf("%s[%d].output.pass is required when output.credentials is true", group, index)
		}
	}

	return nil
}

func normalizeOS(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func runtimeOS() string {
	if runtime.GOOS == "windows" {
		return "windows"
	}
	return "linux"
}

func normalizePathForOS(pathValue string, osName string) string {
	if pathValue == "" {
		return ""
	}

	normalized := pathValue
	if osName == "windows" {
		normalized = strings.ReplaceAll(normalized, "/", "\\")
	} else if osName == "linux" {
		normalized = strings.ReplaceAll(normalized, "\\", "/")
	}

	return filepath.Clean(normalized)
}

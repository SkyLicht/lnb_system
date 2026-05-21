package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	PollIntervalMs int          `json:"pollIntervalMs"`
	LogFile        string       `json:"logFile"`
	Paths          []PathConfig `json:"paths"`
}

type PathConfig struct {
	Name      string   `json:"name"`
	Path      string   `json:"path"`
	Recursive bool     `json:"recursive"`
	Ignore    []string `json:"ignore"`
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
	if c.PollIntervalMs <= 0 {
		c.PollIntervalMs = 1000
	}

	for i := range c.Paths {
		if strings.TrimSpace(c.Paths[i].Name) == "" {
			c.Paths[i].Name = c.Paths[i].Path
		}
	}
}

func (c Config) Validate() error {
	if c.PollIntervalMs < 100 {
		return errors.New("pollIntervalMs must be at least 100 milliseconds")
	}

	if len(c.Paths) == 0 {
		return errors.New("at least one path must be configured")
	}

	for index, pathConfig := range c.Paths {
		if strings.TrimSpace(pathConfig.Path) == "" {
			return fmt.Errorf("paths[%d].path is required", index)
		}
		if strings.TrimSpace(pathConfig.Name) == "" {
			return fmt.Errorf("paths[%d].name is required", index)
		}
	}

	return nil
}

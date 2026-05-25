package main

import (
	"fmt"
	"os"
	"strings"

	"lnb/internal/config"
	"lnb/internal/envfile"
)

func main() {
	configPath, err := configPathFromEnv(".env")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load .env configuration: %v\n", err)
		os.Exit(1)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load watcher config: %v\n", err)
		os.Exit(1)
	}

	cfg.ApplyDefaults()
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "invalid watcher config: %v\n", err)
		os.Exit(1)
	}

	printConfig(configPath, cfg)
}

func configPathFromEnv(path string) (string, error) {
	values, err := envfile.Load(path)
	if err != nil {
		return "", fmt.Errorf("load %s: %w", path, err)
	}
	return envfile.Required(values, "WATCHER_CONFIG")
}

func printConfig(path string, cfg config.Config) {
	fmt.Println("Watcher Configuration")
	fmt.Println("=====================")
	fmt.Printf("Config file:          %s\n", path)
	fmt.Printf("OS:                   %s\n", cfg.OS)
	fmt.Printf("Log file:             %s\n", emptyValue(cfg.LogFile))
	fmt.Printf("Shutdown timeout:     %d ms\n", cfg.ShutdownTimeoutMs)
	fmt.Printf("Creation watchers:    %d\n", len(cfg.WatcherOnFileCreation))
	fmt.Printf("File watchers:        %d\n", len(cfg.WatcherOnFile))
	fmt.Println()

	if len(cfg.WatcherOnFileCreation) > 0 {
		fmt.Println("watcher_on_file_creation")
		fmt.Println("------------------------")
		for index, watcherConfig := range cfg.WatcherOnFileCreation {
			printWatchTarget(index+1, watcherConfig.WatchTarget, "")
		}
	}

	if len(cfg.WatcherOnFile) > 0 {
		fmt.Println("watcher_on_file")
		fmt.Println("---------------")
		for index, watcherConfig := range cfg.WatcherOnFile {
			printWatchTarget(index+1, watcherConfig.WatchTarget, watcherConfig.File)
		}
	}
}

func printWatchTarget(index int, target config.WatchTarget, file string) {
	fmt.Printf("%d. %s\n", index, target.Name)
	fmt.Printf("   ID:             %s\n", emptyValue(target.ID))
	if file != "" {
		fmt.Printf("   File:           %s\n", file)
	}
	fmt.Printf("   Function:       %s\n", target.Function)
	fmt.Printf("   Poll interval:  %d ms\n", target.PollIntervalMs)
	fmt.Printf("   Retry:          %d\n", target.Retry)
	fmt.Printf("   Ignore conn:    %t\n", target.IgnoreFailedConnections)
	fmt.Printf("   Recursive:      %t\n", target.Recursive)
	fmt.Printf("   Accepted:       %s\n", listValue(target.Accepted))
	fmt.Printf("   Ignore:         %s\n", listValue(target.Ignore))
	fmt.Println("   Input:")
	fmt.Printf("     Path:         %s\n", target.Input.Path)
	fmt.Printf("     Samba:        %t\n", target.Input.Samba)
	fmt.Printf("     Credentials:  %t\n", target.Input.Credentials)
	if target.Input.Credentials {
		fmt.Printf("     User:         %s\n", emptyValue(target.Input.User))
		fmt.Printf("     Pass:         %s\n", maskSecret(target.Input.Pass))
	}
	fmt.Println("   Output:")
	fmt.Printf("     Path:         %s\n", emptyValue(target.Output.Path))
	fmt.Println()
}

func emptyValue(value string) string {
	if strings.TrimSpace(value) == "" {
		return "(empty)"
	}
	return value
}

func listValue(values []string) string {
	if len(values) == 0 {
		return "(all)"
	}
	return strings.Join(values, ", ")
}

func maskSecret(value string) string {
	if strings.TrimSpace(value) == "" {
		return "(empty)"
	}
	return "********"
}

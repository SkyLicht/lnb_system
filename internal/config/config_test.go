package config

import "testing"

func TestValidateAllowsWatcherOnFileCreation(t *testing.T) {
	cfg := Config{
		WatcherOnFileCreation: []WatcherOnFileCreationConfig{
			{
				WatchTarget: WatchTarget{
					PollIntervalMs: 1000,
					Name:           "input",
					Path:           "watched/input",
					Function:       "log_to_console",
					Recursive:      true,
				},
			},
		},
	}

	cfg.ApplyDefaults()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected watcher_on_file_creation config to be valid, got error: %v", err)
	}
}

func TestValidateAllowsWatcherOnFile(t *testing.T) {
	cfg := Config{
		WatcherOnFile: []WatcherOnFileConfig{
			{
				WatchTarget: WatchTarget{
					PollIntervalMs: 1000,
					Name:           "input",
					Path:           "watched/input",
					Function:       "log_to_console",
					Recursive:      true,
				},
				File: "log_ss",
			},
		},
	}

	cfg.ApplyDefaults()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected watcher_on_file config to be valid, got error: %v", err)
	}
}

func TestValidateRequiresAtLeastOneWatcher(t *testing.T) {
	cfg := Config{}
	cfg.ApplyDefaults()

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error when no watcher_on_file_creation or watcher_on_file is configured")
	}
}

func TestValidateRequiresFileInWatcherOnFile(t *testing.T) {
	cfg := Config{
		WatcherOnFile: []WatcherOnFileConfig{
			{
				WatchTarget: WatchTarget{
					PollIntervalMs: 1000,
					Name:           "input",
					Path:           "watched/input",
					Function:       "log_to_console",
					Recursive:      true,
				},
			},
		},
	}

	cfg.ApplyDefaults()
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error when watcher_on_file.file is empty")
	}
}

func TestApplyDefaultsSetsFunction(t *testing.T) {
	cfg := Config{
		WatcherOnFileCreation: []WatcherOnFileCreationConfig{
			{
				WatchTarget: WatchTarget{
					PollIntervalMs: 1000,
					Name:           "input",
					Path:           "watched/input",
				},
			},
		},
	}

	cfg.ApplyDefaults()
	if cfg.WatcherOnFileCreation[0].Function != "log_to_console" {
		t.Fatalf("expected default function to be log_to_console, got %q", cfg.WatcherOnFileCreation[0].Function)
	}
}

func TestApplyDefaultsSetsShutdownTimeout(t *testing.T) {
	cfg := Config{
		WatcherOnFileCreation: []WatcherOnFileCreationConfig{
			{
				WatchTarget: WatchTarget{
					PollIntervalMs: 1000,
					Name:           "input",
					Path:           "watched/input",
				},
			},
		},
	}

	cfg.ApplyDefaults()
	if cfg.ShutdownTimeoutMs != 15000 {
		t.Fatalf("expected default shutdownTimeoutMs to be 15000, got %d", cfg.ShutdownTimeoutMs)
	}
}

func TestValidateRejectsLowShutdownTimeout(t *testing.T) {
	cfg := Config{
		ShutdownTimeoutMs: 500,
		WatcherOnFileCreation: []WatcherOnFileCreationConfig{
			{
				WatchTarget: WatchTarget{
					PollIntervalMs: 1000,
					Name:           "input",
					Path:           "watched/input",
					Function:       "log_to_console",
				},
			},
		},
	}

	cfg.ApplyDefaults()
	cfg.ShutdownTimeoutMs = 500
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error when shutdownTimeoutMs is less than 1000")
	}
}

func TestApplyDefaultsSetsOSFromRuntime(t *testing.T) {
	cfg := Config{
		WatcherOnFileCreation: []WatcherOnFileCreationConfig{
			{
				WatchTarget: WatchTarget{
					PollIntervalMs: 1000,
					Name:           "input",
					Path:           "watched/input",
				},
			},
		},
	}

	cfg.ApplyDefaults()
	if cfg.OS == "" {
		t.Fatal("expected default os to be set")
	}
}

func TestValidateRejectsUnknownOS(t *testing.T) {
	cfg := Config{
		OS: "mac",
		WatcherOnFileCreation: []WatcherOnFileCreationConfig{
			{
				WatchTarget: WatchTarget{
					PollIntervalMs: 1000,
					Name:           "input",
					Path:           "watched/input",
					Function:       "log_to_console",
				},
			},
		},
	}

	cfg.ApplyDefaults()
	cfg.OS = "mac"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error when os is not windows or linux")
	}
}

func TestValidateRequiresInputCredentialsFields(t *testing.T) {
	cfg := Config{
		WatcherOnFileCreation: []WatcherOnFileCreationConfig{
			{
				WatchTarget: WatchTarget{
					PollIntervalMs: 1000,
					Name:           "input",
					Path:           "watched/input",
					Function:       "log_to_console",
					Input: IOConfig{
						Credentials: true,
					},
				},
			},
		},
	}

	cfg.ApplyDefaults()
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error when input.credentials is true and input.user/input.pass are empty")
	}
}

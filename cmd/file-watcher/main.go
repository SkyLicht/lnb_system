package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"lnb/internal/app"
	"lnb/internal/config"
	"lnb/internal/events"
)

func main() {
	configPath := flag.String("config", "configs/watcher.config.json", "path to the JSON configuration file")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	cfg.ApplyDefaults()
	if err := cfg.Validate(); err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	eventLogger, err := events.NewJSONLLogger(cfg.LogFile, os.Stdout)
	if err != nil {
		logger.Error("failed to initialize event logger", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := eventLogger.Close(); err != nil {
			logger.Error("failed to close event logger", "error", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signals := make(chan os.Signal, 2)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)

	go func() {
		firstSignal := <-signals
		logger.Info("shutdown requested", "signal", firstSignal.String())
		cancel()

		secondSignal := <-signals
		logger.Error("forcing shutdown after second signal", "signal", secondSignal.String())
		os.Exit(1)
	}()

	runner := app.NewRunner(cfg, logger, eventLogger)
	startedAt := time.Now()
	totalFeatures := len(cfg.WatcherOnFileCreation) + len(cfg.WatcherOnFile)

	logger.Info(
		"starting file watcher",
		"features", totalFeatures,
		"watcher_on_file_creation", len(cfg.WatcherOnFileCreation),
		"watcher_on_file", len(cfg.WatcherOnFile),
		"shutdown_timeout_ms", cfg.ShutdownTimeoutMs,
		"config", *configPath,
	)
	if err := runner.Run(ctx); err != nil {
		logger.Error("file watcher stopped with an error", "error", err)
		os.Exit(1)
	}

	logger.Info("file watcher stopped", "uptime", time.Since(startedAt).String())
}

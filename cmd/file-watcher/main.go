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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	runner := app.NewRunner(cfg, logger, eventLogger)
	startedAt := time.Now()

	logger.Info("starting file watcher", "watchers", len(cfg.Paths), "config", *configPath)
	if err := runner.Run(ctx); err != nil {
		logger.Error("file watcher stopped with an error", "error", err)
		os.Exit(1)
	}

	logger.Info("file watcher stopped", "uptime", time.Since(startedAt).String())
}

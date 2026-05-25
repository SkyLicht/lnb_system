package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"lnb/internal/app"
	"lnb/internal/config"
	"lnb/internal/envfile"
	"lnb/internal/events"
	"lnb/internal/health"
	"lnb/internal/logging"
	"lnb/internal/ui"
)

func main() {
	logFormat := flag.String("log-format", "console", "log format: console or json")
	noColor := flag.Bool("no-color", false, "disable colored console logs")
	flag.Parse()

	envValues, err := envfile.Load(".env")
	logStore := logging.NewLogStore(600)
	logger := newLogger(*logFormat, !*noColor, logStore)
	if err != nil {
		logger.Error("failed to load .env configuration", "error", err)
		os.Exit(1)
	}

	configPath, err := envfile.Required(envValues, "WATCHER_CONFIG")
	if err != nil {
		logger.Error("failed to load .env configuration", "error", err)
		os.Exit(1)
	}
	httpURL, err := envfile.Required(envValues, "WATCHER_HTTP_URL")
	if err != nil {
		logger.Error("failed to load .env configuration", "error", err)
		os.Exit(1)
	}
	httpAddress, err := addressFromHTTPURL(httpURL)
	if err != nil {
		logger.Error("invalid WATCHER_HTTP_URL", "url", httpURL, "error", err)
		os.Exit(1)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	cfg.ApplyDefaults()
	if err := cfg.Validate(); err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	healthStore := health.NewStore()
	healthServer := startHTTPServer(httpAddress, httpURL, healthStore, logStore, logger)
	defer shutdownHealthServer(healthServer, logger)

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

	runner := app.NewRunner(cfg, logger, eventLogger, healthStore)
	startedAt := time.Now()
	totalFeatures := len(cfg.WatcherOnFileCreation) + len(cfg.WatcherOnFile)

	logger.Info(
		"starting file watcher",
		"features", totalFeatures,
		"watcher_on_file_creation", len(cfg.WatcherOnFileCreation),
		"watcher_on_file", len(cfg.WatcherOnFile),
		"shutdown_timeout_ms", cfg.ShutdownTimeoutMs,
		"config", configPath,
		"dashboard", httpURL,
	)
	if err := runner.Run(ctx); err != nil {
		logger.Error("file watcher stopped with an error", "error", err)
		os.Exit(1)
	}

	logger.Info("file watcher stopped", "uptime", time.Since(startedAt).String())
}

func startHTTPServer(address string, publicURL string, healthStore *health.Store, logStore *logging.LogStore, logger *slog.Logger) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/", ui.Handler())
	mux.Handle("/health", healthStore.Handler())
	mux.HandleFunc("/logs", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(map[string]any{
			"logs": logStore.Entries(),
		})
	})

	server := &http.Server{
		Addr:              address,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("watcher dashboard started", "address", address, "url", publicURL, "health", publicURL+"/health")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("watcher dashboard stopped with an error", "error", err)
		}
	}()

	return server
}

func shutdownHealthServer(server *http.Server, logger *slog.Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("failed to stop health endpoint", "error", err)
	}
}

func addressFromHTTPURL(value string) (string, error) {
	parsed, err := url.Parse(value)
	if err != nil {
		return "", err
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("url must start with http:// or https://")
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("url host is required")
	}
	return parsed.Host, nil
}

func newLogger(format string, color bool, logStore *logging.LogStore) *slog.Logger {
	var handler slog.Handler
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	default:
		handler = logging.NewConsoleHandler(os.Stdout, slog.LevelInfo, color)
	}
	return slog.New(logging.NewTeeHandler(handler, logging.NewMemoryHandler(logStore, slog.LevelInfo)))
}

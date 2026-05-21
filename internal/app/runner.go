package app

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"lnb/internal/config"
	"lnb/internal/events"
	"lnb/internal/watcher"
)

type Runner struct {
	config      config.Config
	logger      *slog.Logger
	eventLogger events.Logger
}

func NewRunner(cfg config.Config, logger *slog.Logger, eventLogger events.Logger) Runner {
	return Runner{
		config:      cfg,
		logger:      logger,
		eventLogger: eventLogger,
	}
}

func (r Runner) Run(ctx context.Context) error {
	interval := time.Duration(r.config.PollIntervalMs) * time.Millisecond

	var wg sync.WaitGroup
	for _, pathConfig := range r.config.Paths {
		pathConfig := pathConfig
		wg.Add(1)

		go func() {
			defer wg.Done()

			engine := watcher.New(pathConfig, interval, r.eventLogger, r.logger)
			engine.Run(ctx)
		}()
	}

	<-ctx.Done()
	r.logger.Info("shutdown signal received")
	wg.Wait()
	return nil
}

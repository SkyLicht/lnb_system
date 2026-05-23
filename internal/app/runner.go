package app

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"lnb/internal/config"
	"lnb/internal/events"
	"lnb/internal/features"
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
	featureRunners := features.Build(r.config, r.logger, r.eventLogger)

	var wg sync.WaitGroup
	for _, featureRunner := range featureRunners {
		featureRunner := featureRunner
		wg.Add(1)

		go func() {
			defer wg.Done()
			featureRunner.Run(ctx)
		}()
	}

	<-ctx.Done()
	r.logger.Info("shutdown signal received")

	waitDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitDone)
	}()

	shutdownTimeout := time.Duration(r.config.ShutdownTimeoutMs) * time.Millisecond
	select {
	case <-waitDone:
		r.logger.Info("graceful shutdown completed")
		return nil
	case <-time.After(shutdownTimeout):
		return fmt.Errorf("graceful shutdown timed out after %s", shutdownTimeout)
	}
}

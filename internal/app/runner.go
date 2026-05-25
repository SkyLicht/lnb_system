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
	"lnb/internal/health"
)

type Runner struct {
	config      config.Config
	logger      *slog.Logger
	eventLogger events.Logger
	healthStore *health.Store
}

func NewRunner(cfg config.Config, logger *slog.Logger, eventLogger events.Logger, healthStore *health.Store) Runner {
	return Runner{
		config:      cfg,
		logger:      logger,
		eventLogger: eventLogger,
		healthStore: healthStore,
	}
}

func (r Runner) Run(ctx context.Context) error {
	featureRunners := features.Build(r.config, r.logger, r.eventLogger, r.healthStore)
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	errs := make(chan error, len(featureRunners))
	for _, featureRunner := range featureRunners {
		featureRunner := featureRunner
		wg.Add(1)

		go func() {
			defer wg.Done()
			if err := featureRunner.Run(runCtx); err != nil {
				errs <- err
				cancel()
			}
		}()
	}

	select {
	case err := <-errs:
		r.logger.Error("feature runner stopped with an error", "error", err)
		if waitErr := r.waitForRunners(&wg); waitErr != nil {
			return fmt.Errorf("%w; %v", err, waitErr)
		}
		return err
	case <-ctx.Done():
		r.logger.Info("shutdown signal received")
	}

	return r.waitForRunners(&wg)
}

func (r Runner) waitForRunners(wg *sync.WaitGroup) error {
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

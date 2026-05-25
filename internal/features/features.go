package features

import (
	"context"
	"log/slog"

	"lnb/internal/config"
	"lnb/internal/events"
	"lnb/internal/features/watcheronfile"
	"lnb/internal/features/watcheronfilecreation"
	"lnb/internal/functions"
	"lnb/internal/health"
)

type Runner interface {
	Run(ctx context.Context) error
}

func Build(cfg config.Config, logger *slog.Logger, eventLogger events.Logger, healthStore *health.Store) []Runner {

	runners := make([]Runner, 0, len(cfg.WatcherOnFileCreation)+len(cfg.WatcherOnFile))
	functionExecutor := functions.NewExecutor(logger)

	for _, watcherConfig := range cfg.WatcherOnFileCreation {
		runners = append(runners, watcheronfilecreation.New(watcherConfig, logger, functionExecutor, healthStore))
	}

	for _, watcherConfig := range cfg.WatcherOnFile {
		runners = append(runners, watcheronfile.New(watcherConfig, logger, eventLogger, functionExecutor, healthStore))
	}

	return runners
}

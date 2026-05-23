package features

import (
	"context"
	"log/slog"

	"lnb/internal/config"
	"lnb/internal/events"
	"lnb/internal/features/watcheronfile"
	"lnb/internal/features/watcheronfilecreation"
	"lnb/internal/functions"
)

type Runner interface {
	Run(ctx context.Context)
}

func Build(cfg config.Config, logger *slog.Logger, eventLogger events.Logger) []Runner {

	runners := make([]Runner, 0, len(cfg.WatcherOnFileCreation)+len(cfg.WatcherOnFile))
	functionExecutor := functions.NewExecutor(logger)

	for _, watcherConfig := range cfg.WatcherOnFileCreation {
		runners = append(runners, watcheronfilecreation.New(watcherConfig, logger, functionExecutor))
	}

	for _, watcherConfig := range cfg.WatcherOnFile {
		runners = append(runners, watcheronfile.New(watcherConfig, logger, eventLogger, functionExecutor))
	}

	return runners
}

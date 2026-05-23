package functions

import "log/slog"

type LogToConsole struct {
	logger *slog.Logger
}

func NewLogToConsole(logger *slog.Logger) LogToConsole {
	return LogToConsole{logger: logger}
}

func (l LogToConsole) Handle(payload Payload) error {
	l.logger.Info(
		"feature function executed",
		"function", FunctionLogToConsole,
		"feature", payload.Feature,
		"watcher", payload.Watcher,
		"path", payload.Path,
		"content", payload.Content,
	)

	return nil
}

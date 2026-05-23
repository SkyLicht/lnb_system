package functions

import (
	"fmt"
	"log/slog"
)

const (
	FunctionLogToConsole = "log_to_console"
	FunctionLNBType1Logs = "lnb_type1_logs"
)

type Payload struct {
	Ref        string
	Feature    string
	Watcher    string
	Path       string
	OutputPath string
	Content    string
}

type Handler interface {
	Handle(Payload) error
}

type Executor struct {
	handlers map[string]Handler
}

func NewExecutor(logger *slog.Logger) Executor {
	return Executor{
		handlers: map[string]Handler{
			FunctionLogToConsole: NewLogToConsole(logger),
			FunctionLNBType1Logs: NewLNBType1Logs(logger),
		},
	}
}

func (e Executor) Execute(name string, payload Payload) error {
	handler, exists := e.handlers[name]
	if !exists {
		return fmt.Errorf("unknown function %q", name)
	}

	return handler.Handle(payload)
}

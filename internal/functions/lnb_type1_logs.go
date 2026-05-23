package functions

import (
	"log/slog"

	"lnb/internal/functions/lnb"
)

type LNBType1Logs struct {
	handler lnb.Type1Logs
}

func NewLNBType1Logs(logger *slog.Logger) LNBType1Logs {
	return LNBType1Logs{
		handler: lnb.NewType1Logs(logger),
	}
}

func (h LNBType1Logs) Handle(payload Payload) error {
	return h.handler.HandleFile(payload.Ref, payload.Path, payload.OutputPath, payload.Content)
}

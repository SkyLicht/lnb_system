package type1events

import "log/slog"

type Event struct {
	EventCode       string
	EventDetailCode string
}

type Action func(logger *slog.Logger, ref string, filePath string, outputPath string, event Event, content string) error

func BuildKey(eventCode string, eventDetailCode string) string {
	return eventCode + ":" + eventDetailCode
}

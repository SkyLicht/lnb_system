package lnb

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"lnb/internal/functions/lnb/type1events"
	"log/slog"
	"strings"
)

type Type1Logs struct {
	logger         *slog.Logger
	lnbType1Events map[string]type1events.Action
}

func NewType1Logs(logger *slog.Logger) Type1Logs {
	return Type1Logs{
		logger:         logger,
		lnbType1Events: type1events.LNBType1Events,
	}
}

func (h Type1Logs) HandleFile(ref string, path string, outputPath string, content string) error {
	event, err := parseType1Event(content)
	if err != nil {
		return err
	}

	eventKey := type1events.BuildKey(event.EventCode, event.EventDetailCode)
	action, exists := h.lnbType1Events[eventKey]
	if !exists {
		h.logger.Warn(
			"type1 event not mapped",
			"event_code", event.EventCode,
			"event_detail_code", event.EventDetailCode,
		)
		return nil
	}

	return action(h.logger, ref, path, outputPath, event, content)
}

func parseType1Event(content string) (type1events.Event, error) {
	decoder := xml.NewDecoder(strings.NewReader(content))
	values := map[string]string{}
	currentElement := ""

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return type1events.Event{}, fmt.Errorf("failed to decode lnb xml: %w", err)
		}

		switch token := token.(type) {
		case xml.StartElement:
			currentElement = token.Name.Local
		case xml.CharData:
			if currentElement == "" {
				continue
			}
			value := strings.TrimSpace(string(token))
			if value == "" {
				continue
			}
			values[currentElement] = value
		case xml.EndElement:
			if token.Name.Local == currentElement {
				currentElement = ""
			}
		}
	}

	eventCode := strings.TrimSpace(values["EventCode"])
	eventDetailCode := strings.TrimSpace(values["EventDetailCode"])
	if eventCode == "" || eventDetailCode == "" {
		return type1events.Event{}, errorsMissingEventFields(eventCode, eventDetailCode)
	}

	return type1events.Event{
		EventCode:       eventCode,
		EventDetailCode: eventDetailCode,
	}, nil
}

func errorsMissingEventFields(eventCode string, eventDetailCode string) error {
	if eventCode == "" && eventDetailCode == "" {
		return errors.New("xml does not include EventCode and EventDetailCode")
	}
	if eventCode == "" {
		return errors.New("xml does not include EventCode")
	}
	return errors.New("xml does not include EventDetailCode")
}

package type1events

import (
	"encoding/xml"
	"fmt"
	"log/slog"
	"path/filepath"
)

type Event04000000Details struct {
	Date               string `xml:"Date"`
	MDLN               string `xml:"MDLN"`
	EventSerial        string `xml:"EventSerial"`
	EventCode          string `xml:"EventCode"`
	EventDetailCode    string `xml:"EventDetailCode"`
	PcbSerial          string `xml:"PcbSerial"`
	Stage              string `xml:"Stage"`
	Lane               string `xml:"Lane"`
	CurrentPcbPosition string `xml:"CurrentPcbPosition"`
	ProductBoardCount  string `xml:"ProductBoardCount"`
	CycleTime1         string `xml:"CycleTime1"`
	CycleTime2         string `xml:"CycleTime2"`
	Lot                string `xml:"Lot"`
	ProductMode        string `xml:"ProductMode"`
	MCNo               string `xml:"MCNo"`
}
type event04000000Envelope struct {
	Element Event04000000Details `xml:"Element"`
}

func HandleEvent04000000(logger *slog.Logger, ref string, filePath string, outputPath string, event Event, content string) error {
	var parsed event04000000Envelope
	if err := xml.Unmarshal([]byte(content), &parsed); err != nil {
		return fmt.Errorf("failed to parse event 04/000000 xml: %w", err)
	}

	logger.Info(
		"type1 event handled",
		"ref", ref,
		"event_code", event.EventCode,
		"event_detail_code", event.EventDetailCode,
		"file_name", filepath.Base(filePath),
		"pcb_serial", parsed.Element.PcbSerial,
	)

	return nil
}

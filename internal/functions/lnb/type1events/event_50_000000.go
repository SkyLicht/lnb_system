package type1events

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Event50000000Details struct {
	Date               string `xml:"Date"`
	MDLN               string `xml:"MDLN"`
	EventSerial        string `xml:"EventSerial"`
	EventCode          string `xml:"EventCode"`
	EventDetailCode    string `xml:"EventDetailCode"`
	PcbSerial          string `xml:"PcbSerial"`
	Stage              string `xml:"Stage"`
	Lane               string `xml:"Lane"`
	RedLightStatus     string `xml:"RedLightStatus"`
	YellowLightStatus  string `xml:"YellowLightStatus"`
	GreenLightStatus   string `xml:"GreenLightStatus"`
	ReserveLightStatus string `xml:"ReserveLightStatus"`
	BuzzerStatus       string `xml:"BuzzerStatus"`
	MCNo               string `xml:"MCNo"`
}

type event50000000Envelope struct {
	Element Event50000000Details `xml:"Element"`
}

type lightStatus struct {
	Red    int `json:"red"`
	Green  int `json:"green"`
	Yellow int `json:"yellow"`
	Buzzer int `json:"buzzer"`
}

type machineTowerState struct {
	Tower towerLightStatus            `json:"tower"`
	Stage map[string]stageLightStatus `json:"stage"`
}

type npmTowerState struct {
	Machine1 machineTowerState `json:"machine_1"`
	Machine2 machineTowerState `json:"machine_2"`
	Machine3 machineTowerState `json:"machine_3"`
	Machine4 machineTowerState `json:"machine_4"`
}

type towerLightStatus struct {
	Red       int    `json:"red"`
	Green     int    `json:"green"`
	Yellow    int    `json:"yellow"`
	Buzzer    int    `json:"buzzer"`
	Timestamp string `json:"timestamp"`
}

type stageLightStatus struct {
	Red       int    `json:"red"`
	Green     int    `json:"green"`
	Yellow    int    `json:"yellow"`
	Buzzer    int    `json:"buzzer"`
	Timestamp string `json:"timestamp"`
}

func HandleEvent50000000(logger *slog.Logger, ref string, filePath string, outputPath string, event Event, content string) error {
	var parsed event50000000Envelope
	if err := xml.Unmarshal([]byte(content), &parsed); err != nil {
		return fmt.Errorf("failed to parse event 50/000000 xml: %w", err)
	}

	jsonFile := strings.TrimSpace(ref)
	if jsonFile == "" {
		jsonFile = "npm_tower"
	}

	jsonPath, err := resolveTowerJSONPath(outputPath, jsonFile+".json", filePath)
	if err != nil {
		return err
	}
	state, err := readNPMTowerState(jsonPath)
	if err != nil {
		return err
	}

	stageKey := buildStageKey(parsed.Element.Stage, parsed.Element.Lane)
	machineKey := buildMachineKey(parsed.Element.MCNo)
	values := toLightStatus(parsed.Element)
	lastUpdatedAt := time.Now().UTC().Format(time.RFC3339)

	if err := applyTowerState(&state, machineKey, stageKey, values, lastUpdatedAt); err != nil {
		return err
	}

	if err := writeNPMTowerState(jsonPath, state); err != nil {
		return err
	}

	logger.Info(
		"type1 event handled",
		"ref", ref,
		"event_code", event.EventCode,
		"event_detail_code", event.EventDetailCode,
		"file_name", filepath.Base(filePath),
		"machine", machineKey,
		"stage", stageKey,
		"last_updated_at", lastUpdatedAt,
		"json_path", jsonPath,
	)

	return nil
}

func resolveTowerJSONPath(path string, file string, sourceFilePath string) (string, error) {
	dirPath := strings.TrimSpace(path)
	if dirPath == "" {
		dirPath = filepath.Dir(sourceFilePath)
	}
	if strings.TrimSpace(dirPath) == "" || dirPath == "." {
		dirPath = "assets"
	}

	jsonPath := filepath.Join(dirPath, file)
	if _, err := os.Stat(jsonPath); err == nil {
		return jsonPath, nil
	}

	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", dirPath, err)
	}

	initialState := npmTowerState{
		Machine1: newDefaultMachineTowerState(),
		Machine2: newDefaultMachineTowerState(),
		Machine3: newDefaultMachineTowerState(),
		Machine4: newDefaultMachineTowerState(),
	}

	if err := writeNPMTowerState(jsonPath, initialState); err != nil {
		return "", err
	}

	return jsonPath, nil
}

func readNPMTowerState(path string) (npmTowerState, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return npmTowerState{}, fmt.Errorf("failed to read %s: %w", path, err)
	}

	var state npmTowerState
	if err := json.Unmarshal(content, &state); err != nil {
		return npmTowerState{}, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	return state, nil
}

func writeNPMTowerState(path string, state npmTowerState) error {
	encoded, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode %s: %w", path, err)
	}

	if err := os.WriteFile(path, encoded, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}

	return nil
}

func newDefaultMachineTowerState() machineTowerState {
	return machineTowerState{
		Tower: towerLightStatus{
			Red:       0,
			Green:     0,
			Yellow:    0,
			Buzzer:    0,
			Timestamp: "",
		},
		Stage: map[string]stageLightStatus{
			"00_00": newDefaultStageLightStatus(),
			"00_01": newDefaultStageLightStatus(),
			"01_00": newDefaultStageLightStatus(),
			"01_01": newDefaultStageLightStatus(),
			"02_00": newDefaultStageLightStatus(),
			"02_01": newDefaultStageLightStatus(),
		},
	}
}

func newDefaultStageLightStatus() stageLightStatus {
	return stageLightStatus{
		Red:       0,
		Green:     0,
		Yellow:    0,
		Buzzer:    0,
		Timestamp: "",
	}
}

func applyTowerState(state *npmTowerState, machineKey string, stageKey string, values lightStatus, timestamp string) error {
	var machine *machineTowerState
	switch machineKey {
	case "machine_1":
		machine = &state.Machine1
	case "machine_2":
		machine = &state.Machine2
	case "machine_3":
		machine = &state.Machine3
	case "machine_4":
		machine = &state.Machine4
	default:
		return fmt.Errorf("unsupported MCNo mapping for %q", machineKey)
	}

	machine.Tower = towerLightStatus{
		Red:       values.Red,
		Green:     values.Green,
		Yellow:    values.Yellow,
		Buzzer:    values.Buzzer,
		Timestamp: timestamp,
	}
	if machine.Stage == nil {
		machine.Stage = map[string]stageLightStatus{}
	}
	machine.Stage[stageKey] = stageLightStatus{
		Red:       values.Red,
		Green:     values.Green,
		Yellow:    values.Yellow,
		Buzzer:    values.Buzzer,
		Timestamp: timestamp,
	}
	return nil
}

func buildMachineKey(mcNo string) string {
	value := parseIntOrDefault(mcNo, 0)
	return fmt.Sprintf("machine_%d", value)
}

func buildStageKey(stage string, lane string) string {
	return fmt.Sprintf("%s_%s", normalizeTwoDigits(stage), normalizeTwoDigits(lane))
}

func toLightStatus(details Event50000000Details) lightStatus {
	return lightStatus{
		Red:    parseIntOrDefault(details.RedLightStatus, 0),
		Green:  parseIntOrDefault(details.GreenLightStatus, 0),
		Yellow: parseIntOrDefault(details.YellowLightStatus, 0),
		Buzzer: parseIntOrDefault(details.BuzzerStatus, 0),
	}
}

func normalizeTwoDigits(value string) string {
	parsed := parseIntOrDefault(value, -1)
	if parsed < 0 {
		trimmed := strings.TrimSpace(value)
		if len(trimmed) == 1 {
			return "0" + trimmed
		}
		return trimmed
	}
	return fmt.Sprintf("%02d", parsed)
}

func parseIntOrDefault(value string, fallback int) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return fallback
	}
	return parsed
}

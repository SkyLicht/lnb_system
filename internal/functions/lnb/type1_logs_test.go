package lnb

import (
	"testing"

	type1events "lnb/internal/functions/lnb/type1events"
)

func TestParseType1EventExtractsCodes(t *testing.T) {
	content := `<?xml version="1.0" encoding="UTF-8"?><MachineEvent><Element><EventCode>50</EventCode><EventDetailCode>000000</EventDetailCode></Element></MachineEvent>`

	event, err := parseType1Event(content)
	if err != nil {
		t.Fatalf("expected xml to parse, got error: %v", err)
	}

	if event.EventCode != "50" {
		t.Fatalf("expected EventCode 50, got %q", event.EventCode)
	}
	if event.EventDetailCode != "000000" {
		t.Fatalf("expected EventDetailCode 000000, got %q", event.EventDetailCode)
	}
}

func TestParseType1EventReturnsErrorWhenCodesMissing(t *testing.T) {
	content := `<?xml version="1.0" encoding="UTF-8"?><MachineEvent><Element><MCNo>3</MCNo></Element></MachineEvent>`

	_, err := parseType1Event(content)
	if err == nil {
		t.Fatal("expected error when EventCode and EventDetailCode are missing")
	}
}

func TestBuildType1EventKey(t *testing.T) {
	key := type1events.BuildKey("04", "000000")

	if key != "04:000000" {
		t.Fatalf("expected event key 04:000000, got %q", key)
	}
}

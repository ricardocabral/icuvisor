package tools

import (
	"context"
	"encoding/json"
	"math"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/safety"
)

func TestUpdateSportSettingsOmittedZonesDoesNotWriteZones(t *testing.T) {
	t.Parallel()

	client := newFakeSportSettingsClient(intervals.SportSettings{ID: 7, Types: []string{"Ride"}, FTP: 250, PaceUnits: "MINS_KM"})
	ftp := 275
	client.setting = intervals.SportSettings{ID: 7, Type: "Ride", FTP: ftp}
	tool := newUpdateSportSettingsTool(client, client, "test", "UTC", false, safety.NewCapability(safety.ModeSafe))

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"sport":"Ride","ftp":275}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 1 {
		t.Fatalf("write calls = %d, want 1", len(client.calls))
	}
	call := client.calls[0]
	if call.ZonesProvided || len(call.Zones) != 0 {
		t.Fatalf("zones write = provided %v zones %#v, want omitted", call.ZonesProvided, call.Zones)
	}
	if call.FTP == nil || *call.FTP != ftp || call.SportSettingID != 7 || !call.RecalcHRZones {
		t.Fatalf("write call = %+v, want FTP-only sport setting update", call)
	}
	meta := resultMap(t, result)["_meta"].(map[string]any)
	if meta["zones_provided"] != false {
		t.Fatalf("meta = %#v, want zones_provided=false", meta)
	}
}

func TestUpdateSportSettingsSafeModeRejectsZonesBeforeWrite(t *testing.T) {
	t.Parallel()

	client := newFakeSportSettingsClient(intervals.SportSettings{ID: 7, Types: []string{"Run"}})
	tool := newUpdateSportSettingsTool(client, client, "test", "UTC", false, safety.NewCapability(safety.ModeSafe))

	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"sport":"Run","zones":[{"kind":"pace","boundaries":[77.5,100],"names":["Easy","Threshold"]}]}`)})
	if err == nil || !strings.Contains(err.Error(), "zones overwrite prior") {
		t.Fatalf("Handler() error = %v, want zone gate user error", err)
	}
	if message, ok := PublicErrorMessage(err); !ok || !strings.Contains(message, "ICUVISOR_DELETE_MODE=full") {
		t.Fatalf("PublicErrorMessage() = %q, %v; want typed public gate message", message, ok)
	}
	if client.fakeProfileClient.calls != 0 || len(client.calls) != 0 {
		t.Fatalf("profile calls = %d, write calls = %#v; want neither in safe mode", client.fakeProfileClient.calls, client.calls)
	}
}

func TestValidateSportSettingsPaceZonePercentages(t *testing.T) {
	tests := []struct {
		name       string
		boundaries []float64
		wantErr    bool
	}{
		{name: "valid percentages", boundaries: []float64{77.5, 100}},
		{name: "zero", boundaries: []float64{0, 100}, wantErr: true},
		{name: "non finite", boundaries: []float64{77.5, math.Inf(1)}, wantErr: true},
		{name: "above maximum", boundaries: []float64{77.5, 200.1}, wantErr: true},
		{name: "duplicate", boundaries: []float64{77.5, 77.5}, wantErr: true},
		{name: "descending", boundaries: []float64{100, 77.5}, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateSportSettingsZones([]updateSportSettingsZoneRequest{{Kind: "pace", Boundaries: tc.boundaries}})
			if (err != nil) != tc.wantErr {
				t.Fatalf("validateSportSettingsZones() error = %v, wantErr %t", err, tc.wantErr)
			}
		})
	}
	if err := validateSportSettingsZones([]updateSportSettingsZoneRequest{{Kind: "power", Boundaries: []float64{0, 100}}, {Kind: "hr", Boundaries: []float64{0, 120}}}); err != nil {
		t.Fatalf("validateSportSettingsZones() changed power/hr zero handling: %v", err)
	}
}

func TestUpdateSportSettingsFullModeAppliesZonesAndResponseMeta(t *testing.T) {
	client := newFakeSportSettingsClient(intervals.SportSettings{ID: 7, Types: []string{"Ride"}, FTP: 250})
	client.setting = intervals.SportSettings{ID: 7, Type: "Ride", FTP: 280, PowerZones: []int{100, 200}, PowerZoneNames: []string{"Z1", "Z2"}}
	tool := newUpdateSportSettingsTool(client, client, "v1.2.3", "UTC", false, safety.NewCapability(safety.ModeFull), responseShaping{deleteMode: safety.ModeFull, toolset: safety.ToolsetCore})

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"sport":"Ride","ftp":280,"zones":[{"kind":"power","boundaries":[100,200],"names":["Z1","Z2"]}]}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 1 || !client.calls[0].ZonesProvided || len(client.calls[0].Zones) != 1 {
		t.Fatalf("write calls = %#v, want one gated zone overwrite", client.calls)
	}
	payload := resultMap(t, result)
	settings := payload["sport_settings"].(map[string]any)
	if settings["zone_definitions_overwritten"] != true || len(settings["zones"].([]any)) != 1 {
		t.Fatalf("settings = %#v, want zone echo", settings)
	}
	meta := payload["_meta"].(map[string]any)
	if meta["delete_mode"] != "full" || meta["server_version"] != "v1.2.3" || meta["hr_zone_recalculation_requested"] != true {
		t.Fatalf("meta = %#v, want delete_mode/full server version and recompute", meta)
	}
	units := meta["units"].(map[string]any)
	if units["system"] != "metric" || units["pace"] == "" {
		t.Fatalf("units = %#v, want unit metadata", units)
	}
}

func TestUpdateSportSettingsFullModeRoundTripsRunPaceZones(t *testing.T) {
	client := newFakeSportSettingsClient(intervals.SportSettings{ID: 8, Type: "Run", Types: []string{"Run"}, PaceUnits: "MINS_MILE"})
	client.setting = intervals.SportSettings{ID: 8, Type: "Run", PaceUnits: "MINS_MILE", PaceZones: []float64{77.5, 100}, PaceZoneNames: []string{"Easy", "Threshold"}}
	tool := newUpdateSportSettingsTool(client, client, "test", "UTC", false, safety.NewCapability(safety.ModeFull), responseShaping{deleteMode: safety.ModeFull, toolset: safety.ToolsetCore})

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"sport":"Run","zones":[{"kind":"pace","boundaries":[77.5,100],"names":["Easy","Threshold"]}]}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 1 || !client.calls[0].ZonesProvided || len(client.calls[0].Zones) != 1 {
		t.Fatalf("write calls = %#v, want one gated pace-zone overwrite", client.calls)
	}
	zone := client.calls[0].Zones[0]
	if zone.Kind != "pace" || len(zone.Boundaries) != 2 || zone.Boundaries[0] != 77.5 || zone.Boundaries[1] != 100 || len(zone.Names) != 2 || zone.Names[0] != "Easy" || zone.Names[1] != "Threshold" {
		t.Fatalf("pace zone call = %#v, want percentage/name round trip", zone)
	}
	payload := resultMap(t, result)
	settings := payload["sport_settings"].(map[string]any)
	zones := settings["zones"].([]any)
	echo := zones[0].(map[string]any)
	if echo["kind"] != "pace" || echo["boundaries"].([]any)[0] != 77.5 || echo["boundaries"].([]any)[1] != 100.0 || echo["names"].([]any)[1] != "Threshold" || settings["zone_definitions_overwritten"] != true {
		t.Fatalf("settings = %#v, want pace-zone percentage/name echo", settings)
	}
	meta := payload["_meta"].(map[string]any)
	if meta["delete_mode"] != "full" || meta["zones_provided"] != true {
		t.Fatalf("meta = %#v, want full-mode zone metadata", meta)
	}
}

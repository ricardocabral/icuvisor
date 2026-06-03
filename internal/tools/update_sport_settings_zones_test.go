package tools

import (
	"context"
	"encoding/json"
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

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"sport":"Ride","effective_date":"2026-05-01","ftp":275}`)})
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
	if call.FTP == nil || *call.FTP != ftp || call.SportSettingID != 7 || call.EffectiveDate != "2026-05-01" {
		t.Fatalf("write call = %+v, want FTP-only sport setting update", call)
	}
	meta := resultMap(t, result)["_meta"].(map[string]any)
	if meta["zones_provided"] != false {
		t.Fatalf("meta = %#v, want zones_provided=false", meta)
	}
}

func TestUpdateSportSettingsSafeModeRejectsZonesBeforeWrite(t *testing.T) {
	t.Parallel()

	client := newFakeSportSettingsClient(intervals.SportSettings{ID: 7, Types: []string{"Ride"}})
	tool := newUpdateSportSettingsTool(client, client, "test", "UTC", false, safety.NewCapability(safety.ModeSafe))

	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"sport":"Ride","effective_date":"2026-05-01","zones":[{"kind":"power","boundaries":[100,200],"names":["Z1","Z2"]}]}`)})
	if err == nil || !strings.Contains(err.Error(), "zones overwrite prior") {
		t.Fatalf("Handler() error = %v, want zone gate user error", err)
	}
	if message, ok := PublicErrorMessage(err); !ok || !strings.Contains(message, "ICUVISOR_DELETE_MODE=full") {
		t.Fatalf("PublicErrorMessage() = %q, %v; want typed public gate message", message, ok)
	}
	if len(client.calls) != 0 {
		t.Fatalf("write calls = %#v, want none in safe mode", client.calls)
	}
}

func TestUpdateSportSettingsFullModeAppliesZonesAndResponseMeta(t *testing.T) {
	client := newFakeSportSettingsClient(intervals.SportSettings{ID: 7, Types: []string{"Ride"}, FTP: 250})
	client.setting = intervals.SportSettings{ID: 7, Type: "Ride", FTP: 280, PowerZones: []int{100, 200}, PowerZoneNames: []string{"Z1", "Z2"}}
	tool := newUpdateSportSettingsTool(client, client, "v1.2.3", "UTC", false, safety.NewCapability(safety.ModeFull), responseShaping{deleteMode: safety.ModeFull, toolset: safety.ToolsetCore})

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"sport":"Ride","effective_date":"2026-05-01","ftp":280,"zones":[{"kind":"power","boundaries":[100,200],"names":["Z1","Z2"]}]}`)})
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
	if meta["delete_mode"] != "full" || meta["server_version"] != "v1.2.3" || meta["recompute_pending"] != true {
		t.Fatalf("meta = %#v, want delete_mode/full server version and recompute", meta)
	}
	units := meta["units"].(map[string]any)
	if units["system"] != "metric" || units["pace"] == "" {
		t.Fatalf("units = %#v, want unit metadata", units)
	}
}

func TestUpdateSportSettingsFullModeRoundTripsRunPaceZones(t *testing.T) {
	client := newFakeSportSettingsClient(intervals.SportSettings{ID: 8, Type: "Run", Types: []string{"Run"}, PaceUnits: "MINS_MILE"})
	client.setting = intervals.SportSettings{ID: 8, Type: "Run", PaceUnits: "MINS_MILE", PaceZones: []float64{480, 420}, PaceZoneNames: []string{"Aerobic", "Threshold"}}
	tool := newUpdateSportSettingsTool(client, client, "test", "UTC", false, safety.NewCapability(safety.ModeFull), responseShaping{deleteMode: safety.ModeFull, toolset: safety.ToolsetCore})

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"sport":"Run","effective_date":"2026-05-01","zones":[{"kind":"pace","boundaries":[480,420],"names":["Aerobic","Threshold"]}]}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 1 || !client.calls[0].ZonesProvided || len(client.calls[0].Zones) != 1 {
		t.Fatalf("write calls = %#v, want one gated pace-zone overwrite", client.calls)
	}
	zone := client.calls[0].Zones[0]
	if zone.Kind != "pace" || len(zone.Boundaries) != 2 || zone.Boundaries[0] != 480 || zone.Boundaries[1] != 420 || len(zone.Names) != 2 || zone.Names[0] != "Aerobic" || zone.Names[1] != "Threshold" {
		t.Fatalf("pace zone call = %#v, want boundary/name round trip", zone)
	}
	payload := resultMap(t, result)
	settings := payload["sport_settings"].(map[string]any)
	zones := settings["zones"].([]any)
	echo := zones[0].(map[string]any)
	if echo["kind"] != "pace" || echo["boundaries"].([]any)[0] != 480.0 || echo["names"].([]any)[1] != "Threshold" || settings["zone_definitions_overwritten"] != true {
		t.Fatalf("settings = %#v, want pace-zone boundary/name echo", settings)
	}
	meta := payload["_meta"].(map[string]any)
	if meta["delete_mode"] != "full" || meta["zones_provided"] != true {
		t.Fatalf("meta = %#v, want full-mode zone metadata", meta)
	}
}

package tools

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

type fakeDeleteToolsClient struct {
	fakeProfileClient
	event                   intervals.Event
	activity                intervals.Activity
	customItem              intervals.CustomItem
	gear                    intervals.Gear
	events                  []intervals.Event
	listParams              intervals.ListEventsParams
	deletedEventIDs         []string
	deletedActivityIDs      []string
	deletedCustomItemIDs    []string
	deletedSportSettingsIDs []string
	deletedGearIDs          []string
	deleteErr               error
	getErr                  error
}

func (f *fakeDeleteToolsClient) GetEvent(ctx context.Context, eventID string) (intervals.Event, error) {
	if f.getErr != nil {
		return intervals.Event{}, f.getErr
	}
	return f.event, nil
}

func (f *fakeDeleteToolsClient) DeleteEvent(ctx context.Context, eventID string) error {
	f.deletedEventIDs = append(f.deletedEventIDs, eventID)
	return f.deleteErr
}

func (f *fakeDeleteToolsClient) GetActivity(ctx context.Context, activityID string) (intervals.Activity, error) {
	if f.getErr != nil {
		return intervals.Activity{}, f.getErr
	}
	return f.activity, nil
}

func (f *fakeDeleteToolsClient) DeleteActivity(ctx context.Context, activityID string) error {
	f.deletedActivityIDs = append(f.deletedActivityIDs, activityID)
	return f.deleteErr
}

func (f *fakeDeleteToolsClient) GetCustomItem(ctx context.Context, itemID string) (intervals.CustomItem, error) {
	if f.getErr != nil {
		return intervals.CustomItem{}, f.getErr
	}
	return f.customItem, nil
}

func (f *fakeDeleteToolsClient) DeleteCustomItem(ctx context.Context, itemID string) error {
	f.deletedCustomItemIDs = append(f.deletedCustomItemIDs, itemID)
	return f.deleteErr
}

func (f *fakeDeleteToolsClient) DeleteSportSettings(ctx context.Context, sportSettingsID string) error {
	f.deletedSportSettingsIDs = append(f.deletedSportSettingsIDs, sportSettingsID)
	return f.deleteErr
}

func (f *fakeDeleteToolsClient) GetGear(ctx context.Context, gearID string) (intervals.Gear, error) {
	if f.getErr != nil {
		return intervals.Gear{}, f.getErr
	}
	return f.gear, nil
}

func (f *fakeDeleteToolsClient) DeleteGear(ctx context.Context, gearID string) error {
	f.deletedGearIDs = append(f.deletedGearIDs, gearID)
	return f.deleteErr
}

func (f *fakeDeleteToolsClient) ListEvents(ctx context.Context, params intervals.ListEventsParams) ([]intervals.Event, error) {
	f.listParams = params
	if f.getErr != nil {
		return nil, f.getErr
	}
	return append([]intervals.Event(nil), f.events...), nil
}

func TestDeletePerIDToolsSuccessAndEcho(t *testing.T) {
	t.Parallel()

	profile := intervals.AthleteWithSportSettings{ID: "12345", PreferredUnits: "metric", Timezone: "UTC", SportSettings: []intervals.SportSettings{{ID: 7, Type: "Ride", FTP: 280, LTHR: 172}}}
	tests := []struct {
		name       string
		tool       func(*fakeDeleteToolsClient) Tool
		arguments  string
		wantID     string
		wantMetaID string
		calls      func(*fakeDeleteToolsClient) []string
	}{
		{name: deleteEventName, tool: func(c *fakeDeleteToolsClient) Tool { return newDeleteEventTool(c, c, "test", "UTC", false) }, arguments: `{"event_id":" e-1 "}`, wantID: "e-1", wantMetaID: "event_id", calls: func(c *fakeDeleteToolsClient) []string { return c.deletedEventIDs }},
		{name: deleteActivityName, tool: func(c *fakeDeleteToolsClient) Tool { return newDeleteActivityTool(c, c, "test", "UTC", false) }, arguments: `{"activity_id":" a-1 "}`, wantID: "a-1", wantMetaID: "activity_id", calls: func(c *fakeDeleteToolsClient) []string { return c.deletedActivityIDs }},
		{name: deleteCustomItemName, tool: func(c *fakeDeleteToolsClient) Tool { return newDeleteCustomItemTool(c, c, "test", "UTC", false) }, arguments: `{"item_id":" ci-1 "}`, wantID: "ci-1", wantMetaID: "id", calls: func(c *fakeDeleteToolsClient) []string { return c.deletedCustomItemIDs }},
		{name: deleteSportSettingsName, tool: func(c *fakeDeleteToolsClient) Tool { return newDeleteSportSettingsTool(c, c, "test", "UTC", false) }, arguments: `{"sport_settings_id":" 7 "}`, wantID: "7", wantMetaID: "sport_settings_id", calls: func(c *fakeDeleteToolsClient) []string { return c.deletedSportSettingsIDs }},
		{name: deleteGearName, tool: func(c *fakeDeleteToolsClient) Tool { return newDeleteGearTool(c, c, "test", "UTC", false) }, arguments: `{"gear_id":" g-1 "}`, wantID: "g-1", wantMetaID: "gear_id", calls: func(c *fakeDeleteToolsClient) []string { return c.deletedGearIDs }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := &fakeDeleteToolsClient{fakeProfileClient: fakeProfileClient{profile: profile}, event: mustEvent(t, `{"id":"e-1","category":"WORKOUT","name":"Endurance","start_date_local":"2026-05-13"}`), activity: mustActivity(t, `{"id":"a-1","name":"Run","type":"Run","start_date_local":"2026-05-13T07:00:00","moving_time":3600,"icu_distance":10000}`), customItem: mustCustomItem(t, `{"id":"ci-1","type":"FITNESS_CHART","name":"CTL"}`), gear: mustGear(t, `{"id":"g-1","name":"Race Bike","type":"Bike","brand":"Acme"}`)}
			tool := tc.tool(client)
			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(tc.arguments)})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			if got := tc.calls(client); len(got) != 1 || got[0] != tc.wantID {
				t.Fatalf("delete calls = %#v, want [%s]", got, tc.wantID)
			}
			out := resultMap(t, result)
			if out["deleted_id"] != tc.wantID || out["status"] != "deleted" {
				t.Fatalf("response = %#v, want deleted_id/status", out)
			}
			meta := out["_meta"].(map[string]any)
			deleted := meta["deleted"].(map[string]any)
			if deleted[tc.wantMetaID] == nil || strings.TrimSpace(deleted[tc.wantMetaID].(string)) == "" {
				t.Fatalf("_meta.deleted = %#v, missing %s", deleted, tc.wantMetaID)
			}
		})
	}
}

func TestDeletePerIDToolsRejectConfirmArgument(t *testing.T) {
	t.Parallel()

	client := &fakeDeleteToolsClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{Timezone: "UTC"}}}
	tests := []struct {
		name      string
		tool      Tool
		arguments string
	}{
		{name: deleteEventName, tool: newDeleteEventTool(client, client, "test", "UTC", false), arguments: `{"event_id":"e-1","confirm":true}`},
		{name: deleteActivityName, tool: newDeleteActivityTool(client, client, "test", "UTC", false), arguments: `{"activity_id":"a-1","confirm":true}`},
		{name: deleteCustomItemName, tool: newDeleteCustomItemTool(client, client, "test", "UTC", false), arguments: `{"item_id":"ci-1","confirm":true}`},
		{name: deleteSportSettingsName, tool: newDeleteSportSettingsTool(client, client, "test", "UTC", false), arguments: `{"sport_settings_id":"7","confirm":true}`},
		{name: deleteGearName, tool: newDeleteGearTool(client, client, "test", "UTC", false), arguments: `{"gear_id":"g-1","confirm":true}`},
		{name: deleteEventsByDateRangeName, tool: newDeleteEventsByDateRangeTool(client, client, "test", "UTC", false), arguments: `{"start_date":"2026-05-13","end_date":"2026-05-13","confirm":true}`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if _, err := tc.tool.Handler(context.Background(), Request{Name: tc.tool.Name, Arguments: json.RawMessage(tc.arguments)}); err == nil {
				t.Fatal("Handler() error = nil, want strict argument rejection")
			}
		})
	}
}

func TestDeleteEventsByDateRangeValidationAndTimezone(t *testing.T) {
	t.Parallel()

	client := &fakeDeleteToolsClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "12345", PreferredUnits: "metric", Timezone: "America/Sao_Paulo"}}, events: []intervals.Event{mustEvent(t, `{"id":"11","category":"WORKOUT","name":"Morning","start_date_local":"2026-05-13T00:00:00"}`), mustEvent(t, `{"id":"12","category":"NOTE","name":"Evening","start_date_local":"2026-05-13T23:59:59"}`)}}
	tool := newDeleteEventsByDateRangeTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"start_date":"2026-05-13","end_date":"2026-05-13","category":"WORKOUT"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if client.listParams.Oldest != "2026-05-13" || client.listParams.Newest != "2026-05-13" || client.listParams.Category != "WORKOUT" {
		t.Fatalf("list params = %#v, want exact athlete-local boundary dates and category", client.listParams)
	}
	if got := client.deletedEventIDs; !slices.Equal(got, []string{"11", "12"}) {
		t.Fatalf("deleted event IDs = %#v, want sorted listed IDs", got)
	}
	out := resultMap(t, result)
	meta := out["_meta"].(map[string]any)
	if meta["timezone"] != "America/Sao_Paulo" || int(meta["deleted_count"].(float64)) != 2 {
		t.Fatalf("meta = %#v, want timezone and deleted_count", meta)
	}
	ids := out["deleted_ids"].([]any)
	if len(ids) != 2 || ids[0] != "11" || ids[1] != "12" {
		t.Fatalf("deleted_ids = %#v, want ID list", ids)
	}

	for _, raw := range []string{
		`{"start_date":"2026-05-13"}`,
		`{"end_date":"2026-05-13"}`,
		`{"start_date":"2026-05-13","end_date":"2026-06-13"}`,
		`{"start_date":"2026-05-14","end_date":"2026-05-13"}`,
	} {
		if _, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(raw)}); err == nil {
			t.Fatalf("Handler(%s) error = nil, want validation error", raw)
		}
	}
}

func TestDeleteToolsRegistrationMetadata(t *testing.T) {
	t.Parallel()

	client := &fakeDeleteToolsClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "12345", PreferredUnits: "metric", Timezone: "UTC"}}}
	registrar := &collectingRegistrar{}
	if err := NewRegistry(client, "test", "UTC").Register(context.Background(), registrar); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	for _, name := range []string{deleteEventName, deleteActivityName, deleteCustomItemName, deleteSportSettingsName, deleteGearName, deleteEventsByDateRangeName} {
		t.Run(name, func(t *testing.T) {
			tool := findTool(t, registrar.tools, name)
			if tool.Requirement != RequirementDelete || !tool.RequiresDelete() {
				t.Fatalf("requirement = %q delete=%v, want delete", tool.Requirement, tool.RequiresDelete())
			}
			props := tool.InputSchema.(map[string]any)["properties"].(map[string]any)
			if _, ok := props["confirm"]; ok || strings.Contains(strings.ToLower(tool.Description), "confirm:true") {
				t.Fatalf("schema/description includes model-controlled confirm: %#v %q", props, tool.Description)
			}
		})
	}
}

func TestDeleteToolErrorsPreserveNotFoundSentinel(t *testing.T) {
	t.Parallel()

	client := &fakeDeleteToolsClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{Timezone: "UTC"}}, getErr: intervals.ErrNotFound}
	tool := newDeleteEventTool(client, client, "test", "UTC", false)
	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"event_id":"missing"}`)})
	if err == nil || !errors.Is(err, intervals.ErrNotFound) {
		t.Fatalf("Handler() error = %v, want wrapped intervals.ErrNotFound", err)
	}
}

func mustEvent(t *testing.T, raw string) intervals.Event {
	t.Helper()
	var event intervals.Event
	if err := json.Unmarshal([]byte(raw), &event); err != nil {
		t.Fatalf("unmarshal event: %v", err)
	}
	return event
}

func mustActivity(t *testing.T, raw string) intervals.Activity {
	t.Helper()
	var activity intervals.Activity
	if err := json.Unmarshal([]byte(raw), &activity); err != nil {
		t.Fatalf("unmarshal activity: %v", err)
	}
	return activity
}

func mustCustomItem(t *testing.T, raw string) intervals.CustomItem {
	t.Helper()
	var item intervals.CustomItem
	if err := json.Unmarshal([]byte(raw), &item); err != nil {
		t.Fatalf("unmarshal custom item: %v", err)
	}
	return item
}

func mustGear(t *testing.T, raw string) intervals.Gear {
	t.Helper()
	var gear intervals.Gear
	if err := json.Unmarshal([]byte(raw), &gear); err != nil {
		t.Fatalf("unmarshal gear: %v", err)
	}
	return gear
}

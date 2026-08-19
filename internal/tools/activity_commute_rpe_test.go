package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

func TestActivityCommuteRPEListRowsAreSourceLabelledAndNullSafe(t *testing.T) {
	t.Parallel()

	activities := loadSubjectiveActivityListFixture(t)
	client := &fakeActivitiesProfileClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		activities:        activities,
	}
	tool := newGetActivitiesToolWithGear(client, client, nil, nil, nil, nil, "test", "UTC", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-08-17","newest":"2026-08-18"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}

	payload := resultMap(t, result)
	rows := payload["activities"].([]any)
	if len(rows) != 2 {
		t.Fatalf("activities = %#v, want two fixture rows", rows)
	}
	commute := rows[0].(map[string]any)
	if commute["commute"] != true || commute["feel"] != float64(4) || commute["rpe"] != float64(5) {
		t.Fatalf("commute row = %#v, want source commute/feel/rpe", commute)
	}
	if commute["timezone"] != "Europe/Lisbon" {
		t.Fatalf("commute timezone = %#v, want activity timezone", commute["timezone"])
	}
	assertActivitySubjectiveScales(t, commute)

	training := rows[1].(map[string]any)
	if training["commute"] != false {
		t.Fatalf("training commute = %#v, want explicit false preserved", training["commute"])
	}
	for _, key := range []string{"feel", "rpe"} {
		if _, ok := training[key]; ok {
			t.Fatalf("training row emitted null %s: %#v", key, training)
		}
	}
	if _, ok := training["custom_fields"]; ok {
		t.Fatalf("training row emitted unrequested custom fields: %#v", training)
	}

	meta := payload["_meta"].(map[string]any)
	semantics := meta["field_semantics"].(map[string]any)
	for _, key := range []string{"commute", "feel", "rpe"} {
		if strings.TrimSpace(semantics[key].(string)) == "" {
			t.Fatalf("field_semantics missing %s: %#v", key, semantics)
		}
	}
	if len(client.listCalls) == 0 {
		t.Fatal("ListActivities was not called")
	}
	fields := client.listCalls[0].Fields
	for _, want := range []string{"commute", "feel", "icu_rpe"} {
		found := false
		for _, field := range fields {
			if field == want {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("list fields = %#v, want %q requested", fields, want)
		}
	}
}

func TestActivityCommuteRPEDetailPreservesRawFullPayloadAndTimezone(t *testing.T) {
	t.Parallel()

	activity := loadSubjectiveActivityDetailFixture(t)
	client := &fakeActivityReadClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "imperial", Timezone: "America/Sao_Paulo"}},
		activity:          activity,
	}
	tool := newGetActivityDetailsToolWithGear(client, client, nil, nil, nil, nil, "test", "UTC", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"activity-detail-subjective","include_full":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}

	payload := resultMap(t, result)
	row := payload["activity"].(map[string]any)
	if row["commute"] != false || row["feel"] != float64(2) || row["rpe"] != float64(8) {
		t.Fatalf("detail row = %#v, want source commute/feel/rpe", row)
	}
	if row["timezone"] != "Europe/Lisbon" {
		t.Fatalf("detail timezone = %#v, want activity timezone", row["timezone"])
	}
	assertActivitySubjectiveScales(t, payload)
	full := row["full"].(map[string]any)
	if full["commute"] != false || full["feel"] != float64(2) || full["icu_rpe"] != float64(8) || full["custom_note"] != "synthetic fixture" {
		t.Fatalf("full payload = %#v, want raw upstream keys and custom field", full)
	}
}

func TestActivityCommuteRPEMalformedValuesAreNotInferred(t *testing.T) {
	t.Parallel()

	activity := decodeActivityFixture(t, `{"id":"malformed-subjective","name":"Bike commute","type":"Ride","tags":["commute"],"commute":"yes","feel":"4","icu_rpe":7.5}`)
	client := &fakeActivityReadClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		activity:          activity,
	}
	tool := newGetActivityDetailsToolWithGear(client, client, nil, nil, nil, nil, "test", "UTC", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"malformed-subjective","include_full":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	row := resultMap(t, result)["activity"].(map[string]any)
	for _, key := range []string{"commute", "feel", "rpe"} {
		if _, ok := row[key]; ok {
			t.Fatalf("malformed activity inferred %s from invalid upstream value/name/tag: %#v", key, row)
		}
	}
	full := row["full"].(map[string]any)
	if full["commute"] != "yes" || full["feel"] != "4" || full["icu_rpe"] != 7.5 {
		t.Fatalf("full malformed values = %#v, want raw values preserved", full)
	}
}

func assertActivitySubjectiveScales(t *testing.T, row map[string]any) {
	t.Helper()
	meta, ok := row["_meta"].(map[string]any)
	if !ok {
		t.Fatalf("row metadata = %#v, want scale metadata", row["_meta"])
	}
	scales, ok := meta["scales"].(map[string]any)
	if !ok {
		t.Fatalf("row scales = %#v, want feel and rpe labels", meta["scales"])
	}
	if scales["feel"] != "1-5 (athlete-reported feel)" {
		t.Fatalf("feel scale = %#v", scales["feel"])
	}
	if scales["rpe"] != "1-10 (rating of perceived exertion)" {
		t.Fatalf("rpe scale = %#v", scales["rpe"])
	}
}

func loadSubjectiveActivityListFixture(t *testing.T) []intervals.Activity {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "intervals", "testdata", "activity_list_subjective.json"))
	if err != nil {
		t.Fatalf("read activity list fixture: %v", err)
	}
	var activities []intervals.Activity
	if err := json.Unmarshal(data, &activities); err != nil {
		t.Fatalf("decode activity list fixture: %v", err)
	}
	return activities
}

func loadSubjectiveActivityDetailFixture(t *testing.T) intervals.Activity {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "intervals", "testdata", "activity_detail_subjective.json"))
	if err != nil {
		t.Fatalf("read activity detail fixture: %v", err)
	}
	var activity intervals.Activity
	if err := json.Unmarshal(data, &activity); err != nil {
		t.Fatalf("decode activity detail fixture: %v", err)
	}
	return activity
}

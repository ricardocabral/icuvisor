package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

func (f *fakeActivitiesProfileClient) ListCustomItems(ctx context.Context) ([]intervals.CustomItem, error) {
	f.customItemsCalls++
	if f.customItemsErr != nil {
		return nil, f.customItemsErr
	}
	return append([]intervals.CustomItem(nil), f.customItems...), nil
}

func (f *fakeActivityReadClient) ListCustomItems(ctx context.Context) ([]intervals.CustomItem, error) {
	return append([]intervals.CustomItem(nil), f.customItems...), nil
}

func decodeCustomItems(t *testing.T, raw ...string) []intervals.CustomItem {
	t.Helper()
	items := make([]intervals.CustomItem, 0, len(raw))
	for _, entry := range raw {
		var item intervals.CustomItem
		if err := json.Unmarshal([]byte(entry), &item); err != nil {
			t.Fatalf("decoding custom item: %v", err)
		}
		items = append(items, item)
	}
	return items
}

func TestActivityCustomFieldCodes(t *testing.T) {
	t.Parallel()

	items := decodeCustomItems(t,
		`{"id":"c1","type":"ACTIVITY_FIELD","content":{"field":"fueling_score"}}`,
		`{"id":"c2","type":"ACTIVITY_FIELD","content":{"field":"breakfast"}}`,
		`{"id":"c3","type":"activity_field","content":{"field":"fueling_score"}}`,
		`{"id":"c4","type":"INTERVAL_FIELD","content":{"field":"lactate"}}`,
		`{"id":"c5","type":"INPUT_FIELD","content":{"field":"travel_fatigue"}}`,
		`{"id":"c6","type":"ACTIVITY_CHART","content":{"field":"ignored"}}`,
		`{"id":"c7","type":"ACTIVITY_FIELD","content":{}}`,
	)

	got := activityCustomFieldCodes(items)
	if want := []string{"breakfast", "fueling_score"}; !slices.Equal(got, want) {
		t.Fatalf("activityCustomFieldCodes() = %#v, want %#v", got, want)
	}
	if codes := activityCustomFieldCodes(nil); codes != nil {
		t.Fatalf("activityCustomFieldCodes(nil) = %#v, want nil", codes)
	}
}

func TestTerseActivityFieldsWithCustom(t *testing.T) {
	t.Parallel()

	if base := terseActivityFieldsWithCustom(nil); !slices.Equal(base, terseActivityFields) {
		t.Fatalf("terseActivityFieldsWithCustom(nil) = %#v, want terseActivityFields", base)
	}
	got := terseActivityFieldsWithCustom([]string{"fueling_score", "name", "breakfast", "fueling_score"})
	if len(got) != len(terseActivityFields)+2 {
		t.Fatalf("len(got) = %d, want %d", len(got), len(terseActivityFields)+2)
	}
	if !slices.Contains(got, "fueling_score") || !slices.Contains(got, "breakfast") {
		t.Fatalf("terseActivityFieldsWithCustom missing custom codes: %#v", got)
	}
	if !slices.Equal(terseActivityFieldsWithCustom(nil), terseActivityFields) {
		t.Fatal("terseActivityFieldsWithCustom mutated terseActivityFields")
	}
}

func TestGetActivitiesSurfacesCustomFieldsAndTimezone(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, []string{
		`{"id":"a1","name":"Long Ride","type":"Ride","start_date_local":"2026-01-03T07:00:00","distance":40000,"moving_time":3600,"fueling_score":8.5,"breakfast":"oats","empty_custom":null}`,
	}, "metric")
	client.profile.Timezone = "Australia/Brisbane"
	client.customItems = decodeCustomItems(t,
		`{"id":"c1","type":"ACTIVITY_FIELD","content":{"field":"fueling_score"}}`,
		`{"id":"c2","type":"ACTIVITY_FIELD","content":{"field":"breakfast"}}`,
		`{"id":"c3","type":"ACTIVITY_FIELD","content":{"field":"empty_custom"}}`,
	)
	tool := newGetActivitiesToolWithGear(client, client, nil, nil, client, newCustomFieldCache(), "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-01","custom_fields":["fueling_score","breakfast","empty_custom"]}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	payload := resultMap(t, result)
	if meta := payload["_meta"].(map[string]any); meta["timezone"] != "Australia/Brisbane" {
		t.Fatalf("_meta.timezone = %#v, want Australia/Brisbane", meta["timezone"])
	}
	row := payload["activities"].([]any)[0].(map[string]any)
	custom, ok := row["custom_fields"].(map[string]any)
	if !ok {
		t.Fatalf("custom_fields missing from row: %#v", row)
	}
	if custom["fueling_score"] != 8.5 {
		t.Fatalf("custom_fields.fueling_score = %#v, want 8.5", custom["fueling_score"])
	}
	if custom["breakfast"] != "oats" {
		t.Fatalf("custom_fields.breakfast = %#v, want oats", custom["breakfast"])
	}
	if _, ok := custom["empty_custom"]; ok {
		t.Fatalf("null custom field should be dropped: %#v", custom)
	}
	if len(client.listCalls) == 0 {
		t.Fatal("no ListActivities calls recorded")
	}
	for _, code := range []string{"fueling_score", "breakfast", "empty_custom"} {
		if !slices.Contains(client.listCalls[0].Fields, code) {
			t.Fatalf("ListActivities fields %#v missing custom code %q", client.listCalls[0].Fields, code)
		}
	}
}

func TestGetActivityDetailsSurfacesCustomFieldsAndTimezone(t *testing.T) {
	t.Parallel()

	activity := decodeActivityFixture(t, `{"id":"a1","icu_athlete_id":"i12345","name":"Ride","type":"Ride","start_date_local":"2026-01-02T07:00:00","fueling_score":7,"notes_field":"felt strong"}`)
	client := &fakeActivityReadClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "Australia/Brisbane"}}, activity: activity}
	client.customItems = decodeCustomItems(t,
		`{"id":"c1","type":"ACTIVITY_FIELD","content":{"field":"fueling_score"}}`,
		`{"id":"c2","type":"ACTIVITY_FIELD","content":{"field":"notes_field"}}`,
	)
	tool := newGetActivityDetailsToolWithGear(client, client, nil, nil, client, newCustomFieldCache(), "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"a1","custom_fields":["fueling_score","notes_field"]}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	payload := resultMap(t, result)
	if meta := payload["_meta"].(map[string]any); meta["timezone"] != "Australia/Brisbane" {
		t.Fatalf("_meta.timezone = %#v, want Australia/Brisbane", meta["timezone"])
	}
	row := payload["activity"].(map[string]any)
	custom, ok := row["custom_fields"].(map[string]any)
	if !ok {
		t.Fatalf("custom_fields missing from activity detail: %#v", row)
	}
	if custom["fueling_score"] != float64(7) {
		t.Fatalf("custom_fields.fueling_score = %#v, want 7", custom["fueling_score"])
	}
	if custom["notes_field"] != "felt strong" {
		t.Fatalf("custom_fields.notes_field = %#v, want felt strong", custom["notes_field"])
	}
}

func TestGetActivitiesDefaultOmitsCustomFieldLookup(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, []string{
		`{"id":"a1","name":"Ride","type":"Ride","start_date_local":"2026-01-03T07:00:00","fueling_score":8.5}`,
	}, "metric")
	client.customItemsErr = errors.New("custom item lookup failed")
	tool := newGetActivitiesToolWithGear(client, client, nil, nil, client, newCustomFieldCache(), "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-01"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v, want graceful success without custom fields", err)
	}
	row := resultMap(t, result)["activities"].([]any)[0].(map[string]any)
	if _, ok := row["custom_fields"]; ok {
		t.Fatalf("custom_fields should be absent by default: %#v", row)
	}
	if client.customItemsCalls != 0 {
		t.Fatalf("ListCustomItems calls = %d, want 0 without explicit custom_fields", client.customItemsCalls)
	}
}

func TestGetActivitiesRejectsTooManyCustomFields(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, nil, "metric")
	tool := newGetActivitiesToolWithGear(client, client, nil, nil, client, newCustomFieldCache(), "test", "UTC", false)
	fields := make([]string, 0, maxSelectedActivityCustomFields+1)
	for i := 0; i <= maxSelectedActivityCustomFields; i++ {
		fields = append(fields, fmt.Sprintf("field_%02d", i))
	}
	rawFields, err := json.Marshal(fields)
	if err != nil {
		t.Fatalf("marshal fields: %v", err)
	}

	_, err = tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-01","custom_fields":` + string(rawFields) + `}`)})
	message, ok := PublicErrorMessage(err)
	if !ok || !strings.Contains(message, "at most 20") {
		t.Fatalf("Handler() error = %v, public=%q ok=%v; want max custom_fields hint", err, message, ok)
	}
	if client.customItemsCalls != 0 {
		t.Fatalf("ListCustomItems calls = %d, want validation before lookup", client.customItemsCalls)
	}
}

func TestGetActivitiesKnownButAbsentCustomFieldIsOmitted(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, []string{
		`{"id":"a1","name":"Ride","type":"Ride","start_date_local":"2026-01-03T07:00:00"}`,
	}, "metric")
	client.customItems = decodeCustomItems(t, `{"id":"c1","type":"ACTIVITY_FIELD","content":{"field":"vo2max_est"}}`)
	tool := newGetActivitiesToolWithGear(client, client, nil, nil, client, newCustomFieldCache(), "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-01","custom_fields":["vo2max_est"]}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	row := resultMap(t, result)["activities"].([]any)[0].(map[string]any)
	if _, ok := row["custom_fields"]; ok {
		t.Fatalf("absent custom field should be omitted: %#v", row)
	}
}

func TestGetActivityDetailsUnknownCustomFieldHintsAvailableFields(t *testing.T) {
	t.Parallel()

	activity := decodeActivityFixture(t, `{"id":"a1","icu_athlete_id":"i12345","name":"Ride","type":"Ride","start_date_local":"2026-01-02T07:00:00"}`)
	client := &fakeActivityReadClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}}, activity: activity}
	client.customItems = decodeCustomItems(t, `{"id":"c1","type":"ACTIVITY_FIELD","content":{"field":"vo2max_est"}}`)
	tool := newGetActivityDetailsToolWithGear(client, client, nil, nil, client, newCustomFieldCache(), "test", "UTC", false)

	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"a1","custom_fields":["unknown_field"]}`)})
	message, ok := PublicErrorMessage(err)
	if !ok || !strings.Contains(message, "unknown_field") || !strings.Contains(message, "vo2max_est") {
		t.Fatalf("Handler() error = %v, public=%q ok=%v; want public unknown-field hint", err, message, ok)
	}
}

func TestGetActivitiesUnknownCustomFieldHintsAvailableFields(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, []string{
		`{"id":"a1","name":"Ride","type":"Ride","start_date_local":"2026-01-03T07:00:00"}`,
	}, "metric")
	client.customItems = decodeCustomItems(t, `{"id":"c1","type":"ACTIVITY_FIELD","content":{"field":"vo2max_est"}}`)
	tool := newGetActivitiesToolWithGear(client, client, nil, nil, client, newCustomFieldCache(), "test", "UTC", false)

	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-01","custom_fields":["unknown_field"]}`)})
	message, ok := PublicErrorMessage(err)
	if !ok || !strings.Contains(message, "unknown_field") || !strings.Contains(message, "vo2max_est") {
		t.Fatalf("Handler() error = %v, public=%q ok=%v; want public unknown-field hint", err, message, ok)
	}
}

func TestCustomFieldCacheMemoizesPerAthlete(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, []string{
		`{"id":"a1","name":"Ride","type":"Ride","start_date_local":"2026-01-03T07:00:00"}`,
	}, "metric")
	client.customItems = decodeCustomItems(t, `{"id":"c1","type":"ACTIVITY_FIELD","content":{"field":"fueling_score"}}`)
	tool := newGetActivitiesToolWithGear(client, client, nil, nil, client, newCustomFieldCache(), "test", "UTC", false)

	for i := 0; i < 3; i++ {
		if _, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-01","custom_fields":["fueling_score"]}`)}); err != nil {
			t.Fatalf("Handler() error = %v", err)
		}
	}
	if client.customItemsCalls != 1 {
		t.Fatalf("ListCustomItems calls = %d, want 1 (memoized)", client.customItemsCalls)
	}
}

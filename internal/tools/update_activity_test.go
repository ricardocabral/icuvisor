package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

type fakeActivityUpdaterClient struct {
	fakeProfileClient
	activity intervals.Activity
	calls    []intervals.UpdateActivityParams
	err      error
}

func (f *fakeActivityUpdaterClient) UpdateActivity(ctx context.Context, params intervals.UpdateActivityParams) (intervals.Activity, error) {
	f.calls = append(f.calls, params)
	return f.activity, f.err
}

func decodeActivity(t *testing.T, raw string) intervals.Activity {
	t.Helper()
	var activity intervals.Activity
	if err := json.Unmarshal([]byte(raw), &activity); err != nil {
		t.Fatalf("decode activity: %v", err)
	}
	return activity
}

func TestUpdateActivitySuccessSparseFields(t *testing.T) {
	t.Parallel()

	description := "## Completed strength\n\n- Back squat: 3 × 5 @ 80 kg total.\n- Split squat: 3 × 8/side @ 16 kg per hand."
	client := &fakeActivityUpdaterClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345"}},
		activity:          decodeActivity(t, `{"id":"a1","name":"Strength A","description":"## Completed strength\n\n- Back squat: 3 × 5 @ 80 kg total.\n- Split squat: 3 × 8/side @ 16 kg per hand.","extra":null}`),
	}
	tool := newUpdateActivityTool(client, client, "test", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":" a1 ","name":" Strength A ","description":"## Completed strength\n\n- Back squat: 3 × 5 @ 80 kg total.\n- Split squat: 3 × 8/side @ 16 kg per hand.","carbs_ingested_g":90,"include_full":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 1 {
		t.Fatalf("calls = %#v, want 1", client.calls)
	}
	call := client.calls[0]
	if call.ActivityID != "a1" || call.Name != "Strength A" || !call.NameSet || call.Description != description || !call.DescriptionSet || call.CarbsIngested != 90 || !call.CarbsIngestedSet {
		t.Fatalf("call = %#v, want trimmed sparse update", call)
	}
	out := resultMap(t, result)
	if out["activity_id"] != "a1" || out["status"] != "updated" {
		t.Fatalf("response = %#v, want updated confirmation", out)
	}
	fields, ok := out["fields_updated"].([]any)
	if !ok || len(fields) != 3 || fields[0] != "carbs_ingested_g" || fields[1] != "description" || fields[2] != "name" {
		t.Fatalf("fields_updated = %#v, want [carbs_ingested_g description name]", out["fields_updated"])
	}
	meta := out["_meta"].(map[string]any)
	if meta["athleteId"] != "i12345" || meta["destructive"] != false || meta["append_only"] != false || meta["source_endpoint"] != "/activity/{activityId}" {
		t.Fatalf("meta = %#v, want non-destructive metadata", meta)
	}
	full := out["full"].(map[string]any)
	assertKeyPresentNil(t, full, "extra")
}

func TestUpdateActivityClearsDescriptionWhenExplicitEmpty(t *testing.T) {
	t.Parallel()

	client := &fakeActivityUpdaterClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345"}},
		activity:          decodeActivity(t, `{"id":"a1"}`),
	}
	tool := newUpdateActivityTool(client, client, "test", false)

	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"a1","description":""}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	call := client.calls[0]
	if call.NameSet || !call.DescriptionSet || call.Description != "" {
		t.Fatalf("call = %#v, want explicit description clear without name change", call)
	}
}

func TestUpdateActivitySetsCarbsIngested(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		raw  string
		want int
	}{
		{name: "positive grams", raw: `{"activity_id":"a1","carbs_ingested_g":90}`, want: 90},
		{name: "logged zero", raw: `{"activity_id":"a1","carbs_ingested_g":0}`, want: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := &fakeActivityUpdaterClient{activity: decodeActivity(t, `{"id":"a1"}`)}
			tool := newUpdateActivityTool(client, client, "test", false)

			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(tc.raw)})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			call := client.calls[0]
			if !call.CarbsIngestedSet || call.CarbsIngested != tc.want {
				t.Fatalf("call = %#v, want carbs_ingested=%d set", call, tc.want)
			}
			fields := resultMap(t, result)["fields_updated"].([]any)
			if len(fields) != 1 || fields[0] != "carbs_ingested_g" {
				t.Fatalf("fields_updated = %#v, want [carbs_ingested_g]", fields)
			}
		})
	}
}

func TestUpdateActivityRejectsBadArguments(t *testing.T) {
	t.Parallel()

	client := &fakeActivityUpdaterClient{}
	tool := newUpdateActivityTool(client, client, "test", false)
	for _, raw := range []string{
		`{"activity_id":""}`,
		`{"activity_id":"a1"}`,
		`{"activity_id":"a1","name":""}`,
		`{"activity_id":"a1","name":"   "}`,
		`{"activity_id":"a1","carbs_ingested_g":null}`,
		`{"activity_id":"a1","carbs_ingested_g":-1}`,
		`{"activity_id":"a1","carbs_ingested_g":2147483648}`,
		`{"activity_id":"a1","carbs_ingested_g":1.5}`,
		`{"activity_id":"a1","carbs_ingested_g":"90"}`,
		`{"activity_id":"a1","name":"x","confirm":true}`,
	} {
		if _, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(raw)}); err == nil {
			t.Fatalf("Handler(%s) error = nil, want validation error", raw)
		}
	}
}

func TestUpdateActivityRejectsInvalidRPEAndUnverifiedAliasesBeforeUpdater(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		raw  string
	}{
		{name: "rpe zero", raw: `{"activity_id":"a1","rpe":0}`},
		{name: "rpe above range", raw: `{"activity_id":"a1","rpe":11}`},
		{name: "rpe fractional", raw: `{"activity_id":"a1","rpe":1.5}`},
		{name: "rpe string", raw: `{"activity_id":"a1","rpe":"7"}`},
		{name: "rpe null", raw: `{"activity_id":"a1","rpe":null}`},
		{name: "perceived exertion alias", raw: `{"activity_id":"a1","perceived_exertion":7}`},
		{name: "native upstream key", raw: `{"activity_id":"a1","icu_rpe":7}`},
		{name: "unsupported field is atomic with description", raw: `{"activity_id":"a1","description":"note","icu_rpe":7}`},
		{name: "unsupported field is atomic with carbs", raw: `{"activity_id":"a1","carbs_ingested_g":90,"perceived_exertion":7}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client := &fakeActivityUpdaterClient{}
			tool := newUpdateActivityTool(client, client, "test", false)
			_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(tc.raw)})
			if err == nil {
				t.Fatalf("Handler(%s) error = nil, want unsupported RPE rejection", tc.raw)
			}
			if len(client.calls) != 0 {
				t.Fatalf("updater calls = %#v, want zero for rejected request", client.calls)
			}
		})
	}
}

func TestUpdateActivityPublicError(t *testing.T) {
	t.Parallel()

	client := &fakeActivityUpdaterClient{err: errors.New("upstream detail")}
	tool := newUpdateActivityTool(client, client, "test", false)
	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"a1","name":"x"}`)})
	if message, ok := PublicErrorMessage(err); !ok || message != updateActivityMessage {
		t.Fatalf("PublicErrorMessage = %q, %v; err = %v", message, ok, err)
	}
}

func TestUpdateActivityRegistrationMetadata(t *testing.T) {
	t.Parallel()

	client := &fakeActivityUpdaterClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345"}}}
	tool := newUpdateActivityTool(client, client, "test", false)
	if tool.Requirement != RequirementWrite {
		t.Fatalf("requirement = %q, want write", tool.Requirement)
	}
	description := strings.ToLower(tool.Description)
	if !strings.Contains(description, "non-destructive") || strings.Contains(description, "confirm") {
		t.Fatalf("description = %q, want non-destructive language without confirm", tool.Description)
	}
	props := tool.InputSchema.(map[string]any)["properties"].(map[string]any)
	for _, name := range []string{"activity_id", "name", "description", "carbs_ingested_g", "include_full", "tags", "commute", "feel", "rpe", "sub_type", "distance_m"} {
		if _, ok := props[name]; !ok {
			t.Fatalf("schema missing %s", name)
		}
	}
	carbs := props["carbs_ingested_g"].(map[string]any)
	carbsDescription := strings.ToLower(carbs["description"].(string))
	if carbs["type"] != "integer" || carbs["minimum"] != 0 || carbs["maximum"] != 2147483647 {
		t.Fatalf("carbs_ingested_g schema = %#v, want bounded integer", carbs)
	}
	for _, phrase := range []string{"logged zero", "omit", "clearing is not supported", "carbs_used_g", "strava"} {
		if !strings.Contains(carbsDescription, phrase) {
			t.Fatalf("carbs_ingested_g description = %q, want %q", carbsDescription, phrase)
		}
	}
	for _, name := range []string{"carbs_used_g", "confirm", "perceived_exertion", "icu_rpe", "moving_time", "elapsed_time", "start_date"} {
		if _, ok := props[name]; ok {
			t.Fatalf("schema includes unsupported or read-only property %q", name)
		}
	}
}

func TestUpdateActivityWritableMetadataSparseValues(t *testing.T) {
	for _, tc := range []struct {
		name, raw string
		field     string
	}{
		{name: "clear tags", raw: `{"activity_id":"a1","tags":[]}`, field: "tags"},
		{name: "false commute", raw: `{"activity_id":"a1","commute":false}`, field: "commute"},
		{name: "feel", raw: `{"activity_id":"a1","feel":1}`, field: "feel"},
		{name: "rpe", raw: `{"activity_id":"a1","rpe":10}`, field: "rpe"},
		{name: "subtype", raw: `{"activity_id":"a1","sub_type":"NONE"}`, field: "sub_type"},
		{name: "distance metres", raw: `{"activity_id":"a1","distance_m":0}`, field: "distance_m"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := &fakeActivityUpdaterClient{activity: decodeActivity(t, strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(tc.raw, "activity_id", "id"), "\"rpe\"", "\"icu_rpe\""), "distance_m", "distance"))}
			tool := newUpdateActivityTool(client, nil, "test", false)
			result, err := tool.Handler(context.Background(), Request{Arguments: json.RawMessage(tc.raw)})
			if err != nil {
				t.Fatal(err)
			}
			if len(client.calls) != 1 {
				t.Fatalf("calls=%v", client.calls)
			}
			call := client.calls[0]
			switch tc.field {
			case "tags":
				if call.Tags == nil || len(*call.Tags) != 0 {
					t.Fatalf("tags=%v", call.Tags)
				}
			case "commute":
				if call.Commute == nil || *call.Commute {
					t.Fatalf("commute=%v", call.Commute)
				}
			case "feel":
				if call.Feel == nil || *call.Feel != 1 {
					t.Fatalf("feel=%v", call.Feel)
				}
			case "rpe":
				if call.RPE == nil || *call.RPE != 10 {
					t.Fatalf("rpe=%v", call.RPE)
				}
			case "sub_type":
				if call.SubType == nil || *call.SubType != "NONE" {
					t.Fatalf("subtype=%v", call.SubType)
				}
			case "distance_m":
				if call.DistanceM == nil || *call.DistanceM != 0 {
					t.Fatalf("distance=%v", call.DistanceM)
				}
			}
			if tc.field == "feel" || tc.field == "rpe" {
				meta := resultMap(t, result)["_meta"].(map[string]any)
				scales, ok := meta["scales"].(map[string]any)
				if !ok || scales[tc.field] == nil {
					t.Fatalf("scale metadata=%v", meta)
				}
			}
			fields := resultMap(t, result)["fields_updated"].([]any)
			if len(fields) != 1 || fields[0] != tc.field {
				t.Fatalf("fields=%v", fields)
			}
		})
	}
}

func TestUpdateActivityWritableMetadataRejectsInvalidValues(t *testing.T) {
	for _, field := range []string{`"tags":null`, `"commute":null`, `"feel":null`, `"rpe":null`, `"sub_type":null`, `"distance_m":null`, `"feel":0`, `"feel":6`, `"rpe":0`, `"rpe":11`, `"distance_m":-1`, `"sub_type":"race"`, `"tags":[""]`, `"distance_m":1e100`, `"moving_time":20`, `"elapsed_time":20`, `"start_date":"2026-01-01"`} {
		client := &fakeActivityUpdaterClient{}
		tool := newUpdateActivityTool(client, nil, "test", false)
		_, err := tool.Handler(context.Background(), Request{Arguments: json.RawMessage(`{"activity_id":"a1",` + field + `}`)})
		if err == nil || len(client.calls) != 0 {
			t.Fatalf("field=%s err=%v calls=%v", field, err, client.calls)
		}
	}
}

func TestUpdateActivityDoesNotClaimUnverifiedMetadata(t *testing.T) {
	for _, tc := range []struct{ name, raw, wantStatus string }{
		{name: "missing", raw: `{"id":"a1"}`, wantStatus: "verification_unavailable"},
		{name: "different", raw: `{"id":"a1","commute":true}`, wantStatus: "verification_mismatch"},
		{name: "confirmed", raw: `{"id":"a1","commute":false}`, wantStatus: "updated"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := &fakeActivityUpdaterClient{activity: decodeActivity(t, tc.raw)}
			tool := newUpdateActivityTool(client, nil, "test", false)
			result, err := tool.Handler(context.Background(), Request{Arguments: json.RawMessage(`{"activity_id":"a1","commute":false}`)})
			if err != nil {
				t.Fatal(err)
			}
			out := resultMap(t, result)
			if out["status"] != tc.wantStatus {
				t.Fatalf("result=%v", out)
			}
			if tc.wantStatus != "updated" {
				if fields, ok := out["fields_updated"].([]any); ok && len(fields) > 0 {
					t.Fatalf("unverified update claimed: %v", out)
				}
			}
		})
	}
}

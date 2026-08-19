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

	client := &fakeActivityUpdaterClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345"}},
		activity:          decodeActivity(t, `{"id":"a1","name":"Threshold ride","description":"Held target W","extra":null}`),
	}
	tool := newUpdateActivityTool(client, client, "test", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":" a1 ","name":" Threshold ride ","description":"Held target W","carbs_ingested_g":90,"include_full":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 1 {
		t.Fatalf("calls = %#v, want 1", client.calls)
	}
	call := client.calls[0]
	if call.ActivityID != "a1" || call.Name != "Threshold ride" || !call.NameSet || call.Description != "Held target W" || !call.DescriptionSet || call.CarbsIngested != 90 || !call.CarbsIngestedSet {
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

func TestUpdateActivityRejectsUnsupportedRPEBeforeUpdater(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		raw  string
	}{
		{name: "rpe zero", raw: `{"activity_id":"a1","rpe":0}`},
		{name: "rpe in range", raw: `{"activity_id":"a1","rpe":7}`},
		{name: "rpe upper bound", raw: `{"activity_id":"a1","rpe":10}`},
		{name: "rpe above range", raw: `{"activity_id":"a1","rpe":11}`},
		{name: "rpe fractional", raw: `{"activity_id":"a1","rpe":1.5}`},
		{name: "rpe string", raw: `{"activity_id":"a1","rpe":"7"}`},
		{name: "rpe null", raw: `{"activity_id":"a1","rpe":null}`},
		{name: "perceived exertion alias", raw: `{"activity_id":"a1","perceived_exertion":7}`},
		{name: "native upstream key", raw: `{"activity_id":"a1","icu_rpe":7}`},
		{name: "unsupported field is atomic with name", raw: `{"activity_id":"a1","name":"renamed","rpe":7}`},
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
	for _, name := range []string{"activity_id", "name", "description", "carbs_ingested_g", "include_full"} {
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
	for _, name := range []string{"carbs_used_g", "confirm", "rpe", "perceived_exertion", "icu_rpe", "feel"} {
		if _, ok := props[name]; ok {
			t.Fatalf("schema includes unsupported or read-only property %q", name)
		}
	}
}

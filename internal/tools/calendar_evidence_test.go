package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/safety"
)

func TestCalendarConstraintsSurviveTerseReadSurfaces(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, fields string
		present      bool
	}{
		{"omitted", "", false},
		{"null", `,"training_availability":null,"can_train_sports":null,"max_training_time":null`, false},
		{"explicit zero and empty", `,"training_availability":"UNAVAILABLE","can_train_sports":[],"max_training_time":0`, true},
		{"limited sports", `,"training_availability":"LIMITED","can_train_sports":["Swim"],"max_training_time":45`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			events := decodeToolEvents(t, `{"id":"constraint","category":"NOTE","name":"Restrictions","start_date_local":"2026-05-26"`+tc.fields+`}`)
			profile := fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", Timezone: "UTC"}}
			for _, surface := range []string{"list", "detail", "planning"} {
				t.Run(surface, func(t *testing.T) {
					client := &fakeEventsTrainingPlanClient{fakeProfileClient: profile, events: events, eventDetail: events[0]}
					var tool Tool
					var args string
					switch surface {
					case "list":
						tool = newGetEventsTool(client, client, "test", "UTC", false)
						args = `{"oldest":"2026-05-25","newest":"2026-05-31"}`
					case "detail":
						tool = newGetEventByIDTool(client, client, "test", "UTC", false)
						args = `{"event_id":"constraint"}`
					case "planning":
						p := &fakePlanningContextClient{fakeProfileClient: profile, weekEvents: events}
						tool = newGetPlanningContextToolWithClock(p, p, "test", "UTC", false, fixedTodayClock())
						args = `{"week_start":"2026-05-25"}`
					}
					result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(args)})
					if err != nil {
						t.Fatal(err)
					}
					out := resultMap(t, result)
					var row map[string]any
					switch surface {
					case "list":
						row = out["events"].([]any)[0].(map[string]any)
					case "detail":
						row = out["event"].(map[string]any)
					case "planning":
						row = out["week_events"].(map[string]any)["notes"].([]any)[0].(map[string]any)
					}
					for _, key := range []string{"training_availability", "can_train_sports", "max_training_time"} {
						if _, ok := row[key]; ok != tc.present {
							t.Fatalf("%s presence = %t, want %t; row %#v", key, ok, tc.present, row)
						}
					}
					if tc.present {
						var want map[string]any
						if err := json.Unmarshal([]byte("{"+tc.fields[1:]+"}"), &want); err != nil {
							t.Fatal(err)
						}
						for key, value := range want {
							gotJSON, _ := json.Marshal(row[key])
							wantJSON, _ := json.Marshal(value)
							if string(gotJSON) != string(wantJSON) {
								t.Fatalf("%s = %s, want %s", key, gotJSON, wantJSON)
							}
						}
					}
					if _, ok := row["full"]; ok {
						t.Fatalf("terse row leaked full: %#v", row)
					}
				})
			}
		})
	}
}

type durationEvidenceWriter struct {
	fakeEventWriterClient
	detail      intervals.Event
	detailErr   error
	detailCalls []string
}

func (f *durationEvidenceWriter) GetEvent(_ context.Context, id string) (intervals.Event, error) {
	f.detailCalls = append(f.detailCalls, id)
	return f.detail, f.detailErr
}

func TestEventWriteDurationReadbackEvidence(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, fields, status string
		readErr              error
	}{
		{"strength collapsed to warmup", `,"moving_time":600,"time_target":3600`, "mismatch", nil},
		{"stored target disagrees", `,"moving_time":3600,"time_target":600`, "mismatch", nil},
		{"target mismatch without parsed duration", `,"time_target":600`, "mismatch", nil},
		{"readback canceled after accepted write", "", "unavailable", context.Canceled},
		{"stored duration agrees", `,"moving_time":3600,"time_target":3600`, "verified", nil},
		{"explicit stored zero", `,"moving_time":0,"time_target":3600`, "mismatch", nil},
		{"target alone cannot verify parsed duration", `,"time_target":3600`, "unavailable", nil},
		{"readback fails", "", "unavailable", errors.New("upstream secret detail")},
		{"wrong event identity", `,"moving_time":3600`, "unavailable", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			id := "written"
			if tc.name == "wrong event identity" {
				id = "other"
			}
			client := &durationEvidenceWriter{fakeEventWriterClient: fakeEventWriterClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", Timezone: "UTC"}}, event: decodeToolEvents(t, `{"id":"written","category":"WORKOUT","type":"WeightTraining","start_date_local":"2026-05-26","moving_time":3600}`)[0]}, detail: decodeToolEvents(t, `{"id":"`+id+`","category":"WORKOUT","start_date_local":"2026-05-26"`+tc.fields+`}`)[0], detailErr: tc.readErr}
			tool := newAddOrUpdateEventTool(client, client, "test", "UTC", false)
			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"date":"2026-05-26","category":"WORKOUT","type":"WeightTraining","name":"Strength","description":"- Warmup 10m\n- Squat 3 sets","moving_time_seconds":3600}`)})
			if err != nil {
				t.Fatal(err)
			}
			out := resultMap(t, result)
			meta := out["_meta"].(map[string]any)
			evidence, ok := meta["duration_verification"].(map[string]any)
			if !ok {
				t.Fatalf("missing duration evidence: %#v", meta)
			}
			encoded, _ := json.Marshal(out)
			if strings.Contains(string(encoded), "upstream secret detail") {
				t.Fatal("readback error leaked internal detail")
			}
			if evidence["status"] != tc.status {
				t.Fatalf("evidence %#v, want %s", evidence, tc.status)
			}
			if tc.status != "verified" && evidence["warning"] == nil {
				t.Fatalf("missing actionable warning: %#v", evidence)
			}
			if len(client.calls) != 1 || len(client.detailCalls) != 1 || client.detailCalls[0] != "written" {
				t.Fatalf("writes/reads = %#v/%#v", client.calls, client.detailCalls)
			}
			if meta["confirmation_status"] != writeReturnedByUpstreamStatus || out["event"].(map[string]any)["event_id"] != "written" {
				t.Fatalf("write identity lost: %#v", out)
			}
		})
	}
}

func TestEventDurationEvidencePreservesRetryIdentityAndNewerEdits(t *testing.T) {
	t.Parallel()
	for _, interrupted := range []bool{false, true} {
		t.Run(map[bool]string{false: "retry after newer user edit", true: "interrupted create"}[interrupted], func(t *testing.T) {
			stored := decodeToolEvents(t, `{"id":"kept","external_id":"stable","category":"WORKOUT","type":"WeightTraining","name":"User's newer title","description":"User's newer instructions","start_date_local":"2026-05-26","moving_time":600,"time_target":3600}`)
			var client EventWriterClient
			var profile ProfileClient
			normal := &fakeEventWriterClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", Timezone: "UTC"}}, events: stored}
			recovered := &interruptedEventWriter{fakeProfileClient: normal.fakeProfileClient, listResults: [][]intervals.Event{nil, stored}}
			if interrupted {
				client = recovered
				profile = recovered
			} else {
				client = normal
				profile = normal
			}
			tool := newAddOrUpdateEventTool(client, profile, "test", "UTC", false)
			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"date":"2026-05-26","category":"WORKOUT","type":"WeightTraining","external_id":"stable","name":"Original title","description":"Original instructions","moving_time_seconds":3600}`)})
			if err != nil {
				t.Fatal(err)
			}
			out := resultMap(t, result)
			row := out["event"].(map[string]any)
			meta := out["_meta"].(map[string]any)
			if row["event_id"] != "kept" || row["name"] != "User's newer title" || row["description"] != "User's newer instructions" {
				t.Fatalf("newer edit lost: %#v", out)
			}
			evidence, ok := meta["duration_verification"].(map[string]any)
			if !ok || evidence["status"] != "mismatch" || evidence["source"] != "same_day_event_list" {
				t.Fatalf("duration mismatch lost: %#v", meta)
			}
			wantStatus := skippedExistingEventStatus
			if interrupted {
				wantStatus = verifiedAfterInterruptedWriteStatus
			}
			if meta["confirmation_status"] != wantStatus {
				t.Fatalf("write identity status changed: %#v", meta)
			}
			if len(normal.calls) != 0 {
				t.Fatal("retry overwrote newer user edit")
			}
			if interrupted && (len(recovered.writeCalls) != 1 || len(recovered.listCalls) != 2) {
				t.Fatalf("interrupted create retried: %#v", recovered)
			}
		})
	}
}

func TestEventWriteDurationExpectationSources(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, args, readFields string
		wantRead               bool
		field                  string
		requested, stored      int
	}{
		{"elapsed explicit zero", `"elapsed_time_seconds":0`, `,"elapsed_time":0`, true, "elapsed_time_seconds", 0, 0},
		{"timed repeated steps", `"workout_doc":{"steps":[{"reps":3,"steps":[{"duration":60,"freeride":true},{"duration":30,"freeride":true}]}]}`, `,"moving_time":270`, true, "moving_time_seconds", 270, 270},
		{"distance steps have no promised duration", `"workout_doc":{"steps":[{"distance":{"value":1000,"unit":"m"},"freeride":true}]}`, ``, false, "", 0, 0},
		{"mixed steps cannot verify total timed duration", `"workout_doc":{"steps":[{"duration":60,"freeride":true},{"distance":{"value":1000,"unit":"m"},"freeride":true}]}`, ``, false, "", 0, 0},
		{"lap duration is open ended", `"workout_doc":{"steps":[{"duration":60,"press_lap":true,"freeride":true}]}`, ``, false, "", 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := &durationEvidenceWriter{fakeEventWriterClient: fakeEventWriterClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", Timezone: "UTC"}}, event: decodeToolEvents(t, `{"id":"kept","category":"WORKOUT","type":"Ride","start_date_local":"2026-05-26"}`)[0]}, detail: decodeToolEvents(t, `{"id":"kept"`+tc.readFields+`}`)[0]}
			tool := newAddOrUpdateEventTool(client, client, "test", "UTC", false)
			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"date":"2026-05-26","category":"WORKOUT","type":"Ride",` + tc.args + `}`)})
			if err != nil {
				t.Fatal(err)
			}
			meta := resultMap(t, result)["_meta"].(map[string]any)
			evidence, ok := meta["duration_verification"].(map[string]any)
			if ok != tc.wantRead || (len(client.detailCalls) == 1) != tc.wantRead {
				t.Fatalf("evidence/read = %#v/%#v", meta, client.detailCalls)
			}
			if tc.wantRead {
				fields := evidence["fields"].([]any)
				field := fields[0].(map[string]any)
				if evidence["status"] != "verified" || field["field"] != tc.field || field["requested_seconds"] != float64(tc.requested) || field["stored_seconds"] != float64(tc.stored) {
					t.Fatalf("duration evidence = %#v", evidence)
				}
			}
		})
	}
}

func TestApplyTrainingPlanCannotReplaceCalendarRestrictions(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ name, fields string }{
		{"unavailable", `,"training_availability":"UNAVAILABLE"`},
		{"limited", `,"training_availability":"LIMITED"`},
		{"empty sports", `,"can_train_sports":[]`},
		{"sport restriction", `,"can_train_sports":["Swim"]`},
		{"zero time cap", `,"max_training_time":0`},
		{"time cap", `,"max_training_time":45`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := newApplyTrainingPlanTestClient(t)
			client.workouts = client.workouts[:1]
			client.events = decodeToolEvents(t, `{"id":"protected","category":"WORKOUT","type":"Ride","start_date_local":"2026-06-01"`+tc.fields+`}`)
			tool := newApplyTrainingPlanTool(client, client, "test", "UTC", false, safety.NewCapability(safety.ModeFull))
			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"plan_id":"plan-1","start_date":"2026-06-01","dry_run":false,"conflict_policy":"replace_existing"}`)})
			if err != nil {
				t.Fatal(err)
			}
			if len(client.writeCalls) != 0 || len(client.deleteCalls) != 0 {
				t.Fatalf("calendar restriction replaced: writes/deletes %#v/%#v", client.writeCalls, client.deleteCalls)
			}
			meta := resultMap(t, result)["_meta"].(map[string]any)
			skipped := meta["skipped"].([]any)
			conflict := skipped[0].(map[string]any)["conflicts"].([]any)[0].(map[string]any)
			if conflict["protected"] != true || conflict["reason"] != "calendar_constraints" {
				t.Fatalf("missing explicit protection: %#v", conflict)
			}
			var want map[string]any
			if err := json.Unmarshal([]byte("{"+tc.fields[1:]+"}"), &want); err != nil {
				t.Fatal(err)
			}
			for key, value := range want {
				gotJSON, _ := json.Marshal(conflict[key])
				wantJSON, _ := json.Marshal(value)
				if string(gotJSON) != string(wantJSON) {
					t.Fatalf("%s = %s, want %s", key, gotJSON, wantJSON)
				}
			}
		})
	}
}

func TestApplyTrainingPlanReportsStoredDurationMismatch(t *testing.T) {
	t.Parallel()
	client := newApplyTrainingPlanTestClient(t)
	client.workouts = client.workouts[:1]
	client.created = decodeToolEvents(t, `{"id":"kept","category":"WORKOUT","type":"Ride","start_date_local":"2026-06-01","time_target":3600}`)
	writer := &applyDurationEvidenceClient{fakeApplyTrainingPlanClient: client, event: decodeToolEvents(t, `{"id":"kept","moving_time":600,"time_target":3600}`)[0]}
	tool := newApplyTrainingPlanTool(writer, writer, "test", "UTC", false, safety.NewCapability(safety.ModeSafe))
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"plan_id":"plan-1","start_date":"2026-06-01","dry_run":false}`)})
	if err != nil {
		t.Fatal(err)
	}
	meta := resultMap(t, result)["_meta"].(map[string]any)
	evidence, ok := meta["duration_verifications"].([]any)
	if !ok || len(evidence) != 1 {
		t.Fatalf("missing duration evidence: %#v", meta)
	}
	row := evidence[0].(map[string]any)
	if row["event_id"] != "kept" || row["verification"].(map[string]any)["status"] != "mismatch" || meta["created_count"] != float64(1) {
		t.Fatalf("lost write identity or duration mismatch: %#v", meta)
	}
	if len(client.writeCalls) != 1 || writer.reads != 1 {
		t.Fatalf("writes/reads %d/%d", len(client.writeCalls), writer.reads)
	}
}

type applyDurationEvidenceClient struct {
	*fakeApplyTrainingPlanClient
	event intervals.Event
	reads int
}

func (f *applyDurationEvidenceClient) GetEvent(_ context.Context, _ string) (intervals.Event, error) {
	f.reads++
	return f.event, nil
}

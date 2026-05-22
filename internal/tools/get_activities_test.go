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

type fakeActivitiesProfileClient struct {
	fakeProfileClient
	activities         []intervals.Activity
	activityPages      [][]intervals.Activity
	applyListParams    bool
	rejectInvalidRange bool
	listCalls          []intervals.ListActivitiesParams
	listErr            error
	gear               []intervals.Gear
	gearByTarget       map[string][]intervals.Gear
	gearErr            error
	gearCalls          int
}

func (f *fakeActivitiesProfileClient) ListGear(ctx context.Context) ([]intervals.Gear, error) {
	f.gearCalls++
	if f.gearErr != nil {
		return nil, f.gearErr
	}
	if f.gearByTarget != nil {
		target, _ := intervals.TargetAthleteIDFromContext(ctx)
		return f.gearByTarget[target], nil
	}
	return f.gear, nil
}

func (f *fakeActivitiesProfileClient) ListActivities(ctx context.Context, params intervals.ListActivitiesParams) ([]intervals.Activity, error) {
	callIndex := len(f.listCalls)
	f.listCalls = append(f.listCalls, params)
	if f.listErr != nil {
		return nil, f.listErr
	}
	if f.rejectInvalidRange && activityBoundaryBefore(params.Newest, params.Oldest) {
		return nil, errors.New("invalid upstream range")
	}
	if len(f.activityPages) > 0 {
		pageIndex := min(callIndex, len(f.activityPages)-1)
		return append([]intervals.Activity(nil), f.activityPages[pageIndex]...), nil
	}
	out := append([]intervals.Activity(nil), f.activities...)
	if f.applyListParams {
		out = applyFakeActivityListParams(out, params)
	}
	return out, nil
}

func applyFakeActivityListParams(activities []intervals.Activity, params intervals.ListActivitiesParams) []intervals.Activity {
	filtered := make([]intervals.Activity, 0, len(activities))
	for _, activity := range activities {
		if params.Newest != "" && activitySortDate(activity) > params.Newest {
			continue
		}
		filtered = append(filtered, activity)
	}
	if params.Limit > 0 && len(filtered) > params.Limit {
		filtered = filtered[:params.Limit]
	}
	return filtered
}

func TestGetActivitiesRegistrationMetadata(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, nil, "metric")
	tool := newGetActivitiesToolWithGear(client, client, nil, nil, "test", "UTC", false)
	if !strings.Contains(tool.Description, "List activities for a date range") {
		t.Fatalf("description = %q, want distinguishing activity-list sentence", tool.Description)
	}
	properties := tool.InputSchema.(map[string]any)["properties"].(map[string]any)
	for _, name := range []string{"oldest", "newest", "include_unnamed", "page_size", "next_page_token", "include_full"} {
		if _, ok := properties[name]; !ok {
			t.Fatalf("schema missing %s", name)
		}
	}
}

func TestGetActivitiesCaloriesBurnedSemanticsAndNullStripping(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, []string{
		`{"id":"zero","name":"Recovery Ride","type":"Ride","start_date_local":"2026-01-03T07:00:00","calories":0}`,
		`{"id":"absent","name":"Untyped Ride","type":"Ride","start_date_local":"2026-01-02T07:00:00"}`,
	}, "metric")
	tool := newGetActivitiesToolWithGear(client, client, nil, nil, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-01"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	payload := resultMap(t, result)
	rows := payload["activities"].([]any)
	zero := rows[0].(map[string]any)
	if zero["calories_burned"] != float64(0) {
		t.Fatalf("zero calories row = %#v, want present calories_burned=0", zero)
	}
	for _, key := range []string{"calories", "calories_intake", "calories_total", "carbs_g", "protein_g", "fat_g"} {
		if _, ok := zero[key]; ok {
			t.Fatalf("zero calories row emitted unsupported/ambiguous %s: %#v", key, zero)
		}
	}
	absent := rows[1].(map[string]any)
	if _, ok := absent["calories_burned"]; ok {
		t.Fatalf("absent calories row emitted calories_burned: %#v", absent)
	}
	semantics := payload["_meta"].(map[string]any)["field_semantics"].(map[string]any)
	if !strings.Contains(semantics["calories_burned"].(string), "Active/exercise calories") {
		t.Fatalf("field_semantics = %#v, want active calories label", semantics)
	}
}

func TestGetActivitiesActivityNutritionFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		rawActivity   string
		wantIngestedG *int
		wantUsedG     *int
		wantAbsent    []string
		wantSemantics []string
	}{
		{
			name:          "terse default with carbs_ingested and carbs_used",
			rawActivity:   `{"id":"a1","name":"Long Ride","type":"Ride","start_date_local":"2026-05-14T07:00:00","calories":1850,"carbs_ingested":210,"carbs_used":390}`,
			wantIngestedG: intPtr(210),
			wantUsedG:     intPtr(390),
			wantAbsent:    []string{"carbs", "carbohydrates", "carbs_ingested", "carbs_used", "kcal_consumed", "calories_intake"},
			wantSemantics: []string{"calories_burned", "carbs_ingested_g", "carbs_used_g"},
		},
		{
			name:          "null-stripping when carb fields absent",
			rawActivity:   `{"id":"a2","name":"Easy Run","type":"Run","start_date_local":"2026-05-14T08:00:00","calories":400}`,
			wantIngestedG: nil,
			wantUsedG:     nil,
			wantAbsent:    []string{"carbs_ingested_g", "carbs_used_g", "carbs", "carbohydrates"},
		},
		{
			name:          "only carbs_ingested present",
			rawActivity:   `{"id":"a3","name":"Race","type":"Ride","start_date_local":"2026-05-14T09:00:00","calories":2100,"carbs_ingested":280}`,
			wantIngestedG: intPtr(280),
			wantUsedG:     nil,
			wantAbsent:    []string{"carbs_used_g", "carbs", "carbohydrates"},
			wantSemantics: []string{"carbs_ingested_g"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			client := newFakeActivitiesClient(t, []string{tc.rawActivity}, "metric")
			tool := newGetActivitiesToolWithGear(client, client, nil, nil, "test", "UTC", false)

			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-05-01"}`)})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			payload := resultMap(t, result)
			rows := payload["activities"].([]any)
			if len(rows) == 0 {
				t.Fatalf("activities is empty")
			}
			row := rows[0].(map[string]any)

			if tc.wantIngestedG != nil {
				if row["carbs_ingested_g"] != float64(*tc.wantIngestedG) {
					t.Fatalf("carbs_ingested_g = %v, want %d", row["carbs_ingested_g"], *tc.wantIngestedG)
				}
			}
			if tc.wantUsedG != nil {
				if row["carbs_used_g"] != float64(*tc.wantUsedG) {
					t.Fatalf("carbs_used_g = %v, want %d", row["carbs_used_g"], *tc.wantUsedG)
				}
			}
			for _, key := range tc.wantAbsent {
				if _, ok := row[key]; ok {
					t.Fatalf("activity row emitted disallowed/ambiguous key %s: %#v", key, row)
				}
			}
			if len(tc.wantSemantics) > 0 {
				semantics := payload["_meta"].(map[string]any)["field_semantics"].(map[string]any)
				for _, key := range tc.wantSemantics {
					if label, ok := semantics[key].(string); !ok || label == "" {
						t.Fatalf("field_semantics missing %s: %#v", key, semantics)
					}
				}
			}
		})
	}
}

func TestGetActivitiesResolvesGearNames(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, []string{
		`{"id":"a1","name":"Ride","type":"Ride","start_date_local":"2026-01-02T07:00:00","gear_id":"g-1"}`,
	}, "metric")
	client.gear = decodeToolGear(t, `{"id":"g-1","name":"Race Bike"}`)
	cache := newGearListCache()
	tool := newGetActivitiesToolWithGear(client, client, client, cache, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-01"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	row := resultMap(t, result)["activities"].([]any)[0].(map[string]any)
	if row["gear_id"] != "g-1" || row["gear_name"] != "Race Bike" || row["gear_resolution"] != gearResolutionResolved {
		t.Fatalf("row = %#v, want resolved gear", row)
	}
	if client.gearCalls != 1 {
		t.Fatalf("gear calls = %d, want one lookup", client.gearCalls)
	}
}

func TestGetActivitiesSkipsGearFetchWithoutGearIDs(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, []string{
		`{"id":"a1","name":"Run","type":"Run","start_date_local":"2026-01-02T07:00:00"}`,
	}, "metric")
	tool := newGetActivitiesToolWithGear(client, client, client, newGearListCache(), "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-01"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	row := resultMap(t, result)["activities"].([]any)[0].(map[string]any)
	if _, ok := row["gear_id"]; ok {
		t.Fatalf("row = %#v, want no gear fields", row)
	}
	if client.gearCalls != 0 {
		t.Fatalf("gear calls = %d, want no lookup", client.gearCalls)
	}
}

func TestGetActivitiesMarksUnknownUnnamedAndLookupUnavailableGear(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, []string{
		`{"id":"unknown","name":"Ride","type":"Ride","start_date_local":"2026-01-04T07:00:00","gear_id":"missing"}`,
		`{"id":"unnamed","name":"Run","type":"Run","start_date_local":"2026-01-03T07:00:00","gear_id":"shoe-1"}`,
	}, "metric")
	client.gear = decodeToolGear(t, `{"id":"shoe-1"}`)
	tool := newGetActivitiesToolWithGear(client, client, client, newGearListCache(), "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-01"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	rows := resultMap(t, result)["activities"].([]any)
	statuses := map[string]string{}
	for _, rawRow := range rows {
		row := rawRow.(map[string]any)
		statuses[row["activity_id"].(string)] = row["gear_resolution"].(string)
	}
	if statuses["unknown"] != gearResolutionUnresolved || statuses["unnamed"] != gearResolutionNameMissing {
		t.Fatalf("statuses = %#v, want unknown and name_missing", statuses)
	}

	client = newFakeActivitiesClient(t, []string{`{"id":"a1","name":"Ride","type":"Ride","start_date_local":"2026-01-02T07:00:00","gear_id":"g-1"}`}, "metric")
	client.gearErr = errors.New("gear upstream down")
	tool = newGetActivitiesToolWithGear(client, client, client, newGearListCache(), "test", "UTC", false)
	result, err = tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-01"}`)})
	if err != nil {
		t.Fatalf("lookup unavailable Handler() error = %v", err)
	}
	row := resultMap(t, result)["activities"].([]any)[0].(map[string]any)
	if row["gear_id"] != "g-1" || row["gear_resolution"] != gearResolutionLookupUnavailable {
		t.Fatalf("row = %#v, want lookup_unavailable with gear_id", row)
	}
}

func TestGetActivitiesPreservesGearLookupCancellation(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, []string{`{"id":"a1","name":"Ride","type":"Ride","start_date_local":"2026-01-02T07:00:00","gear_id":"g-1"}`}, "metric")
	client.gearErr = context.Canceled
	tool := newGetActivitiesToolWithGear(client, client, client, newGearListCache(), "test", "UTC", false)

	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-01"}`)})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Handler() error = %v, want context.Canceled", err)
	}
}

func TestGetActivitiesGearCacheReuseAndTargetIsolation(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, []string{`{"id":"a1","name":"Ride","type":"Ride","start_date_local":"2026-01-02T07:00:00","gear_id":"g-1"}`}, "metric")
	client.gear = decodeToolGear(t, `{"id":"g-1","name":"Race Bike"}`)
	cache := newGearListCache()
	gearTool := newGetGearListTool(client, cache, "test", false)
	activityTool := newGetActivitiesToolWithGear(client, client, client, cache, "test", "UTC", false)
	ctx111 := intervals.WithTargetAthleteID(context.Background(), "i111")

	if _, err := gearTool.Handler(ctx111, Request{Name: gearTool.Name, Arguments: json.RawMessage(`{}`)}); err != nil {
		t.Fatalf("gear Handler() error = %v", err)
	}
	if _, err := activityTool.Handler(ctx111, Request{Name: activityTool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-01"}`)}); err != nil {
		t.Fatalf("activity Handler() error = %v", err)
	}
	if client.gearCalls != 1 {
		t.Fatalf("gear calls = %d, want activity read to reuse get_gear_list cache", client.gearCalls)
	}

	client.gearByTarget = map[string][]intervals.Gear{
		"i111": decodeToolGear(t, `{"id":"g-1","name":"A Bike"}`),
		"i222": decodeToolGear(t, `{"id":"g-1","name":"B Bike"}`),
	}
	client.gear = nil
	cache = newGearListCache()
	activityTool = newGetActivitiesToolWithGear(client, client, client, cache, "test", "UTC", false)
	client.gearCalls = 0
	first, err := activityTool.Handler(ctx111, Request{Name: activityTool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-01"}`)})
	if err != nil {
		t.Fatalf("athlete 111 Handler() error = %v", err)
	}
	second, err := activityTool.Handler(intervals.WithTargetAthleteID(context.Background(), "i222"), Request{Name: activityTool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-01"}`)})
	if err != nil {
		t.Fatalf("athlete 222 Handler() error = %v", err)
	}
	firstName := resultMap(t, first)["activities"].([]any)[0].(map[string]any)["gear_name"]
	secondName := resultMap(t, second)["activities"].([]any)[0].(map[string]any)["gear_name"]
	if firstName != "A Bike" || secondName != "B Bike" || client.gearCalls != 2 {
		t.Fatalf("names/calls = %v/%v/%d, want target-isolated cache", firstName, secondName, client.gearCalls)
	}
}

func TestGetActivitiesPaginationFiltersAndTokenRoundTrip(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, []string{
		`{"id":"a3","name":"Tempo","type":"Run","start_date_local":"2026-01-03T07:00:00","distance":5000,"moving_time":1500}`,
		`{"id":"a2","name":"","type":"Run","start_date_local":"2026-01-03T07:00:00","distance":1000,"moving_time":300}`,
		`{"id":"a1","name":"Easy","type":"Run","start_date_local":"2026-01-02T07:00:00","distance":2000,"moving_time":600}`,
	}, "metric")
	tool := newGetActivitiesToolWithGear(client, client, nil, nil, "test", "UTC", false)

	first, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-01","newest":"2026-01-04","page_size":1}`)})
	if err != nil {
		t.Fatalf("first Handler() error = %v", err)
	}
	firstMap := resultMap(t, first)
	firstRows := firstMap["activities"].([]any)
	if got := firstRows[0].(map[string]any)["activity_id"]; got != "a3" {
		t.Fatalf("first activity_id = %v, want a3", got)
	}
	meta := firstMap["_meta"].(map[string]any)
	token, ok := meta["next_page_token"].(string)
	if !ok || token == "" {
		t.Fatalf("next_page_token = %#v, want non-empty", meta["next_page_token"])
	}
	if len(client.listCalls) == 0 || client.listCalls[0].Limit != 3 {
		t.Fatalf("ListActivities limit calls = %#v, want first limit 3", client.listCalls)
	}

	second, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"next_page_token":"` + token + `"}`)})
	if err != nil {
		t.Fatalf("second Handler() error = %v", err)
	}
	secondRows := resultMap(t, second)["activities"].([]any)
	if got := secondRows[0].(map[string]any)["activity_id"]; got != "a1" {
		t.Fatalf("second activity_id = %v, want a1 after unnamed same-timestamp row is filtered", got)
	}
}

func TestGetActivitiesBoundaryResponseShapeGoldenFixtures(t *testing.T) {
	t.Parallel()

	const exactFullWindowToken = `eyJ2IjoxLCJvbGRlc3QiOiIyMDI2LTAxLTAxIiwiaW5jbHVkZV91bm5hbWVkIjpmYWxzZSwiaW5jbHVkZV9mdWxsIjpmYWxzZSwicGFnZV9zaXplIjoxLCJmaWVsZHMiOlsiaWQiLCJuYW1lIiwidHlwZSIsInN1Yl90eXBlIiwic3RhcnRfZGF0ZV9sb2NhbCIsInN0YXJ0X2RhdGUiLCJ0aW1lem9uZSIsInNvdXJjZSIsIl9ub3RlIiwiaWN1X2F0aGxldGVfaWQiLCJleHRlcm5hbF9pZCIsInN0cmVhbV90eXBlcyIsImRpc3RhbmNlIiwiaWN1X2Rpc3RhbmNlIiwibW92aW5nX3RpbWUiLCJlbGFwc2VkX3RpbWUiLCJhdmVyYWdlX3NwZWVkIiwibWF4X3NwZWVkIiwidG90YWxfZWxldmF0aW9uX2dhaW4iLCJ0b3RhbF9lbGV2YXRpb25fbG9zcyIsImljdV90cmFpbmluZ19sb2FkIiwiYXZlcmFnZV9oZWFydHJhdGUiLCJtYXhfaGVhcnRyYXRlIiwiYXZlcmFnZV9jYWRlbmNlIiwiY2Fsb3JpZXMiLCJjYXJic19pbmdlc3RlZCIsImNhcmJzX3VzZWQiLCJkZXZpY2VfbmFtZSIsImdlYXJfaWQiXSwiYmVmb3JlX3N0YXJ0X2RhdGVfbG9jYWwiOiIyMDI2LTAxLTAzVDA3OjAwOjAwIiwiYmVmb3JlX2lkIjoiZjMiLCJza2lwX2lkc19hdF9ib3VuZGFyeSI6WyJmMyJdfQ`
	const identicalTimestampStallToken = `eyJ2IjoxLCJvbGRlc3QiOiIyMDI2LTAxLTAxIiwiaW5jbHVkZV91bm5hbWVkIjpmYWxzZSwiaW5jbHVkZV9mdWxsIjpmYWxzZSwicGFnZV9zaXplIjoxLCJmaWVsZHMiOlsiaWQiLCJuYW1lIiwidHlwZSIsInN1Yl90eXBlIiwic3RhcnRfZGF0ZV9sb2NhbCIsInN0YXJ0X2RhdGUiLCJ0aW1lem9uZSIsInNvdXJjZSIsIl9ub3RlIiwiaWN1X2F0aGxldGVfaWQiLCJleHRlcm5hbF9pZCIsInN0cmVhbV90eXBlcyIsImRpc3RhbmNlIiwiaWN1X2Rpc3RhbmNlIiwibW92aW5nX3RpbWUiLCJlbGFwc2VkX3RpbWUiLCJhdmVyYWdlX3NwZWVkIiwibWF4X3NwZWVkIiwidG90YWxfZWxldmF0aW9uX2dhaW4iLCJ0b3RhbF9lbGV2YXRpb25fbG9zcyIsImljdV90cmFpbmluZ19sb2FkIiwiYXZlcmFnZV9oZWFydHJhdGUiLCJtYXhfaGVhcnRyYXRlIiwiYXZlcmFnZV9jYWRlbmNlIiwiY2Fsb3JpZXMiLCJjYXJic19pbmdlc3RlZCIsImNhcmJzX3VzZWQiLCJkZXZpY2VfbmFtZSJdLCJiZWZvcmVfc3RhcnRfZGF0ZV9sb2NhbCI6IjIwMjYtMDEtMDNUMDc6MDA6MDAiLCJiZWZvcmVfaWQiOiJzMyIsInNraXBfaWRzX2F0X2JvdW5kYXJ5IjpbInMzIl19`

	tests := []struct {
		name            string
		rawActivities   []string
		arguments       string
		applyListParams bool
		wantIDs         []string
		wantPageSize    float64
		wantMore        bool
		wantToken       string
	}{
		{
			name:         "empty upstream page",
			arguments:    `{"oldest":"2026-01-01","page_size":2}`,
			wantPageSize: 2,
		},
		{
			name: "partial page",
			rawActivities: []string{
				`{"id":"p2","name":"Tempo","type":"Run","start_date_local":"2026-01-03T07:00:00","distance":1000,"moving_time":300}`,
				`{"id":"p1","name":"Easy","type":"Run","start_date_local":"2026-01-02T07:00:00","distance":1000,"moving_time":300}`,
			},
			arguments:    `{"oldest":"2026-01-01","page_size":3}`,
			wantIDs:      []string{"p2", "p1"},
			wantPageSize: 3,
		},
		{
			name: "exact full window",
			rawActivities: []string{
				`{"id":"f2","name":"Middle","type":"Run","start_date_local":"2026-01-02T07:00:00","distance":1000,"moving_time":300}`,
				`{"id":"f1","name":"Oldest","type":"Run","start_date_local":"2026-01-01T07:00:00","distance":1000,"moving_time":300}`,
				`{"id":"f3","name":"Newest","type":"Run","start_date_local":"2026-01-03T07:00:00","distance":1000,"moving_time":300}`,
			},
			arguments:    `{"oldest":"2026-01-01","page_size":1}`,
			wantIDs:      []string{"f3"},
			wantPageSize: 1,
			wantMore:     true,
			wantToken:    exactFullWindowToken,
		},
		{
			name: "identical-timestamp stall",
			rawActivities: []string{
				`{"id":"s3","name":"Named","type":"Run","start_date_local":"2026-01-03T07:00:00","distance":1000,"moving_time":300}`,
				`{"id":"s2","name":"","type":"Run","start_date_local":"2026-01-03T07:00:00","distance":1000,"moving_time":300}`,
				`{"id":"s1","name":"","type":"Run","start_date_local":"2026-01-03T07:00:00","distance":1000,"moving_time":300}`,
			},
			arguments:       fmt.Sprintf(`{"next_page_token":"%s"}`, identicalTimestampStallToken),
			applyListParams: true,
			wantPageSize:    1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := newFakeActivitiesClient(t, tc.rawActivities, "metric")
			client.applyListParams = tc.applyListParams
			tool := newGetActivitiesToolWithGear(client, client, nil, nil, "test", "UTC", false)

			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(tc.arguments)})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			resultMap := resultMap(t, result)
			rows := resultMap["activities"].([]any)
			if got := activityRowIDs(rows); !slices.Equal(got, tc.wantIDs) {
				t.Fatalf("activity IDs = %#v, want %#v", got, tc.wantIDs)
			}
			meta := resultMap["_meta"].(map[string]any)
			if meta["page_size"] != tc.wantPageSize || meta["include_full"] != false || meta["more_available"] != tc.wantMore {
				t.Fatalf("_meta = %#v, want page_size=%v include_full=false more_available=%v", meta, tc.wantPageSize, tc.wantMore)
			}
			token, hasToken := meta["next_page_token"].(string)
			if hasToken != tc.wantMore {
				t.Fatalf("next_page_token present = %v, want %v; _meta = %#v", hasToken, tc.wantMore, meta)
			}
			if token != tc.wantToken {
				t.Fatalf("next_page_token = %q, want byte-identical %q", token, tc.wantToken)
			}
		})
	}
}

func TestGetActivitiesStopsBeforeInvertedLowerBoundAfterFilteredBoundary(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, []string{
		`{"id":"u3","name":"","type":"Run","start_date_local":"2026-01-03T07:00:00","distance":1000,"moving_time":300}`,
		`{"id":"u2","name":"","type":"Run","start_date_local":"2026-01-03T07:00:00","distance":1000,"moving_time":300}`,
		`{"id":"u1","name":"","type":"Run","start_date_local":"2026-01-03T07:00:00","distance":1000,"moving_time":300}`,
	}, "metric")
	client.applyListParams = true
	client.rejectInvalidRange = true
	tool := newGetActivitiesToolWithGear(client, client, nil, nil, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-03T07:00:00","page_size":1}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	rows := resultMap(t, result)["activities"].([]any)
	if len(rows) != 0 {
		t.Fatalf("activities = %#v, want empty terminal page", rows)
	}
}

func TestGetActivitiesErrorsInsteadOfSkippingBeyondMaxSameTimestampWindow(t *testing.T) {
	t.Parallel()

	rawActivities := make([]string, 0, maxActivityFetchLimit+2)
	for i := maxActivityFetchLimit; i >= 1; i-- {
		rawActivities = append(rawActivities, fmt.Sprintf(`{"id":"u%03d","name":"","type":"Run","start_date_local":"2026-01-03T07:00:00","distance":1000,"moving_time":300}`, i))
	}
	rawActivities = append(rawActivities,
		`{"id":"named-same","name":"Same timestamp","type":"Run","start_date_local":"2026-01-03T07:00:00","distance":1000,"moving_time":300}`,
		`{"id":"older","name":"Older","type":"Run","start_date_local":"2026-01-02T07:00:00","distance":1000,"moving_time":300}`,
	)
	client := newFakeActivitiesClient(t, rawActivities, "metric")
	client.applyListParams = true
	tool := newGetActivitiesToolWithGear(client, client, nil, nil, "test", "UTC", false)

	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-01","page_size":1}`)})
	if message, ok := PublicErrorMessage(err); !ok || !strings.Contains(message, "same-timestamp filtered rows") {
		t.Fatalf("PublicErrorMessage = %q, %v; err = %v, want bounded pagination error", message, ok, err)
	}
}

func TestGetActivitiesWidensLookaheadBeforeCrossingSameTimestampBoundary(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, []string{
		`{"id":"u3","name":"","type":"Run","start_date_local":"2026-01-03T07:00:00","distance":1000,"moving_time":300}`,
		`{"id":"u2","name":"","type":"Run","start_date_local":"2026-01-03T07:00:00","distance":1000,"moving_time":300}`,
		`{"id":"u1","name":"","type":"Run","start_date_local":"2026-01-03T07:00:00","distance":1000,"moving_time":300}`,
		`{"id":"named-same","name":"Same timestamp","type":"Run","start_date_local":"2026-01-03T07:00:00","distance":1000,"moving_time":300}`,
		`{"id":"older","name":"Older","type":"Run","start_date_local":"2026-01-02T07:00:00","distance":1000,"moving_time":300}`,
	}, "metric")
	client.applyListParams = true
	tool := newGetActivitiesToolWithGear(client, client, nil, nil, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-01","page_size":1}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	rows := resultMap(t, result)["activities"].([]any)
	if len(rows) != 1 || rows[0].(map[string]any)["activity_id"] != "named-same" {
		t.Fatalf("activities = %#v, want eligible same-timestamp activity before older rows", rows)
	}
	if len(client.listCalls) < 3 || client.listCalls[2].Limit != maxActivityFetchLimit {
		t.Fatalf("list calls = %#v, want widened boundary lookahead", client.listCalls)
	}
}

func TestGetActivitiesAdvancesPastFullyFilteredSameTimestampWindow(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, []string{
		`{"id":"u3","name":"","type":"Run","start_date_local":"2026-01-03T07:00:00","distance":1000,"moving_time":300}`,
		`{"id":"u2","name":"","type":"Run","start_date_local":"2026-01-03T07:00:00","distance":1000,"moving_time":300}`,
		`{"id":"u1","name":"","type":"Run","start_date_local":"2026-01-03T07:00:00","distance":1000,"moving_time":300}`,
		`{"id":"named","name":"Older","type":"Run","start_date_local":"2026-01-02T07:00:00","distance":1000,"moving_time":300}`,
	}, "metric")
	client.applyListParams = true
	tool := newGetActivitiesToolWithGear(client, client, nil, nil, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-01","page_size":1}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	rows := resultMap(t, result)["activities"].([]any)
	if len(rows) != 1 || rows[0].(map[string]any)["activity_id"] != "named" {
		t.Fatalf("activities = %#v, want older named activity after filtered boundary", rows)
	}
	if len(client.listCalls) < 3 {
		t.Fatalf("list calls = %#v, want repeated boundary fetch then older fetch", client.listCalls)
	}
}

func TestGetActivitiesReturnsTokenWhenFilteredWindowsHitFetchCap(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, nil, "metric")
	for page := 0; page < maxActivityPageFetches; page++ {
		client.activityPages = append(client.activityPages, decodeActivityPage(t,
			fmtActivity(page, 3),
			fmtActivity(page, 2),
			fmtActivity(page, 1),
		))
	}
	client.activityPages = append(client.activityPages, decodeActivityPage(t,
		`{"id":"eligible","name":"Recovered","type":"Run","start_date_local":"2026-01-01T07:00:00","distance":1000,"moving_time":300}`,
	))
	tool := newGetActivitiesToolWithGear(client, client, nil, nil, "test", "UTC", false)

	first, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-01","page_size":1}`)})
	if err != nil {
		t.Fatalf("first Handler() error = %v", err)
	}
	firstMap := resultMap(t, first)
	if rows := firstMap["activities"].([]any); len(rows) != 0 {
		t.Fatalf("first activities = %#v, want empty filtered page", rows)
	}
	token, ok := firstMap["_meta"].(map[string]any)["next_page_token"].(string)
	if !ok || token == "" {
		t.Fatalf("next_page_token = %#v, want continuation after capped filtered windows", firstMap["_meta"])
	}
	if len(client.listCalls) != maxActivityPageFetches {
		t.Fatalf("first list calls = %d, want fetch cap %d", len(client.listCalls), maxActivityPageFetches)
	}

	second, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"next_page_token":"` + token + `"}`)})
	if err != nil {
		t.Fatalf("second Handler() error = %v", err)
	}
	rows := resultMap(t, second)["activities"].([]any)
	if got := rows[0].(map[string]any)["activity_id"]; got != "eligible" {
		t.Fatalf("second activity_id = %v, want eligible", got)
	}
}

func TestGetActivitiesPreservesProfileCancellation(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, nil, "metric")
	client.err = context.Canceled
	tool := newGetActivitiesToolWithGear(client, client, nil, nil, "test", "UTC", false)

	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-01"}`)})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Handler() error = %v, want context.Canceled", err)
	}
}

func TestGetActivitiesDoesNotTokenizeDateLessStravaStubsWithoutCursorProgress(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, []string{`{}`, `{}`}, "metric")
	tool := newGetActivitiesToolWithGear(client, client, nil, nil, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-01","page_size":1}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	resultMap := resultMap(t, result)
	if rows := resultMap["activities"].([]any); len(rows) != 1 {
		t.Fatalf("activities = %#v, want first Strava stub only", rows)
	}
	if token, ok := resultMap["_meta"].(map[string]any)["next_page_token"]; ok {
		t.Fatalf("next_page_token = %#v, want no continuation without date/id cursor progress", token)
	}
}

func TestGetActivitiesUsesProfileTimezoneFallback(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, []string{
		`{"id":"run1","name":"Run","type":"Run","start_date_local":"2026-01-03T07:00:00","distance":1000,"moving_time":300}`,
	}, "metric")
	client.profile.Timezone = "America/Sao_Paulo"
	tool := newGetActivitiesToolWithGear(client, client, nil, nil, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-01"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	row := resultMap(t, result)["activities"].([]any)[0].(map[string]any)
	if got := row["timezone"]; got != "America/Sao_Paulo" {
		t.Fatalf("timezone = %v, want athlete profile timezone fallback", got)
	}
}

func TestGetActivitiesDoesNotLoopOnSameTimestampFilteredLookahead(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, []string{
		`{"id":"a3","name":"Named","type":"Run","start_date_local":"2026-01-03T07:00:00","distance":1000,"moving_time":300}`,
		`{"id":"a2","name":"","type":"Run","start_date_local":"2026-01-03T07:00:00","distance":1000,"moving_time":300}`,
		`{"id":"a1","name":"","type":"Run","start_date_local":"2026-01-03T07:00:00","distance":1000,"moving_time":300}`,
	}, "metric")
	client.applyListParams = true
	tool := newGetActivitiesToolWithGear(client, client, nil, nil, "test", "UTC", false)

	first, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-01","page_size":1}`)})
	if err != nil {
		t.Fatalf("first Handler() error = %v", err)
	}
	token := resultMap(t, first)["_meta"].(map[string]any)["next_page_token"].(string)
	second, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"next_page_token":"` + token + `"}`)})
	if err != nil {
		t.Fatalf("second Handler() error = %v", err)
	}
	secondMap := resultMap(t, second)
	if rows := secondMap["activities"].([]any); len(rows) != 0 {
		t.Fatalf("second activities = %#v, want empty terminal page", rows)
	}
	if token, ok := secondMap["_meta"].(map[string]any)["next_page_token"]; ok {
		t.Fatalf("second next_page_token = %#v, want no continuation without cursor progress", token)
	}
}

func TestGetActivitiesDetectsDocumentedStravaStubShapes(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, []string{
		`{"id":"stub-wahoo","icu_athlete_id":"i12345","start_date_local":"2026-01-03T07:00:00","external_id":"wahoo:12345"}`,
		`{}`,
		`{"id":"stub1","icu_athlete_id":"i12345","start_date_local":"2026-01-02T07:00:00"}`,
		`{"id":"stub2","icu_athlete_id":"i12345","start_date_local":"2026-01-01T07:00:00","name":null}`,
	}, "metric")
	tool := newGetActivitiesToolWithGear(client, client, nil, nil, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-01"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	rows := resultMap(t, result)["activities"].([]any)
	if len(rows) != 4 {
		t.Fatalf("activities = %#v, want all documented Strava stubs surfaced", rows)
	}
	wantWorkarounds := map[string]string{
		"stub-wahoo": wantWahooStravaWorkaround,
		"":           wantUnknownStravaWorkaround,
		"stub1":      wantUnknownStravaWorkaround,
		"stub2":      wantUnknownStravaWorkaround,
	}
	for _, rawRow := range rows {
		row := rawRow.(map[string]any)
		activityID, _ := row["activity_id"].(string)
		assertUnavailableReasonAndWorkaround(t, row, "strava_blocked", wantWorkarounds[activityID])
	}
}

func TestGetActivitiesKeepsStravaRowsWhenUnnamedFilteringIsDefault(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, []string{
		`{"id":"hidden1","source":"Strava","_note":"Strava activity hidden","start_date_local":"2026-01-02T07:00:00","name":null}`,
		`{"id":"unnamed","name":"","type":"Run","start_date_local":"2026-01-01T07:00:00","distance":1000,"moving_time":300}`,
	}, "metric")
	tool := newGetActivitiesToolWithGear(client, client, nil, nil, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-01"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	rows := resultMap(t, result)["activities"].([]any)
	if len(rows) != 1 {
		t.Fatalf("activities = %#v, want only Strava unavailable row", rows)
	}
	row := rows[0].(map[string]any)
	if row["activity_id"] != "hidden1" {
		t.Fatalf("row = %#v, want visible Strava unavailable marker", row)
	}
	assertUnavailableReasonAndWorkaround(t, row, "strava_blocked", wantUnknownStravaWorkaround)
}

func TestGetActivitiesDoesNotEmitPaceForCycling(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, []string{
		`{"id":"ride1","name":"Ride","type":"Ride","start_date_local":"2026-01-03T07:00:00","distance":20000,"moving_time":2400,"average_speed":8.333333}`,
	}, "metric")
	tool := newGetActivitiesToolWithGear(client, client, nil, nil, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-01"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	row := resultMap(t, result)["activities"].([]any)[0].(map[string]any)
	if _, ok := row["distance_km"]; !ok {
		t.Fatalf("row = %#v, want converted distance", row)
	}
	if _, ok := row["average_speed_kmh"]; !ok {
		t.Fatalf("row = %#v, want converted average speed", row)
	}
	if _, ok := row["pace_seconds_per_km"]; ok {
		t.Fatalf("row = %#v, want no run pace for cycling", row)
	}
}

func TestGetActivitiesShapesStravaFullAndUnits(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, []string{
		`{"id":"run1","name":"Run","type":"Run","start_date_local":"2026-01-03T07:00:00","distance":1609.344,"moving_time":480,"average_speed":3.3528,"name_null":null}`,
		`{"id":"hidden1","source":"Strava","_note":"Strava activity hidden","start_date_local":"2026-01-02T07:00:00","name":null}`,
	}, "imperial")
	tool := newGetActivitiesToolWithGear(client, client, nil, nil, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-01","include_unnamed":true,"include_full":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	rows := resultMap(t, result)["activities"].([]any)
	run := rows[0].(map[string]any)
	if _, ok := run["distance_mi"]; !ok {
		t.Fatalf("run row keys = %#v, want distance_mi", run)
	}
	if _, ok := run["pace_seconds_per_mile"]; !ok {
		t.Fatalf("run row keys = %#v, want pace_seconds_per_mile", run)
	}
	full := run["full"].(map[string]any)
	if value, ok := full["name_null"]; !ok || value != nil {
		t.Fatalf("full name_null = %#v present %v, want preserved nil", value, ok)
	}
	hidden := rows[1].(map[string]any)
	assertUnavailableReasonAndWorkaround(t, hidden, "strava_blocked", wantUnknownStravaWorkaround)
}

func newFakeActivitiesClient(t *testing.T, rawActivities []string, preferredUnits string) *fakeActivitiesProfileClient {
	t.Helper()
	client := &fakeActivitiesProfileClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: preferredUnits, Timezone: "UTC"}}}
	for _, raw := range rawActivities {
		var activity intervals.Activity
		if err := json.Unmarshal([]byte(raw), &activity); err != nil {
			t.Fatalf("decode activity fixture: %v", err)
		}
		client.activities = append(client.activities, activity)
	}
	return client
}

func decodeActivityPage(t *testing.T, rawActivities ...string) []intervals.Activity {
	t.Helper()
	activities := make([]intervals.Activity, 0, len(rawActivities))
	for _, raw := range rawActivities {
		var activity intervals.Activity
		if err := json.Unmarshal([]byte(raw), &activity); err != nil {
			t.Fatalf("decode activity page fixture: %v", err)
		}
		activities = append(activities, activity)
	}
	return activities
}

func fmtActivity(page int, ordinal int) string {
	return fmt.Sprintf(`{"id":"filtered-%02d-%d","name":"","type":"Run","start_date_local":"2026-01-%02dT07:0%d:00","distance":1000,"moving_time":300}`, page, ordinal, 10-page, ordinal)
}

func findTool(t *testing.T, tools []Tool, name string) Tool {
	t.Helper()
	for _, tool := range tools {
		if tool.Name == name {
			return tool
		}
	}
	t.Fatalf("tool %s not found in %#v", name, tools)
	return Tool{}
}

func activityIDs(activities []intervals.Activity) []string {
	ids := make([]string, 0, len(activities))
	for _, activity := range activities {
		ids = append(ids, activity.ID)
	}
	return ids
}

func activityRowIDs(rows []any) []string {
	ids := make([]string, 0, len(rows))
	for _, rawRow := range rows {
		row := rawRow.(map[string]any)
		id, _ := row["activity_id"].(string)
		ids = append(ids, id)
	}
	return ids
}

func resultMap(t *testing.T, result Result) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal([]byte(resultText(t, result)), &out); err != nil {
		t.Fatalf("decode result: %v", err)
	}
	return out
}

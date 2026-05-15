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
	registrar := &collectingRegistrar{}
	if err := NewRegistry(client, "test", "UTC").Register(context.Background(), registrar); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if len(registrar.tools) != 3 {
		t.Fatalf("registered tool count = %d, want profile + activities + advanced capabilities", len(registrar.tools))
	}
	tool := findTool(t, registrar.tools, getActivitiesName)
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

func TestGetActivitiesPaginationFiltersAndTokenRoundTrip(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, []string{
		`{"id":"a3","name":"Tempo","type":"Run","start_date_local":"2026-01-03T07:00:00","distance":5000,"moving_time":1500}`,
		`{"id":"a2","name":"","type":"Run","start_date_local":"2026-01-03T07:00:00","distance":1000,"moving_time":300}`,
		`{"id":"a1","name":"Easy","type":"Run","start_date_local":"2026-01-02T07:00:00","distance":2000,"moving_time":600}`,
	}, "metric")
	tool := newGetActivitiesTool(client, client, "test", "UTC", false)

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

func TestFetchActivitiesPageBoundaryGoldenFixtures(t *testing.T) {
	t.Parallel()

	const exactFullWindowToken = `eyJ2IjoxLCJvbGRlc3QiOiIyMDI2LTAxLTAxIiwiaW5jbHVkZV91bm5hbWVkIjpmYWxzZSwiaW5jbHVkZV9mdWxsIjpmYWxzZSwicGFnZV9zaXplIjoxLCJmaWVsZHMiOlsiaWQiLCJuYW1lIiwidHlwZSIsInN1Yl90eXBlIiwic3RhcnRfZGF0ZV9sb2NhbCIsInN0YXJ0X2RhdGUiLCJ0aW1lem9uZSIsInNvdXJjZSIsIl9ub3RlIiwiaWN1X2F0aGxldGVfaWQiLCJleHRlcm5hbF9pZCIsInN0cmVhbV90eXBlcyIsImRpc3RhbmNlIiwiaWN1X2Rpc3RhbmNlIiwibW92aW5nX3RpbWUiLCJlbGFwc2VkX3RpbWUiLCJhdmVyYWdlX3NwZWVkIiwibWF4X3NwZWVkIiwidG90YWxfZWxldmF0aW9uX2dhaW4iLCJ0b3RhbF9lbGV2YXRpb25fbG9zcyIsImljdV90cmFpbmluZ19sb2FkIiwiYXZlcmFnZV9oZWFydHJhdGUiLCJtYXhfaGVhcnRyYXRlIiwiYXZlcmFnZV9jYWRlbmNlIiwiY2Fsb3JpZXMiLCJkZXZpY2VfbmFtZSJdLCJiZWZvcmVfc3RhcnRfZGF0ZV9sb2NhbCI6IjIwMjYtMDEtMDNUMDc6MDA6MDAiLCJiZWZvcmVfaWQiOiJmMyIsInNraXBfaWRzX2F0X2JvdW5kYXJ5IjpbImYzIl19`
	const identicalTimestampStallToken = `eyJ2IjoxLCJvbGRlc3QiOiIyMDI2LTAxLTAxIiwiaW5jbHVkZV91bm5hbWVkIjpmYWxzZSwiaW5jbHVkZV9mdWxsIjpmYWxzZSwicGFnZV9zaXplIjoxLCJmaWVsZHMiOlsiaWQiLCJuYW1lIiwidHlwZSIsInN1Yl90eXBlIiwic3RhcnRfZGF0ZV9sb2NhbCIsInN0YXJ0X2RhdGUiLCJ0aW1lem9uZSIsInNvdXJjZSIsIl9ub3RlIiwiaWN1X2F0aGxldGVfaWQiLCJleHRlcm5hbF9pZCIsInN0cmVhbV90eXBlcyIsImRpc3RhbmNlIiwiaWN1X2Rpc3RhbmNlIiwibW92aW5nX3RpbWUiLCJlbGFwc2VkX3RpbWUiLCJhdmVyYWdlX3NwZWVkIiwibWF4X3NwZWVkIiwidG90YWxfZWxldmF0aW9uX2dhaW4iLCJ0b3RhbF9lbGV2YXRpb25fbG9zcyIsImljdV90cmFpbmluZ19sb2FkIiwiYXZlcmFnZV9oZWFydHJhdGUiLCJtYXhfaGVhcnRyYXRlIiwiYXZlcmFnZV9jYWRlbmNlIiwiY2Fsb3JpZXMiLCJkZXZpY2VfbmFtZSJdLCJiZWZvcmVfc3RhcnRfZGF0ZV9sb2NhbCI6IjIwMjYtMDEtMDNUMDc6MDA6MDAiLCJiZWZvcmVfaWQiOiJzMyIsInNraXBfaWRzX2F0X2JvdW5kYXJ5IjpbInMzIl19`

	tests := []struct {
		name            string
		rawActivities   []string
		args            GetActivitiesRequest
		nextPageToken   string
		applyListParams bool
		wantIDs         []string
		wantToken       string
	}{
		{
			name: "empty upstream page",
			args: GetActivitiesRequest{Oldest: "2026-01-01", PageSize: 2},
		},
		{
			name: "partial page",
			rawActivities: []string{
				`{"id":"p2","name":"Tempo","type":"Run","start_date_local":"2026-01-03T07:00:00","distance":1000,"moving_time":300}`,
				`{"id":"p1","name":"Easy","type":"Run","start_date_local":"2026-01-02T07:00:00","distance":1000,"moving_time":300}`,
			},
			args:    GetActivitiesRequest{Oldest: "2026-01-01", PageSize: 3},
			wantIDs: []string{"p2", "p1"},
		},
		{
			name: "exact full window",
			rawActivities: []string{
				`{"id":"f2","name":"Middle","type":"Run","start_date_local":"2026-01-02T07:00:00","distance":1000,"moving_time":300}`,
				`{"id":"f1","name":"Oldest","type":"Run","start_date_local":"2026-01-01T07:00:00","distance":1000,"moving_time":300}`,
				`{"id":"f3","name":"Newest","type":"Run","start_date_local":"2026-01-03T07:00:00","distance":1000,"moving_time":300}`,
			},
			args:      GetActivitiesRequest{Oldest: "2026-01-01", PageSize: 1},
			wantIDs:   []string{"f3"},
			wantToken: exactFullWindowToken,
		},
		{
			name: "identical-timestamp stall",
			rawActivities: []string{
				`{"id":"s3","name":"Named","type":"Run","start_date_local":"2026-01-03T07:00:00","distance":1000,"moving_time":300}`,
				`{"id":"s2","name":"","type":"Run","start_date_local":"2026-01-03T07:00:00","distance":1000,"moving_time":300}`,
				`{"id":"s1","name":"","type":"Run","start_date_local":"2026-01-03T07:00:00","distance":1000,"moving_time":300}`,
			},
			args:            GetActivitiesRequest{Oldest: "2026-01-01", PageSize: 1},
			nextPageToken:   identicalTimestampStallToken,
			applyListParams: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := newFakeActivitiesClient(t, tc.rawActivities, "metric")
			client.applyListParams = tc.applyListParams
			var token *activitiesPageToken
			if tc.nextPageToken != "" {
				parsed, err := parseActivitiesPageToken(tc.nextPageToken)
				if err != nil {
					t.Fatalf("parseActivitiesPageToken() error = %v", err)
				}
				token = parsed
			}

			activities, nextToken, err := fetchActivitiesPage(context.Background(), client, tc.args, token)
			if err != nil {
				t.Fatalf("fetchActivitiesPage() error = %v", err)
			}
			if got := activityIDs(activities); !slices.Equal(got, tc.wantIDs) {
				t.Fatalf("activity IDs = %#v, want %#v", got, tc.wantIDs)
			}
			if nextToken != tc.wantToken {
				t.Fatalf("next token = %q, want %q", nextToken, tc.wantToken)
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
	tool := newGetActivitiesTool(client, client, "test", "UTC", false)

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
	tool := newGetActivitiesTool(client, client, "test", "UTC", false)

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
	tool := newGetActivitiesTool(client, client, "test", "UTC", false)

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
	tool := newGetActivitiesTool(client, client, "test", "UTC", false)

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
	tool := newGetActivitiesTool(client, client, "test", "UTC", false)

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
	tool := newGetActivitiesTool(client, client, "test", "UTC", false)

	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-01"}`)})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Handler() error = %v, want context.Canceled", err)
	}
}

func TestGetActivitiesRejectsMismatchedToken(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, []string{
		`{"id":"a2","name":"Tempo","start_date_local":"2026-01-03T07:00:00"}`,
		`{"id":"a1","name":"Easy","start_date_local":"2026-01-02T07:00:00"}`,
	}, "metric")
	tool := newGetActivitiesTool(client, client, "test", "UTC", false)
	first, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-01","page_size":1}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	token := resultMap(t, first)["_meta"].(map[string]any)["next_page_token"].(string)
	_, err = tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-02-01","next_page_token":"` + token + `"}`)})
	if err == nil {
		t.Fatal("Handler() error = nil, want mismatched token error")
	}
	if message, ok := PublicErrorMessage(err); !ok || !strings.Contains(message, "invalid get_activities arguments") {
		t.Fatalf("PublicErrorMessage = %q, %v; want invalid arguments", message, ok)
	}
}

func TestGetActivitiesDoesNotTokenizeDateLessStravaStubsWithoutCursorProgress(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, []string{`{}`, `{}`}, "metric")
	tool := newGetActivitiesTool(client, client, "test", "UTC", false)

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
	tool := newGetActivitiesTool(client, client, "test", "UTC", false)

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
	tool := newGetActivitiesTool(client, client, "test", "UTC", false)

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
		`{}`,
		`{"id":"stub1","icu_athlete_id":"i12345","start_date_local":"2026-01-02T07:00:00"}`,
		`{"id":"stub2","icu_athlete_id":"i12345","start_date_local":"2026-01-01T07:00:00","name":null}`,
	}, "metric")
	tool := newGetActivitiesTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-01"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	rows := resultMap(t, result)["activities"].([]any)
	if len(rows) != 3 {
		t.Fatalf("activities = %#v, want all documented Strava stubs surfaced", rows)
	}
	for _, rawRow := range rows {
		row := rawRow.(map[string]any)
		if row["unavailable"].(map[string]any)["reason"] != "strava_tos" {
			t.Fatalf("row = %#v, want Strava unavailable marker", row)
		}
	}
}

func TestGetActivitiesKeepsStravaRowsWhenUnnamedFilteringIsDefault(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, []string{
		`{"id":"hidden1","source":"Strava","_note":"Strava activity hidden","start_date_local":"2026-01-02T07:00:00","name":null}`,
		`{"id":"unnamed","name":"","type":"Run","start_date_local":"2026-01-01T07:00:00","distance":1000,"moving_time":300}`,
	}, "metric")
	tool := newGetActivitiesTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-01"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	rows := resultMap(t, result)["activities"].([]any)
	if len(rows) != 1 {
		t.Fatalf("activities = %#v, want only Strava unavailable row", rows)
	}
	row := rows[0].(map[string]any)
	if row["activity_id"] != "hidden1" || row["unavailable"].(map[string]any)["reason"] != "strava_tos" {
		t.Fatalf("row = %#v, want visible Strava unavailable marker", row)
	}
}

func TestGetActivitiesDoesNotEmitPaceForCycling(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, []string{
		`{"id":"ride1","name":"Ride","type":"Ride","start_date_local":"2026-01-03T07:00:00","distance":20000,"moving_time":2400,"average_speed":8.333333}`,
	}, "metric")
	tool := newGetActivitiesTool(client, client, "test", "UTC", false)

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
	tool := newGetActivitiesTool(client, client, "test", "UTC", false)

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
	unavailable := hidden["unavailable"].(map[string]any)
	if unavailable["reason"] != "strava_tos" || hidden["strava_imported"] != true {
		t.Fatalf("hidden row = %#v, want Strava unavailable", hidden)
	}
}

func newFakeActivitiesClient(t *testing.T, rawActivities []string, preferredUnits string) *fakeActivitiesProfileClient {
	t.Helper()
	client := &fakeActivitiesProfileClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "12345", PreferredUnits: preferredUnits, Timezone: "UTC"}}}
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

func resultMap(t *testing.T, result Result) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal([]byte(resultText(t, result)), &out); err != nil {
		t.Fatalf("decode result: %v", err)
	}
	return out
}

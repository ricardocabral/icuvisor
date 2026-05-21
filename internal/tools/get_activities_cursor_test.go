package tools

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestFetchActivitiesPageBoundaryGoldenFixtures(t *testing.T) {
	t.Parallel()

	const exactFullWindowToken = `eyJ2IjoxLCJvbGRlc3QiOiIyMDI2LTAxLTAxIiwiaW5jbHVkZV91bm5hbWVkIjpmYWxzZSwiaW5jbHVkZV9mdWxsIjpmYWxzZSwicGFnZV9zaXplIjoxLCJmaWVsZHMiOlsiaWQiLCJuYW1lIiwidHlwZSIsInN1Yl90eXBlIiwic3RhcnRfZGF0ZV9sb2NhbCIsInN0YXJ0X2RhdGUiLCJ0aW1lem9uZSIsInNvdXJjZSIsIl9ub3RlIiwiaWN1X2F0aGxldGVfaWQiLCJleHRlcm5hbF9pZCIsInN0cmVhbV90eXBlcyIsImRpc3RhbmNlIiwiaWN1X2Rpc3RhbmNlIiwibW92aW5nX3RpbWUiLCJlbGFwc2VkX3RpbWUiLCJhdmVyYWdlX3NwZWVkIiwibWF4X3NwZWVkIiwidG90YWxfZWxldmF0aW9uX2dhaW4iLCJ0b3RhbF9lbGV2YXRpb25fbG9zcyIsImljdV90cmFpbmluZ19sb2FkIiwiYXZlcmFnZV9oZWFydHJhdGUiLCJtYXhfaGVhcnRyYXRlIiwiYXZlcmFnZV9jYWRlbmNlIiwiY2Fsb3JpZXMiLCJkZXZpY2VfbmFtZSIsImdlYXJfaWQiXSwiYmVmb3JlX3N0YXJ0X2RhdGVfbG9jYWwiOiIyMDI2LTAxLTAzVDA3OjAwOjAwIiwiYmVmb3JlX2lkIjoiZjMiLCJza2lwX2lkc19hdF9ib3VuZGFyeSI6WyJmMyJdfQ`
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

			activities, nextToken, err := fetchActivitiesPage(context.Background(), client, tc.args, token, "")
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

func TestFetchActivitiesPageTokenBindsAthlete(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, []string{
		`{"id":"a2","name":"Second","type":"Run","start_date_local":"2026-01-02T07:00:00","distance":1000,"moving_time":300}`,
		`{"id":"a1","name":"First","type":"Run","start_date_local":"2026-01-01T07:00:00","distance":1000,"moving_time":300}`,
	}, "metric")
	_, nextToken, err := fetchActivitiesPage(context.Background(), client, GetActivitiesRequest{Oldest: "2026-01-01", PageSize: 1}, nil, "i222")
	if err != nil {
		t.Fatalf("fetchActivitiesPage() error = %v", err)
	}
	if nextToken == "" {
		t.Fatal("next token = empty, want token")
	}
	parsed, err := parseActivitiesPageToken(nextToken)
	if err != nil {
		t.Fatalf("parseActivitiesPageToken() error = %v", err)
	}
	if parsed.AthleteID != "i222" {
		t.Fatalf("token athlete_id = %q, want i222", parsed.AthleteID)
	}
}

func TestGetActivitiesRejectsMismatchedToken(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, []string{
		`{"id":"a2","name":"Tempo","start_date_local":"2026-01-03T07:00:00"}`,
		`{"id":"a1","name":"Easy","start_date_local":"2026-01-02T07:00:00"}`,
	}, "metric")
	tool := newGetActivitiesToolWithGear(client, client, nil, nil, "test", "UTC", false)
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

func TestGetActivitiesUnsupportedTokenVersionIsInvalidInput(t *testing.T) {
	t.Parallel()

	token, err := encodeActivitiesPageToken(activitiesPageToken{Version: 2, Oldest: "2026-01-01", PageSize: 10})
	if err != nil {
		t.Fatalf("encodeActivitiesPageToken() error = %v", err)
	}
	_, _, err = decodeGetActivitiesRequest(json.RawMessage(`{"next_page_token":"` + token + `"}`))
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("decodeGetActivitiesRequest() error = %v, want ErrInvalidInput", err)
	}
}

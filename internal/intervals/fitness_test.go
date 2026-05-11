package intervals

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFitnessMetricClientEndpoints(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		wantPath  string
		wantQuery string
		body      string
		call      func(context.Context, *Client) error
	}{
		{
			name:      "athlete summary",
			wantPath:  "/athlete/i12345/athlete-summary.json",
			wantQuery: "end=2026-05-07&start=2026-05-01",
			body:      `[{"date":"2026-05-01","fitness":70,"fatigue":80,"form":-10,"timeInZones":[10,20],"byCategory":[{"category":"Ride","time":3600}]}]`,
			call: func(ctx context.Context, client *Client) error {
				rows, err := client.ListAthleteSummary(ctx, AthleteSummaryParams{Start: "2026-05-01", End: "2026-05-07"})
				if err != nil {
					return err
				}
				if len(rows) != 1 || rows[0].Fitness != 70 || len(rows[0].ByCategory) != 1 {
					t.Fatalf("summary rows = %+v", rows)
				}
				return nil
			},
		},
		{
			name:      "athlete power curves",
			wantPath:  "/athlete/i12345/power-curves.json",
			wantQuery: "curves=r.2026-05-01.2026-05-07&secs=60%2C300&type=Ride",
			body:      `{"list":[{"id":"r","secs":[60,300],"values":[320,260],"activity_id":["a1","a2"]}],"activities":{}}`,
			call: func(ctx context.Context, client *Client) error {
				set, err := client.ListAthletePowerCurves(ctx, CurveParams{Sport: "Ride", CurveSpec: "r.2026-05-01.2026-05-07", DurationSeconds: []int{60, 300}})
				if err != nil {
					return err
				}
				if len(set.List) != 1 || len(set.List[0].Values) != 2 || set.List[0].ActivityID[1] != "a2" {
					t.Fatalf("curve set = %+v", set)
				}
				return nil
			},
		},
		{
			name:      "activity power vs hr",
			wantPath:  "/activity/a1/power-vs-hr.json",
			wantQuery: "",
			body:      `{"powerHr":1.2,"decoupling":4.5}`,
			call: func(ctx context.Context, client *Client) error {
				got, err := client.GetActivityPowerVsHR(ctx, "a1")
				if err != nil {
					return err
				}
				if got.PowerHR == nil || *got.PowerHR != 1.2 || got.Decoupling == nil || *got.Decoupling != 4.5 {
					t.Fatalf("power-vs-hr = %+v", got)
				}
				return nil
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if got := r.URL.Path; got != tc.wantPath {
					t.Fatalf("path = %q, want %q", got, tc.wantPath)
				}
				if got := r.URL.RawQuery; got != tc.wantQuery {
					t.Fatalf("query = %q, want %q", got, tc.wantQuery)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()
			client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
			if err := tc.call(context.Background(), client); err != nil {
				t.Fatalf("call error = %v", err)
			}
		})
	}
}

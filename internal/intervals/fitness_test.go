package intervals

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
)

func TestFitnessMetricClientEndpoints(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		wantPath   string
		wantQuery  string
		wantParams map[string]string
		wantAbsent []string
		body       string
		call       func(context.Context, *Client) error
	}{
		{
			name:      "athlete summary",
			wantPath:  "/athlete/i12345/athlete-summary.json",
			wantQuery: "end=2026-05-07&start=2026-05-01",
			wantParams: map[string]string{
				"start": "2026-05-01",
				"end":   "2026-05-07",
			},
			wantAbsent: []string{"oldest", "newest"},
			body:       `[{"athlete_id":"i12345","date":"2026-05-01","fitness":70,"fatigue":80,"form":-10,"timeInZones":[10,20],"byCategory":[{"category":"Ride","time":3600}]}]`,
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
			name:      "athlete power curves with durability curve specs",
			wantPath:  "/athlete/i12345/power-curves.json",
			wantQuery: "curves=r.2026-05-01.2026-05-07%2Cr.2026-05-01.2026-05-07-kj0&secs=60%2C300&type=Ride",
			wantParams: map[string]string{
				"type":   "Ride",
				"curves": "r.2026-05-01.2026-05-07,r.2026-05-01.2026-05-07-kj0",
				"secs":   "60,300",
			},
			wantAbsent: []string{"sport", "curve", "duration_seconds", "distances"},
			body:       `{"list":[{"id":"r","secs":[60,300],"values":[320,260],"activity_id":["a1","a2"]},{"id":"r-kj0","after_kj":1000,"secs":[60,300],"values":[300,240],"activity_id":["a3","a4"]}],"activities":{}}`,
			call: func(ctx context.Context, client *Client) error {
				set, err := client.ListAthletePowerCurves(ctx, CurveParams{Sport: "Ride", CurveSpec: "r.2026-05-01.2026-05-07,r.2026-05-01.2026-05-07-kj0", DurationSeconds: []int{60, 300}})
				if err != nil {
					return err
				}
				if len(set.List) != 2 || len(set.List[0].Values) != 2 || set.List[0].ActivityID[1] != "a2" {
					t.Fatalf("curve set = %+v", set)
				}
				if set.List[1].AfterKJ == nil || *set.List[1].AfterKJ != 1000 || set.List[1].Values[1] != 240 || set.List[1].ActivityID[1] != "a4" {
					t.Fatalf("durability curve = %+v", set.List[1])
				}
				return nil
			},
		},
		{
			name:      "athlete hr curves with sport and secs",
			wantPath:  "/athlete/i12345/hr-curves.json",
			wantQuery: "curves=r.2026-05-01.2026-05-07&secs=60%2C300&type=Run",
			wantParams: map[string]string{
				"type":   "Run",
				"curves": "r.2026-05-01.2026-05-07",
				"secs":   "60,300",
			},
			wantAbsent: []string{"sport", "curve", "duration_seconds", "distances"},
			body:       `{"list":[{"id":"r","secs":[60,300],"values":[178,165],"activity_id":["a1","a2"]}],"activities":{}}`,
			call: func(ctx context.Context, client *Client) error {
				set, err := client.ListAthleteHRCurves(ctx, CurveParams{Sport: "Run", CurveSpec: "r.2026-05-01.2026-05-07", DurationSeconds: []int{60, 300}})
				if err != nil {
					return err
				}
				if len(set.List) != 1 || len(set.List[0].Secs) != 2 || set.List[0].Values[0] != 178 {
					t.Fatalf("hr curve set = %+v", set)
				}
				return nil
			},
		},
		{
			name:      "athlete pace curves with sport and distances",
			wantPath:  "/athlete/i12345/pace-curves.json",
			wantQuery: "curves=r.2026-05-01.2026-05-07&distances=1000%2C5000&type=Run",
			wantParams: map[string]string{
				"type":      "Run",
				"curves":    "r.2026-05-01.2026-05-07",
				"distances": "1000,5000",
			},
			wantAbsent: []string{"sport", "curve", "distance_meters", "secs"},
			body:       `{"list":[{"id":"r","distance":[1000,5000],"values":[240,1500],"activity_id":["a1","a2"]}],"activities":{}}`,
			call: func(ctx context.Context, client *Client) error {
				set, err := client.ListAthletePaceCurves(ctx, CurveParams{Sport: "Run", CurveSpec: "r.2026-05-01.2026-05-07", DistanceMeters: []int{1000, 5000}})
				if err != nil {
					return err
				}
				if len(set.List) != 1 || len(set.List[0].Distance) != 2 || set.List[0].Values[1] != 1500 {
					t.Fatalf("pace curve set = %+v", set)
				}
				return nil
			},
		},
		{
			name:      "athlete hr curves permit omitted sport",
			wantPath:  "/athlete/i12345/hr-curves.json",
			wantQuery: "curves=r.2026-05-01.2026-05-07&secs=60",
			wantParams: map[string]string{
				"curves": "r.2026-05-01.2026-05-07",
				"secs":   "60",
			},
			wantAbsent: []string{"type", "sport", "distances"},
			body:       `{"list":[{"id":"r","secs":[60],"values":[178],"activity_id":["a1"]}],"activities":{}}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.ListAthleteHRCurves(ctx, CurveParams{CurveSpec: "r.2026-05-01.2026-05-07", DurationSeconds: []int{60}})
				return err
			},
		},
		{
			name:      "athlete pace curves permit omitted sport",
			wantPath:  "/athlete/i12345/pace-curves.json",
			wantQuery: "curves=r.2026-05-01.2026-05-07&distances=1000",
			wantParams: map[string]string{
				"curves":    "r.2026-05-01.2026-05-07",
				"distances": "1000",
			},
			wantAbsent: []string{"type", "sport", "secs"},
			body:       `{"list":[{"id":"r","distance":[1000],"values":[240],"activity_id":["a1"]}],"activities":{}}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.ListAthletePaceCurves(ctx, CurveParams{CurveSpec: "r.2026-05-01.2026-05-07", DistanceMeters: []int{1000}})
				return err
			},
		},
		{
			name:       "activity power vs hr",
			wantPath:   "/activity/a1/power-vs-hr.json",
			wantQuery:  "",
			wantAbsent: []string{"type", "curves", "secs", "distances"},
			body:       `{"powerHr":1.2,"decoupling":4.5}`,
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
				assertQueryParams(t, r, tc.wantParams, tc.wantAbsent)
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

func TestListAthleteSummaryIsolatesTargetAthlete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		ctx         context.Context
		wantPath    string
		body        string
		wantRows    int
		wantAthlete string
		wantErr     error
	}{
		{
			name:        "configured athlete filters followed athletes",
			ctx:         context.Background(),
			wantPath:    "/athlete/i12345/athlete-summary.json",
			body:        `[{"athlete_id":"i99999","athlete_name":"Followed Rider","date":"2026-05-01","fitness":99},{"athlete_id":"12345","date":"2026-05-01","fitness":88},{"athlete_id":" I12345 ","date":"2026-05-01","fitness":70}]`,
			wantRows:    1,
			wantAthlete: " I12345 ",
		},
		{
			name:        "request scoped coach target filters configured athlete",
			ctx:         WithTargetAthleteID(context.Background(), "i67890"),
			wantPath:    "/athlete/i67890/athlete-summary.json",
			body:        `[{"athlete_id":"i12345","date":"2026-05-01","fitness":70},{"athlete_id":"I67890","date":"2026-05-01","fitness":55}]`,
			wantRows:    1,
			wantAthlete: "I67890",
		},
		{
			name:     "missing ownership fails closed",
			ctx:      context.Background(),
			wantPath: "/athlete/i12345/athlete-summary.json",
			body:     `[{"date":"2026-05-01","fitness":70}]`,
			wantErr:  ErrTargetAthleteMismatch,
		},
		{
			name:     "invalid ownership fails closed",
			ctx:      context.Background(),
			wantPath: "/athlete/i12345/athlete-summary.json",
			body:     `[{"athlete_id":"not-an-athlete","date":"2026-05-01","fitness":70}]`,
			wantErr:  ErrTargetAthleteMismatch,
		},
		{
			name:     "noncanonical ownership key fails closed",
			ctx:      context.Background(),
			wantPath: "/athlete/i12345/athlete-summary.json",
			body:     `[{"Athlete_ID":"i12345","date":"2026-05-01","fitness":70}]`,
			wantErr:  ErrTargetAthleteMismatch,
		},
		{
			name:     "non object ownership fails closed",
			ctx:      context.Background(),
			wantPath: "/athlete/i12345/athlete-summary.json",
			body:     `[null]`,
			wantErr:  ErrTargetAthleteMismatch,
		},
		{
			name:     "canonical foreign ownership overrides case variant target key",
			ctx:      context.Background(),
			wantPath: "/athlete/i12345/athlete-summary.json",
			body:     `[{"athlete_id":"i99999","Athlete_ID":"i12345","athlete_name":"Followed Rider","date":"2026-05-01","fitness":99}]`,
			wantRows: 0,
		},
		{
			name:        "malformed valid foreign row is discarded before typed decode",
			ctx:         context.Background(),
			wantPath:    "/athlete/i12345/athlete-summary.json",
			body:        `[{"athlete_id":"i99999","athlete_name":"Followed Rider","date":"2026-05-01","count":"not-an-integer"},{"athlete_id":"i12345","date":"2026-05-01","fitness":70}]`,
			wantRows:    1,
			wantAthlete: "i12345",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tc.wantPath {
					t.Fatalf("path = %q, want %q", r.URL.Path, tc.wantPath)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()

			client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
			rows, err := client.ListAthleteSummary(tc.ctx, AthleteSummaryParams{Start: "2026-05-01", End: "2026-05-01"})
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("ListAthleteSummary() error = %v, want %v", err, tc.wantErr)
			}
			if tc.wantErr != nil {
				return
			}
			if len(rows) != tc.wantRows {
				t.Fatalf("ListAthleteSummary() rows = %#v, want %d row(s)", rows, tc.wantRows)
			}
			if tc.wantRows > 0 {
				if got := rows[0].Raw["athlete_id"]; got != tc.wantAthlete {
					t.Fatalf("retained athlete_id = %#v, want %q", got, tc.wantAthlete)
				}
				if _, leaked := rows[0].Raw["athlete_name"]; leaked {
					t.Fatalf("retained row leaked followed-athlete PII: %#v", rows[0].Raw)
				}
			}
		})
	}
}

func TestListAthleteSummaryHonorsRequestedDateWindow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		call      func(context.Context, *Client, AthleteSummaryParams) ([]string, error)
		wantDates []string
	}{
		{
			name:      "typed rows enforce requested anchors",
			wantDates: []string{"2026-08-11", "2026-08-24"},
			call: func(ctx context.Context, client *Client, params AthleteSummaryParams) ([]string, error) {
				rows, err := client.ListAthleteSummary(ctx, params)
				if err != nil {
					return nil, err
				}
				dates := make([]string, 0, len(rows))
				for _, row := range rows {
					dates = append(dates, row.Date)
				}
				return dates, nil
			},
		},
		{
			name:      "raw rows preserve out of window evidence",
			wantDates: []string{"2026-08-10", "2026-08-11", "2026-08-24", "2026-08-25"},
			call: func(ctx context.Context, client *Client, params AthleteSummaryParams) ([]string, error) {
				rows, err := client.ListAthleteSummaryRaw(ctx, params)
				if err != nil {
					return nil, err
				}
				dates := make([]string, 0, len(rows))
				for _, row := range rows {
					date, _ := row.Raw["date"].(string)
					dates = append(dates, date)
				}
				return dates, nil
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`[
					{"athlete_id":"i12345","date":"2026-08-10","training_load":10},
					{"athlete_id":"i12345","date":"2026-08-11","training_load":20},
					{"athlete_id":"i12345","date":"2026-08-24","training_load":30},
					{"athlete_id":"i12345","date":"2026-08-25","training_load":40}
				]`))
			}))
			defer server.Close()

			client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
			dates, err := tc.call(context.Background(), client, AthleteSummaryParams{Start: "2026-08-11", End: "2026-08-24"})
			if err != nil {
				t.Fatalf("athlete summary call error = %v", err)
			}
			if !slices.Equal(dates, tc.wantDates) {
				t.Fatalf("summary dates = %v, want %v", dates, tc.wantDates)
			}
		})
	}
}

func TestListAthleteSummaryValidatesDateContracts(t *testing.T) {
	t.Parallel()

	t.Run("invalid bounds fail before request", func(t *testing.T) {
		t.Parallel()
		var requests int
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests++
			_, _ = w.Write([]byte(`[]`))
		}))
		defer server.Close()
		client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
		for _, params := range []AthleteSummaryParams{
			{Start: "2026-08-11T00:00:00", End: "2026-08-24"},
			{Start: "2026-08-11", End: "not-a-date"},
			{Start: "2026-08-24", End: "2026-08-11"},
		} {
			if _, err := client.ListAthleteSummary(context.Background(), params); err == nil {
				t.Fatalf("ListAthleteSummary(%+v) error = nil, want invalid date contract", params)
			}
			if _, err := client.ListAthleteSummaryRaw(context.Background(), params); err == nil {
				t.Fatalf("ListAthleteSummaryRaw(%+v) error = nil, want invalid date contract", params)
			}
		}
		if requests != 0 {
			t.Fatalf("invalid bounds made %d upstream request(s), want zero", requests)
		}
	})

	t.Run("typed target row rejects invalid date while raw retains evidence", func(t *testing.T) {
		t.Parallel()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[
				{"athlete_id":"i12345","training_load":10},
				{"athlete_id":"i12345","date":"2026-08-11T00:00:00","training_load":20}
			]`))
		}))
		defer server.Close()
		client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
		params := AthleteSummaryParams{Start: "2026-08-11", End: "2026-08-24"}
		if _, err := client.ListAthleteSummary(context.Background(), params); err == nil {
			t.Fatal("ListAthleteSummary() error = nil, want invalid target-row date")
		}
		rows, err := client.ListAthleteSummaryRaw(context.Background(), params)
		if err != nil {
			t.Fatalf("ListAthleteSummaryRaw() error = %v", err)
		}
		if len(rows) != 2 {
			t.Fatalf("raw rows = %#v, want invalid-date evidence retained", rows)
		}
	})
}

func assertQueryParams(t *testing.T, r *http.Request, want map[string]string, absent []string) {
	t.Helper()
	query := r.URL.Query()
	for key, wantValue := range want {
		if got := query.Get(key); got != wantValue {
			t.Fatalf("query %s = %q, want %q", key, got, wantValue)
		}
	}
	for _, key := range absent {
		if got := query.Get(key); got != "" {
			t.Fatalf("query %s = %q, want absent", key, got)
		}
	}
}

func TestListAthleteSummaryRawPreservesElementsAndDecodeMarkers(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"athlete_id":"i12345","date":"2026-05-01","training_load":0,"count":"optional-type-error"},{"athlete_id":"I12345","date":"2026-05-02","training_load":12.5}]`))
	}))
	defer server.Close()
	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	rows, err := client.ListAthleteSummaryRaw(context.Background(), AthleteSummaryParams{Start: "2026-05-01", End: "2026-05-02"})
	if err != nil {
		t.Fatalf("ListAthleteSummaryRaw() error = %v", err)
	}
	if len(rows) != 2 || rows[0].Raw["training_load"] != float64(0) || string(rows[1].RawJSON) != `{"athlete_id":"I12345","date":"2026-05-02","training_load":12.5}` {
		t.Fatalf("raw rows = %#v", rows)
	}
	if rows[0].DecodeError == "" || rows[1].DecodeError == "" {
		t.Fatalf("decode markers = %#v, want markers preserved for typed-field errors", rows)
	}
}

func TestListAthleteSummaryRawIsolatesTargetAthlete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		ctx              context.Context
		wantPath         string
		body             string
		wantRows         int
		wantAthleteID    string
		wantTrainingLoad float64
		wantErr          error
	}{
		{
			name:     "filters followed athlete raw rows",
			ctx:      context.Background(),
			wantPath: "/athlete/i12345/athlete-summary.json",
			body:     `[{"athlete_id":"i99999","athlete_name":"Followed Rider","date":"2026-05-01","training_load":200},{"athlete_id":"i12345","date":"2026-05-01","training_load":50}]`,
			wantRows: 1,
		},
		{
			name:             "request scoped coach target filters configured athlete raw rows",
			ctx:              WithTargetAthleteID(context.Background(), "i67890"),
			wantPath:         "/athlete/i67890/athlete-summary.json",
			body:             `[{"athlete_id":"i12345","date":"2026-05-01","training_load":200},{"athlete_id":"I67890","date":"2026-05-01","training_load":50}]`,
			wantRows:         1,
			wantAthleteID:    "I67890",
			wantTrainingLoad: 50,
		},
		{
			name:     "missing ownership fails closed",
			ctx:      context.Background(),
			wantPath: "/athlete/i12345/athlete-summary.json",
			body:     `[{"date":"2026-05-01","training_load":50}]`,
			wantErr:  ErrTargetAthleteMismatch,
		},
		{
			name:     "invalid ownership fails closed",
			ctx:      context.Background(),
			wantPath: "/athlete/i12345/athlete-summary.json",
			body:     `[{"athlete_id":12345,"date":"2026-05-01","training_load":50}]`,
			wantErr:  ErrTargetAthleteMismatch,
		},
		{
			name:     "non object ownership fails closed",
			ctx:      context.Background(),
			wantPath: "/athlete/i12345/athlete-summary.json",
			body:     `[null]`,
			wantErr:  ErrTargetAthleteMismatch,
		},
		{
			name:     "noncanonical ownership key fails closed",
			ctx:      context.Background(),
			wantPath: "/athlete/i12345/athlete-summary.json",
			body:     `[{"Athlete_ID":"i12345","date":"2026-05-01","training_load":50}]`,
			wantErr:  ErrTargetAthleteMismatch,
		},
		{
			name:     "canonical foreign ownership overrides case variant target key",
			ctx:      context.Background(),
			wantPath: "/athlete/i12345/athlete-summary.json",
			body:     `[{"athlete_id":"i99999","Athlete_ID":"i12345","athlete_name":"Followed Rider","date":"2026-05-01","training_load":200}]`,
			wantRows: 0,
		},
		{
			name:     "malformed valid foreign row is discarded before typed marker decode",
			ctx:      context.Background(),
			wantPath: "/athlete/i12345/athlete-summary.json",
			body:     `[{"athlete_id":"i99999","athlete_name":"Followed Rider","date":"2026-05-01","count":"not-an-integer"},{"athlete_id":"i12345","date":"2026-05-01","training_load":50}]`,
			wantRows: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tc.wantPath {
					t.Fatalf("path = %q, want %q", r.URL.Path, tc.wantPath)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()

			client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
			rows, err := client.ListAthleteSummaryRaw(tc.ctx, AthleteSummaryParams{Start: "2026-05-01", End: "2026-05-01"})
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("ListAthleteSummaryRaw() error = %v, want %v", err, tc.wantErr)
			}
			if tc.wantErr != nil {
				return
			}
			if len(rows) != tc.wantRows {
				t.Fatalf("ListAthleteSummaryRaw() rows = %#v, want %d row(s)", rows, tc.wantRows)
			}
			if tc.wantRows > 0 {
				if tc.wantAthleteID != "" {
					if got := rows[0].Raw["athlete_id"]; got != tc.wantAthleteID {
						t.Fatalf("raw athlete_id = %#v, want request target %q", got, tc.wantAthleteID)
					}
					if got := rows[0].Raw["training_load"]; got != tc.wantTrainingLoad {
						t.Fatalf("raw training_load = %#v, want request-target value %v", got, tc.wantTrainingLoad)
					}
				}
				if _, leaked := rows[0].Raw["athlete_name"]; leaked {
					t.Fatalf("raw row leaked followed-athlete PII: %#v", rows[0].Raw)
				}
			}
		})
	}
}

func TestPowerCurveSeasonQueryUsesAthleteLocalAnchor(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/athlete/i12345/power-curves.json" {
			t.Errorf("path=%s", r.URL.Path)
		}
		for key, want := range map[string]string{"curves": "s0,s0-kj0,s0-kj1", "now": "2026-10-26", "type": "Ride", "secs": "7200,18000"} {
			if got := r.URL.Query().Get(key); got != want {
				t.Errorf("query %s=%q, want %q", key, got, want)
			}
		}
		_, _ = w.Write([]byte(`{"list":[{"id":"s0","start_date_local":"2026-03-01","secs":[7200],"values":[250],"watts_per_kg":[4.25],"wkg_activity_id":["light"],"powerModels":[{"type":"MS_2P","criticalPower":271.25,"wPrime":18234.5}]}],"activities":{}}`))
	}))
	defer server.Close()
	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	set, err := client.ListAthletePowerCurves(context.Background(), CurveParams{Sport: "Ride", Now: "2026-10-26", CurveSpec: "s0,s0-kj0,s0-kj1", DurationSeconds: []int{7200, 18000}})
	if err != nil {
		t.Fatal(err)
	}
	curve := set.List[0]
	if curve.WattsPerKG[0] == nil || *curve.WattsPerKG[0] != 4.25 || curve.WKGActivityID[0] != "light" {
		t.Fatalf("W/kg=%#v", curve)
	}
	if len(curve.PowerModels) != 1 || curve.PowerModels[0].CriticalPower == nil || *curve.PowerModels[0].CriticalPower != 271.25 {
		t.Fatalf("models=%#v", curve.PowerModels)
	}
}

func TestCurveRejectsTimestampSeasonAnchorBeforeHTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Error("invalid local date made HTTP request") }))
	defer server.Close()
	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	_, err := client.ListAthletePowerCurves(context.Background(), CurveParams{Sport: "Ride", Now: "2026-10-27T01:00:00Z", CurveSpec: "s0"})
	if err == nil {
		t.Fatal("timestamp accepted as local date")
	}
}

package intervals

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListActivitiesSendsQueryAndPreservesRawNulls(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/athlete/i12345/activities"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		query := r.URL.Query()
		checks := map[string]string{
			"oldest":   "2026-01-01",
			"newest":   "2026-01-31",
			"route_id": "99",
			"limit":    "51",
			"fields":   "id,name,start_date_local,source,_note",
		}
		for key, want := range checks {
			if got := query.Get(key); got != want {
				t.Fatalf("query %s = %q, want %q", key, got, want)
			}
		}
		for _, key := range []string{"routeId", "start", "end"} {
			if got := query.Get(key); got != "" {
				t.Fatalf("query %s = %q, want absent", key, got)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{"id":"a1","name":null,"type":"Run","start_date_local":"2026-01-30T07:00:00","distance":5000,"icu_training_load":42,"stream_types":["time","distance"],"has_weather":true,"average_weather_temp":22.5,"average_wind_speed":4.1,"prevailing_wind_deg":180},
			{"id":"a2","source":"strava","_note":"Strava activity hidden"}
		]`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	activities, err := client.ListActivities(context.Background(), ListActivitiesParams{
		Oldest:  "2026-01-01",
		Newest:  "2026-01-31",
		RouteID: 99,
		Limit:   51,
		Fields:  []string{"id", "name", "start_date_local", "source", "_note"},
	})
	if err != nil {
		t.Fatalf("ListActivities() error = %v", err)
	}
	if len(activities) != 2 {
		t.Fatalf("activity count = %d, want 2", len(activities))
	}
	if activities[0].Name != nil {
		t.Fatalf("Name = %q, want nil pointer for upstream null", *activities[0].Name)
	}
	if rawName, ok := activities[0].Raw["name"]; !ok || rawName != nil {
		rawJSON, _ := json.Marshal(activities[0].Raw)
		t.Fatalf("raw name = %#v (present %v), raw = %s; want present nil", rawName, ok, rawJSON)
	}
	if got, want := activities[0].StreamTypes, []string{"time", "distance"}; len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("StreamTypes = %#v, want %#v", got, want)
	}
	if activities[0].HasWeather == nil || !*activities[0].HasWeather || activities[0].AverageWeatherTemp == nil || *activities[0].AverageWeatherTemp != 22.5 || activities[0].AverageWindSpeed == nil || *activities[0].AverageWindSpeed != 4.1 || activities[0].PrevailingWindDeg == nil || *activities[0].PrevailingWindDeg != 180 {
		t.Fatalf("weather fields = %+v, want decoded activity weather", activities[0])
	}
	if activities[1].Source == nil || *activities[1].Source != "strava" {
		t.Fatalf("Source = %#v, want strava", activities[1].Source)
	}
}

func TestActivitySubjectiveFieldsPreserveSourceTypesAndRawPayload(t *testing.T) {
	t.Parallel()

	var activity Activity
	if err := json.Unmarshal([]byte(`{"id":"subjective","commute":true,"feel":4,"icu_rpe":8,"name":null}`), &activity); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}
	if activity.Commute == nil || !*activity.Commute || activity.Feel == nil || *activity.Feel != 4 || activity.RPE == nil || *activity.RPE != 8 {
		t.Fatalf("normalized fields = %+v, want commute=true feel=4 rpe=8", activity)
	}
	if activity.Raw["commute"] != true || activity.Raw["feel"] != float64(4) || activity.Raw["icu_rpe"] != float64(8) {
		t.Fatalf("raw subjective fields = %#v, want upstream keys preserved", activity.Raw)
	}
	if value, ok := activity.Raw["name"]; !ok || value != nil {
		t.Fatalf("raw name = %#v present %v, want preserved null", value, ok)
	}
}

func TestActivitySubjectiveFieldsIgnoreMalformedOptionalValues(t *testing.T) {
	t.Parallel()

	var activity Activity
	if err := json.Unmarshal([]byte(`{"id":"malformed","commute":"yes","feel":"4","icu_rpe":7.5,"name":"still readable"}`), &activity); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v, want malformed optional values tolerated", err)
	}
	if activity.Commute != nil || activity.Feel != nil || activity.RPE != nil {
		t.Fatalf("normalized malformed fields = %+v, want omitted", activity)
	}
	for key, want := range map[string]any{"commute": "yes", "feel": "4", "icu_rpe": 7.5} {
		if got, ok := activity.Raw[key]; !ok || got != want {
			t.Fatalf("raw[%q] = %#v present %v, want %#v preserved", key, got, ok, want)
		}
	}
}

func TestListActivitiesRequiresOldest(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, "https://example.invalid", http.DefaultClient, RetryConfig{})
	if _, err := client.ListActivities(context.Background(), ListActivitiesParams{}); err == nil {
		t.Fatal("ListActivities() error = nil, want required oldest error")
	}
}

func TestListActivitiesAroundSendsActivityIDAndLimit(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/athlete/i12345/activities-around"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		query := r.URL.Query()
		checks := map[string]string{
			"activity_id": "a-ref",
			"limit":       "7",
		}
		for key, want := range checks {
			if got := query.Get(key); got != want {
				t.Fatalf("query %s = %q, want %q", key, got, want)
			}
		}
		for _, key := range []string{"id", "count", "activityId"} {
			if got := query.Get(key); got != "" {
				t.Fatalf("query %s = %q, want absent", key, got)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{"id":"near1","name":null,"type":"Ride","start_date_local":"2026-01-30T07:00:00","stream_types":["time"],"custom_metric":123},
			{"id":"near2","source":"strava","_note":"Strava activity hidden"}
		]`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	activities, err := client.ListActivitiesAround(context.Background(), ActivitiesAroundParams{ActivityID: " a-ref ", Limit: 7})
	if err != nil {
		t.Fatalf("ListActivitiesAround() error = %v", err)
	}
	if len(activities) != 2 {
		t.Fatalf("activity count = %d, want 2", len(activities))
	}
	if activities[0].Name != nil {
		t.Fatalf("Name = %q, want nil pointer for upstream null", *activities[0].Name)
	}
	if rawName, ok := activities[0].Raw["name"]; !ok || rawName != nil {
		rawJSON, _ := json.Marshal(activities[0].Raw)
		t.Fatalf("raw name = %#v (present %v), raw = %s; want present nil", rawName, ok, rawJSON)
	}
	if got := activities[0].Raw["custom_metric"]; got != float64(123) {
		t.Fatalf("raw custom_metric = %#v, want preserved unknown field", got)
	}
	if activities[1].Source == nil || *activities[1].Source != "strava" {
		t.Fatalf("Source = %#v, want strava", activities[1].Source)
	}
}

func TestListActivitiesAroundRequiresActivityID(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, "https://example.invalid", http.DefaultClient, RetryConfig{})
	if _, err := client.ListActivitiesAround(context.Background(), ActivitiesAroundParams{}); err == nil {
		t.Fatal("ListActivitiesAround() error = nil, want required activity ID error")
	}
}

func TestLinkActivityToEventUsesPutActivityPairedEventID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPut; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/activity/i147866949"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		var decoded map[string]any
		if err := json.Unmarshal(body, &decoded); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if decoded["paired_event_id"] != float64(1001) || len(decoded) != 1 {
			t.Fatalf("body = %#v, want only paired_event_id", decoded)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"i147866949","paired_event_id":1001,"start_date_local":"2026-05-10T07:00:00"}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	activity, err := client.LinkActivityToEvent(context.Background(), LinkActivityToEventParams{ActivityID: " i147866949 ", EventID: " 1001 "})
	if err != nil {
		t.Fatalf("LinkActivityToEvent() error = %v", err)
	}
	if activity.ID != "i147866949" || activity.Raw["paired_event_id"] != float64(1001) {
		t.Fatalf("activity = %+v raw=%#v, want linked response", activity, activity.Raw)
	}
}

func TestUpdateActivitySendsSparsePutPayload(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPut; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/activity/i147866949"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		var decoded map[string]any
		if err := json.Unmarshal(body, &decoded); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if decoded["name"] != "Threshold ride" || decoded["description"] != "Felt strong; held target W" || len(decoded) != 2 {
			t.Fatalf("body = %#v, want name+description only", decoded)
		}
		for _, unsupported := range []string{"rpe", "perceived_exertion", "icu_rpe", "feel", "commute", "gear_id"} {
			if _, ok := decoded[unsupported]; ok {
				t.Fatalf("body = %#v, must not send read-only/unsupported field %q", decoded, unsupported)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"i147866949","name":"Threshold ride","description":"Felt strong; held target W"}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	activity, err := client.UpdateActivity(context.Background(), UpdateActivityParams{
		ActivityID:     " i147866949 ",
		Name:           "Threshold ride",
		NameSet:        true,
		Description:    "Felt strong; held target W",
		DescriptionSet: true,
	})
	if err != nil {
		t.Fatalf("UpdateActivity() error = %v", err)
	}
	if activity.ID != "i147866949" || activity.Name == nil || *activity.Name != "Threshold ride" {
		t.Fatalf("activity = %+v, want updated name", activity)
	}
}

func TestUpdateActivityRequiresAtLeastOneField(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, "https://example.invalid", http.DefaultClient, RetryConfig{})
	if _, err := client.UpdateActivity(context.Background(), UpdateActivityParams{ActivityID: "a1"}); err == nil {
		t.Fatal("UpdateActivity() error = nil, want validation error")
	}
	if _, err := client.UpdateActivity(context.Background(), UpdateActivityParams{Name: "x", NameSet: true}); err == nil {
		t.Fatal("UpdateActivity() error = nil, want activity ID required")
	}
	for _, grams := range []int{-1, 2147483648} {
		if _, err := client.UpdateActivity(context.Background(), UpdateActivityParams{ActivityID: "a1", CarbsIngested: grams, CarbsIngestedSet: true}); err == nil {
			t.Fatalf("UpdateActivity(carbs_ingested=%d) error = nil, want range validation error", grams)
		}
	}
}

func TestUpdateActivitySendsSparseCarbsIngested(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		grams int
	}{
		{name: "positive grams", grams: 90},
		{name: "logged zero", grams: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatalf("read body: %v", err)
				}
				var decoded map[string]any
				if err := json.Unmarshal(body, &decoded); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				got, ok := decoded["carbs_ingested"]
				if !ok || got != float64(tc.grams) {
					t.Fatalf("carbs_ingested = %#v, present %v, want %d", got, ok, tc.grams)
				}
				if len(decoded) != 1 {
					t.Fatalf("body = %#v, want carbs_ingested only", decoded)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"id":"a1"}`))
			}))
			defer server.Close()

			client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
			_, err := client.UpdateActivity(context.Background(), UpdateActivityParams{
				ActivityID:       "a1",
				CarbsIngested:    tc.grams,
				CarbsIngestedSet: true,
			})
			if err != nil {
				t.Fatalf("UpdateActivity() error = %v", err)
			}
		})
	}
}

func TestUpdateActivitySendsExplicitEmptyDescription(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var decoded map[string]any
		_ = json.Unmarshal(body, &decoded)
		if _, ok := decoded["description"]; !ok {
			t.Fatalf("body = %s, want description sent", body)
		}
		if decoded["description"] != "" {
			t.Fatalf("description = %#v, want empty string to clear", decoded["description"])
		}
		_, _ = w.Write([]byte(`{"id":"a1","description":""}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	if _, err := client.UpdateActivity(context.Background(), UpdateActivityParams{ActivityID: "a1", DescriptionSet: true}); err != nil {
		t.Fatalf("UpdateActivity() error = %v", err)
	}
}

func TestLinkActivityToEventRequiresIDs(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, "https://example.invalid", http.DefaultClient, RetryConfig{})
	for _, params := range []LinkActivityToEventParams{
		{EventID: "1001"},
		{ActivityID: "a1"},
		{ActivityID: "a1", EventID: "evt-1"},
	} {
		if _, err := client.LinkActivityToEvent(context.Background(), params); err == nil {
			t.Fatalf("LinkActivityToEvent(%#v) error = nil, want validation error", params)
		}
	}
}

func TestUpdateActivityWritableMetadataPayload(t *testing.T) {
	for _, tc := range []struct {
		name, source, fileType string
		distance               bool
		wantError              bool
	}{
		{name: "metadata only", source: "UPLOAD", fileType: "fit"},
		{name: "manual distance", source: "MANUAL", distance: true},
		{name: "file distance rejected", source: "UPLOAD", fileType: "fit", distance: true, wantError: true},
		{name: "unknown distance rejected", distance: true, wantError: true},
		{name: "manual with file rejected", source: "MANUAL", fileType: "fit", distance: true, wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			writes := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.Method == http.MethodGet {
					_, _ = w.Write([]byte(`{"id":"a1","source":"` + tc.source + `","file_type":"` + tc.fileType + `"}`))
					return
				}
				writes++
				var payload map[string]any
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					t.Fatal(err)
				}
				wantLen := 5
				if tc.distance {
					wantLen++
				}
				if len(payload) != wantLen || payload["commute"] != false || payload["feel"] != float64(1) || payload["icu_rpe"] != float64(10) || payload["sub_type"] != "NONE" {
					t.Fatalf("payload=%v", payload)
				}
				if tags, ok := payload["tags"].([]any); !ok || len(tags) != 0 {
					t.Fatalf("tags=%v", payload["tags"])
				}
				if tc.distance && payload["distance"] != float64(0) {
					t.Fatalf("distance=%v", payload)
				}
				_, _ = w.Write([]byte(`{"id":"a1"}`))
			}))
			defer server.Close()
			client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
			tags := []string{}
			commute := false
			feel, rpe := 1, 10
			subtype := "NONE"
			distance := 0.0
			params := UpdateActivityParams{ActivityID: "a1", Tags: &tags, Commute: &commute, Feel: &feel, RPE: &rpe, SubType: &subtype}
			if tc.distance {
				params.DistanceM = &distance
			}
			_, err := client.UpdateActivity(context.Background(), params)
			if (err != nil) != tc.wantError {
				t.Fatalf("err=%v wantError=%v", err, tc.wantError)
			}
			if tc.wantError && writes != 0 {
				t.Fatalf("writes=%d", writes)
			}
		})
	}
}

func TestUpdateActivityReadbackVerification(t *testing.T) {
	for _, tc := range []struct {
		name, receipt, readback, wantStatus, wantSource string
		readbackStatus, wantReads                       int
	}{
		{name: "response confirms", receipt: `{"id":"a1","commute":false}`, wantStatus: "verified", wantSource: "write_response"},
		{name: "readback confirms", receipt: `{"id":"a1"}`, readback: `{"id":"a1","commute":false}`, wantStatus: "verified", wantSource: "readback", wantReads: 1},
		{name: "readback differs", receipt: `{"id":"a1","commute":true}`, readback: `{"id":"a1","commute":true}`, wantStatus: "verification_mismatch", wantSource: "readback", wantReads: 1},
		{name: "readback omits", receipt: `{"id":"a1"}`, readback: `{"id":"a1"}`, wantStatus: "verification_unavailable", wantSource: "readback", wantReads: 1},
		{name: "readback fails", receipt: `{"id":"a1"}`, readbackStatus: 404, wantStatus: "verification_unavailable", wantSource: "write_response", wantReads: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			writes, reads := 0, 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/activity/a1" {
					t.Errorf("path=%s", r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				if r.Method == http.MethodPut {
					writes++
					_, _ = w.Write([]byte(tc.receipt))
					return
				}
				reads++
				if tc.readbackStatus != 0 {
					w.WriteHeader(tc.readbackStatus)
				}
				body := tc.readback
				if body == "" {
					body = `{}`
				}
				_, _ = w.Write([]byte(body))
			}))
			defer server.Close()
			client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
			commute := false
			activity, err := client.UpdateActivity(context.Background(), UpdateActivityParams{ActivityID: "a1", Commute: &commute})
			if err != nil {
				t.Fatal(err)
			}
			if writes != 1 || reads != tc.wantReads {
				t.Fatalf("writes=%d reads=%d", writes, reads)
			}
			v := activity.MetadataVerification
			if v == nil || v.Status != tc.wantStatus || v.Source != tc.wantSource {
				t.Fatalf("verification=%+v", v)
			}
			if tc.wantStatus == "verified" {
				if len(v.Confirmed) != 1 || v.Confirmed[0] != "commute" {
					t.Fatalf("verification=%+v", v)
				}
			} else {
				if len(v.Confirmed) > 0 {
					t.Fatalf("unverified field claimed=%+v", v)
				}
			}
		})
	}
}

func TestUpdateActivityVerificationRequiresRequestedIdentity(t *testing.T) {
	for _, tc := range []struct {
		name, receipt, readback, wantStatus string
		coach                               bool
	}{
		{name: "wrong receipt recovers from requested readback", receipt: `{"id":"other","commute":false,"icu_athlete_id":"i67890"}`, readback: `{"id":"a1","commute":false,"icu_athlete_id":"i67890"}`, wantStatus: "verified"},
		{name: "missing receipt identity recovers from requested readback", receipt: `{"commute":false,"icu_athlete_id":"i67890"}`, readback: `{"id":"a1","commute":false,"icu_athlete_id":"i67890"}`, wantStatus: "verified"},
		{name: "wrong receipt and readback", receipt: `{"id":"other","commute":false,"icu_athlete_id":"i67890"}`, readback: `{"id":"other","commute":false,"icu_athlete_id":"i67890"}`, wantStatus: "verification_mismatch"},
		{name: "missing receipt and readback identity", receipt: `{"commute":false,"icu_athlete_id":"i67890"}`, readback: `{"commute":false,"icu_athlete_id":"i67890"}`, wantStatus: "verification_unavailable"},
		{name: "sparse receipt wrong readback", receipt: `{"id":"a1","icu_athlete_id":"i67890"}`, readback: `{"id":"other","commute":false,"icu_athlete_id":"i67890"}`, wantStatus: "verification_mismatch"},
		{name: "coach same owner wrong activity", coach: true, receipt: `{"id":"other","commute":false,"icu_athlete_id":"i67890"}`, readback: `{"id":"other","commute":false,"icu_athlete_id":"i67890"}`, wantStatus: "verification_mismatch"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			writes, reads := 0, 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/activity/a1" {
					t.Errorf("path=%s", r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				if r.Method == http.MethodPut {
					writes++
					_, _ = w.Write([]byte(tc.receipt))
					return
				}
				reads++
				if tc.coach && reads == 1 {
					_, _ = w.Write([]byte(`{"id":"a1","icu_athlete_id":"i67890"}`))
					return
				}
				_, _ = w.Write([]byte(tc.readback))
			}))
			defer server.Close()
			client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
			commute := false
			ctx := context.Background()
			if tc.coach {
				ctx = WithTargetAthleteID(ctx, "i67890")
			}
			activity, err := client.UpdateActivity(ctx, UpdateActivityParams{ActivityID: " a1 ", Commute: &commute})
			if err != nil {
				t.Fatal(err)
			}
			wantReads := 1
			if tc.coach {
				wantReads++
			}
			if writes != 1 || reads != wantReads {
				t.Errorf("writes=%d reads=%d; want one write and %d reads", writes, reads, wantReads)
			}
			v := activity.MetadataVerification
			if v == nil || v.Status != tc.wantStatus || v.Source != "readback" {
				t.Fatalf("verification=%+v", v)
			}
			if tc.wantStatus != "verified" && len(v.Confirmed) > 0 {
				t.Fatalf("wrong activity was confirmed: %+v", v)
			}
		})
	}
}

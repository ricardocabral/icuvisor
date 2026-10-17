package tools

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/intervals"
)

type calendarEventsFixture struct {
	SameDayDistinct    []intervals.Event `json:"same_day_distinct"`
	OccupiedReschedule []intervals.Event `json:"occupied_reschedule"`
	MultiDayRace       []intervals.Event `json:"multi_day_race"`
}

type calendarComplianceFixture struct {
	Events     []intervals.Event    `json:"events"`
	Activities []intervals.Activity `json:"activities"`
}

type calendarTimezoneFixture struct {
	Name             string          `json:"name"`
	Timezone         string          `json:"timezone"`
	Oldest           string          `json:"oldest"`
	Newest           string          `json:"newest"`
	Event            intervals.Event `json:"event"`
	WantUpdatedLocal string          `json:"want_updated_local"`
}

type interruptedCreateFixture struct {
	Request        json.RawMessage `json:"request"`
	CommittedEvent intervals.Event `json:"committed_event"`
}

type calendarLinkingFixture struct {
	Event                 intervals.Event    `json:"event"`
	AlreadyLinkedActivity intervals.Activity `json:"already_linked_activity"`
	DifferentLinkActivity intervals.Activity `json:"different_link_activity"`
	UnlinkedActivity      intervals.Activity `json:"unlinked_activity"`
}

type partialRangeFixture struct {
	Request           json.RawMessage `json:"request"`
	FirstCreatedEvent intervals.Event `json:"first_created_event"`
}

type interruptedEventWriter struct {
	fakeProfileClient
	listResults [][]intervals.Event
	writeCalls  []intervals.WriteEventParams
	listCalls   []intervals.ListEventsParams
}

func (f *interruptedEventWriter) AddOrUpdateEvent(_ context.Context, params intervals.WriteEventParams) (intervals.Event, error) {
	f.writeCalls = append(f.writeCalls, params)
	return intervals.Event{}, errors.New("connection reset after request body")
}

func (f *interruptedEventWriter) ListEvents(_ context.Context, params intervals.ListEventsParams) ([]intervals.Event, error) {
	f.listCalls = append(f.listCalls, params)
	if len(f.listResults) == 0 {
		return nil, nil
	}
	result := f.listResults[0]
	f.listResults = f.listResults[1:]
	return append([]intervals.Event(nil), result...), nil
}

func TestCalendarReconciliationSameDaySessionsAndMultiDayRaceStayDistinct(t *testing.T) {
	t.Parallel()

	fixture := readCalendarFixture[calendarEventsFixture](t, "events.json")
	client := &fakeEventsTrainingPlanClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		events:            append(append([]intervals.Event(nil), fixture.SameDayDistinct...), fixture.MultiDayRace...),
	}
	tool := newGetEventsTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-05-25","newest":"2026-07-12","limit":10}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	rows := rowsByEventID(resultMap(t, result)["events"].([]any))
	if len(rows) != 3 {
		t.Fatalf("events = %#v, want two distinct sessions and one race row", rows)
	}
	if rows["session-am"]["name"] != "AM aerobic run" || rows["session-pm"]["name"] != "PM endurance ride" {
		t.Fatalf("same-day rows = %#v / %#v, want separately identified sessions", rows["session-am"], rows["session-pm"])
	}
	race := rows["stage-race"]
	if race["start_date_local"] != "2026-07-10T08:00:00" || race["end_date_local"] != "2026-07-12T16:00:00" || race["load_target"] != float64(420) {
		t.Fatalf("race = %#v, want one multi-day race with one load target", race)
	}
}

func TestCalendarReconciliationRescheduleOntoOccupiedDateReportsConflicts(t *testing.T) {
	t.Parallel()

	fixture := readCalendarFixture[calendarEventsFixture](t, "events.json")
	client := &fakeEventWriterClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		events:            fixture.OccupiedReschedule,
		event:             decodeToolEvents(t, `{"id":"moving-run","category":"WORKOUT","type":"Run","name":"Rescheduled intervals","start_date_local":"2026-06-12T00:00:00","time_target":3000}`)[0],
	}
	tool := newAddOrUpdateEventTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"event_id":"moving-run","date":"2026-06-12","category":"WORKOUT","type":"Run","name":"Rescheduled intervals","moving_time_seconds":3000}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 1 || client.calls[0].EventID != "moving-run" {
		t.Fatalf("write calls = %#v, want only explicit event_id updated", client.calls)
	}
	meta := resultMap(t, result)["_meta"].(map[string]any)
	conflicts, ok := meta["same_day_conflicts"].([]any)
	if !ok {
		t.Fatalf("meta = %#v, want same_day_conflicts array", meta)
	}
	if len(conflicts) != 2 || meta["duplicate_warning"] != "Target date already contains other events; only the requested event_id was updated. Verify the reschedule did not create an unintended overlap." {
		t.Fatalf("meta = %#v, want two occupied-date conflicts and actionable warning", meta)
	}
}

func TestCalendarReconciliationLinkRetrySkipsVerifiedExistingPair(t *testing.T) {
	t.Parallel()

	fixture := readCalendarFixture[calendarLinkingFixture](t, "linking.json")
	client := &fakeLinkActivityToEventClient{activity: fixture.AlreadyLinkedActivity, event: fixture.Event, linked: fixture.AlreadyLinkedActivity}
	tool := newLinkActivityToEventTool(client, client, client, "test", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"activity-1","event_id":"1001"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 0 {
		t.Fatalf("link calls = %#v, want retry skipped before a second write", client.calls)
	}
	out := resultMap(t, result)
	if out["status"] != "already_linked" {
		t.Fatalf("status = %#v, want already_linked", out["status"])
	}
	meta := out["_meta"].(map[string]any)
	if meta["confirmation_status"] != "verified_existing_link" {
		t.Fatalf("meta = %#v, want verified existing-link confirmation", meta)
	}
}

func TestCalendarReconciliationLinkRefusesToOverwriteDifferentPair(t *testing.T) {
	t.Parallel()

	fixture := readCalendarFixture[calendarLinkingFixture](t, "linking.json")
	client := &fakeLinkActivityToEventClient{activity: fixture.DifferentLinkActivity, event: fixture.Event, linked: fixture.DifferentLinkActivity}
	tool := newLinkActivityToEventTool(client, client, client, "test", false)

	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"activity-1","event_id":"1001"}`)})
	if err == nil {
		t.Fatal("Handler() error = nil, want existing-pair conflict")
	}
	if len(client.calls) != 0 {
		t.Fatalf("link calls = %#v, want conflict before write", client.calls)
	}
	message, ok := PublicErrorMessage(err)
	if !ok || message != "activity is already linked to event 999; unlink it before linking event 1001" {
		t.Fatalf("PublicErrorMessage(%v) = %q/%v, want actionable existing-pair conflict", err, message, ok)
	}
}

func TestCalendarReconciliationLinkDoesNotReportUnverifiedWriteAsSuccess(t *testing.T) {
	t.Parallel()

	fixture := readCalendarFixture[calendarLinkingFixture](t, "linking.json")
	client := &fakeLinkActivityToEventClient{activity: fixture.UnlinkedActivity, event: fixture.Event, linked: fixture.UnlinkedActivity}
	tool := newLinkActivityToEventTool(client, client, client, "test", false)

	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"activity-1","event_id":"1001"}`)})
	if err == nil {
		t.Fatal("Handler() error = nil, want unverified link write to remain an error")
	}
	if len(client.calls) != 1 {
		t.Fatalf("link calls = %#v, want one attempted write", client.calls)
	}
	message, ok := PublicErrorMessage(err)
	if !ok || message != linkActivityToEventMessage {
		t.Fatalf("PublicErrorMessage(%v) = %q/%v, want terse link guidance", err, message, ok)
	}
}

func TestCalendarReconciliationLinkedCompletionAndProviderDuplicateCountOnce(t *testing.T) {
	t.Parallel()

	fixture := readCalendarFixture[calendarComplianceFixture](t, "compliance.json")
	client := newFakeComputeClient()
	client.events = fixture.Events
	client.activities = fixture.Activities
	tool := newComputeComplianceRateToolWithClock(client, client, client, client, "test", "UTC", false, fixedTodayClock())

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"start_date":"2026-05-20","end_date":"2026-05-20","target_metric":"time","include_full":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	out := resultMap(t, result)
	summary := out["result"].(map[string]any)
	if summary["scheduled_count"] != float64(2) || summary["completed_count"] != float64(2) || summary["completed_linked_count"] != float64(2) || summary["delta_sample_count"] != float64(2) {
		t.Fatalf("result = %#v, want two linked sessions counted once despite provider duplicate", summary)
	}
	rows := rowsByEventID(out["series"].([]any))
	if rows["planned-strength"]["paired_activity_id"] != "strength-direct" || rows["planned-ride"]["paired_activity_id"] != "ride-direct" {
		t.Fatalf("rows = %#v, want explicit links to win over provider duplicate", rows)
	}
}

func TestCalendarReconciliationTimezoneOffsetsAndDST(t *testing.T) {
	t.Parallel()

	fixtures := readCalendarFixture[[]calendarTimezoneFixture](t, "timezones.json")
	for _, fixture := range fixtures {
		t.Run(fixture.Name, func(t *testing.T) {
			t.Parallel()
			client := &fakeEventsTrainingPlanClient{
				fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: fixture.Timezone}},
				events:            []intervals.Event{fixture.Event},
			}
			tool := newGetEventsTool(client, client, "test", "UTC", false)
			args := json.RawMessage(`{"oldest":"` + fixture.Oldest + `","newest":"` + fixture.Newest + `","limit":10}`)

			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: args})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			row := resultMap(t, result)["events"].([]any)[0].(map[string]any)
			if row["updated_local"] != fixture.WantUpdatedLocal {
				t.Fatalf("updated_local = %#v, want %q", row["updated_local"], fixture.WantUpdatedLocal)
			}
			meta := resultMap(t, result)["_meta"].(map[string]any)
			if meta["timezone"] != fixture.Timezone {
				t.Fatalf("timezone = %#v, want %q", meta["timezone"], fixture.Timezone)
			}
		})
	}
}

func TestCalendarReconciliationInterruptedCreateRequiresReadAfterWriteVerification(t *testing.T) {
	t.Parallel()

	fixture := readCalendarFixture[interruptedCreateFixture](t, "interrupted_create.json")
	client := &interruptedEventWriter{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		listResults:       [][]intervals.Event{nil, {fixture.CommittedEvent}},
	}
	tool := newAddOrUpdateEventTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: fixture.Request})
	if err != nil {
		t.Fatalf("Handler() error = %v, want verified interrupted create recovery", err)
	}
	if len(client.writeCalls) != 1 || len(client.listCalls) != 2 {
		t.Fatalf("writes=%d lists=%d, want preflight, one write, and bounded verification", len(client.writeCalls), len(client.listCalls))
	}
	meta := resultMap(t, result)["_meta"].(map[string]any)
	if meta["operation"] != "create_recovered" || meta["confirmation_status"] != "verified_after_interrupted_write" || meta["duplicate_event_id"] != "committed-after-timeout" {
		t.Fatalf("meta = %#v, want verified interrupted-write recovery", meta)
	}
}

func TestCalendarReconciliationRetrySkipsPreviouslyCommittedCreate(t *testing.T) {
	t.Parallel()

	fixture := readCalendarFixture[interruptedCreateFixture](t, "interrupted_create.json")
	client := &fakeEventWriterClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		events:            []intervals.Event{fixture.CommittedEvent},
	}
	tool := newAddOrUpdateEventTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: fixture.Request})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 0 {
		t.Fatalf("write calls = %#v, want retry skipped before a duplicate create", client.calls)
	}
	meta := resultMap(t, result)["_meta"].(map[string]any)
	if meta["operation"] != "skip_duplicate" || meta["confirmation_status"] != "skipped_existing_event" || meta["duplicate_event_id"] != "committed-after-timeout" {
		t.Fatalf("meta = %#v, want idempotent retry result", meta)
	}
}

func TestCalendarReconciliationPartialRangeWriteIsNotReportedSuccessful(t *testing.T) {
	t.Parallel()

	fixture := readCalendarFixture[partialRangeFixture](t, "partial_range.json")
	client := &fakeUnavailableDateRangeClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		created:           []intervals.Event{fixture.FirstCreatedEvent},
		writeErrors:       []error{nil, errors.New("upstream interrupted second write")},
	}
	tool := newAddUnavailableDateRangeTool(client, client, "test", "UTC", false)

	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: fixture.Request})
	if err == nil {
		t.Fatal("Handler() error = nil, want partial range write to remain an error")
	}
	if len(client.calls) != 2 || client.calls[0].Date != "2026-08-10" || client.calls[1].Date != "2026-08-11" {
		t.Fatalf("write calls = %#v, want first success and second interrupted attempt", client.calls)
	}
	message, ok := PublicErrorMessage(err)
	if !ok || message != writeUnavailableDateRangeMessage || strings.Contains(message, "second write") {
		t.Fatalf("PublicErrorMessage(%v) = %q/%v, want terse partial-write error", err, message, ok)
	}
}

func TestCalendarReconciliationInterruptedCreateWithoutVerificationReturnsActionableError(t *testing.T) {
	t.Parallel()

	fixture := readCalendarFixture[interruptedCreateFixture](t, "interrupted_create.json")
	client := &interruptedEventWriter{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		listResults:       [][]intervals.Event{nil, nil},
	}
	tool := newAddOrUpdateEventTool(client, client, "test", "UTC", false)

	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: fixture.Request})
	if err == nil {
		t.Fatal("Handler() error = nil, want unverified interrupted write to remain an error")
	}
	message, ok := PublicErrorMessage(err)
	if !ok || message != writeEventMessage || strings.Contains(message, "connection reset") {
		t.Fatalf("PublicErrorMessage(%v) = %q/%v, want terse write guidance without transport detail", err, message, ok)
	}
}

func readCalendarFixture[T any](t *testing.T, name string) T {
	t.Helper()
	path := filepath.Join("testdata", "calendar_reconciliation", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var fixture T
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	return fixture
}

func TestCalendarReconciliationUnconfirmedCreateResponseRequiresVerification(t *testing.T) {
	t.Parallel()
	fixture := readCalendarFixture[interruptedCreateFixture](t, "interrupted_create.json")
	for _, committed := range []bool{false, true} {
		t.Run(map[bool]string{false: "not verified", true: "verified and retry skipped"}[committed], func(t *testing.T) {
			t.Parallel()
			writes, reads := 0, 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch {
				case r.Method == http.MethodGet && r.URL.Path == "/athlete/i12345/events":
					reads++
					if r.URL.Query().Get("oldest") != "2026-06-13" || r.URL.Query().Get("newest") != "2026-06-13" {
						t.Errorf("verification must stay bounded to the written date: %s", r.URL)
					}
					if committed && writes > 0 {
						_ = json.NewEncoder(w).Encode([]map[string]any{fixture.CommittedEvent.Raw})
					} else {
						_, _ = w.Write([]byte("[]"))
					}
				case r.Method == http.MethodPost && r.URL.Path == "/athlete/i12345/events/bulk":
					writes++
					_, _ = w.Write([]byte("[null]"))
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL)
					w.WriteHeader(http.StatusBadRequest)
				}
			}))
			defer server.Close()
			client, err := intervals.NewClient(intervals.Options{Config: config.Config{APIKey: "test-key", AthleteID: "i12345", APIBaseURL: server.URL, HTTPTimeout: time.Second}, HTTPClient: server.Client(), Version: "test"})
			if err != nil {
				t.Fatal(err)
			}
			profile := &fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", Timezone: "UTC"}}
			tool := newAddOrUpdateEventTool(client, profile, "test", "UTC", false)
			request := Request{Name: tool.Name, Arguments: fixture.Request}
			result, err := tool.Handler(context.Background(), request)
			if reads != 2 || writes != 1 {
				t.Fatalf("reads=%d writes=%d, want preflight + one verification and no repeated write", reads, writes)
			}
			if !committed {
				if message, ok := PublicErrorMessage(err); !ok || message != writeEventMessage {
					t.Fatalf("error = %v, want actionable unconfirmed-write guidance", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if meta := resultMap(t, result)["_meta"].(map[string]any); meta["confirmation_status"] != verifiedAfterInterruptedWriteStatus {
				t.Fatalf("meta = %#v, want verified recovery", meta)
			}
			result, err = tool.Handler(context.Background(), request)
			if err != nil {
				t.Fatal(err)
			}
			if meta := resultMap(t, result)["_meta"].(map[string]any); meta["confirmation_status"] != skippedExistingEventStatus || writes != 1 || reads != 3 {
				t.Fatalf("retry meta = %#v, writes=%d reads=%d; want verified duplicate skip", meta, writes, reads)
			}
		})
	}
}

func TestCalendarReconciliationLoadUsesUpstreamAggregatesWithoutAddingCalendarTargets(t *testing.T) {
	t.Parallel()
	cases := readCalendarFixture[[]struct {
		Name      string                    `json:"name"`
		Summary   intervals.SummaryWithCats `json:"summary"`
		WantCount float64                   `json:"want_count"`
		WantLoad  float64                   `json:"want_load"`
		WantCTL   float64                   `json:"want_ctl"`
	}](t, "aggregates.json")
	compliance := readCalendarFixture[calendarComplianceFixture](t, "compliance.json")
	calendar := readCalendarFixture[calendarEventsFixture](t, "events.json")
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			client := newFakeComputeClient()
			client.summaries = []intervals.SummaryWithCats{tc.Summary}
			client.activities = compliance.Activities
			client.events = append(append([]intervals.Event(nil), compliance.Events...), calendar.MultiDayRace...)
			args, err := json.Marshal(map[string]any{"start_date": tc.Summary.Date, "end_date": tc.Summary.Date, "include_full": true})
			if err != nil {
				t.Fatal(err)
			}
			summaryTool := newGetTrainingSummaryTool(client, client, "test", "UTC", false)
			result, err := summaryTool.Handler(context.Background(), Request{Name: summaryTool.Name, Arguments: args})
			if err != nil {
				t.Fatal(err)
			}
			summary := resultMap(t, result)["summary"].(map[string]any)
			if summary["count"] != tc.WantCount || summary["training_load"] != tc.WantLoad {
				t.Fatalf("summary = %#v, want count %v and load %v", summary, tc.WantCount, tc.WantLoad)
			}
			balanceTool := newComputeLoadBalanceTool(client, client, client, client, "test", "UTC", false)
			result, err = balanceTool.Handler(context.Background(), Request{Name: balanceTool.Name, Arguments: args})
			if err != nil {
				t.Fatal(err)
			}
			balance := resultMap(t, result)["result"].(map[string]any)
			if balance["training_load_total"] != tc.WantLoad {
				t.Fatalf("load balance = %#v, want load %v", balance, tc.WantLoad)
			}
			fitnessTool := newGetFitnessTool(client, client, "test", "UTC", false)
			result, err = fitnessTool.Handler(context.Background(), Request{Name: fitnessTool.Name, Arguments: args})
			if err != nil {
				t.Fatal(err)
			}
			rows := resultMap(t, result)["fitness"].([]any)
			if len(rows) != 1 || rows[0].(map[string]any)["ctl"] != tc.WantCTL {
				t.Fatalf("fitness = %#v, want one upstream CTL %v", rows, tc.WantCTL)
			}
			if client.activityCalls != 0 || client.summaryCalls != 3 {
				t.Fatalf("activity reads=%d summary reads=%d, want only upstream aggregates", client.activityCalls, client.summaryCalls)
			}
		})
	}
}

type calendarIdentityTransport func(*http.Request) (*http.Response, error)

func (transport calendarIdentityTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return transport(req)
}

func TestCalendarReconciliationDuplicateIdentityMustBeUnambiguous(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		ids         []string
		nullRow     bool
		distinctRow bool
		wantError   bool
	}{
		{name: "missing identity", ids: []string{""}, wantError: true},
		{name: "whitespace identity", ids: []string{"  "}, wantError: true},
		{name: "two identities", ids: []string{"existing-1", "existing-2"}, wantError: true},
		{name: "valid then missing identity", ids: []string{"existing-1", ""}, wantError: true},
		{name: "missing then valid identity", ids: []string{"", "existing-1"}, wantError: true},
		{name: "null row", nullRow: true, wantError: true},
		{name: "valid then null row", ids: []string{"existing-1"}, nullRow: true, wantError: true},
		{name: "one identity", ids: []string{"existing-1"}},
		{name: "repeated same identity", ids: []string{"existing-1", "existing-1"}},
		{name: "distinct same-day session", ids: []string{"existing-1"}, distinctRow: true},
	}
	for _, tc := range cases {
		for _, match := range []string{"exact shape", "external ID"} {
			for _, phase := range []string{"preflight", "recovery"} {
				t.Run(tc.name+"/"+match+"/"+phase, func(t *testing.T) {
					t.Parallel()
					fixture := readCalendarFixture[interruptedCreateFixture](t, "interrupted_create.json")
					var args map[string]any
					if err := json.Unmarshal(fixture.Request, &args); err != nil {
						t.Fatal(err)
					}
					if match == "exact shape" {
						delete(args, "external_id")
					}
					requestBody, err := json.Marshal(args)
					if err != nil {
						t.Fatal(err)
					}
					rows := make([]map[string]any, 0, len(tc.ids)+1)
					for _, id := range tc.ids {
						row := make(map[string]any, len(fixture.CommittedEvent.Raw))
						for key, value := range fixture.CommittedEvent.Raw {
							row[key] = value
						}
						delete(row, "id")
						if id != "" {
							row["id"] = id
						}
						if match == "external ID" {
							row["name"] = "Edited upstream title"
						}
						rows = append(rows, row)
					}
					if tc.nullRow {
						rows = append(rows, nil)
					}
					if tc.distinctRow {
						rows = append(rows, map[string]any{"id": "another-session", "category": "WORKOUT", "type": "Ride", "name": "Evening ride", "start_date_local": "2026-06-13T18:00:00"})
					}
					readBody, err := json.Marshal(rows)
					if err != nil {
						t.Fatal(err)
					}
					writes, reads := 0, 0
					transport := calendarIdentityTransport(func(req *http.Request) (*http.Response, error) {
						body := string(readBody)
						switch req.Method {
						case http.MethodGet:
							reads++
							if req.URL.Path != "/athlete/i12345/events" || req.URL.Query().Get("oldest") != args["date"] || req.URL.Query().Get("newest") != args["date"] {
								t.Fatalf("unexpected calendar read: %s", req.URL)
							}
							if phase == "recovery" && reads == 1 {
								body = "[]"
							}
						case http.MethodPost:
							writes++
							if req.URL.Path != "/athlete/i12345/events/bulk" {
								t.Fatalf("unexpected write: %s", req.URL)
							}
							body = "[null]"
						default:
							t.Fatalf("unexpected method: %s", req.Method)
						}
						return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
					})
					client, err := intervals.NewClient(intervals.Options{Config: config.Config{APIKey: "test-key", AthleteID: "i12345", APIBaseURL: "https://example.invalid", HTTPTimeout: time.Second}, HTTPClient: &http.Client{Transport: transport}, Version: "test"})
					if err != nil {
						t.Fatal(err)
					}
					profile := &fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", Timezone: "UTC"}}
					tool := newAddOrUpdateEventTool(client, profile, "test", "UTC", false)
					wantReads, wantWrites := 1, 0
					if phase == "recovery" {
						wantReads, wantWrites = 2, 1
					}
					for attempt := 0; attempt < 2; attempt++ {
						result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: requestBody})
						if tc.wantError {
							message, public := PublicErrorMessage(err)
							if !public || !strings.Contains(message, "inspect") || !strings.Contains(message, "event_id") {
								t.Fatalf("result=%#v error=%v; want actionable ambiguous-identity error", result.StructuredContent, err)
							}
						} else {
							if err != nil {
								t.Fatal(err)
							}
							meta := resultMap(t, result)["_meta"].(map[string]any)
							wantStatus := skippedExistingEventStatus
							if phase == "recovery" && attempt == 0 {
								wantStatus = verifiedAfterInterruptedWriteStatus
							}
							if meta["confirmation_status"] != wantStatus || meta["duplicate_event_id"] != "existing-1" {
								t.Fatalf("meta=%#v, want %s for existing-1", meta, wantStatus)
							}
						}
						if reads != wantReads+attempt || writes != wantWrites {
							t.Fatalf("reads=%d writes=%d; want %d/%d with no second write", reads, writes, wantReads+attempt, wantWrites)
						}
					}
				})
			}
		}
	}
}

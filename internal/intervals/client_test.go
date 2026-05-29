package intervals

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ricardocabral/icuvisor/internal/config"
)

func TestDoJSONSetsAuthUserAgentPathAndDecodes(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/athlete/i12345"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		username, password, ok := r.BasicAuth()
		if !ok || username != basicAuthUsername || password != "x" {
			t.Fatalf("basic auth = (%q, %q, %v), want configured API key", username, password, ok)
		}
		if got, want := r.UserAgent(), "icuvisor/v0.1-test"; got != want {
			t.Fatalf("User-Agent = %q, want %q", got, want)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"i12345","name":"Example Athlete","timezone":"UTC","sportSettings":[{"id":7,"athlete_id":"i12345","types":["Ride"],"ftp":250,"power_zones":[100,200]}]}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	var got AthleteWithSportSettings
	if err := client.doJSON(context.Background(), &got, "athlete", client.athleteID); err != nil {
		t.Fatalf("doJSON() error = %v", err)
	}
	if got.ID != "i12345" || got.Name != "Example Athlete" || len(got.SportSettings) != 1 || got.SportSettings[0].FTP != 250 {
		t.Fatalf("decoded athlete = %+v", got)
	}
}

func TestDoJSONUsesContextTargetAthleteWithoutMutatingClient(t *testing.T) {
	t.Parallel()

	paths := make(chan string, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths <- r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"ok"}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	var got map[string]any
	if err := client.doJSON(WithTargetAthleteID(context.Background(), "i67890"), &got, "athlete", client.athleteID); err != nil {
		t.Fatalf("doJSON() with target error = %v", err)
	}
	if err := client.doJSON(context.Background(), &got, "athlete", client.athleteID); err != nil {
		t.Fatalf("doJSON() default error = %v", err)
	}
	if first := <-paths; first != "/athlete/i67890" {
		t.Fatalf("first path = %q, want target athlete", first)
	}
	if second := <-paths; second != "/athlete/i12345" {
		t.Fatalf("second path = %q, want original configured athlete", second)
	}
}

func TestActivityIDEndpointsRequireResolvedTargetOwnership(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		call func(context.Context, *Client) error
	}{
		{name: "get activity", call: func(ctx context.Context, c *Client) error { _, err := c.GetActivity(ctx, "a1"); return err }},
		{name: "streams", call: func(ctx context.Context, c *Client) error {
			_, err := c.GetActivityStreams(ctx, ActivityStreamsParams{ActivityID: "a1"})
			return err
		}},
		{name: "add message", call: func(ctx context.Context, c *Client) error {
			_, err := c.AddActivityMessage(ctx, AddActivityMessageParams{ActivityID: "a1", Content: "nice"})
			return err
		}},
		{name: "link event", call: func(ctx context.Context, c *Client) error {
			_, err := c.LinkActivityToEvent(ctx, LinkActivityToEventParams{ActivityID: "a1", EventID: "42"})
			return err
		}},
		{name: "delete activity", call: func(ctx context.Context, c *Client) error { return c.DeleteActivity(ctx, "a1") }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.Method == http.MethodGet && r.URL.Path == "/activity/a1" {
					_, _ = w.Write([]byte(`{"id":"a1","icu_athlete_id":"i222"}`))
					return
				}
				switch {
				case r.Method == http.MethodGet && r.URL.Path == "/athlete/i222/events/42":
					_, _ = w.Write([]byte(`{"id":"42"}`))
				case r.Method == http.MethodGet && r.URL.Path == "/activity/a1/streams":
					_, _ = w.Write([]byte(`[]`))
				case r.Method == http.MethodPost && r.URL.Path == "/activity/a1/messages":
					_, _ = w.Write([]byte(`{"id":1}`))
				case r.Method == http.MethodPut && r.URL.Path == "/activity/a1":
					_, _ = w.Write([]byte(`{"id":"a1","icu_athlete_id":"i222"}`))
				case r.Method == http.MethodDelete && r.URL.Path == "/activity/a1/tombstone":
					w.WriteHeader(http.StatusNoContent)
				default:
					t.Fatalf("unexpected request %s %s", r.Method, r.URL.String())
				}
			}))
			defer server.Close()

			client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
			if err := tc.call(WithTargetAthleteID(context.Background(), "i222"), client); err != nil {
				t.Fatalf("matching target call error = %v", err)
			}
			if err := tc.call(WithTargetAthleteID(context.Background(), "i333"), client); !errors.Is(err, ErrTargetAthleteMismatch) {
				t.Fatalf("mismatched target error = %v, want ErrTargetAthleteMismatch", err)
			}
		})
	}
}

func TestLinkActivityToEventPreflightsTargetEventBeforeWrite(t *testing.T) {
	t.Parallel()

	var sawPUT atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/activity/a1":
			_, _ = w.Write([]byte(`{"id":"a1","icu_athlete_id":"i222"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/athlete/i222/events/404":
			w.WriteHeader(http.StatusNotFound)
		case r.Method == http.MethodPut && r.URL.Path == "/activity/a1":
			sawPUT.Store(true)
			_, _ = w.Write([]byte(`{"id":"a1","icu_athlete_id":"i222"}`))
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	_, err := client.LinkActivityToEvent(WithTargetAthleteID(context.Background(), "i222"), LinkActivityToEventParams{ActivityID: "a1", EventID: "404"})
	if err == nil {
		t.Fatal("LinkActivityToEvent() error = nil, want event preflight failure")
	}
	if sawPUT.Load() {
		t.Fatal("LinkActivityToEvent() sent PUT after event preflight failed")
	}
}

func TestSportSettingsDecodesPaceUnitsVerbatim(t *testing.T) {
	t.Parallel()

	var got AthleteWithSportSettings
	if err := json.Unmarshal([]byte(`{"sportSettings":[{"pace_units":"SECS_100M"},{"pace_units":"mins_km"}]}`), &got); err != nil {
		t.Fatalf("unmarshal profile: %v", err)
	}
	if len(got.SportSettings) != 2 {
		t.Fatalf("sport settings = %d, want 2", len(got.SportSettings))
	}
	if got.SportSettings[0].PaceUnits != "SECS_100M" || got.SportSettings[1].PaceUnits != "mins_km" {
		t.Fatalf("pace units decoded = %#v, want verbatim upstream strings", got.SportSettings)
	}
}

func TestDoJSONRetriesRateLimitAndServerErrorsForGET(t *testing.T) {
	t.Parallel()

	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempt := atomic.AddInt32(&attempts, 1)
		switch attempt {
		case 1:
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
		case 2:
			w.WriteHeader(http.StatusBadGateway)
		default:
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"i12345"}`))
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{MaxAttempts: 3, BaseDelay: time.Nanosecond, MaxDelay: time.Millisecond})
	var got AthleteWithSportSettings
	if err := client.doJSON(context.Background(), &got, "athlete", client.athleteID); err != nil {
		t.Fatalf("doJSON() error = %v", err)
	}
	if attempts != 3 {
		t.Fatalf("attempts = %d, want 3", attempts)
	}
}

func TestDoJSONDoesNotRetryClientErrorsAndClassifiesStatus(t *testing.T) {
	t.Parallel()

	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not here"}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{MaxAttempts: 3, BaseDelay: time.Nanosecond, MaxDelay: time.Millisecond})
	var got AthleteWithSportSettings
	err := client.doJSON(context.Background(), &got, "athlete", client.athleteID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("doJSON() error = %v, want ErrNotFound", err)
	}
	var apiErr *Error
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusNotFound {
		t.Fatalf("doJSON() structured error = %#v, want 404", apiErr)
	}
	if strings.Contains(err.Error(), "x") || strings.Contains(err.Error(), "not here") {
		t.Fatalf("error %q leaked secret or response body", err)
	}
	if attempts != 1 {
		t.Fatalf("attempts = %d, want 1", attempts)
	}
}

func TestGetAthleteProfileDecodesFixture(t *testing.T) {
	t.Parallel()

	fixture, err := os.ReadFile("testdata/athlete_profile.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/athlete/i12345"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fixture)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	got, err := client.GetAthleteProfile(context.Background())
	if err != nil {
		t.Fatalf("GetAthleteProfile() error = %v", err)
	}
	if got.ID != "i12345" || got.Timezone != "America/Sao_Paulo" || got.MeasurementPreference != "METRIC" {
		t.Fatalf("profile identity/units = %+v", got)
	}
	if len(got.SportSettings) != 1 || got.SportSettings[0].IndoorFTP != 240 || got.SportSettings[0].PaceUnits != "MINS_KM" {
		t.Fatalf("sport settings = %+v", got.SportSettings)
	}
}

func TestDoJSONClosesResponseBody(t *testing.T) {
	t.Parallel()

	var closed atomic.Bool
	body := &closeTrackingBody{Reader: strings.NewReader(`{"id":"i12345"}`), closed: &closed}
	httpClient := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       body,
		}, nil
	})}
	client := newTestClient(t, "https://example.invalid", httpClient, RetryConfig{})

	var got AthleteWithSportSettings
	if err := client.doJSON(context.Background(), &got, "athlete", client.athleteID); err != nil {
		t.Fatalf("doJSON() error = %v", err)
	}
	if !closed.Load() {
		t.Fatal("response body was not closed")
	}
}

func TestDoJSONClosesResponseBodyAcrossPaths(t *testing.T) {
	t.Parallel()

	quickRetry := RetryConfig{MaxAttempts: 2, BaseDelay: time.Nanosecond, MaxDelay: time.Nanosecond}
	tests := []struct {
		name       string
		ctx        func() (context.Context, context.CancelFunc)
		retry      RetryConfig
		respond    func(int32, *atomic.Int32) (*http.Response, error)
		wantErr    error
		wantClosed int32
	}{
		{
			name:  "success",
			retry: RetryConfig{MaxAttempts: 1},
			respond: func(_ int32, closed *atomic.Int32) (*http.Response, error) {
				return testResponse(http.StatusOK, `{"id":"i12345"}`, closed), nil
			},
			wantClosed: 1,
		},
		{
			name:  "retry then success",
			retry: quickRetry,
			respond: func(attempt int32, closed *atomic.Int32) (*http.Response, error) {
				if attempt == 1 {
					return testResponse(http.StatusServiceUnavailable, `temporary`, closed), nil
				}
				return testResponse(http.StatusOK, `{"id":"i12345"}`, closed), nil
			},
			wantClosed: 2,
		},
		{
			name:  "retry exhaustion",
			retry: quickRetry,
			respond: func(_ int32, closed *atomic.Int32) (*http.Response, error) {
				return testResponse(http.StatusServiceUnavailable, `temporary`, closed), nil
			},
			wantErr:    ErrUpstream,
			wantClosed: 2,
		},
		{
			name:  "oversize",
			retry: RetryConfig{MaxAttempts: 1},
			respond: func(_ int32, closed *atomic.Int32) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       &countingReadCloser{Reader: io.LimitReader(zeroReader{}, maxResponseBodyBytes+1), closed: closed},
				}, nil
			},
			wantErr:    ErrResponseTooLarge,
			wantClosed: 1,
		},
		{
			name:  "4xx",
			retry: quickRetry,
			respond: func(_ int32, closed *atomic.Int32) (*http.Response, error) {
				return testResponse(http.StatusBadRequest, `bad request`, closed), nil
			},
			wantErr:    ErrUpstream,
			wantClosed: 1,
		},
		{
			name: "context canceled with response",
			ctx: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx, func() {}
			},
			retry: quickRetry,
			respond: func(_ int32, closed *atomic.Int32) (*http.Response, error) {
				return testResponse(http.StatusServiceUnavailable, `temporary`, closed), nil
			},
			wantErr:    ErrUpstream,
			wantClosed: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var attempts atomic.Int32
			var closed atomic.Int32
			httpClient := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return tc.respond(attempts.Add(1), &closed)
			})}
			client := newTestClient(t, "https://example.invalid", httpClient, tc.retry)
			ctx := context.Background()
			cancel := func() {}
			if tc.ctx != nil {
				ctx, cancel = tc.ctx()
			}
			defer cancel()

			var got AthleteWithSportSettings
			err := client.doJSON(ctx, &got, "athlete", client.athleteID)
			if tc.wantErr == nil && err != nil {
				t.Fatalf("doJSON() error = %v", err)
			}
			if tc.wantErr != nil && !errors.Is(err, tc.wantErr) {
				t.Fatalf("doJSON() error = %v, want errors.Is %v", err, tc.wantErr)
			}
			if got := closed.Load(); got != tc.wantClosed {
				t.Fatalf("closed bodies = %d, want %d", got, tc.wantClosed)
			}
		})
	}
}

func TestDoJSONOversizeBodyReturnsSentinel(t *testing.T) {
	t.Parallel()

	var closed atomic.Int32
	httpClient := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       &countingReadCloser{Reader: io.LimitReader(zeroReader{}, maxResponseBodyBytes+1), closed: &closed},
		}, nil
	})}
	client := newTestClient(t, "https://example.invalid", httpClient, RetryConfig{MaxAttempts: 1})

	var got AthleteWithSportSettings
	err := client.doJSON(context.Background(), &got, "athlete", client.athleteID)
	if !errors.Is(err, ErrResponseTooLarge) {
		t.Fatalf("doJSON() error = %v, want ErrResponseTooLarge", err)
	}
	if got := closed.Load(); got != 1 {
		t.Fatalf("closed bodies = %d, want 1", got)
	}
}

func TestDoJSONRetriesRateLimitThenSucceeds(t *testing.T) {
	t.Parallel()

	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempt := atomic.AddInt32(&attempts, 1)
		if attempt == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"i12345"}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{MaxAttempts: 2, BaseDelay: time.Nanosecond, MaxDelay: time.Nanosecond})
	var got AthleteWithSportSettings
	if err := client.doJSON(context.Background(), &got, "athlete", client.athleteID); err != nil {
		t.Fatalf("doJSON() error = %v", err)
	}
	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}
}

func TestDoJSONRetriesServerErrorThenSucceeds(t *testing.T) {
	t.Parallel()

	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempt := atomic.AddInt32(&attempts, 1)
		if attempt == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"i12345"}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{MaxAttempts: 2, BaseDelay: time.Nanosecond, MaxDelay: time.Nanosecond})
	var got AthleteWithSportSettings
	if err := client.doJSON(context.Background(), &got, "athlete", client.athleteID); err != nil {
		t.Fatalf("doJSON() error = %v", err)
	}
	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}
}

func TestDoJSONDoesNotRetryBadRequest(t *testing.T) {
	t.Parallel()

	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad"}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{MaxAttempts: 3, BaseDelay: time.Nanosecond, MaxDelay: time.Nanosecond})
	var got AthleteWithSportSettings
	err := client.doJSON(context.Background(), &got, "athlete", client.athleteID)
	if !errors.Is(err, ErrUpstream) {
		t.Fatalf("doJSON() error = %v, want ErrUpstream", err)
	}
	if attempts != 1 {
		t.Fatalf("attempts = %d, want 1", attempts)
	}
}

func TestDoJSONRetryBudgetExhaustionReturnsLastStatus(t *testing.T) {
	t.Parallel()

	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`temporary`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{MaxAttempts: 2, BaseDelay: time.Nanosecond, MaxDelay: time.Nanosecond})
	var got AthleteWithSportSettings
	err := client.doJSON(context.Background(), &got, "athlete", client.athleteID)
	if !errors.Is(err, ErrUpstream) || !strings.Contains(err.Error(), "HTTP 503") {
		t.Fatalf("doJSON() error = %v, want wrapped HTTP 503 upstream error", err)
	}
	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}
}

func TestAddJitterProducesUniqueSamples(t *testing.T) {
	t.Parallel()

	const samples = 1000
	const minUnique = 900

	seen := make(map[time.Duration]struct{}, samples)
	for range samples {
		seen[addJitter(time.Second, 1)] = struct{}{}
	}
	if got := len(seen); got < minUnique {
		t.Fatalf("unique jitter samples = %d, want at least %d", got, minUnique)
	}
}

func TestRetryConfigWithDefaults(t *testing.T) {
	t.Parallel()

	zero := RetryConfig{}.WithDefaults()
	if zero.MaxAttempts != defaultMaxAttempts || zero.BaseDelay != defaultBaseDelay || zero.MaxDelay != defaultMaxDelay || zero.Jitter != defaultJitter {
		t.Fatalf("zero WithDefaults() = %+v, want package defaults", zero)
	}

	partial := RetryConfig{MaxAttempts: 5}.WithDefaults()
	if partial.MaxAttempts != 5 || partial.BaseDelay != defaultBaseDelay || partial.MaxDelay != defaultMaxDelay || partial.Jitter != 0 {
		t.Fatalf("partial WithDefaults() = %+v, want explicit max attempts, default delays, zero jitter", partial)
	}

	negativeJitter := RetryConfig{MaxAttempts: 5, Jitter: -1}.WithDefaults()
	if negativeJitter.Jitter != 0 {
		t.Fatalf("negative jitter WithDefaults() = %+v, want jitter clamped to 0", negativeJitter)
	}
}

func TestSleepBeforeRetryHonorsContextCancellation(t *testing.T) {
	t.Parallel()

	client := &Client{retry: RetryConfig{MaxAttempts: 3, BaseDelay: time.Hour, MaxDelay: time.Hour}.WithDefaults()}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := client.sleepBeforeRetry(ctx, time.Millisecond)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("sleepBeforeRetry() error = %v, want context.Canceled", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type closeTrackingBody struct {
	*strings.Reader
	closed *atomic.Bool
}

func (b *closeTrackingBody) Close() error {
	b.closed.Store(true)
	return nil
}

func testResponse(statusCode int, body string, closed *atomic.Int32) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Header:     make(http.Header),
		Body:       &countingReadCloser{Reader: strings.NewReader(body), closed: closed},
	}
}

type countingReadCloser struct {
	io.Reader
	closed *atomic.Int32
}

func (b *countingReadCloser) Close() error {
	b.closed.Add(1)
	return nil
}

type zeroReader struct{}

func (zeroReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 'x'
	}
	return len(p), nil
}

func newTestClient(t *testing.T, baseURL string, httpClient *http.Client, retry RetryConfig) *Client {
	t.Helper()
	client, err := NewClient(Options{
		Config: config.Config{
			APIKey:      "x",
			AthleteID:   "i12345",
			APIBaseURL:  baseURL,
			HTTPTimeout: time.Second,
		},
		Version:    "v0.1-test",
		HTTPClient: httpClient,
		Retry:      retry,
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	return client
}

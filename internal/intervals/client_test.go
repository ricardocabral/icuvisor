package intervals

import (
	"context"
	"errors"
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

func TestSleepBeforeRetryHonorsContextCancellation(t *testing.T) {
	t.Parallel()

	client := &Client{retry: normalizeRetryConfig(RetryConfig{MaxAttempts: 3, BaseDelay: time.Hour, MaxDelay: time.Hour})}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := client.sleepBeforeRetry(ctx, 1, 0)
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

func newTestClient(t *testing.T, baseURL string, httpClient *http.Client, retry RetryConfig) *Client {
	t.Helper()
	client, err := NewClient(Options{
		Config: config.Config{
			APIKey:      "x",
			AthleteID:   "12345",
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

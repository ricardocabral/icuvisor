package intervals

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestActivityStreamUnmarshalPreservesRawFields(t *testing.T) {
	t.Parallel()

	var got ActivityStream
	if err := json.Unmarshal([]byte(`{"type":"Power","name":"Watts","data":[250,260.5],"data2":[1,2],"valueTypeIsArray":true,"anomalies":[3,5],"custom":true,"allNull":false,"extra":{"unit":"W"}}`), &got); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}
	if got.Type != "Power" || got.Name != "Watts" || len(got.Data) != 2 || got.Data[1] != 260.5 || len(got.Data2) != 2 {
		t.Fatalf("typed stream = %+v", got)
	}
	if !got.ValueTypeIsArray || !got.Custom || got.AllNull || len(got.Anomalies) != 2 {
		t.Fatalf("flags/anomalies = %+v", got)
	}
	extra, ok := got.Raw["extra"].(map[string]any)
	if !ok || extra["unit"] != "W" {
		t.Fatalf("Raw extra = %#v, want preserved upstream fields", got.Raw["extra"])
	}
}

func TestGetActivityStreamsBuildsQuery(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/activity/a1/streams"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		query := r.URL.Query()
		if got, want := query.Get("types"), "watts,heartrate"; got != want {
			t.Fatalf("types query = %q, want %q", got, want)
		}
		if got, want := query.Get("includeDefaults"), "true"; got != want {
			t.Fatalf("includeDefaults query = %q, want %q", got, want)
		}
		if got := query.Get("include_defaults"); got != "" {
			t.Fatalf("include_defaults query = %q, want absent", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"type":"watts","data":[1],"upstream_extra":"kept"}]`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	got, err := client.GetActivityStreams(context.Background(), ActivityStreamsParams{ActivityID: " a1 ", Types: []string{" watts ", "", "heartrate"}, IncludeDefaults: true})
	if err != nil {
		t.Fatalf("GetActivityStreams() error = %v", err)
	}
	if len(got) != 1 || got[0].Type != "watts" || got[0].Raw["upstream_extra"] != "kept" {
		t.Fatalf("streams = %#v", got)
	}
}

func TestGetActivityStreamsRequiresActivityID(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, "https://example.invalid", http.DefaultClient, RetryConfig{})
	_, err := client.GetActivityStreams(context.Background(), ActivityStreamsParams{ActivityID: " \t "})
	if err == nil || !strings.Contains(err.Error(), "activity ID is required") {
		t.Fatalf("GetActivityStreams() error = %v, want required activity ID", err)
	}
}

func TestGetActivityStreamsWrapsHTTPError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{MaxAttempts: 1})
	_, err := client.GetActivityStreams(context.Background(), ActivityStreamsParams{ActivityID: "a1"})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("GetActivityStreams() error = %v, want ErrUnauthorized", err)
	}
	if !strings.Contains(err.Error(), "getting activity a1 streams") {
		t.Fatalf("GetActivityStreams() error = %q, want activity context", err.Error())
	}
}

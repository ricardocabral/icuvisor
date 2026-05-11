package intervals

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestGetActivityMessagesSendsQueryAndPreservesRawNulls(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/activity/a123/messages"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		if got := r.URL.Query().Get("sinceId"); got != "10" {
			t.Fatalf("sinceId = %q, want 10", got)
		}
		if got := r.URL.Query().Get("limit"); got != "25" {
			t.Fatalf("limit = %q, want 25", got)
		}
		fixture, err := os.ReadFile("testdata/activity_messages.json")
		if err != nil {
			t.Fatalf("read fixture: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fixture)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	messages, err := client.GetActivityMessages(context.Background(), ActivityMessagesParams{ActivityID: "a123", SinceID: 10, Limit: 25})
	if err != nil {
		t.Fatalf("GetActivityMessages() error = %v", err)
	}
	if len(messages) != 1 || messages[0].ID != 11 {
		t.Fatalf("messages = %#v, want one message", messages)
	}
	if value, ok := messages[0].Raw["extra"]; !ok || value != nil {
		rawJSON, _ := json.Marshal(messages[0].Raw)
		t.Fatalf("raw extra = %#v present %v raw=%s, want nil", value, ok, rawJSON)
	}
}

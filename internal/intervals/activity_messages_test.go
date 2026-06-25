package intervals

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestAddActivityMessagePostsAppendOnlyContent(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPost; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/activity/a123/messages"; got != want {
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
		if decoded["content"] != "  keep body verbatim  " || len(decoded) != 1 {
			t.Fatalf("body = %#v, want content only", decoded)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":99,"new_chat":{"id":7},"extra":null}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	message, err := client.AddActivityMessage(context.Background(), AddActivityMessageParams{ActivityID: " a123 ", Content: "  keep body verbatim  "})
	if err != nil {
		t.Fatalf("AddActivityMessage() error = %v", err)
	}
	if message.ID != 99 || message.Raw["extra"] != nil {
		t.Fatalf("message = %+v raw=%#v, want decoded raw response", message, message.Raw)
	}
}

func TestAddActivityMessageRequiresInputs(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, "https://example.invalid", http.DefaultClient, RetryConfig{})
	for _, params := range []AddActivityMessageParams{{Content: "msg"}, {ActivityID: "a1"}, {ActivityID: "a1", Content: "   "}} {
		if _, err := client.AddActivityMessage(context.Background(), params); err == nil {
			t.Fatalf("AddActivityMessage(%#v) error = nil, want validation error", params)
		}
	}
}

func TestGetActivityMessagesSendsQueryAndPreservesRawNulls(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/activity/a123/messages"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		query := r.URL.Query()
		if got := query.Get("sinceId"); got != "10" {
			t.Fatalf("sinceId = %q, want 10", got)
		}
		if got := query.Get("limit"); got != "25" {
			t.Fatalf("limit = %q, want 25", got)
		}
		if got := query.Get("since_id"); got != "" {
			t.Fatalf("since_id = %q, want absent", got)
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

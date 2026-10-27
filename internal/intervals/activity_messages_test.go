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

func TestActivityMessageDeletedTimestamp(t *testing.T) {
	for _, tc := range []struct {
		name, raw string
		deleted   *bool
		timestamp string
	}{
		{name: "absent", raw: `{}`}, {name: "null", raw: `{"deleted":null}`},
		{name: "timestamp", raw: `{"deleted":"2026-10-21T10:00:00Z"}`, deleted: boolPointer(true), timestamp: "2026-10-21T10:00:00Z"},
		{name: "legacy false", raw: `{"deleted":false}`, deleted: boolPointer(false)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var m ActivityMessage
			if err := json.Unmarshal([]byte(tc.raw), &m); err != nil {
				t.Fatal(err)
			}
			if m.DeletedAt != tc.timestamp {
				t.Fatalf("timestamp=%q want %q", m.DeletedAt, tc.timestamp)
			}
			if (m.Deleted == nil) != (tc.deleted == nil) || (m.Deleted != nil && *m.Deleted != *tc.deleted) {
				t.Fatalf("deleted=%v want %v", m.Deleted, tc.deleted)
			}
		})
	}
}
func boolPointer(v bool) *bool { return &v }

func TestChangeActivityMessageSafetyAndRouting(t *testing.T) {
	for _, tc := range []struct {
		name, activity, message string
		remove                  bool
		status                  int
		wantError               bool
		wantStatus              string
		wantWrites              int
	}{
		{name: "edit own", wantStatus: "updated", wantWrites: 1},
		{name: "delete own", remove: true, wantStatus: "deleted", wantWrites: 1},
		{name: "wrong activity", message: `{"id":11,"athlete_id":"i12345","activity_id":"other","content":"private"}`, wantError: true},
		{name: "other author", message: `{"id":11,"athlete_id":"i999","activity_id":"a1","content":"private"}`, wantError: true},
		{name: "wrong athlete", activity: `{"id":"a1","icu_athlete_id":"i999","icu_chat_id":7}`, wantError: true},
		{name: "no chat", activity: `{"id":"a1","icu_athlete_id":"i12345"}`, wantError: true},
		{name: "missing message", message: `{"id":12,"athlete_id":"i12345","activity_id":"a1"}`, wantError: true},
		{name: "already deleted", message: `{"id":11,"athlete_id":"i12345","activity_id":"a1","deleted":"2026-10-21T10:00:00Z"}`, remove: true, wantStatus: "already_deleted"},
		{name: "edit deleted", message: `{"id":11,"athlete_id":"i12345","activity_id":"a1","deleted":"2026-10-21T10:00:00Z"}`, wantError: true},
		{name: "forbidden", status: 403, wantError: true, wantWrites: 1},
		{name: "not found", status: 404, wantError: true, wantWrites: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			writes := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch r.URL.Path {
				case "/athlete/i12345", "/athlete/0":
					_, _ = w.Write([]byte(`{"id":"i12345","timezone":"America/Sao_Paulo"}`))
				case "/activity/a1":
					body := tc.activity
					if body == "" {
						body = `{"id":"a1","icu_athlete_id":"i12345","icu_chat_id":7}`
					}
					_, _ = w.Write([]byte(body))
				case "/activity/a1/messages":
					if r.URL.Query().Get("sinceId") != "10" {
						t.Errorf("query=%v", r.URL.Query())
					}
					body := tc.message
					if body == "" {
						body = `{"id":11,"athlete_id":"i12345","activity_id":"a1","content":"original"}`
					}
					_, _ = w.Write([]byte("[" + body + "]"))
				case "/chats/7/messages/11":
					writes++
					if tc.remove {
						if r.Method != http.MethodDelete {
							t.Errorf("method=%s", r.Method)
						}
					} else {
						if r.Method != http.MethodPut {
							t.Errorf("method=%s", r.Method)
						}
						var body map[string]any
						if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
							t.Error(err)
						}
						if len(body) != 1 || body["content"] != "corrected" {
							t.Errorf("payload=%v", body)
						}
					}
					if tc.status != 0 {
						w.WriteHeader(tc.status)
					}
					_, _ = w.Write([]byte(`{}`))
				default:
					t.Errorf("unexpected path %s", r.URL.Path)
					w.WriteHeader(404)
				}
			}))
			defer server.Close()
			client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
			params := ChangeActivityMessageParams{ActivityID: " a1 ", MessageID: 11, Content: "corrected"}
			var got ActivityMessageChange
			var err error
			if tc.remove {
				got, err = client.DeleteActivityMessage(context.Background(), params)
			} else {
				got, err = client.UpdateActivityMessage(context.Background(), params)
			}
			if (err != nil) != tc.wantError {
				t.Fatalf("result=%+v error=%v wantError=%v", got, err, tc.wantError)
			}
			if writes != tc.wantWrites {
				t.Fatalf("writes=%d want %d", writes, tc.wantWrites)
			}
			if err == nil && got.Timezone != "America/Sao_Paulo" {
				t.Errorf("selected athlete timezone=%q", got.Timezone)
			}
			if err == nil && got.Status != tc.wantStatus {
				t.Errorf("status=%s", got.Status)
			}
			if err == nil && tc.wantWrites == 1 && got.Before.Content != "original" {
				t.Errorf("recovery=%+v", got.Before)
			}
		})
	}
}

func TestCoachCommentUsesPrincipalNotSelectedAthlete(t *testing.T) {
	for _, author := range []string{"i12345", "i67890"} {
		t.Run(author, func(t *testing.T) {
			writes := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch r.URL.Path {
				case "/athlete/0":
					_, _ = w.Write([]byte(`{"id":"i12345"}`))
				case "/athlete/i67890":
					_, _ = w.Write([]byte(`{"id":"i67890"}`))
				case "/activity/a1":
					_, _ = w.Write([]byte(`{"id":"a1","icu_athlete_id":"i67890","icu_chat_id":7}`))
				case "/activity/a1/messages":
					_, _ = w.Write([]byte(`[{"id":11,"athlete_id":"` + author + `","activity_id":"a1","content":"before","type":"TEXT"}]`))
				case "/chats/7/messages/11":
					writes++
					_, _ = w.Write([]byte(`{}`))
				default:
					t.Errorf("unexpected path %s", r.URL.Path)
					w.WriteHeader(404)
				}
			}))
			defer server.Close()
			client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
			client.athleteID = "0"
			_, err := client.UpdateActivityMessage(WithTargetAthleteID(context.Background(), "i67890"), ChangeActivityMessageParams{ActivityID: "a1", MessageID: 11, Content: "after"})
			if author == "i12345" {
				if err != nil || writes != 1 {
					t.Fatalf("err=%v writes=%d", err, writes)
				}
			} else {
				if err == nil || writes != 0 {
					t.Fatalf("err=%v writes=%d", err, writes)
				}
			}
		})
	}
}

func TestMixedActivityMessageDeletionStates(t *testing.T) {
	var messages []ActivityMessage
	if err := json.Unmarshal([]byte(`[{"id":1,"content":"live","deleted":null},{"id":2,"content":"(deleted)","deleted":"2026-10-21T10:00:00Z"},{"id":3,"content":"absent"}]`), &messages); err != nil {
		t.Fatal(err)
	}
	if len(messages) != 3 || messages[0].Deleted != nil || messages[1].Deleted == nil || !*messages[1].Deleted || messages[1].DeletedAt != "2026-10-21T10:00:00Z" || messages[2].Deleted != nil {
		t.Fatalf("messages=%+v", messages)
	}
}

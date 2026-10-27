package tools

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/intervals"
)

func TestAddSharedWorkoutFolderTargetAndRedaction(t *testing.T) {
	t.Parallel()
	client := &fakeSharedLibraryClient{folder: decodeToolWorkoutFolders(t, `{"id":10,"athlete_id":"i999","name":"Base","shareToken":"secret","owner":{"email":"private@example.test"}}`)[0]}
	tool := newAddSharedWorkoutFolderTool(client, "test", false)
	result, err := tool.Handler(intervals.WithTargetAthleteID(context.Background(), "i222"), Request{Arguments: json.RawMessage(`{"folder_id":"10"}`)})
	if err != nil {
		t.Fatal(err)
	}
	if client.target != "i222" || client.folderID != "10" {
		t.Fatalf("target %s folder %s", client.target, client.folderID)
	}
	out := resultMap(t, result)
	row := out["folder"].(map[string]any)
	if row["folder_id"] != "10" || out["_meta"].(map[string]any)["verified_in_library"] != true {
		t.Fatalf("output = %#v", out)
	}
	if _, ok := row["shareToken"]; ok {
		t.Fatal("share token leaked")
	}
	if _, ok := row["owner"]; ok {
		t.Fatal("owner leaked")
	}
}
func TestAddSharedWorkoutFolderRejectsInvalidAndUnverified(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{`{}`, `{"folder_id":"../10"}`, `{"folder_id":"0"}`, `{"folder_id":"10","athlete_id":"i333"}`, `{"folder_id":"10","include_full":true}`} {
		t.Run(raw, func(t *testing.T) {
			client := &fakeSharedLibraryClient{}
			tool := newAddSharedWorkoutFolderTool(client, "test", false)
			if _, err := tool.Handler(context.Background(), Request{Arguments: json.RawMessage(raw)}); err == nil {
				t.Fatal("wanted error")
			}
			if client.calls != 0 {
				t.Fatal("invalid request reached client")
			}
		})
	}
	for _, id := range []string{"", "11"} {
		client := &fakeSharedLibraryClient{folder: intervals.WorkoutFolder{ID: id}}
		tool := newAddSharedWorkoutFolderTool(client, "test", false)
		if _, err := tool.Handler(context.Background(), Request{Arguments: json.RawMessage(`{"folder_id":"10"}`)}); err == nil {
			t.Fatal("wrong folder reported success")
		}
	}
	for _, cause := range []error{context.Canceled, context.DeadlineExceeded, errors.New("403 details")} {
		client := &fakeSharedLibraryClient{err: cause}
		tool := newAddSharedWorkoutFolderTool(client, "test", false)
		if _, err := tool.Handler(context.Background(), Request{Arguments: json.RawMessage(`{"folder_id":"10"}`)}); !errors.Is(err, cause) {
			t.Fatalf("error = %v, want %v", err, cause)
		}
	}
}

type sharedLibraryInterruptedTransport struct {
	cause     error
	transport http.RoundTripper
	cancel    context.CancelFunc
}

func (r sharedLibraryInterruptedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Method == http.MethodGet {
		if errors.Is(r.cause, context.Canceled) {
			r.cancel()
		}
		return nil, r.cause
	}
	return r.transport.RoundTrip(req)
}

func TestAddSharedWorkoutFolderInterruptedReadbackRetainsRecoveryGuidance(t *testing.T) {
	t.Parallel()
	for _, cause := range []error{context.Canceled, context.DeadlineExceeded} {
		t.Run(cause.Error(), func(t *testing.T) {
			var posts atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != "/athlete/i12345/folders/10/share-with-me" {
					t.Errorf("request = %s %s", r.Method, r.URL.Path)
				}
				posts.Add(1)
				_, _ = io.WriteString(w, `[{"id":10}]`)
			}))
			defer server.Close()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client, err := intervals.NewClient(intervals.Options{
				Config:     config.Config{AthleteID: "i12345", APIKey: "test-key", APIBaseURL: server.URL},
				HTTPClient: &http.Client{Transport: sharedLibraryInterruptedTransport{cause: cause, transport: http.DefaultTransport, cancel: cancel}},
				Retry:      intervals.RetryConfig{MaxAttempts: 1},
			})
			if err != nil {
				t.Fatal(err)
			}
			tool := newAddSharedWorkoutFolderTool(client, "test", false)
			_, err = tool.Handler(ctx, Request{Arguments: json.RawMessage(`{"folder_id":"10"}`)})
			if !errors.Is(err, cause) {
				t.Fatalf("error = %v, want cause %v", err, cause)
			}
			message, ok := PublicErrorMessage(err)
			if !ok || !strings.Contains(message, "inspect get_workout_library before retrying") {
				t.Fatalf("public error = %q, available %t; wanted uncertain-write recovery guidance", message, ok)
			}
			if posts.Load() != 1 {
				t.Fatalf("posts = %d, want applied write without retry", posts.Load())
			}
		})
	}
}

func TestAddSharedWorkoutFolderInputExamples(t *testing.T) {
	t.Parallel()
	examples := schemaInputExamples(t, addSharedWorkoutFolderInputSchema())
	if len(examples) == 0 {
		t.Fatal("missing shared-folder write input example")
	}
}

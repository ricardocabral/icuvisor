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

func TestCustomItemsClientListsAndGetsItems(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/athlete/i12345/custom-item":
			_, _ = w.Write([]byte(`[{"id":7,"type":"FITNESS_CHART","name":"CTL Chart","content":{"series":[{"field":"ctl"}]}}]`))
		case "/athlete/i12345/custom-item/7":
			_, _ = w.Write([]byte(`{"id":7,"type":"FITNESS_CHART","name":"CTL Chart","content":{"series":[{"field":"ctl"}],"layout":{"height":240}}}`))
		default:
			t.Fatalf("path = %q, want custom-item list or detail", r.URL.Path)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	items, err := client.ListCustomItems(context.Background())
	if err != nil {
		t.Fatalf("ListCustomItems() error = %v", err)
	}
	if len(items) != 1 || items[0].ID != "7" || items[0].Content == nil {
		t.Fatalf("items = %+v, want raw list item content", items)
	}
	item, err := client.GetCustomItem(context.Background(), "7")
	if err != nil {
		t.Fatalf("GetCustomItem() error = %v", err)
	}
	content, ok := item.Content.(map[string]any)
	if !ok || content["layout"] == nil {
		t.Fatalf("content = %#v, want verbatim detail payload", item.Content)
	}
}

func TestCustomItemsClientCreatesAndUpdatesItem(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/athlete/i12345/custom-item" && r.Method == http.MethodPost:
			if body["type"] != "FITNESS_CHART" || body["name"] != "New CTL" || body["content"].(map[string]any)["layout"] == nil {
				t.Fatalf("create body = %#v, want custom item create payload", body)
			}
			_, _ = w.Write([]byte(`{"id":9,"type":"FITNESS_CHART","name":"New CTL","content":{"layout":{"height":260}}}`))
		case r.URL.Path == "/athlete/i12345/custom-item/9" && r.Method == http.MethodPut:
			if body["name"] != "Renamed" || body["type"] != nil {
				t.Fatalf("update body = %#v, want sparse custom item update payload", body)
			}
			_, _ = w.Write([]byte(`{"id":9,"type":"FITNESS_CHART","name":"Renamed","content":{"layout":{"height":260}}}`))
		default:
			t.Fatalf("request = %s %s, want custom-item create or update", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	item, err := client.CreateCustomItem(context.Background(), WriteCustomItemParams{ItemType: "FITNESS_CHART", Name: "New CTL", NameSet: true, Content: map[string]any{"layout": map[string]any{"height": 260}}, ContentSet: true})
	if err != nil {
		t.Fatalf("CreateCustomItem() error = %v", err)
	}
	if item.ID != "9" || item.Content == nil {
		t.Fatalf("item = %+v, want created custom item", item)
	}
	item, err = client.UpdateCustomItem(context.Background(), WriteCustomItemParams{ItemID: "9", Name: "Renamed", NameSet: true})
	if err != nil {
		t.Fatalf("UpdateCustomItem() error = %v", err)
	}
	if item.ID != "9" || item.Name == nil || *item.Name != "Renamed" {
		t.Fatalf("item = %+v, want updated custom item", item)
	}
}

func TestCreateCustomItem422ReturnsValidationError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		body      string
		wantField string
	}{
		{
			name:      "unknown field in content",
			body:      "Unknown field: WhoopStrain",
			wantField: "WhoopStrain",
		},
		{
			name:      "json error no field",
			body:      `{"error":"content is invalid"}`,
			wantField: "",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusUnprocessableEntity)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()

			client := newTestClient(t, server.URL, server.Client(), RetryConfig{MaxAttempts: 1})
			_, err := client.CreateCustomItem(context.Background(), WriteCustomItemParams{
				ItemType:   "INPUT_FIELD",
				Name:       "Test",
				NameSet:    true,
				Content:    map[string]any{"field": "WhoopStrain"},
				ContentSet: true,
			})
			if !errors.Is(err, ErrValidation) {
				t.Fatalf("CreateCustomItem() error = %v, want ErrValidation", err)
			}
			var ve *ValidationError
			if !errors.As(err, &ve) {
				t.Fatalf("CreateCustomItem() error = %v, want *ValidationError", err)
			}
			if ve.Field != tc.wantField {
				t.Fatalf("ValidationError.Field = %q, want %q", ve.Field, tc.wantField)
			}
			if tc.body != "" && strings.Contains(err.Error(), tc.body) {
				t.Fatalf("error %q leaked raw upstream body", err)
			}
		})
	}
}

package intervals

import (
	"context"
	"net/http"
	"net/http/httptest"
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

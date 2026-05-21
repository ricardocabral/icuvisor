package intervals

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestListGearSendsCollectionPathAndDecodesFixture(t *testing.T) {
	t.Parallel()

	var method, path string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.Path
		data, err := os.ReadFile("testdata/gear_list.json")
		if err != nil {
			t.Fatalf("read gear fixture: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	gear, err := client.ListGear(context.Background())
	if err != nil {
		t.Fatalf("ListGear() error = %v", err)
	}
	if method != http.MethodGet || path != "/athlete/i12345/gear" {
		t.Fatalf("request = %s %s, want GET gear collection path", method, path)
	}
	if len(gear) != 2 {
		t.Fatalf("gear len = %d, want 2", len(gear))
	}
	if gear[0].ID != "123" || gear[0].Name == nil || *gear[0].Name != "Race Bike" {
		t.Fatalf("gear[0] = %#v, want numeric ID normalized and name decoded", gear[0])
	}
	if gear[1].ID != "shoe-7" || gear[1].Name != nil || gear[1].Retired == nil || !*gear[1].Retired {
		t.Fatalf("gear[1] = %#v, want string ID, absent name, and retired flag", gear[1])
	}
}

func TestListGearDecodesEmptyList(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := os.ReadFile("testdata/gear_list_empty.json")
		if err != nil {
			t.Fatalf("read empty gear fixture: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	gear, err := client.ListGear(context.Background())
	if err != nil {
		t.Fatalf("ListGear() error = %v", err)
	}
	if len(gear) != 0 {
		t.Fatalf("gear len = %d, want empty", len(gear))
	}
}

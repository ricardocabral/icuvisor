package intervals

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeleteMethodsSendDeletePaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		call     func(*Client) error
		wantPath string
	}{
		{name: "event", call: func(c *Client) error { return c.DeleteEvent(context.Background(), " e-1 ") }, wantPath: "/athlete/i12345/events/e-1"},
		{name: "activity", call: func(c *Client) error { return c.DeleteActivity(context.Background(), " a-1 ") }, wantPath: "/activity/a-1/tombstone"},
		{name: "custom item singular endpoint", call: func(c *Client) error { return c.DeleteCustomItem(context.Background(), " ci-1 ") }, wantPath: "/athlete/i12345/custom-item/ci-1"},
		{name: "sport settings", call: func(c *Client) error { return c.DeleteSportSettings(context.Background(), " 7 ") }, wantPath: "/athlete/i12345/sport-settings/7"},
		{name: "gear", call: func(c *Client) error { return c.DeleteGear(context.Background(), " g-1 ") }, wantPath: "/athlete/i12345/gear/g-1"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var method, path string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				method = r.Method
				path = r.URL.Path
				w.WriteHeader(http.StatusNoContent)
			}))
			defer server.Close()

			client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
			if err := tc.call(client); err != nil {
				t.Fatalf("delete call error = %v", err)
			}
			if method != http.MethodDelete || path != tc.wantPath {
				t.Fatalf("request = %s %s, want DELETE %s", method, path, tc.wantPath)
			}
		})
	}
}

func TestDeleteMethodsRequireID(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, "https://example.invalid", http.DefaultClient, RetryConfig{})
	for _, tc := range []struct {
		name string
		call func() error
	}{
		{name: "event", call: func() error { return client.DeleteEvent(context.Background(), " ") }},
		{name: "activity", call: func() error { return client.DeleteActivity(context.Background(), " ") }},
		{name: "custom item", call: func() error { return client.DeleteCustomItem(context.Background(), " ") }},
		{name: "sport settings", call: func() error { return client.DeleteSportSettings(context.Background(), " ") }},
		{name: "gear", call: func() error { return client.DeleteGear(context.Background(), " ") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := tc.call(); err == nil {
				t.Fatal("delete error = nil, want required ID error")
			}
		})
	}
}

func TestDeleteMethodsMapNotFoundToTypedError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		call func(*Client) error
	}{
		{name: "event", call: func(c *Client) error { return c.DeleteEvent(context.Background(), "missing") }},
		{name: "activity", call: func(c *Client) error { return c.DeleteActivity(context.Background(), "missing") }},
		{name: "custom item", call: func(c *Client) error { return c.DeleteCustomItem(context.Background(), "missing") }},
		{name: "sport settings", call: func(c *Client) error { return c.DeleteSportSettings(context.Background(), "missing") }},
		{name: "gear", call: func(c *Client) error { return c.DeleteGear(context.Background(), "missing") }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			}))
			defer server.Close()

			client := newTestClient(t, server.URL, server.Client(), RetryConfig{MaxAttempts: 1})
			err := tc.call(client)
			if err == nil || !errors.Is(err, ErrNotFound) {
				t.Fatalf("delete error = %v, want errors.Is ErrNotFound", err)
			}
		})
	}
}

func TestGetGearSendsReadPath(t *testing.T) {
	t.Parallel()

	var method, path string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"g-1","name":"Race Bike","type":"Bike"}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	gear, err := client.GetGear(context.Background(), " g-1 ")
	if err != nil {
		t.Fatalf("GetGear() error = %v", err)
	}
	if method != http.MethodGet || path != "/athlete/i12345/gear/g-1" {
		t.Fatalf("request = %s %s, want GET gear path", method, path)
	}
	if gear.ID != "g-1" || gear.Name == nil || *gear.Name != "Race Bike" {
		t.Fatalf("gear = %#v, want decoded gear", gear)
	}
}

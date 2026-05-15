package mcp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/intervals"
)

func newProtocolIntervalsClient(t *testing.T, handler http.Handler) (*intervals.Client, func()) {
	t.Helper()
	server := httptest.NewServer(handler)
	client, err := intervals.NewClient(intervals.Options{
		Config: config.Config{
			APIKey:      strings.Repeat("x", 8),
			AthleteID:   "12345",
			APIBaseURL: server.URL,
			HTTPTimeout: time.Second,
		},
		Version:    "test",
		HTTPClient: server.Client(),
	})
	if err != nil {
		server.Close()
		t.Fatalf("NewClient() error = %v", err)
	}
	return client, server.Close
}

func newNoNetworkProtocolClient(t *testing.T) *intervals.Client {
	t.Helper()
	client, cleanup := newProtocolIntervalsClient(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("test intervals client must not perform HTTP during registration")
	}))
	t.Cleanup(cleanup)
	return client
}

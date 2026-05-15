package tools

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/intervals"
)

type panicRoundTripper struct{}

func (panicRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	panic("test intervals client must not perform HTTP during registration")
}

func newNoNetworkIntervalsClient(t *testing.T) *intervals.Client {
	t.Helper()
	client, err := intervals.NewClient(intervals.Options{
		Config: config.Config{
			APIKey:      strings.Repeat("x", 8),
			AthleteID:   "12345",
			APIBaseURL: "http://127.0.0.1",
			HTTPTimeout: time.Second,
		},
		Version:    "test",
		HTTPClient: &http.Client{Transport: panicRoundTripper{}},
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	return client
}

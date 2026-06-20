package intervals

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestListUpstreamAthletesSanitizesRowsAndRedactsSensitiveFields(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/athletes"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{
				"athlete_id":" i12345 ",
				"athlete_name":" Ada Rider ",
				"email":"ada@example.test",
				"icu_api_key":"icu-secret",
				"api_key":"api-secret",
				"invite_token":"invite-secret",
				"private_wellness_key":"wellness-secret",
				"provider_authorized":true,
				"nested":{"access_token":"nested-secret"}
			},
			{
				"id":"67890",
				"name":" Bob Runner ",
				"access_token":"access-secret",
				"refresh_token":"refresh-secret",
				"oauth_state":"oauth-secret",
				"extra_unknown_field":"ignored"
			}
		]`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	got, err := client.ListUpstreamAthletes(context.Background())
	if err != nil {
		t.Fatalf("ListUpstreamAthletes() error = %v", err)
	}

	want := []UpstreamAthlete{{AthleteID: "i12345", Name: "Ada Rider"}, {AthleteID: "67890", Name: "Bob Runner"}}
	if len(got) != len(want) {
		t.Fatalf("ListUpstreamAthletes() len = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ListUpstreamAthletes()[%d] = %#v, want %#v", i, got[i], want[i])
		}
	}

	publicJSON, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal public athletes: %v", err)
	}
	public := string(publicJSON)
	for _, forbidden := range []string{
		"email", "ada@example.test", "icu_api_key", "icu-secret", "api_key", "api-secret",
		"invite_token", "invite-secret", "private_wellness_key", "wellness-secret", "provider_authorized",
		"access_token", "access-secret", "refresh_token", "refresh-secret", "oauth_state", "oauth-secret", "nested-secret",
	} {
		if strings.Contains(public, forbidden) {
			t.Fatalf("public athletes JSON %s contains forbidden %q", public, forbidden)
		}
	}
}

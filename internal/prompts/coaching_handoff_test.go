package prompts

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestCoachingHandoffPromptContract(t *testing.T) {
	t.Parallel()

	text := renderPromptText(t, CoachingHandoffPrompt(), nil)
	for _, want := range []string{
		"lookback_days=28, race_context_days=90",
		"Handoff scope; Conversation-stated context (Goals, Constraints, Accepted decisions); Icuvisor evidence; Current plan state; Data gaps and unresolved questions; Next actions",
		"explicitly stated or accepted",
		"Claim | Source tool | Athlete-local evidence date/window | Freshness/as-of",
		"resolve_calendar_dates",
		"not provided",
		"never invent fetched_at",
		"_meta.stale",
		"_meta.missing_fields",
		"Strava-blocked",
		"current-day partial",
		"next_page_token",
		"icuvisor_list_advanced_capabilities",
		"instead of calculating a substitute in chat",
		"manually copy it into a fresh Claude, ChatGPT, Cursor, or other client conversation",
		"raw athlete identifiers",
		"local or config paths",
		"Omit health details, precise locations, and private free-text notes by default",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("coaching handoff prompt missing %q:\n%s", want, text)
		}
	}

	toolsLine := lineWithPrefix(t, text, "Tools: ")
	for _, forbidden := range []string{"add_", "update_", "delete_", "get_activity_streams"} {
		if strings.Contains(toolsLine, forbidden) {
			t.Fatalf("coaching handoff tools include forbidden route %q: %s", forbidden, toolsLine)
		}
	}
}

func TestCoachingHandoffPromptRendersBoundedArguments(t *testing.T) {
	t.Parallel()

	text := renderPromptText(t, CoachingHandoffPrompt(), map[string]string{
		"lookback_days":     "42",
		"race_context_days": "180",
	})
	if !strings.Contains(text, "Scope: lookback_days=42, race_context_days=180.") {
		t.Fatalf("coaching handoff prompt did not render bounded arguments:\n%s", text)
	}
}

func TestCoachingHandoffPromptRejectsInvalidArguments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		arguments map[string]string
		want      string
	}{
		{name: "lookback is not an integer", arguments: map[string]string{"lookback_days": "four weeks"}, want: "invalid lookback_days; provide an integer from 1 to 90"},
		{name: "lookback exceeds maximum", arguments: map[string]string{"lookback_days": "91"}, want: "invalid lookback_days; provide an integer from 1 to 90"},
		{name: "race context is zero", arguments: map[string]string{"race_context_days": "0"}, want: "invalid race_context_days; provide an integer from 1 to 365"},
		{name: "race context exceeds maximum", arguments: map[string]string{"race_context_days": "366"}, want: "invalid race_context_days; provide an integer from 1 to 365"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := CoachingHandoffPrompt().Handler(context.Background(), Request{
				Name:      CoachingHandoffName,
				Arguments: tc.arguments,
			})
			var userErr *UserError
			if !errors.As(err, &userErr) {
				t.Fatalf("Handler() error = %v, want UserError", err)
			}
			if userErr.Message != tc.want {
				t.Fatalf("UserError.Message = %q, want %q", userErr.Message, tc.want)
			}
		})
	}
}

func lineWithPrefix(t *testing.T, text, prefix string) string {
	t.Helper()
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, prefix) {
			return line
		}
	}
	t.Fatalf("prompt missing line prefix %q", prefix)
	return ""
}

package coach

import "testing"

func TestEvaluator(t *testing.T) {
	t.Parallel()

	evaluator := NewEvaluator(true, Config{Athletes: []Athlete{
		{ID: "i1", AllowedTools: []string{"get_*"}, DeniedTools: []string{"get_activity_streams"}},
		{ID: "i2", AllowedTools: []string{"*"}, DeniedTools: []string{"delete_*"}},
		{ID: "i3"},
	}})

	tests := []struct {
		name      string
		athleteID string
		toolName  string
		want      bool
	}{
		{name: "allowed prefix", athleteID: "i1", toolName: "get_athlete_profile", want: true},
		{name: "deny overrides allow", athleteID: "i1", toolName: "get_activity_streams", want: false},
		{name: "star allow with delete prefix deny", athleteID: "i2", toolName: "delete_event", want: false},
		{name: "empty allowed denies", athleteID: "i3", toolName: "get_athlete_profile", want: false},
		{name: "unknown athlete denies", athleteID: "i4", toolName: "get_athlete_profile", want: false},
		{name: "non athlete-scoped unaffected", athleteID: "i4", toolName: "icuvisor_list_advanced_capabilities", want: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, _ := evaluator.Evaluate(tc.athleteID, tc.toolName)
			if got != tc.want {
				t.Fatalf("Evaluate() = %t, want %t", got, tc.want)
			}
		})
	}
}

func TestDisabledEvaluatorAllowsAthleteScopedTools(t *testing.T) {
	t.Parallel()

	evaluator := NewEvaluator(false, Config{})
	got, reason := evaluator.Evaluate("i404", "delete_event")
	if !got || reason != "coach_acl_not_applicable" {
		t.Fatalf("Evaluate() = %t %q, want allowed not applicable", got, reason)
	}
}

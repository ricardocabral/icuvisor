package coach

import (
	"errors"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/toolcatalog"
)

func TestToolFilterVisibility(t *testing.T) {
	t.Parallel()

	filter := NewToolFilter(NewEvaluator(true, Config{Athletes: []Athlete{
		{ID: "i111", AllowedTools: []string{toolcatalog.GetAthleteProfile}, DeniedTools: []string{toolcatalog.GetPowerCurves}},
		{ID: "i222", AllowedTools: []string{toolcatalog.GetPowerCurves}},
	}}))

	tests := []struct {
		name      string
		athleteID string
		toolName  string
		want      bool
	}{
		{name: "list athletes always visible", athleteID: "i999", toolName: toolcatalog.ListAthletes, want: true},
		{name: "select athlete always visible", athleteID: "i999", toolName: toolcatalog.SelectAthlete, want: true},
		{name: "advanced capabilities always visible", athleteID: "i999", toolName: toolcatalog.ICUvisorListAdvancedCapabilities, want: true},
		{name: "non athlete scoped allowed", athleteID: "i999", toolName: "server_status", want: true},
		{name: "athlete allow", athleteID: "i111", toolName: toolcatalog.GetAthleteProfile, want: true},
		{name: "athlete deny", athleteID: "i111", toolName: toolcatalog.GetPowerCurves, want: false},
		{name: "unknown athlete denied for athlete scoped", athleteID: "i999", toolName: toolcatalog.GetAthleteProfile, want: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := filter.VisibleForAthlete(tc.athleteID, tc.toolName); got != tc.want {
				t.Fatalf("VisibleForAthlete() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestToolFilterVisibleToolNamesAndAllowedForAny(t *testing.T) {
	t.Parallel()

	filter := NewToolFilter(NewEvaluator(true, Config{Athletes: []Athlete{
		{ID: "i111", AllowedTools: []string{toolcatalog.GetAthleteProfile}},
		{ID: "i222", AllowedTools: []string{toolcatalog.GetPowerCurves}},
	}}))

	got := filter.VisibleToolNamesForAthlete("i111", []string{toolcatalog.GetPowerCurves, toolcatalog.SelectAthlete, toolcatalog.GetAthleteProfile})
	want := []string{toolcatalog.GetAthleteProfile, toolcatalog.SelectAthlete}
	if len(got) != len(want) {
		t.Fatalf("VisibleToolNamesForAthlete() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("VisibleToolNamesForAthlete() = %v, want %v", got, want)
		}
	}
	if !filter.AllowedForAny(toolcatalog.GetPowerCurves) {
		t.Fatal("AllowedForAny(get_power_curves) = false, want true")
	}
	if filter.AllowedForAny(toolcatalog.DeleteEvent) {
		t.Fatal("AllowedForAny(delete_event) = true, want false")
	}
}

func TestToolFilterDisabledEvaluatorAllowsAthleteScopedTools(t *testing.T) {
	t.Parallel()

	filter := NewToolFilter(NewEvaluator(false, Config{}))
	if !filter.VisibleForAthlete("", toolcatalog.GetAthleteProfile) {
		t.Fatal("VisibleForAthlete() = false, want true when evaluator disabled")
	}
	if !filter.AllowedForAny(toolcatalog.GetAthleteProfile) {
		t.Fatal("AllowedForAny() = false, want true when evaluator disabled")
	}
}

func TestToolFilterResolveTarget(t *testing.T) {
	t.Parallel()

	filter := NewToolFilter(NewEvaluator(true, Config{Athletes: []Athlete{
		{ID: "i111", AllowedTools: []string{toolcatalog.GetAthleteProfile}},
		{ID: "i222", AllowedTools: []string{toolcatalog.GetPowerCurves}},
	}}))
	normalize := func(value string) (string, error) {
		switch value {
		case "111", "i111":
			return "i111", nil
		case "222", "i222":
			return "i222", nil
		case "bad":
			return "", errors.New("bad athlete")
		default:
			return value, nil
		}
	}

	got, err := filter.ResolveTarget("", "i111", "i222", toolcatalog.GetPowerCurves, normalize)
	if err != nil {
		t.Fatalf("ResolveTarget(selected) error = %v", err)
	}
	if got != "i222" {
		t.Fatalf("ResolveTarget(selected) = %q, want i222", got)
	}
	got, err = filter.ResolveTarget("111", "i222", "i222", toolcatalog.GetAthleteProfile, normalize)
	if err != nil {
		t.Fatalf("ResolveTarget(override) error = %v", err)
	}
	if got != "i111" {
		t.Fatalf("ResolveTarget(override) = %q, want i111", got)
	}
	if _, err := filter.ResolveTarget("i333", "i111", "", toolcatalog.GetAthleteProfile, normalize); !errors.Is(err, ErrAthleteNotAuthorized) {
		t.Fatalf("ResolveTarget(out of roster) error = %v, want ErrAthleteNotAuthorized", err)
	}
	if _, err := filter.ResolveTarget("bad", "i111", "", toolcatalog.GetAthleteProfile, normalize); !errors.Is(err, ErrInvalidAthleteID) {
		t.Fatalf("ResolveTarget(bad) error = %v, want ErrInvalidAthleteID", err)
	}
	if _, err := filter.ResolveTarget("i111", "i111", "", toolcatalog.GetPowerCurves, normalize); !errors.Is(err, ErrToolNotAllowed) {
		t.Fatalf("ResolveTarget(ACL denied) error = %v, want ErrToolNotAllowed", err)
	}
}

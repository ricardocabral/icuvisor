package tools

import (
	"context"
	"encoding/json"
	"math"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

func TestProposeAnnualTrainingPlanTaperScenarios(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                  string
		arguments             string
		wantPhaseTypes        []string
		wantTaperWeeks        int
		wantTaperLoad         float64
		wantLoadSource        string
		wantWeeksSource       string
		wantTargetSource      string
		wantAdjustment        string
		wantRequestedWeeksNil bool
		wantRequestedPctNil   bool
	}{
		{
			name:                  "omitted inputs preserve one week sixty percent taper",
			arguments:             `{"start_date":"2026-07-13","goal_date":"2026-08-03","current_weekly_load":300,"target_weekly_load":500}`,
			wantPhaseTypes:        []string{"base", "peak", "race_taper"},
			wantTaperWeeks:        1,
			wantTaperLoad:         300,
			wantLoadSource:        "race_taper_60_pct_target",
			wantWeeksSource:       "default",
			wantTargetSource:      "default",
			wantAdjustment:        "none",
			wantRequestedWeeksNil: true,
			wantRequestedPctNil:   true,
		},
		{
			name:                "explicit one week forty percent taper",
			arguments:           `{"start_date":"2026-07-13","goal_date":"2026-08-03","current_weekly_load":300,"target_weekly_load":500,"taper_weeks":1,"taper_target_load_pct":40}`,
			wantPhaseTypes:      []string{"base", "peak", "race_taper"},
			wantTaperWeeks:      1,
			wantTaperLoad:       200,
			wantLoadSource:      "race_taper_target_pct",
			wantWeeksSource:     "input",
			wantTargetSource:    "input",
			wantAdjustment:      "none",
			wantRequestedPctNil: false,
		},
		{
			name:                "explicit two week fifty percent taper",
			arguments:           `{"start_date":"2026-07-13","goal_date":"2026-08-10","current_weekly_load":300,"target_weekly_load":600,"taper_weeks":2,"taper_target_load_pct":50}`,
			wantPhaseTypes:      []string{"base", "peak", "race_taper"},
			wantTaperWeeks:      2,
			wantTaperLoad:       300,
			wantLoadSource:      "race_taper_target_pct",
			wantWeeksSource:     "input",
			wantTargetSource:    "input",
			wantAdjustment:      "none",
			wantRequestedPctNil: false,
		},
		{
			name:                "requested taper is clamped to short horizon",
			arguments:           `{"start_date":"2026-07-13","goal_date":"2026-07-20","current_weekly_load":300,"target_weekly_load":500,"taper_weeks":3}`,
			wantPhaseTypes:      []string{"race_taper"},
			wantTaperWeeks:      2,
			wantTaperLoad:       300,
			wantLoadSource:      "race_taper_60_pct_target",
			wantWeeksSource:     "input",
			wantTargetSource:    "default",
			wantAdjustment:      "taper_weeks_clamped_to_horizon",
			wantRequestedPctNil: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tool := taperScenarioTool()
			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(tc.arguments)})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			out := resultMap(t, result)
			phases := out["phases"].([]any)
			gotPhaseTypes := make([]string, 0, len(phases))
			for _, raw := range phases {
				gotPhaseTypes = append(gotPhaseTypes, raw.(map[string]any)["phase_type"].(string))
			}
			if !equalStrings(gotPhaseTypes, tc.wantPhaseTypes) {
				t.Fatalf("phase types = %v, want %v", gotPhaseTypes, tc.wantPhaseTypes)
			}

			meta := out["_meta"].(map[string]any)
			taper := meta["taper_scenario"].(map[string]any)
			if int(taper["resolved_taper_weeks"].(float64)) != tc.wantTaperWeeks {
				t.Fatalf("resolved taper weeks = %v, want %d", taper["resolved_taper_weeks"], tc.wantTaperWeeks)
			}
			if taper["taper_weeks_source"] != tc.wantWeeksSource || taper["taper_target_load_pct_source"] != tc.wantTargetSource || taper["adjustment_code"] != tc.wantAdjustment {
				t.Fatalf("taper metadata = %#v, want sources %q/%q and adjustment %q", taper, tc.wantWeeksSource, tc.wantTargetSource, tc.wantAdjustment)
			}
			if tc.wantRequestedWeeksNil != (taper["requested_taper_weeks"] == nil) || tc.wantRequestedPctNil != (taper["requested_taper_target_load_pct"] == nil) {
				t.Fatalf("requested taper metadata = %#v, want nil flags %t/%t", taper, tc.wantRequestedWeeksNil, tc.wantRequestedPctNil)
			}

			weeks := out["weekly_targets"].([]any)
			bridge := meta["projection_bridge"].(map[string]any)["weekly_plan_targets"].([]any)
			for i, raw := range weeks {
				week := raw.(map[string]any)
				if week["is_taper_week"] != true {
					continue
				}
				if week["training_load"] != tc.wantTaperLoad || week["load_source"] != tc.wantLoadSource || week["is_recovery_week"] == true {
					t.Fatalf("taper week = %#v, want load %.1f/source %q/non-recovery", week, tc.wantTaperLoad, tc.wantLoadSource)
				}
				bridgeWeek := bridge[i].(map[string]any)
				if bridgeWeek["training_load"] != week["training_load"] {
					t.Fatalf("bridge week %d = %#v, want weekly target load %.1f", i, bridgeWeek, week["training_load"])
				}
			}
		})
	}
}

func TestProposeAnnualTrainingPlanTaperRoundingReconcilesMetadataAndBridge(t *testing.T) {
	t.Parallel()

	tool := taperScenarioTool()
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"start_date":"2026-07-13","goal_date":"2026-07-20","target_weekly_load":425,"taper_target_load_pct":37.555}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	out := resultMap(t, result)
	meta := out["_meta"].(map[string]any)
	taper := meta["taper_scenario"].(map[string]any)
	resolvedPct := taper["resolved_taper_target_load_pct"].(float64)
	if resolvedPct != 37.555 {
		t.Fatalf("resolved taper percentage = %.12g, want exact input 37.555", resolvedPct)
	}
	weeks := out["weekly_targets"].([]any)
	bridge := meta["projection_bridge"].(map[string]any)["weekly_plan_targets"].([]any)
	for i, raw := range weeks {
		week := raw.(map[string]any)
		if week["is_taper_week"] != true {
			continue
		}
		want := math.Round(425*resolvedPct/100*10) / 10
		if week["training_load"] != want {
			t.Fatalf("week %d load = %v, want %.1f from metadata percentage", i, week["training_load"], want)
		}
		if bridge[i].(map[string]any)["training_load"] != week["training_load"] {
			t.Fatalf("bridge week %d load = %v, want %v", i, bridge[i].(map[string]any)["training_load"], week["training_load"])
		}
	}
}

func TestProposeAnnualTrainingPlanTaperRejectsInvalidBounds(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		args string
		want string
	}{
		{name: "zero weeks", args: `{"goal_date":"2026-07-20","taper_weeks":0}`, want: "taper_weeks must be between 1 and 3"},
		{name: "four weeks", args: `{"goal_date":"2026-07-20","taper_weeks":4}`, want: "taper_weeks must be between 1 and 3"},
		{name: "zero percent", args: `{"goal_date":"2026-07-20","taper_target_load_pct":0}`, want: "taper_target_load_pct must be between 1 and 100"},
		{name: "over one hundred percent", args: `{"goal_date":"2026-07-20","taper_target_load_pct":101}`, want: "taper_target_load_pct must be between 1 and 100"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := decodeSeasonPlanProposalRequest(json.RawMessage(tc.args))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("decode() error = %v, want %q", err, tc.want)
			}
		})
	}
}

func taperScenarioTool() Tool {
	client := &fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}}
	return newProposeAnnualTrainingPlanToolWithClock(client, "test", "UTC", false, fixedSeasonPlanClock())
}

func equalStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

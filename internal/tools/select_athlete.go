package tools

import (
	"context"
	"encoding/json"
	"slices"
	"sort"

	"github.com/ricardocabral/icuvisor/internal/coach"
	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/toolcatalog"
)

const (
	selectAthleteName        = "select_athlete"
	selectAthleteDescription = "Select the default target athlete for subsequent coach-mode tool calls in this MCP session."
)

type selectAthleteRequest struct {
	AthleteID string `json:"athlete_id"`
}

type selectAthleteResponse struct {
	PreviousAthleteID string            `json:"previous_athlete_id"`
	NewAthleteID      string            `json:"new_athlete_id"`
	AllowedTools      []string          `json:"allowed_tools"`
	Meta              selectAthleteMeta `json:"_meta"`
}

type selectAthleteMeta struct {
	Scope                   string `json:"scope"`
	RequiresNewConversation bool   `json:"requires_new_conversation"`
}

func newSelectAthleteTool(cfg coach.Config) Tool {
	evaluator := coach.NewEvaluator(true, cfg)
	return coreTool(Tool{Name: selectAthleteName, Description: selectAthleteDescription, InputSchema: selectAthleteInputSchema(), OutputSchema: genericOutputSchema("Selected coach athlete and visible tools."), Handler: selectAthleteHandler(evaluator)})
}

func selectAthleteInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"athlete_id"}, "properties": map[string]any{"athlete_id": map[string]any{"type": "string", "description": "Target athlete to select for this session. Format: i12345 or 12345."}}}
}

func selectAthleteHandler(evaluator coach.Evaluator) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		var args selectAthleteRequest
		if err := json.Unmarshal(req.Arguments, &args); err != nil {
			return Result{}, NewUserError(invalidCoachAthleteTargetMessage, err)
		}
		normalized, err := config.NormalizeAthleteID(args.AthleteID)
		if err != nil || !evaluator.HasAthlete(normalized) {
			return Result{}, NewUserError(invalidCoachAthleteTargetMessage, err)
		}
		selection, ok := coach.SelectionContextFromContext(ctx)
		if !ok || selection.Store == nil {
			return Result{}, NewUserError("select_athlete session state is unavailable", nil)
		}
		visibleTools := visibleToolsForAthlete
		if selection.VisibleTools != nil {
			visibleTools = func(_ coach.Evaluator, athleteID string) []string { return selection.VisibleTools(athleteID) }
		}
		previous := selection.Store.Selected(selection.Key)
		previousTools := visibleTools(evaluator, previous)
		selection.Store.Select(selection.Key, normalized)
		newTools := visibleTools(evaluator, normalized)
		return TextResult(selectAthleteResponse{PreviousAthleteID: previous, NewAthleteID: normalized, AllowedTools: newTools, Meta: selectAthleteMeta{Scope: selection.Scope, RequiresNewConversation: !slices.Equal(previousTools, newTools)}}), nil
	}
}

func visibleToolsForAthlete(evaluator coach.Evaluator, athleteID string) []string {
	out := []string{listAthletesName, selectAthleteName, toolcatalog.ICUvisorListAdvancedCapabilities}
	for _, name := range toolcatalog.AthleteScopedToolNames() {
		if allowed, _ := evaluator.Evaluate(athleteID, name); allowed {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
}

const invalidCoachAthleteTargetMessage = "invalid athlete_id; use a configured target athlete"

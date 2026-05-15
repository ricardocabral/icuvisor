package tools

import (
	"context"
	"encoding/json"
	"sort"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/coach"
)

const (
	listAthletesName        = "list_athletes"
	listAthletesDescription = "List the coach-mode roster configured for this icuvisor server. This tool makes no intervals.icu API calls."
)

type listAthletesResponse struct {
	Athletes []listAthletesRow `json:"athletes"`
	Meta     listAthletesMeta  `json:"_meta"`
}

type listAthletesRow struct {
	AthleteID string `json:"athlete_id"`
	Label     string `json:"label,omitempty"`
	Default   bool   `json:"default,omitempty"`
	Active    bool   `json:"active,omitempty"`
}

type listAthletesMeta struct {
	Source           string `json:"source"`
	Count            int    `json:"count"`
	DefaultAthleteID string `json:"default_athlete_id"`
	ActiveAthleteID  string `json:"active_athlete_id"`
}

func newListAthletesTool(cfg coach.Config) Tool {
	return coreTool(Tool{Name: listAthletesName, Description: listAthletesDescription, InputSchema: noArgsSchema(), OutputSchema: genericOutputSchema("Configured coach roster."), Handler: listAthletesHandler(cfg)})
}

func listAthletesHandler(cfg coach.Config) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		if err := noArgs(req.Arguments, listAthletesName); err != nil {
			return Result{}, err
		}
		active := cfg.DefaultAthleteID
		if selection, ok := coach.SelectionContextFromContext(ctx); ok {
			active = selection.Store.Selected(selection.Key)
		}
		rows := make([]listAthletesRow, 0, len(cfg.Athletes))
		for _, athlete := range cfg.Athletes {
			rows = append(rows, listAthletesRow{AthleteID: athlete.ID, Label: athlete.Label, Default: athlete.ID == cfg.DefaultAthleteID, Active: athlete.ID == active})
		}
		sort.Slice(rows, func(i, j int) bool { return rows[i].AthleteID < rows[j].AthleteID })
		return TextResult(listAthletesResponse{Athletes: rows, Meta: listAthletesMeta{Source: "config", Count: len(rows), DefaultAthleteID: cfg.DefaultAthleteID, ActiveAthleteID: active}}), nil
	}
}

func noArgsSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "properties": map[string]any{}}
}

func noArgs(raw json.RawMessage, toolName string) error {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "{}" || trimmed == "null" {
		return nil
	}
	return NewUserError("invalid "+toolName+" arguments; no arguments are supported", nil)
}

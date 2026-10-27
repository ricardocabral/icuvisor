package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/response"
)

const (
	searchWorkoutLibraryName        = "search_workout_library"
	searchWorkoutLibraryDescription = "Discover publicly shared workout-library folders and training plans by text, sport, weeks and hours per week. This global search does not add folders or schedule workouts. Required bounds are whole weeks and whole hours per week; use limit to bound results. No upstream pagination cursor is available; narrow filters when the limit is reached. Terse rows omit owner details, share tokens and workouts; include_full adds folder descriptions only. Use add_shared_workout_folder separately to add a selected result to your library."
	searchWorkoutLibraryEndpoint    = "/folder-search"
	searchWorkoutLibraryMessage     = "could not search public workout library; check credentials and search filters"
)

// WorkoutLibrarySearcherClient retrieves public folders and training plans without athlete-scoped side effects.
type WorkoutLibrarySearcherClient interface {
	SearchWorkoutLibrary(context.Context, intervals.WorkoutLibrarySearchParams) ([]intervals.WorkoutFolder, error)
}

type searchWorkoutLibraryRequest struct {
	Query           string   `json:"q"`
	Type            string   `json:"type,omitempty"`
	ActivityTypes   []string `json:"activity_types,omitempty"`
	WeeksMin        *int     `json:"weeks_min"`
	WeeksMax        *int     `json:"weeks_max"`
	HoursPerWeekMin *int     `json:"hours_per_week_min"`
	HoursPerWeekMax *int     `json:"hours_per_week_max"`
	WorkoutTargets  []string `json:"workout_targets,omitempty"`
	Limit           *int     `json:"limit,omitempty"`
	IncludeFull     bool     `json:"include_full,omitempty"`
}

type sharedWorkoutFolderRow struct {
	FolderID        string   `json:"folder_id"`
	Name            string   `json:"name,omitempty"`
	Type            string   `json:"type,omitempty"`
	Sports          []string `json:"sports,omitempty"`
	DurationWeeks   *int     `json:"duration_weeks,omitempty"`
	HoursPerWeekMin *int     `json:"hours_per_week_min,omitempty"`
	HoursPerWeekMax *int     `json:"hours_per_week_max,omitempty"`
	WorkoutTargets  []string `json:"workout_targets,omitempty"`
	NumWorkouts     *int     `json:"num_workouts,omitempty"`
	Description     string   `json:"description,omitempty"`
	Blurb           string   `json:"blurb,omitempty"`
}

type searchWorkoutLibraryResponse struct {
	Folders []sharedWorkoutFolderRow `json:"folders"`
	Meta    searchWorkoutLibraryMeta `json:"_meta"`
}

type searchWorkoutLibraryMeta struct {
	SourceEndpoint string `json:"source_endpoint"`
	Count          int    `json:"count"`
	Limit          int    `json:"limit"`
	LimitReached   bool   `json:"limit_reached"`
	Truncated      bool   `json:"truncated"`
	Pagination     string `json:"pagination"`
	IncludeFull    bool   `json:"include_full"`
}

func newSearchWorkoutLibraryTool(client WorkoutLibrarySearcherClient, version string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return fullTool(Tool{Name: searchWorkoutLibraryName, Description: searchWorkoutLibraryDescription, InputSchema: searchWorkoutLibraryInputSchema(), OutputSchema: genericOutputSchema("Public folder summaries with weeks, integer hours per week, workout count and target types. _meta.limit_reached indicates that more results may exist; truncated reports local clipping if upstream exceeds limit. There is no total_count or next-page cursor. include_full adds descriptions/blurb, never share tokens, owner details or child workouts."), Requirement: RequirementRead, Handler: func(ctx context.Context, req Request) (Result, error) {
		args, params, err := decodeSearchWorkoutLibraryRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError("invalid search_workout_library arguments; provide q and ordered week/hour bounds within schema limits", err)
		}
		if client == nil {
			return Result{}, NewUserError(searchWorkoutLibraryMessage, errors.New("missing workout library search client"))
		}
		folders, err := client.SearchWorkoutLibrary(ctx, params)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(searchWorkoutLibraryMessage, err)
		}
		meta := searchWorkoutLibraryMeta{SourceEndpoint: searchWorkoutLibraryEndpoint, Limit: params.Limit, LimitReached: len(folders) >= params.Limit, Truncated: len(folders) > params.Limit, Pagination: "no upstream cursor; narrow filters when limit_reached is true", IncludeFull: args.IncludeFull}
		if len(folders) > params.Limit {
			folders = folders[:params.Limit]
		}
		rows := make([]sharedWorkoutFolderRow, 0, len(folders))
		for _, folder := range folders {
			rows = append(rows, sharedWorkoutFolderToRow(folder, args.IncludeFull))
		}
		meta.Count = len(rows)
		return encodeShaped(searchWorkoutLibraryResponse{Folders: rows, Meta: meta}, args.IncludeFull, []string{"folders"}, version, debugMetadata, searchWorkoutLibraryName, response.UnitSystemMetric, shapeCfg)
	}})
}

func decodeSearchWorkoutLibraryRequest(raw json.RawMessage) (searchWorkoutLibraryRequest, intervals.WorkoutLibrarySearchParams, error) {
	args, err := DecodeStrict[searchWorkoutLibraryRequest](raw)
	if err != nil {
		return args, intervals.WorkoutLibrarySearchParams{}, err
	}
	if args.WeeksMin == nil || args.WeeksMax == nil || args.HoursPerWeekMin == nil || args.HoursPerWeekMax == nil {
		return args, intervals.WorkoutLibrarySearchParams{}, errors.New("weeks_min, weeks_max, hours_per_week_min and hours_per_week_max are required")
	}
	limit := intervals.DefaultWorkoutLibrarySearchLimit
	if args.Limit != nil {
		limit = *args.Limit
		if limit < 1 {
			return args, intervals.WorkoutLibrarySearchParams{}, errors.New("limit must be positive")
		}
	}
	params := intervals.WorkoutLibrarySearchParams{Query: strings.TrimSpace(args.Query), Type: args.Type, ActivityTypes: args.ActivityTypes, WeeksMin: *args.WeeksMin, WeeksMax: *args.WeeksMax, HoursPerWeekMin: *args.HoursPerWeekMin, HoursPerWeekMax: *args.HoursPerWeekMax, WorkoutTargets: args.WorkoutTargets, Limit: limit}
	return args, params, params.Validate()
}

func sharedWorkoutFolderToRow(folder intervals.WorkoutFolder, includeFull bool) sharedWorkoutFolderRow {
	row := sharedWorkoutFolderRow{FolderID: folder.ID, Name: stringValue(folder.Name), Type: stringValue(folder.Type), Sports: folder.ActivityTypes, DurationWeeks: folder.DurationWeeks, HoursPerWeekMin: folder.HoursPerWeekMin, HoursPerWeekMax: folder.HoursPerWeekMax, WorkoutTargets: folder.WorkoutTargets, NumWorkouts: folder.NumWorkouts}
	if includeFull {
		row.Description = stringValue(folder.Description)
		row.Blurb = stringValue(folder.Blurb)
	}
	return row
}

func searchWorkoutLibraryInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"q", "weeks_min", "weeks_max", "hours_per_week_min", "hours_per_week_max"}, "properties": map[string]any{
		"q":                  map[string]any{"type": "string", "minLength": 1, "maxLength": 200, "description": "Required public folder/plan search text, 1-200 nonblank characters."},
		"type":               map[string]any{"type": "string", "enum": []string{"FOLDER", "PLAN"}, "description": "Optional public library item kind: FOLDER for reusable workouts or PLAN for a dated plan template."},
		"activity_types":     map[string]any{"type": "array", "maxItems": 10, "uniqueItems": true, "items": map[string]any{"type": "string", "enum": intervals.WorkoutLibraryActivityTypes()}, "description": "Optional sport filters using intervals.icu activity types, for example Run or Ride; at most 10."},
		"weeks_min":          map[string]any{"type": "integer", "minimum": 0, "maximum": 520, "description": "Required inclusive lower plan-duration filter, in whole weeks (0-520)."},
		"weeks_max":          map[string]any{"type": "integer", "minimum": 0, "maximum": 520, "description": "Required inclusive upper plan-duration filter, in whole weeks; must be at least weeks_min and at most 520."},
		"hours_per_week_min": map[string]any{"type": "integer", "minimum": 0, "maximum": 168, "description": "Required inclusive minimum workload filter, in whole hours per week (0-168)."},
		"hours_per_week_max": map[string]any{"type": "integer", "minimum": 0, "maximum": 168, "description": "Required inclusive maximum workload filter, in whole hours per week; must be at least hours_per_week_min and at most 168."},
		"workout_targets":    map[string]any{"type": "array", "maxItems": 4, "uniqueItems": true, "items": map[string]any{"type": "string", "enum": []string{"AUTO", "POWER", "HR", "PACE"}}, "description": "Optional workout target filters: AUTO, POWER, HR (heart rate) or PACE."},
		"limit":              map[string]any{"type": "integer", "minimum": 1, "maximum": 100, "default": intervals.DefaultWorkoutLibrarySearchLimit, "description": fmt.Sprintf("Maximum number of public folders returned (1-100, default %d). The API has no cursor; narrow filters when this limit is reached.", intervals.DefaultWorkoutLibrarySearchLimit)},
		"include_full":       map[string]any{"type": "boolean", "default": false, "description": "Include folder description and blurb. Child workouts, owner details and share tokens are always omitted; use existing library reads after explicitly adding a folder."},
	}, "examples": []map[string]any{{"q": "base", "type": "PLAN", "activity_types": []string{"Run"}, "weeks_min": 4, "weeks_max": 12, "hours_per_week_min": 3, "hours_per_week_max": 8, "limit": 5}}}
}

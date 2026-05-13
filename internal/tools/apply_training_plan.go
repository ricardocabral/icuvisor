package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/safety"
	"github.com/ricardocabral/icuvisor/internal/workoutdoc"
)

const (
	applyTrainingPlanName                    = "apply_training_plan"
	applyTrainingPlanDescription             = "Apply a workout-library training plan to the athlete calendar from an anchor start date. Defaults to dry_run:true, fetches plan workouts server-side by plan_id, marks conflicts, and only replaces existing events when ICUVISOR_DELETE_MODE=full."
	invalidApplyTrainingPlanArgumentsMessage = "invalid apply_training_plan arguments; provide plan_id, start_date YYYY-MM-DD, optional dry_run, and conflict_policy skip_existing or replace_existing"
	applyTrainingPlanMessage                 = "could not apply training plan; check intervals.icu credentials, athlete ID, plan ID, date range, and delete-mode configuration"
	applyTrainingPlanConflictSkip            = "skip_existing"
	applyTrainingPlanConflictReplace         = "replace_existing"
)

// ApplyTrainingPlanClient fetches workout-library plan content, reads calendar conflicts, and writes events.
type ApplyTrainingPlanClient interface {
	WorkoutLibraryClient
	EventsClient
	EventWriterClient
	EventDeleterClient
}

type applyTrainingPlanRequest struct {
	PlanID         string `json:"plan_id"`
	StartDate      string `json:"start_date"`
	DryRun         *bool  `json:"dry_run,omitempty"`
	ConflictPolicy string `json:"conflict_policy,omitempty"`
}

type applyTrainingPlanResponse struct {
	ProposedEvents []applyTrainingPlanProposedEvent `json:"proposed_events"`
	CreatedEvents  []getEventsRow                   `json:"created_events,omitempty"`
	Meta           applyTrainingPlanMeta            `json:"_meta"`
}

type applyTrainingPlanProposedEvent struct {
	Date      string                      `json:"date"`
	WorkoutID string                      `json:"workout_id"`
	Name      string                      `json:"name,omitempty"`
	Sport     string                      `json:"sport,omitempty"`
	Conflicts []applyTrainingPlanConflict `json:"conflicts"`
}

type applyTrainingPlanConflict struct {
	EventID string `json:"event_id"`
	Reason  string `json:"reason"`
}

type applyTrainingPlanSkipped struct {
	Date      string                      `json:"date"`
	WorkoutID string                      `json:"workout_id"`
	Conflicts []applyTrainingPlanConflict `json:"conflicts"`
}

type applyTrainingPlanReplaced struct {
	Date            string   `json:"date"`
	WorkoutID       string   `json:"workout_id"`
	DeletedEventIDs []string `json:"deleted_event_ids"`
}

type applyTrainingPlanMeta struct {
	PlanID         string                      `json:"plan_id"`
	StartDate      string                      `json:"start_date"`
	DryRun         bool                        `json:"dry_run"`
	ConflictPolicy string                      `json:"conflict_policy"`
	CreatedCount   int                         `json:"created_count"`
	Skipped        []applyTrainingPlanSkipped  `json:"skipped"`
	Replaced       []applyTrainingPlanReplaced `json:"replaced,omitempty"`
	DeleteMode     string                      `json:"delete_mode"`
	Timezone       string                      `json:"timezone"`
}

type applyTrainingPlanWorkout struct {
	Workout intervals.Workout
	Date    string
	Day     int
}

func newApplyTrainingPlanTool(client ApplyTrainingPlanClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, capability safety.Capability) Tool {
	return Tool{Name: applyTrainingPlanName, Description: applyTrainingPlanDescription, InputSchema: applyTrainingPlanInputSchema(capabilityOrSafe(capability)), OutputSchema: applyTrainingPlanOutputSchema(), Requirement: RequirementWrite, Handler: applyTrainingPlanHandler(client, profileClient, version, timezoneFallback, debugMetadata, capabilityOrSafe(capability))}
}

func applyTrainingPlanHandler(client ApplyTrainingPlanClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, capability safety.Capability) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeApplyTrainingPlanRequest(req.Arguments, capabilityOrSafe(capability))
		if err != nil {
			return Result{}, NewUserError(invalidApplyTrainingPlanArgumentsMessage, err)
		}
		unitSystem, timezoneName, err := toolProfile(ctx, profileClient, timezoneFallback)
		if err != nil {
			return Result{}, NewUserError(applyTrainingPlanMessage, err)
		}
		if client == nil {
			return Result{}, NewUserError(applyTrainingPlanMessage, errors.New("missing apply training plan client"))
		}
		payload, err := applyTrainingPlan(ctx, client, args, timezoneName, capabilityOrSafe(capability))
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			var userErr *UserError
			if errors.As(err, &userErr) {
				return Result{}, err
			}
			return Result{}, NewUserError(applyTrainingPlanMessage, err)
		}
		return encodeShaped(payload, false, []string{"proposed_events", "created_events"}, version, debugMetadata, applyTrainingPlanName, unitSystem)
	}
}

func decodeApplyTrainingPlanRequest(raw json.RawMessage, capability safety.Capability) (applyTrainingPlanRequest, error) {
	var args applyTrainingPlanRequest
	if err := decodeStrict(raw, &args); err != nil {
		return args, err
	}
	args.PlanID = strings.TrimSpace(args.PlanID)
	args.StartDate = strings.TrimSpace(args.StartDate)
	args.ConflictPolicy = strings.TrimSpace(args.ConflictPolicy)
	if args.PlanID == "" {
		return args, errors.New("plan_id is required")
	}
	if !validDate(args.StartDate) {
		return args, errors.New("start_date must be athlete-local YYYY-MM-DD")
	}
	if args.ConflictPolicy == "" {
		args.ConflictPolicy = applyTrainingPlanConflictSkip
	}
	if args.ConflictPolicy != applyTrainingPlanConflictSkip && args.ConflictPolicy != applyTrainingPlanConflictReplace {
		return args, errors.New("conflict_policy must be skip_existing or replace_existing")
	}
	if args.ConflictPolicy == applyTrainingPlanConflictReplace && !capabilityOrSafe(capability).CanDelete() {
		return args, errors.New("replace_existing requires ICUVISOR_DELETE_MODE=full")
	}
	return args, nil
}

func applyTrainingPlan(ctx context.Context, client ApplyTrainingPlanClient, args applyTrainingPlanRequest, timezoneName string, capability safety.Capability) (applyTrainingPlanResponse, error) {
	planned, err := planWorkoutsForApply(ctx, client, args.PlanID, args.StartDate)
	if err != nil {
		return applyTrainingPlanResponse{}, err
	}
	if len(planned) == 0 {
		return applyTrainingPlanResponse{}, fmt.Errorf("plan %s has no workouts with relative day metadata", args.PlanID)
	}
	dryRun := true
	if args.DryRun != nil {
		dryRun = *args.DryRun
	}
	oldest := planned[0].Date
	newest := planned[len(planned)-1].Date
	conflictsByDate, err := fetchApplyTrainingPlanConflicts(ctx, client, oldest, newest)
	if err != nil {
		return applyTrainingPlanResponse{}, err
	}
	payload := applyTrainingPlanResponse{ProposedEvents: make([]applyTrainingPlanProposedEvent, 0, len(planned)), Meta: applyTrainingPlanMeta{PlanID: args.PlanID, StartDate: args.StartDate, DryRun: dryRun, ConflictPolicy: args.ConflictPolicy, Skipped: []applyTrainingPlanSkipped{}, DeleteMode: capabilityOrSafe(capability).Mode(), Timezone: timezoneName}}
	for _, plannedWorkout := range planned {
		workout := plannedWorkout.Workout
		conflicts := conflictsByDate[plannedWorkout.Date]
		if conflicts == nil {
			conflicts = []applyTrainingPlanConflict{}
		}
		row := applyTrainingPlanProposedEvent{Date: plannedWorkout.Date, WorkoutID: workout.ID, Name: stringValue(workout.Name), Sport: stringValue(workout.Type), Conflicts: conflicts}
		payload.ProposedEvents = append(payload.ProposedEvents, row)
		if dryRun {
			continue
		}
		if len(conflicts) > 0 && args.ConflictPolicy == applyTrainingPlanConflictSkip {
			payload.Meta.Skipped = append(payload.Meta.Skipped, applyTrainingPlanSkipped{Date: row.Date, WorkoutID: row.WorkoutID, Conflicts: conflicts})
			continue
		}
		if len(conflicts) > 0 && args.ConflictPolicy == applyTrainingPlanConflictReplace {
			replaced := applyTrainingPlanReplaced{Date: row.Date, WorkoutID: row.WorkoutID, DeletedEventIDs: make([]string, 0, len(conflicts))}
			for _, conflict := range conflicts {
				if err := client.DeleteEvent(ctx, conflict.EventID); err != nil {
					return applyTrainingPlanResponse{}, fmt.Errorf("deleting conflicting event %s: %w", conflict.EventID, err)
				}
				replaced.DeletedEventIDs = append(replaced.DeletedEventIDs, conflict.EventID)
			}
			payload.Meta.Replaced = append(payload.Meta.Replaced, replaced)
		}
		params, err := eventParamsFromPlanWorkout(plannedWorkout.Date, workout)
		if err != nil {
			return applyTrainingPlanResponse{}, err
		}
		created, err := client.AddOrUpdateEvent(ctx, params)
		if err != nil {
			return applyTrainingPlanResponse{}, fmt.Errorf("creating event for workout %s on %s: %w", workout.ID, plannedWorkout.Date, err)
		}
		createdRow, err := eventRow(created, false, timezoneName)
		if err != nil {
			return applyTrainingPlanResponse{}, err
		}
		payload.CreatedEvents = append(payload.CreatedEvents, createdRow)
		payload.Meta.CreatedCount++
	}
	return payload, nil
}

func planWorkoutsForApply(ctx context.Context, client ApplyTrainingPlanClient, planID string, startDate string) ([]applyTrainingPlanWorkout, error) {
	folders, err := client.ListWorkoutFolders(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching workout-library folders: %w", err)
	}
	workouts, err := client.ListLibraryWorkouts(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching workout-library workouts: %w", err)
	}
	selected := map[string]intervals.Workout{}
	for _, folder := range folders {
		if folder.ID != planID {
			continue
		}
		for _, child := range folder.Children {
			if child.ID != "" {
				selected[child.ID] = child
			}
		}
		break
	}
	for _, workout := range workouts {
		if workoutFolderID(workout) == planID && workout.ID != "" {
			selected[workout.ID] = workout
		}
	}
	start, _ := time.Parse(time.DateOnly, startDate)
	planned := make([]applyTrainingPlanWorkout, 0, len(selected))
	for _, workout := range selected {
		day := planWorkoutRelativeDay(workout)
		if day <= 0 {
			continue
		}
		planned = append(planned, applyTrainingPlanWorkout{Workout: workout, Day: day, Date: start.AddDate(0, 0, day-1).Format(time.DateOnly)})
	}
	sort.SliceStable(planned, func(i, j int) bool {
		if planned[i].Day != planned[j].Day {
			return planned[i].Day < planned[j].Day
		}
		if stringValue(planned[i].Workout.Name) != stringValue(planned[j].Workout.Name) {
			return stringValue(planned[i].Workout.Name) < stringValue(planned[j].Workout.Name)
		}
		return planned[i].Workout.ID < planned[j].Workout.ID
	})
	return planned, nil
}

func planWorkoutRelativeDay(workout intervals.Workout) int {
	if workout.Day != nil {
		return *workout.Day
	}
	if workout.Days != nil {
		return *workout.Days
	}
	for _, key := range []string{"day", "days", "relative_day", "plan_day"} {
		if day, ok := anyPositiveInt(workout.Raw[key]); ok {
			return day
		}
	}
	return 0
}

func anyPositiveInt(value any) (int, bool) {
	switch typed := value.(type) {
	case int:
		return typed, typed > 0
	case int64:
		return int(typed), typed > 0
	case float64:
		if typed == float64(int(typed)) && typed > 0 {
			return int(typed), true
		}
	case json.Number:
		parsed, err := typed.Int64()
		if err == nil && parsed > 0 {
			return int(parsed), true
		}
	case string:
		parsed, err := time.ParseDuration(strings.TrimSpace(typed) + "h")
		if err == nil {
			day := int(parsed.Hours())
			return day, day > 0
		}
	}
	return 0, false
}

func fetchApplyTrainingPlanConflicts(ctx context.Context, client ApplyTrainingPlanClient, oldest string, newest string) (map[string][]applyTrainingPlanConflict, error) {
	events, err := client.ListEvents(ctx, intervals.ListEventsParams{Oldest: oldest, Newest: newest, Limit: maxEventsLimit})
	if err != nil {
		return nil, fmt.Errorf("fetching calendar conflicts: %w", err)
	}
	conflictsByDate := map[string][]applyTrainingPlanConflict{}
	for _, event := range events {
		date := eventDateOnly(event)
		if date == "" {
			continue
		}
		conflictsByDate[date] = append(conflictsByDate[date], applyTrainingPlanConflict{EventID: event.ID, Reason: "existing_event_on_date"})
	}
	for date := range conflictsByDate {
		sort.SliceStable(conflictsByDate[date], func(i, j int) bool { return conflictsByDate[date][i].EventID < conflictsByDate[date][j].EventID })
	}
	return conflictsByDate, nil
}

func eventDateOnly(event intervals.Event) string {
	for _, value := range []string{stringValue(event.StartDateLocal), anyString(event.Raw["start_date_local"]), anyString(event.Raw["start_date"])} {
		value = strings.TrimSpace(value)
		if len(value) >= len(time.DateOnly) && validDate(value[:len(time.DateOnly)]) {
			return value[:len(time.DateOnly)]
		}
	}
	return ""
}

func eventParamsFromPlanWorkout(date string, workout intervals.Workout) (intervals.WriteEventParams, error) {
	trainingLoad := workoutTrainingLoad(workout)
	args := addOrUpdateEventRequest{Date: date, Category: "WORKOUT", Type: stringValue(workout.Type), Name: stringValue(workout.Name), Tags: append([]string(nil), workout.Tags...), TargetLoad: trainingLoad, DistanceMeters: workout.Distance, MovingTimeSeconds: workout.MovingTime}
	if workout.WorkoutDoc != nil {
		doc, err := workoutDocFromAny(workout.WorkoutDoc)
		if err != nil {
			return intervals.WriteEventParams{}, fmt.Errorf("decoding workout_doc for workout %s: %w", workout.ID, err)
		}
		args.WorkoutDoc = &doc
	} else if workout.Description != nil {
		args.Description = workout.Description
	}
	params, _, err := eventWriteParams(args)
	if err != nil {
		return intervals.WriteEventParams{}, fmt.Errorf("building event params for workout %s: %w", workout.ID, err)
	}
	return params, nil
}

func workoutTrainingLoad(workout intervals.Workout) *float64 {
	if workout.TrainingLoad == nil {
		return nil
	}
	value := float64(*workout.TrainingLoad)
	return &value
}

func workoutDocFromAny(value any) (workoutdoc.WorkoutDoc, error) {
	var doc workoutdoc.WorkoutDoc
	data, err := json.Marshal(value)
	if err != nil {
		return doc, err
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return doc, err
	}
	return doc, nil
}

func applyTrainingPlanInputSchema(capability safety.Capability) map[string]any {
	conflictEnum := []string{applyTrainingPlanConflictSkip}
	if capabilityOrSafe(capability).CanDelete() {
		conflictEnum = append(conflictEnum, applyTrainingPlanConflictReplace)
	}
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"plan_id", "start_date"}, "properties": map[string]any{
		"plan_id":         map[string]any{"type": "string", "description": "Required workout-library folder/plan ID to fetch server-side. Do not pass plan contents in tool arguments."},
		"start_date":      map[string]any{"type": "string", "description": "Required athlete-local YYYY-MM-DD anchor date; workout day 1 is applied to this date and later plan days are relative to it."},
		"dry_run":         map[string]any{"type": "boolean", "default": true, "description": "Safety default is true, even in safe mode. Set dry_run:false explicitly to create or replace calendar events."},
		"conflict_policy": map[string]any{"type": "string", "default": applyTrainingPlanConflictSkip, "enum": conflictEnum, "description": "skip_existing leaves days with calendar conflicts untouched. replace_existing deletes conflicting events before creating plan workouts and is accepted only when ICUVISOR_DELETE_MODE=full."},
	}}
}

func applyTrainingPlanOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Training-plan apply preview or write result with per-day proposed events, conflict markers, created event rows, and _meta created/skipped/replaced/delete-mode counts."}
}

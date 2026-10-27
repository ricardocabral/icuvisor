package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/response"
	"github.com/ricardocabral/icuvisor/internal/workoutdoc"
)

const (
	addOrUpdateEventName                    = "add_or_update_event"
	addOrUpdateEventDescription             = "Create or update a non-destructive calendar event such as a planned workout, race, or note. Omitting event_id creates a new event; providing event_id updates only that event and reports other events on the destination date so reschedules do not silently hide overlaps; use delete_event to remove. Optional external_id is a non-empty upstream idempotency key for retry-safe creates/updates; omit it to leave the upstream key unchanged, and use event_id for intentional edits to an existing event. Creates preflight same-day calendar events, skip matching external_id retries, and skip exact duplicates when upstream fields already match; an interrupted create is reported recovered only after a bounded same-day read verifies the intended event. Requested planned seconds or fully timed workout_doc steps receive _meta.duration_verification from a bounded event readback; mismatch/unavailable warnings preserve the known event_id and must be inspected before further writes. Before workout creates or updates, present a human-readable preview for user approval that covers total duration, key steps, target intensities, load/distance/time changes, and what existing title/prose/tags/structured steps are preserved. `description` is a replacement for the upstream event description/DSL, not append-only notes; omit it on updates to leave the current description unchanged. For WORKOUT updates, supplying `description` without `workout_doc` can replace existing structured steps; include the desired `workout_doc` to preserve or merge structure. When both are supplied, icuvisor merges them into the upstream description DSL and the `<!-- icuvisor:steps -->` sentinel controls serialized-step placement. For WeightTraining, keep exercise prescriptions in free-text description as one Markdown '- ' line per exercise, with no workout_doc; this is not native structured strength data. WORKOUT `type` supplies sport context for structured zone targets, allowing icuvisor to emit metric suffixes when needed, such as `Z2 Power`, `Z2 HR`, or `Z2 Pace`. Prefer `workout_doc` for supported endurance structure, and call `validate_workout` first if uncertain about the DSL syntax (see icuvisor://workout-syntax)."
	descriptionOnlyWorkoutWarning           = "Description was written without workout_doc; if this item previously had structured steps, they may have been replaced. Include workout_doc when preserving or merging workout structure."
	sameDayConflictWarning                  = "Same-day calendar events already exist; verify this create is not an unintended duplicate."
	duplicateCreateSkippedWarning           = "Skipped create because an existing same-day event already matches the requested writable fields."
	duplicateExternalIDSkippedWarning       = "Skipped create because an existing same-day event already has the requested external_id."
	rescheduleOccupiedDateWarning           = "Target date already contains other events; only the requested event_id was updated. Verify the reschedule did not create an unintended overlap."
	writeReturnedByUpstreamStatus           = "write_returned_by_upstream"
	skippedExistingEventStatus              = "skipped_existing_event"
	verifiedAfterInterruptedWriteStatus     = "verified_after_interrupted_write"
	invalidAddOrUpdateEventArgumentsMessage = "invalid add_or_update_event arguments; provide athlete-local date as YYYY-MM-DD, optional start_time as HH:MM, category, type for WORKOUT events, name for NOTE creates, and optional event_id for updates"
	uncertainEventCreateMessage             = "calendar event identity is missing or ambiguous; inspect the target date in intervals.icu and use event_id for an intentional edit"
	writeEventMessage                       = "could not write event; check intervals.icu credentials, athlete ID, event ID, and writable event fields"
)

var errEventCreateIdentityUncertain = errors.New("calendar event identity is missing or ambiguous")

// EventWriterClient creates or updates intervals.icu calendar events for tools.
type EventWriterClient interface {
	AddOrUpdateEvent(context.Context, intervals.WriteEventParams) (intervals.Event, error)
}

type addOrUpdateEventRequest struct {
	Date               string                 `json:"date"`
	StartTime          *string                `json:"start_time,omitempty"`
	EndDateLocal       *string                `json:"end_date_local,omitempty"`
	EventID            string                 `json:"event_id,omitempty"`
	ExternalID         string                 `json:"external_id,omitempty"`
	Category           string                 `json:"category"`
	Type               string                 `json:"type,omitempty"`
	Name               string                 `json:"name,omitempty"`
	Description        *string                `json:"description,omitempty"`
	WorkoutDoc         *workoutdoc.WorkoutDoc `json:"workout_doc,omitempty"`
	Tags               []string               `json:"tags,omitempty"`
	Color              *string                `json:"color,omitempty"`
	NotOnFitnessChart  *bool                  `json:"not_on_fitness_chart,omitempty"`
	Indoor             *bool                  `json:"indoor,omitempty"`
	HideFromAthlete    *bool                  `json:"hide_from_athlete,omitempty"`
	AthleteCannotEdit  *bool                  `json:"athlete_cannot_edit,omitempty"`
	TargetLoad         *float64               `json:"target_load,omitempty"`
	DistanceMeters     *float64               `json:"distance_meters,omitempty"`
	MovingTimeSeconds  *int                   `json:"moving_time_seconds,omitempty"`
	ElapsedTimeSeconds *int                   `json:"elapsed_time_seconds,omitempty"`
	IncludeFull        bool                   `json:"include_full,omitempty"`

	tagsProvided bool
}

type addOrUpdateEventResponse struct {
	Event getEventsRow         `json:"event"`
	Meta  addOrUpdateEventMeta `json:"_meta"`
}

type addOrUpdateEventMeta struct {
	Operation                     string                      `json:"operation"`
	Date                          string                      `json:"date"`
	Timezone                      string                      `json:"timezone"`
	WorkoutDocUploaded            string                      `json:"workout_doc_uploaded,omitempty"`
	WorkoutDocWarning             string                      `json:"workout_doc_warning,omitempty"`
	DescriptionOnlyWorkoutWarning string                      `json:"description_only_workout_warning,omitempty"`
	DuplicateWarning              string                      `json:"duplicate_warning,omitempty"`
	DuplicateEventID              string                      `json:"duplicate_event_id,omitempty"`
	SameDayConflicts              []applyTrainingPlanConflict `json:"same_day_conflicts,omitempty"`
	DurationVerification          *eventDurationVerification  `json:"duration_verification,omitempty"`
	ConfirmationStatus            string                      `json:"confirmation_status"`
	IncludeFull                   bool                        `json:"include_full"`
}

type eventCreatePreflightResult struct {
	Duplicate *intervals.Event
	Conflicts []applyTrainingPlanConflict
}

type eventCalendarReaderClient interface {
	ListEvents(context.Context, intervals.ListEventsParams) ([]intervals.Event, error)
}

func newAddOrUpdateEventTool(client EventWriterClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return coreTool(Tool{Name: addOrUpdateEventName, Description: addOrUpdateEventDescription, InputSchema: addOrUpdateEventInputSchema(), OutputSchema: addOrUpdateEventOutputSchema(), Requirement: RequirementWrite, Handler: addOrUpdateEventHandler(client, profileClient, version, timezoneFallback, debugMetadata, shapeCfg)})
}

func addOrUpdateEventHandler(client EventWriterClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeAddOrUpdateEventRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidAddOrUpdateEventArgumentsMessage, err)
		}
		profile, unitSystem, timezoneName, err := toolProfileDetails(ctx, profileClient, timezoneFallback)
		if err != nil {
			return Result{}, NewUserError(writeEventMessage, err)
		}
		if client == nil {
			return Result{}, NewUserError(writeEventMessage, errors.New("missing event writer client"))
		}
		params, workoutDocUploaded, err := eventWriteParams(args, workoutDocSerializeOptionsForSport(profile, args.Type))
		if err != nil {
			return Result{}, NewUserError(invalidAddOrUpdateEventArgumentsMessage, err)
		}
		var preflight eventCreatePreflightResult
		if args.EventID == "" {
			preflight, err = preflightEventCreateDuplicate(ctx, client, params)
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return Result{}, err
				}
				if errors.Is(err, errEventCreateIdentityUncertain) {
					return Result{}, NewUserError(uncertainEventCreateMessage, err)
				}
				return Result{}, NewUserError(writeEventMessage, err)
			}
			if preflight.Duplicate != nil {
				payload, err := shapeAddOrUpdateEventResponse(*preflight.Duplicate, args, timezoneName, workoutDocUploaded, profile, unitSystem, preflight)
				if err != nil {
					return Result{}, fmt.Errorf("shaping add_or_update_event duplicate response: %w", err)
				}
				payload.Meta.DurationVerification = verifyEventDuration(ctx, client, *preflight.Duplicate, args, false)
				return encodeShaped(payload, args.IncludeFull, nil, version, debugMetadata, addOrUpdateEventName, unitSystem, shapeCfg)
			}
		} else {
			preflight, err = preflightEventUpdate(ctx, client, params)
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return Result{}, err
				}
				return Result{}, NewUserError(writeEventMessage, err)
			}
		}
		event, err := client.AddOrUpdateEvent(ctx, params)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			if errors.Is(err, intervals.ErrEventTimingUnavailable) {
				return Result{}, NewUserError("could not read existing event timing; retry or provide start_time and end_date_local explicitly", err)
			}
			if errors.Is(err, intervals.ErrEventEndBeforeStart) {
				return Result{}, NewUserError("event end would precede its start; provide end_date_local on or after the new start", err)
			}
			if args.EventID == "" {
				recovered, recoveryPreflight, recoveryErr := recoverInterruptedEventCreate(ctx, client, params)
				if recoveryErr != nil && (errors.Is(recoveryErr, context.Canceled) || errors.Is(recoveryErr, context.DeadlineExceeded)) {
					return Result{}, recoveryErr
				}
				if errors.Is(recoveryErr, errEventCreateIdentityUncertain) {
					return Result{}, NewUserError(uncertainEventCreateMessage, recoveryErr)
				}
				if recovered != nil {
					payload, shapeErr := shapeAddOrUpdateEventResponse(*recovered, args, timezoneName, workoutDocUploaded, profile, unitSystem, recoveryPreflight)
					if shapeErr != nil {
						return Result{}, fmt.Errorf("shaping add_or_update_event recovered response: %w", shapeErr)
					}
					payload.Meta.Operation = "create_recovered"
					payload.Meta.ConfirmationStatus = verifiedAfterInterruptedWriteStatus
					payload.Meta.DurationVerification = verifyEventDuration(ctx, client, *recovered, args, false)
					return encodeShaped(payload, args.IncludeFull, nil, version, debugMetadata, addOrUpdateEventName, unitSystem, shapeCfg)
				}
			}
			return Result{}, NewUserError(writeEventMessage, err)
		}
		payload, err := shapeAddOrUpdateEventResponse(event, args, timezoneName, workoutDocUploaded, profile, unitSystem, preflight)
		if err != nil {
			return Result{}, fmt.Errorf("shaping add_or_update_event response: %w", err)
		}
		payload.Meta.DurationVerification = verifyEventDuration(ctx, client, event, args, true)
		return encodeShaped(payload, args.IncludeFull, nil, version, debugMetadata, addOrUpdateEventName, unitSystem, shapeCfg)
	}
}

type eventDurationVerification struct {
	Status  string               `json:"status"`
	Source  string               `json:"source"`
	Fields  []eventDurationField `json:"fields"`
	Warning string               `json:"warning,omitempty"`
}

type eventDurationField struct {
	Field            string `json:"field"`
	RequestedSeconds int    `json:"requested_seconds"`
	StoredSeconds    *int   `json:"stored_seconds,omitempty"`
	TargetSeconds    *int   `json:"target_seconds,omitempty"`
	Status           string `json:"status"`
}

func verifyEventDuration(ctx context.Context, client EventWriterClient, event intervals.Event, args addOrUpdateEventRequest, readBack bool) *eventDurationVerification {
	moving, elapsed := args.MovingTimeSeconds, args.ElapsedTimeSeconds
	if moving == nil && args.WorkoutDoc != nil && fullyTimedEventSteps(args.WorkoutDoc.Steps) {
		seconds := workoutdoc.EstimateDurationSeconds(*args.WorkoutDoc)
		moving = &seconds
	}
	if moving == nil && elapsed == nil {
		return nil
	}
	evidence := &eventDurationVerification{Status: "verified", Source: "same_day_event_list", Fields: []eventDurationField{}}
	if readBack {
		evidence.Source = "event_detail"
		reader, ok := client.(interface {
			GetEvent(context.Context, string) (intervals.Event, error)
		})
		if !ok || strings.TrimSpace(event.ID) == "" {
			evidence.Source = "unavailable"
			event = intervals.Event{}
		} else {
			stored, err := reader.GetEvent(ctx, event.ID)
			if err != nil || stored.ID != event.ID {
				evidence.Source = "unavailable"
				event = intervals.Event{}
			} else {
				event = stored
			}
		}
	}
	for _, value := range []struct {
		name                      string
		requested, stored, target *int
	}{
		{"moving_time_seconds", moving, event.MovingTime, event.TimeTarget},
		{"elapsed_time_seconds", elapsed, event.ElapsedTime, event.ElapsedTimeTarget},
	} {
		if value.requested == nil {
			continue
		}
		field := eventDurationField{Field: value.name, RequestedSeconds: *value.requested, StoredSeconds: value.stored, TargetSeconds: value.target, Status: "verified"}
		switch {
		case value.target != nil && *value.target != *value.requested:
			field.Status = "mismatch"
			evidence.Status = "mismatch"
		case value.stored == nil:
			field.Status = "unavailable"
			if evidence.Status != "mismatch" {
				evidence.Status = "unavailable"
			}
		case *value.stored != *value.requested:
			field.Status = "mismatch"
			evidence.Status = "mismatch"
		}
		evidence.Fields = append(evidence.Fields, field)
	}
	switch evidence.Status {
	case "mismatch":
		evidence.Warning = "Stored event duration differs from the requested duration; inspect the event and its description/steps before further writes. The event identity is confirmed; use event_id for an intentional correction instead of creating it again."
	case "unavailable":
		evidence.Warning = "Stored event duration could not be verified; inspect the event before further writes. A target alone does not confirm parsed workout duration. Preserve the confirmed event_id when correcting or retrying."
	}
	return evidence
}

func fullyTimedEventSteps(steps []workoutdoc.Step) bool {
	if len(steps) == 0 {
		return false
	}
	for _, step := range steps {
		if len(step.Steps) > 0 {
			if !fullyTimedEventSteps(step.Steps) {
				return false
			}
		} else if step.Duration <= 0 || step.Distance != nil || step.PressLap {
			return false
		}
	}
	return true
}

func decodeAddOrUpdateEventRequest(raw json.RawMessage) (addOrUpdateEventRequest, error) {
	fields, err := rawObjectFields(raw)
	if err != nil {
		return addOrUpdateEventRequest{}, err
	}
	var args addOrUpdateEventRequest
	if strings.TrimSpace(string(raw)) == "" {
		return args, errors.New("arguments must be a JSON object")
	}
	decoded, err := DecodeStrict[addOrUpdateEventRequest](raw)
	if err != nil {
		return args, err
	}
	args = decoded
	args.Date = strings.TrimSpace(args.Date)
	args.EventID = strings.TrimSpace(args.EventID)
	args.ExternalID = strings.TrimSpace(args.ExternalID)
	args.Category = strings.TrimSpace(args.Category)
	args.Type = strings.TrimSpace(args.Type)
	args.Name = strings.TrimSpace(args.Name)
	if args.Color != nil {
		color := strings.TrimSpace(*args.Color)
		if color == "" {
			return args, errors.New("color must be non-empty when provided")
		}
		args.Color = &color
	}
	args.tagsProvided = fields["tags"]
	if !validDate(args.Date) {
		return args, errors.New("date must be athlete-local YYYY-MM-DD")
	}
	start, err := intervals.EventStartDateLocal(args.Date, args.StartTime)
	if err != nil {
		return args, err
	}
	if err := intervals.ValidateEventEndDateLocal(start, args.EndDateLocal); err != nil {
		return args, err
	}
	if args.Category == "" {
		return args, errors.New("category is required")
	}
	if strings.EqualFold(args.Category, "WORKOUT") && args.Type == "" {
		return args, errors.New("type is required for WORKOUT events")
	}
	if args.EventID == "" && strings.EqualFold(args.Category, "NOTE") && args.Name == "" {
		return args, errors.New("name is required for NOTE creates")
	}
	if args.MovingTimeSeconds != nil && *args.MovingTimeSeconds < 0 {
		return args, errors.New("moving_time_seconds must be non-negative")
	}
	if args.ElapsedTimeSeconds != nil && *args.ElapsedTimeSeconds < 0 {
		return args, errors.New("elapsed_time_seconds must be non-negative")
	}
	if args.TargetLoad != nil && *args.TargetLoad < 0 {
		return args, errors.New("target_load must be non-negative")
	}
	if args.DistanceMeters != nil && *args.DistanceMeters < 0 {
		return args, errors.New("distance_meters must be non-negative")
	}
	return args, nil
}

func eventWriteParams(args addOrUpdateEventRequest, options workoutdoc.SerializeOptions) (intervals.WriteEventParams, string, error) {
	params := intervals.WriteEventParams{
		EventID:            args.EventID,
		ExternalID:         args.ExternalID,
		Date:               args.Date,
		StartTime:          args.StartTime,
		EndDateLocal:       args.EndDateLocal,
		Category:           args.Category,
		Type:               args.Type,
		Name:               args.Name,
		Description:        args.Description,
		Tags:               append([]string(nil), args.Tags...),
		TagsSet:            args.tagsProvided,
		Color:              args.Color,
		NotOnFitnessChart:  args.NotOnFitnessChart,
		Indoor:             args.Indoor,
		HideFromAthlete:    args.HideFromAthlete,
		AthleteCannotEdit:  args.AthleteCannotEdit,
		TargetLoad:         args.TargetLoad,
		DistanceMeters:     args.DistanceMeters,
		MovingTimeSeconds:  args.MovingTimeSeconds,
		ElapsedTimeSeconds: args.ElapsedTimeSeconds,
	}
	if args.WorkoutDoc == nil {
		return params, "", nil
	}
	prose := ""
	if args.Description != nil {
		prose = *args.Description
	}
	dsl, err := workoutdoc.MergeDescriptionWithOptions(prose, *args.WorkoutDoc, options)
	if err != nil {
		return intervals.WriteEventParams{}, "", fmt.Errorf("merging workout_doc with description: %w", err)
	}
	params.Description = &dsl
	return params, "description_dsl", nil
}

func shapeAddOrUpdateEventResponse(event intervals.Event, args addOrUpdateEventRequest, timezoneName string, workoutDocUploaded string, profile intervals.AthleteWithSportSettings, unitSystem response.UnitSystem, preflight eventCreatePreflightResult) (addOrUpdateEventResponse, error) {
	row, err := eventRow(event, args.IncludeFull, timezoneName, workoutPreviewContextForEvent(event, profile, unitSystem))
	if err != nil {
		return addOrUpdateEventResponse{}, err
	}
	operation := "create"
	if args.EventID != "" {
		operation = "update"
	}
	meta := addOrUpdateEventMeta{Operation: operation, Date: args.Date, Timezone: timezoneName, WorkoutDocUploaded: workoutDocUploaded, WorkoutDocWarning: workoutDocRenderWarning(args.WorkoutDoc, event.WorkoutDoc), DescriptionOnlyWorkoutWarning: addOrUpdateEventDescriptionOnlyWorkoutWarning(args), ConfirmationStatus: writeReturnedByUpstreamStatus, IncludeFull: args.IncludeFull}
	if preflight.Duplicate != nil {
		meta.Operation = "skip_duplicate"
		meta.ConfirmationStatus = skippedExistingEventStatus
		meta.DuplicateEventID = preflight.Duplicate.ID
		meta.DuplicateWarning = duplicateCreateSkippedWarning
		if len(preflight.Conflicts) > 0 && preflight.Conflicts[0].Reason == "matching_external_id" {
			meta.DuplicateWarning = duplicateExternalIDSkippedWarning
		}
		meta.SameDayConflicts = preflight.Conflicts
	} else if len(preflight.Conflicts) > 0 {
		meta.DuplicateWarning = sameDayConflictWarning
		if args.EventID != "" {
			meta.DuplicateWarning = rescheduleOccupiedDateWarning
		}
		meta.SameDayConflicts = preflight.Conflicts
	}
	return addOrUpdateEventResponse{Event: row, Meta: meta}, nil
}

func addOrUpdateEventDescriptionOnlyWorkoutWarning(args addOrUpdateEventRequest) string {
	if args.EventID == "" || args.Description == nil || args.WorkoutDoc != nil || !strings.EqualFold(args.Category, "WORKOUT") {
		return ""
	}
	return descriptionOnlyWorkoutWarning
}

func preflightEventCreateDuplicate(ctx context.Context, client EventWriterClient, params intervals.WriteEventParams) (eventCreatePreflightResult, error) {
	reader, ok := client.(eventCalendarReaderClient)
	if !ok {
		return eventCreatePreflightResult{}, nil
	}
	events, err := reader.ListEvents(ctx, intervals.ListEventsParams{Oldest: params.Date, Newest: params.Date, Limit: maxEventsLimit})
	if err != nil {
		return eventCreatePreflightResult{}, fmt.Errorf("preflighting same-day events: %w", err)
	}
	return eventCreatePreflightFromEvents(params, events)
}

func preflightEventUpdate(ctx context.Context, client EventWriterClient, params intervals.WriteEventParams) (eventCreatePreflightResult, error) {
	reader, ok := client.(eventCalendarReaderClient)
	if !ok {
		return eventCreatePreflightResult{}, nil
	}
	events, err := reader.ListEvents(ctx, intervals.ListEventsParams{Oldest: params.Date, Newest: params.Date, Limit: maxEventsLimit})
	if err != nil {
		return eventCreatePreflightResult{}, fmt.Errorf("preflighting update target date: %w", err)
	}
	result := eventCreatePreflightResult{}
	for _, event := range events {
		if eventDateOnly(event) != params.Date || event.ID == params.EventID {
			continue
		}
		result.Conflicts = append(result.Conflicts, eventWriteConflict(event, "existing_event_on_date"))
	}
	return result, nil
}

func recoverInterruptedEventCreate(ctx context.Context, client EventWriterClient, params intervals.WriteEventParams) (*intervals.Event, eventCreatePreflightResult, error) {
	reader, ok := client.(eventCalendarReaderClient)
	if !ok {
		return nil, eventCreatePreflightResult{}, nil
	}
	events, err := reader.ListEvents(ctx, intervals.ListEventsParams{Oldest: params.Date, Newest: params.Date, Limit: maxEventsLimit})
	if err != nil {
		return nil, eventCreatePreflightResult{}, fmt.Errorf("verifying interrupted event create: %w", err)
	}
	preflight, err := eventCreatePreflightFromEvents(params, events)
	return preflight.Duplicate, preflight, err
}

func eventCreatePreflightFromEvents(params intervals.WriteEventParams, events []intervals.Event) (eventCreatePreflightResult, error) {
	result := eventCreatePreflightResult{}
	duplicateReason := ""
	for _, event := range events {
		if strings.TrimSpace(event.ID) == "" {
			return eventCreatePreflightResult{}, errEventCreateIdentityUncertain
		}
		if eventDateOnly(event) != params.Date {
			continue
		}
		reason := ""
		if eventMatchesExternalID(event, params.ExternalID) {
			reason = "matching_external_id"
		} else if eventMatchesWriteParams(event, params) {
			reason = "duplicate_existing_event"
		}
		if reason == "" {
			result.Conflicts = append(result.Conflicts, eventWriteConflict(event, "existing_event_on_date"))
			continue
		}
		if result.Duplicate != nil && result.Duplicate.ID != event.ID {
			return eventCreatePreflightResult{}, errEventCreateIdentityUncertain
		}
		if result.Duplicate == nil {
			duplicate := event
			result.Duplicate = &duplicate
			duplicateReason = reason
		}
	}
	if result.Duplicate != nil {
		result.Conflicts = []applyTrainingPlanConflict{{EventID: result.Duplicate.ID, Date: eventDateOnly(*result.Duplicate), Reason: duplicateReason}}
	}
	return result, nil
}

func eventWriteConflict(event intervals.Event, reason string) applyTrainingPlanConflict {
	return applyTrainingPlanConflict{
		EventID:  event.ID,
		Date:     eventDateOnly(event),
		Category: firstNonEmpty(stringValue(event.Category), anyString(event.Raw["category"])),
		Type:     firstNonEmpty(stringValue(event.Type), anyString(event.Raw["type"])),
		Name:     firstNonEmpty(stringValue(event.Name), anyString(event.Raw["name"])),
		Reason:   reason,
	}
}

func eventMatchesExternalID(event intervals.Event, externalID string) bool {
	externalID = strings.TrimSpace(externalID)
	if externalID == "" {
		return false
	}
	return strings.TrimSpace(firstNonEmpty(stringValue(event.ExternalID), anyString(event.Raw["external_id"]))) == externalID
}

func eventMatchesWriteParams(event intervals.Event, params intervals.WriteEventParams) bool {
	if eventDateOnly(event) != strings.TrimSpace(params.Date) {
		return false
	}
	wantStart, err := intervals.EventStartDateLocal(params.Date, params.StartTime)
	if err != nil {
		return false
	}
	start, err := intervals.EventStartDateLocal(stringValue(event.StartDateLocal), nil)
	if err != nil || start != wantStart {
		return false
	}
	if !sameText(firstNonEmpty(stringValue(event.Category), anyString(event.Raw["category"])), params.Category) {
		return false
	}
	if !sameText(stringValue(event.Type), params.Type) {
		return false
	}
	if strings.TrimSpace(stringValue(event.Name)) != strings.TrimSpace(params.Name) {
		return false
	}
	if params.Description != nil {
		if stringValue(event.Description) != *params.Description {
			return false
		}
	} else if stringValue(event.Description) != "" {
		return false
	}
	eventColor := strings.TrimSpace(firstNonEmpty(stringValue(event.Color), anyString(event.Raw["color"])))
	if params.Color != nil {
		if eventColor != strings.TrimSpace(*params.Color) {
			return false
		}
	} else if eventColor != "" {
		return false
	}
	if params.NotOnFitnessChart != nil {
		if event.NotOnFitnessChart == nil || *event.NotOnFitnessChart != *params.NotOnFitnessChart {
			return false
		}
	} else if event.NotOnFitnessChart != nil && *event.NotOnFitnessChart {
		return false
	}
	if params.Indoor != nil {
		if event.Indoor == nil || *event.Indoor != *params.Indoor {
			return false
		}
	} else if event.Indoor != nil && *event.Indoor {
		return false
	}
	if params.EndDateLocal != nil {
		if event.EndDateLocal == nil || strings.TrimSuffix(*event.EndDateLocal, "T00:00:00") != strings.TrimSuffix(*params.EndDateLocal, "T00:00:00") {
			return false
		}
	} else if event.EndDateLocal != nil && len(*event.EndDateLocal) >= len(params.Date) && (*event.EndDateLocal)[:len(params.Date)] != params.Date {
		return false
	}
	if params.HideFromAthlete != nil {
		if event.HideFromAthlete == nil || *event.HideFromAthlete != *params.HideFromAthlete {
			return false
		}
	} else if event.HideFromAthlete != nil && *event.HideFromAthlete {
		return false
	}
	if params.AthleteCannotEdit != nil {
		if event.AthleteCannotEdit == nil || *event.AthleteCannotEdit != *params.AthleteCannotEdit {
			return false
		}
	} else if event.AthleteCannotEdit != nil && *event.AthleteCannotEdit {
		return false
	}
	if params.TargetLoad != nil {
		if !sameOptionalFloat(*params.TargetLoad, event.LoadTarget) {
			return false
		}
	} else if nonZeroOptionalFloat(event.LoadTarget) {
		return false
	}
	if params.DistanceMeters != nil {
		if !sameOptionalFloat(*params.DistanceMeters, event.DistanceTarget) {
			return false
		}
	} else if nonZeroOptionalFloat(event.DistanceTarget) {
		return false
	}
	if params.MovingTimeSeconds != nil {
		if !sameOptionalInt(*params.MovingTimeSeconds, event.TimeTarget) {
			return false
		}
	} else if nonZeroOptionalInt(event.TimeTarget) {
		return false
	}
	if params.ElapsedTimeSeconds != nil {
		if !sameOptionalInt(*params.ElapsedTimeSeconds, event.ElapsedTimeTarget) {
			return false
		}
	} else if nonZeroOptionalInt(event.ElapsedTimeTarget) {
		return false
	}
	eventTagValues := []string{}
	if tags := eventTags(event.Raw); tags != nil {
		eventTagValues = *tags
	}
	if !sameStringSlice(eventTagValues, params.Tags) {
		return false
	}
	return true
}

func sameText(left string, right string) bool {
	return strings.EqualFold(strings.TrimSpace(left), strings.TrimSpace(right))
}

func sameOptionalFloat(want float64, values ...*float64) bool {
	for _, value := range values {
		if value != nil {
			return *value == want
		}
	}
	return false
}

func sameOptionalInt(want int, values ...*int) bool {
	for _, value := range values {
		if value != nil {
			return *value == want
		}
	}
	return false
}

func nonZeroOptionalFloat(values ...*float64) bool {
	for _, value := range values {
		if value != nil && *value != 0 {
			return true
		}
	}
	return false
}

func nonZeroOptionalInt(values ...*int) bool {
	for _, value := range values {
		if value != nil && *value != 0 {
			return true
		}
	}
	return false
}

func sameStringSlice(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for idx := range left {
		if left[idx] != right[idx] {
			return false
		}
	}
	return true
}

func addOrUpdateEventInputSchema() map[string]any {
	examples := addOrUpdateEventInputExamples()
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"date", "category"}, "examples": examples, "input_examples": examples, "properties": map[string]any{
		"date":                 map[string]any{"type": "string", "description": "Required athlete-local event date as YYYY-MM-DD; interpreted in the configured athlete timezone."},
		"start_time":           map[string]any{"type": "string", "pattern": "^([01][0-9]|2[0-3]):[0-5][0-9]$", "description": "Optional athlete-local start time as HH:MM in 24-hour format, interpreted in the configured athlete timezone. Omit on creates for a date-only event. Omit on updates or reschedules to preserve the existing clock time; this requires reading the event. Set 00:00 explicitly to reset to the upstream date-only convention."},
		"end_date_local":       map[string]any{"type": "string", "description": "Optional athlete-local end as YYYY-MM-DD or YYYY-MM-DDTHH:MM:SS. Must be on or after the full event start, including start_time or the preserved clock time. Omit on updates to preserve the existing end; do not send for ordinary single-day events."},
		"event_id":             map[string]any{"type": "string", "description": "Optional upstream event ID to update. Omit to create a new event; this tool never deletes events."},
		"external_id":          map[string]any{"type": "string", "description": "Optional non-empty upstream idempotency key. Surrounding whitespace is trimmed; omit or pass blank to leave external_id unchanged/unset. Clearing an existing upstream external_id is not supported."},
		"category":             map[string]any{"type": "string", "description": intervals.EventCategoryReferenceDescription("Required upstream event category.")},
		"type":                 map[string]any{"type": "string", "description": "Required for WORKOUT events: upstream sport/activity type such as Ride, Run, Swim, or the athlete account's configured activity type. Surrounding whitespace is trimmed. Used with athlete sport settings to disambiguate structured workout_doc zone targets."},
		"name":                 map[string]any{"type": "string", "description": "Event title/name shown on the athlete calendar. Required when creating NOTE events; optional for other supported writes and updates unless upstream requires it."},
		"description":          map[string]any{"type": "string", "description": "Optional replacement for the upstream event description/DSL, not append-only notes. Omit on updates to leave unchanged. For WORKOUT updates, supplying description without workout_doc can replace existing structured steps; include the desired workout_doc to preserve or merge structure. Preserved verbatim, including whitespace and line breaks. For a WeightTraining plan, use a blank line before a Markdown list and put each exercise on its own '- ' line (not '•'); include only known sets/reps or time/distance, load with units and basis (total, per hand, or per side), rest, effort, and cues. Keep strength sets out of workout_doc. May be supplied with workout_doc for supported endurance workouts; use the " + workoutdoc.StepsSentinel + " sentinel on its own line to choose where serialized steps are inserted."},
		"workout_doc":          map[string]any{"type": "object", "description": "Optional structured WorkoutDoc. Serialized to the upstream workout DSL and merged with description when both are supplied. For WORKOUT updates, include the desired structured steps when changing prose so the replacement description/DSL preserves the workout structure. In each structured step, description is a label/comment only: do not include duration or distance tokens there; use duration seconds or distance instead. Set press_lap:true on a timed/distance step to serialize the documented Press lap control; do not put that marker in description. Zone targets are serialized using the WORKOUT type and athlete sport settings, adding metric suffixes such as Z2 Power, Z2 HR, or Z2 Pace when needed. Syntax reference: icuvisor://workout-syntax."},
		"tags":                 map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Optional replacement event tags, in caller-provided order. Omit to leave unchanged on updates; provide [] to clear all tags."},
		"color":                map[string]any{"type": "string", "description": "Optional intervals.icu calendar display color. Omit on updates to leave unchanged; surrounding whitespace is trimmed."},
		"not_on_fitness_chart": map[string]any{"type": "boolean", "description": "Optional intervals.icu flag that hides this event from the Fitness chart. Omit on updates to leave unchanged; set false explicitly to show it."},
		"indoor":               map[string]any{"type": "boolean", "description": "Optional planned-event indoor/trainer flag. Set true for indoor trainer rides; commonly paired with type VirtualRide, but this boolean controls intervals.icu's Indoor toggle."},
		"hide_from_athlete":    map[string]any{"type": "boolean", "description": "Coach visibility control. True hides this event from the athlete; false makes it visible. Omit on updates to preserve current visibility. Confirm the intended athlete visibility before writing."},
		"athlete_cannot_edit":  map[string]any{"type": "boolean", "description": "Coach edit lock. True prevents the athlete from editing this event; false restores edit access. Omit on updates to preserve current editability. Confirm the intended athlete editability before writing."},
		"target_load":          map[string]any{"type": "number", "minimum": 0, "description": "Optional planned training load / TSS equivalent when supported upstream."},
		"distance_meters":      map[string]any{"type": "number", "minimum": 0, "description": "Optional planned distance in meters when supported upstream."},
		"moving_time_seconds":  map[string]any{"type": "integer", "minimum": 0, "description": "Optional planned moving duration in seconds when supported upstream."},
		"elapsed_time_seconds": map[string]any{"type": "integer", "minimum": 0, "description": "Optional planned elapsed duration in seconds when supported upstream."},
		"include_full":         map[string]any{"type": "boolean", "default": false, "description": "When true, include the raw upstream event payload under event.full; default response matches get_event_by_id's terse event shape."},
	}}
}

func addOrUpdateEventInputExamples() []map[string]any {
	return []map[string]any{
		{
			"date":        "2026-06-12",
			"start_time":  "07:30",
			"category":    "NOTE",
			"name":        "Race-week nutrition plan",
			"description": "Breakfast: oatmeal and banana. Lunch: rice bowl. Carry 90 g carbs/hour for the long ride.",
		},
		{
			"date":                "2026-07-08",
			"category":            "WORKOUT",
			"type":                "WeightTraining",
			"name":                "Strength A",
			"description":         "## Strength A\n\n- Back squat: 3 × 5 @ 80 kg total; rest 2 min; target RPE 7.\n- Split squat: 3 × 8/side @ 16 kg per hand; rest 90 s.",
			"moving_time_seconds": 2700,
		},
		{
			"date":        "2026-06-15",
			"category":    "NOTE",
			"name":        "Travel logistics",
			"description": "Flight lands at 14:20. Pack pedals, charger, spare cleats, bottles, and race license.",
		},
		{
			"date":        "2026-06-17",
			"category":    "NOTE",
			"name":        "Daily reminder",
			"description": "Take resting HR after waking, do 10 minutes mobility, and log sleep quality before training.",
		},
		{
			"date":        "2026-06-18",
			"category":    "NOTE",
			"name":        "Coach annotation",
			"description": "Athlete reported tight calves; keep Thursday aerobic and reassess before intensity.",
		},
		{
			"date":     "2026-06-16",
			"category": "WORKOUT",
			"type":     "Ride",
			"name":     "Sweet spot repeats",
			"workout_doc": map[string]any{
				"steps": []any{
					map[string]any{"description": "Warm up", "duration": 600, "power": map[string]any{"value": 60, "units": "PERCENT_FTP"}},
					map[string]any{"description": "Main set", "reps": 3, "steps": []any{
						map[string]any{"duration": 480, "power": map[string]any{"value": 88, "units": "PERCENT_FTP"}},
						map[string]any{"duration": 240, "power": map[string]any{"value": 55, "units": "PERCENT_FTP"}},
					}},
				},
			},
			"tags":                []any{"sweet-spot", "indoor"},
			"indoor":              true,
			"target_load":         72,
			"moving_time_seconds": 4200,
		},
		{
			"date":                 "2026-09-13",
			"category":             "RACE_A",
			"type":                 "Run",
			"name":                 "Goal marathon",
			"description":          "A race. Confirm taper, fueling, pacing, and weather plan before adding workouts around it.",
			"tags":                 []any{"race", "goal-race"},
			"distance_meters":      42195,
			"target_load":          210,
			"elapsed_time_seconds": 10800,
		},
		{
			"event_id":             "evt-example-42",
			"date":                 "2026-06-21",
			"category":             "RACE_B",
			"type":                 "Run",
			"name":                 "10K tune-up race",
			"description":          "B race. Practice breakfast, warm-up, and even pacing.",
			"tags":                 []any{"race", "tune-up"},
			"distance_meters":      10000,
			"target_load":          85,
			"elapsed_time_seconds": 3300,
			"include_full":         true,
		},
		{
			"date":                "2026-05-30",
			"category":            "RACE_C",
			"type":                "Ride",
			"name":                "Local criterium practice",
			"description":         "C race. Treat as skills practice; keep surrounding plan priority higher.",
			"tags":                []any{"race", "practice"},
			"distance_meters":     40000,
			"target_load":         95,
			"moving_time_seconds": 5400,
		},
	}
}

func addOrUpdateEventOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Write confirmation containing the same terse event row shape used by get_event_by_id plus operation/date/timezone metadata. _meta.confirmation_status is write_returned_by_upstream when intervals.icu returned a write payload, skipped_existing_event when icuvisor skipped a duplicate create before writing, or verified_after_interrupted_write after a bounded same-day read confirmed a create whose response was interrupted. Updates report other destination-date events under _meta.same_day_conflicts without replacing them. _meta.workout_doc_warning is set when intervals.icu stored the event but did not parse the uploaded workout_doc into a graphically rendered structured workout. _meta.description_only_workout_warning is set for WORKOUT event updates that supplied description without workout_doc."}
}

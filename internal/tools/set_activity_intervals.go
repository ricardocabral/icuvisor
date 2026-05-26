package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/workoutdoc"
)

const (
	setActivityIntervalsName                    = "set_activity_intervals"
	setActivityIntervalsDescription             = "Set or correct the interval structure on one completed activity by writing a structured workout_doc as the activity's description; intervals.icu re-parses the DSL into rendered intervals overlaid on the recording. This is destructive: it overwrites the activity's description (which may include any prior interval DSL). Registered only when ICUVISOR_DELETE_MODE=full. To rename or edit only the free-text description (without changing structured intervals), use update_activity instead."
	invalidSetActivityIntervalsArgumentsMessage = "invalid set_activity_intervals arguments; provide activity_id and a non-empty workout_doc with at least one step"
	setActivityIntervalsMessage                 = "could not set activity intervals; check intervals.icu credentials, activity ID, workout_doc syntax, and delete-mode configuration"
)

type setActivityIntervalsRequest struct {
	ActivityID  string                 `json:"activity_id"`
	WorkoutDoc  *workoutdoc.WorkoutDoc `json:"workout_doc"`
	Prose       string                 `json:"prose,omitempty"`
	IncludeFull bool                   `json:"include_full,omitempty"`

	proseProvided bool
}

type setActivityIntervalsResponse struct {
	ActivityID         string                   `json:"activity_id"`
	Status             string                   `json:"status"`
	WorkoutDocUploaded string                   `json:"workout_doc_uploaded"`
	Full               map[string]any           `json:"full,omitempty"`
	Meta               setActivityIntervalsMeta `json:"_meta"`
}

type setActivityIntervalsMeta struct {
	AthleteID            string `json:"athleteId,omitempty"`
	Destructive          bool   `json:"destructive"`
	SourceEndpoint       string `json:"source_endpoint"`
	IntervalSourceIntent string `json:"interval_source_intent"`
	WorkoutDocWarning    string `json:"workout_doc_warning,omitempty"`
	IncludeFull          bool   `json:"include_full"`
}

func newSetActivityIntervalsTool(client ActivityUpdaterClient, profileClient ProfileClient, version string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return fullTool(Tool{Name: setActivityIntervalsName, Description: setActivityIntervalsDescription, InputSchema: setActivityIntervalsInputSchema(), OutputSchema: setActivityIntervalsOutputSchema(), Requirement: RequirementDelete, Handler: setActivityIntervalsHandler(client, profileClient, version, debugMetadata, shapeCfg)})
}

func setActivityIntervalsHandler(client ActivityUpdaterClient, profileClient ProfileClient, version string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeSetActivityIntervalsRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidSetActivityIntervalsArgumentsMessage, err)
		}
		if client == nil {
			return Result{}, NewUserError(setActivityIntervalsMessage, errors.New("missing activity updater client"))
		}
		description, err := buildActivityIntervalsDescription(*args.WorkoutDoc, args.Prose, args.proseProvided)
		if err != nil {
			return Result{}, NewUserError(invalidSetActivityIntervalsArgumentsMessage, err)
		}
		params := intervals.UpdateActivityParams{ActivityID: args.ActivityID, Description: description, DescriptionSet: true}
		activity, err := client.UpdateActivity(ctx, params)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(setActivityIntervalsMessage, err)
		}
		athleteID := ""
		if profileClient != nil {
			profile, profileErr := profileClient.GetAthleteProfile(ctx)
			if profileErr != nil {
				if errors.Is(profileErr, context.Canceled) || errors.Is(profileErr, context.DeadlineExceeded) {
					return Result{}, profileErr
				}
			} else {
				athleteID = config.NormalizeAthleteIDForDisplay(profile.ID)
			}
		}
		payload := setActivityIntervalsResponse{
			ActivityID:         args.ActivityID,
			Status:             "intervals_set",
			WorkoutDocUploaded: "description_dsl",
			Meta: setActivityIntervalsMeta{
				AthleteID:            athleteID,
				Destructive:          true,
				SourceEndpoint:       "/activity/{activityId}",
				IntervalSourceIntent: "structured_workout",
				WorkoutDocWarning:    activityIntervalsWorkoutDocWarning(activity),
				IncludeFull:          args.IncludeFull,
			},
		}
		if args.IncludeFull {
			payload.Full = activity.Raw
		}
		return encodeShaped(payload, args.IncludeFull, nil, version, debugMetadata, setActivityIntervalsName, "", shapeCfg)
	}
}

func decodeSetActivityIntervalsRequest(raw json.RawMessage) (setActivityIntervalsRequest, error) {
	if strings.TrimSpace(string(raw)) == "" {
		return setActivityIntervalsRequest{}, errors.New("arguments must be a JSON object")
	}
	fields, err := rawObjectFields(raw)
	if err != nil {
		return setActivityIntervalsRequest{}, err
	}
	decoded, err := DecodeStrict[setActivityIntervalsRequest](raw)
	if err != nil {
		return setActivityIntervalsRequest{}, err
	}
	decoded.ActivityID = strings.TrimSpace(decoded.ActivityID)
	if decoded.ActivityID == "" {
		return setActivityIntervalsRequest{}, errors.New("activity_id is required")
	}
	if decoded.WorkoutDoc == nil {
		return setActivityIntervalsRequest{}, errors.New("workout_doc is required and must contain at least one step")
	}
	if len(decoded.WorkoutDoc.Steps) == 0 {
		return setActivityIntervalsRequest{}, errors.New("workout_doc.steps must contain at least one step; use update_activity to clear the description instead")
	}
	decoded.proseProvided = fields["prose"]
	return decoded, nil
}

func buildActivityIntervalsDescription(doc workoutdoc.WorkoutDoc, prose string, proseProvided bool) (string, error) {
	if !proseProvided {
		dsl, err := workoutdoc.Serialize(doc)
		if err != nil {
			return "", fmt.Errorf("serializing workout_doc: %w", err)
		}
		return dsl, nil
	}
	merged, err := workoutdoc.MergeDescription(prose, doc)
	if err != nil {
		return "", fmt.Errorf("merging workout_doc with prose: %w", err)
	}
	return merged, nil
}

func activityIntervalsWorkoutDocWarning(activity intervals.Activity) string {
	// Mirrors workoutDocRenderWarning's intent: if upstream stored the description
	// but did not surface a parsed workout_doc, the structured intervals may not
	// be rendered. The activity payload's icu_intervals field is the canonical
	// post-parse signal; absent or empty means the DSL was kept as plain text.
	if activity.Raw == nil {
		return ""
	}
	if value, ok := activity.Raw["icu_intervals"]; ok {
		if list, ok := value.([]any); ok && len(list) > 0 {
			return ""
		}
	}
	return "intervals.icu accepted the description but did not parse the workout_doc into rendered intervals; verify the DSL with validate_workout (see icuvisor://workout-syntax)"
}

func setActivityIntervalsInputSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"activity_id", "workout_doc"},
		"properties": map[string]any{
			"activity_id":  map[string]any{"type": "string", "description": "Required intervals.icu activity ID whose interval structure will be (re)written. Surrounding whitespace is trimmed; the i-prefix is preserved verbatim."},
			"workout_doc":  map[string]any{"type": "object", "description": "Required structured WorkoutDoc whose steps are serialized to the activity description DSL; intervals.icu re-parses the DSL into rendered intervals. Must contain at least one step. In each structured step, description is a label/comment only: do not include duration or distance tokens there; use duration seconds or distance instead. Syntax reference: icuvisor://workout-syntax. Call validate_workout first if uncertain about the DSL."},
			"prose":        map[string]any{"type": "string", "description": "Optional free-text prose to interleave around the serialized steps; the upstream description retains the prose verbatim around the DSL block. Use the " + workoutdoc.StepsSentinel + " sentinel on its own line to choose where serialized steps are inserted. Omit to send the serialized DSL alone."},
			"include_full": map[string]any{"type": "boolean", "default": false, "description": "When true, include the raw upstream updated-activity payload under full; default returns a terse confirmation."},
		},
	}
}

func setActivityIntervalsOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Destructive activity-intervals write confirmation: activity_id, status, workout_doc_uploaded marker, and _meta with destructive=true, source_endpoint, interval_source_intent=structured_workout (so downstream readers know the interval set is structured, not device auto-laps), workout_doc_warning when upstream stored the description but did not parse the DSL into rendered intervals, and normalized athlete_id when available."}
}

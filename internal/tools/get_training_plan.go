package tools

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

const (
	getTrainingPlanName                    = "get_training_plan"
	getTrainingPlanDescription             = "Fetch the athlete's active training-plan assignment, not calendar events or workout-library templates. Returns assignment and lightweight plan summary fields by default; raw nested plan/workout payloads require include_full:true."
	invalidGetTrainingPlanArgumentsMessage = "invalid get_training_plan arguments; only include_full is supported"
	fetchTrainingPlanMessage               = "could not fetch training plan; check intervals.icu credentials and athlete ID"
	trainingPlanEndpoint                   = "/athlete/{id}/training-plan"
)

// TrainingPlanClient retrieves the active intervals.icu training-plan assignment for tools.
type TrainingPlanClient interface {
	GetTrainingPlan(context.Context) (intervals.TrainingPlan, error)
}

type getTrainingPlanRequest struct {
	IncludeFull bool `json:"include_full,omitempty"`
}

type getTrainingPlanResponse struct {
	TrainingPlan *getTrainingPlanRow         `json:"training_plan,omitempty"`
	Unavailable  *trainingPlanUnavailable    `json:"unavailable,omitempty"`
	Meta         getTrainingPlanResponseMeta `json:"_meta"`
}

type getTrainingPlanRow struct {
	TrainingPlanID string         `json:"training_plan_id,omitempty"`
	PlanID         string         `json:"plan_id,omitempty"`
	Name           string         `json:"name,omitempty"`
	Alias          string         `json:"alias,omitempty"`
	StartDate      string         `json:"training_plan_start_date,omitempty"`
	Timezone       string         `json:"timezone,omitempty"`
	LastApplied    string         `json:"last_applied,omitempty"`
	PlanApplied    string         `json:"plan_applied,omitempty"`
	PlanSummary    map[string]any `json:"plan_summary,omitempty"`
	Full           map[string]any `json:"full,omitempty"`
}

type trainingPlanUnavailable struct {
	Reason string `json:"reason"`
	Detail string `json:"detail"`
}

type getTrainingPlanResponseMeta struct {
	SourceEndpoint      string `json:"source_endpoint"`
	Timezone            string `json:"timezone"`
	IncludeFull         bool   `json:"include_full"`
	AvailabilityCaveat  string `json:"availability_caveat,omitempty"`
	DefaultPayloadScope string `json:"default_payload_scope"`
}

func newGetTrainingPlanTool(client TrainingPlanClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return fullTool(Tool{Name: getTrainingPlanName, Description: getTrainingPlanDescription, InputSchema: getTrainingPlanInputSchema(), OutputSchema: getTrainingPlanOutputSchema(), Handler: getTrainingPlanHandler(client, profileClient, version, timezoneFallback, debugMetadata, shapeCfg)})
}

func getTrainingPlanHandler(client TrainingPlanClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeGetTrainingPlanRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidGetTrainingPlanArgumentsMessage, err)
		}
		unitSystem, timezoneName, err := toolProfile(ctx, profileClient, timezoneFallback)
		if err != nil {
			return Result{}, NewUserError(fetchTrainingPlanMessage, err)
		}
		if client == nil {
			return Result{}, NewUserError(fetchTrainingPlanMessage, errors.New("missing training plan client"))
		}
		plan, err := client.GetTrainingPlan(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(fetchTrainingPlanMessage, err)
		}
		payload := shapeGetTrainingPlanResponse(plan, args.IncludeFull, timezoneName)
		return encodeShaped(payload, args.IncludeFull, nil, version, debugMetadata, getTrainingPlanName, unitSystem, shapeCfg)
	}
}

func decodeGetTrainingPlanRequest(raw json.RawMessage) (getTrainingPlanRequest, error) {
	var args getTrainingPlanRequest
	if len(strings.TrimSpace(string(raw))) == 0 {
		return args, nil
	}
	decoded, err := DecodeStrict[getTrainingPlanRequest](raw)
	if err != nil {
		return args, err
	}
	args = decoded
	return args, nil
}

func shapeGetTrainingPlanResponse(plan intervals.TrainingPlan, includeFull bool, timezoneName string) getTrainingPlanResponse {
	payload := getTrainingPlanResponse{Meta: getTrainingPlanResponseMeta{SourceEndpoint: trainingPlanEndpoint, Timezone: timezoneName, IncludeFull: includeFull, DefaultPayloadScope: "assignment and lightweight plan summary only; raw nested plan/workout payloads require include_full:true"}}
	if !plan.Active {
		payload.Unavailable = &trainingPlanUnavailable{Reason: "no_active_training_plan", Detail: "intervals.icu returned no active training-plan assignment for this athlete"}
		payload.Meta.AvailabilityCaveat = "public API may expose only assignment metadata; calendar workouts remain available through get_events"
		return payload
	}
	row := getTrainingPlanRow{TrainingPlanID: plan.TrainingPlanID, PlanID: firstNonEmpty(plan.PlanID, anyString(plan.Raw["plan_id"])), Name: firstNonEmpty(stringValue(plan.Name), nestedTrainingPlanString(plan, "name")), Alias: stringValue(plan.Alias), StartDate: stringValue(plan.StartDate), Timezone: firstNonEmpty(stringValue(plan.Timezone), timezoneName), LastApplied: stringValue(plan.LastApplied), PlanApplied: stringValue(plan.PlanApplied), PlanSummary: trainingPlanSummary(plan)}
	if includeFull {
		row.Full = cloneJSONMap(plan.Raw)
	}
	payload.TrainingPlan = &row
	if row.PlanSummary == nil {
		payload.Meta.AvailabilityCaveat = "upstream response exposed training-plan assignment metadata without a nested plan summary"
	}
	return payload
}

func trainingPlanSummary(plan intervals.TrainingPlan) map[string]any {
	candidate := plan.TrainingPlan
	if candidate == nil {
		candidate = plan.Plan
	}
	if candidate == nil {
		candidate = firstRaw(plan.Raw, "training_plan", "plan")
	}
	nested, ok := candidate.(map[string]any)
	if !ok || len(nested) == 0 {
		return nil
	}
	summary := map[string]any{}
	for _, key := range []string{"id", "name", "description", "folder_id", "type", "category"} {
		if value := anyString(nested[key]); value != "" {
			summary[key] = value
		}
	}
	if children, ok := nested["children"].([]any); ok {
		summary["child_count"] = len(children)
	}
	if workouts, ok := nested["workouts"].([]any); ok {
		summary["workout_count"] = len(workouts)
	}
	keys := make([]string, 0, len(nested))
	for key := range nested {
		if key == "children" || key == "workout_doc" || key == "workouts" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	if len(keys) > 0 {
		summary["top_level_keys"] = keys
	}
	if len(summary) == 0 {
		return nil
	}
	return summary
}

func nestedTrainingPlanString(plan intervals.TrainingPlan, key string) string {
	for _, candidate := range []any{plan.TrainingPlan, plan.Plan, firstRaw(plan.Raw, "training_plan", "plan")} {
		if nested, ok := candidate.(map[string]any); ok {
			if value := anyString(nested[key]); value != "" {
				return value
			}
		}
	}
	return ""
}

func getTrainingPlanInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "properties": map[string]any{
		"include_full": map[string]any{"type": "boolean", "default": false, "description": "When true, include the raw upstream training-plan assignment and nested plan/workout payload. Default mode returns assignment fields and lightweight plan summary only."},
	}}
}

func getTrainingPlanOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Active training-plan assignment and lightweight plan summary, or a structured no_active_training_plan unavailable response."}
}

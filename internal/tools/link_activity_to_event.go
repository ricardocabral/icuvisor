package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

const (
	linkActivityToEventName                    = "link_activity_to_event"
	linkActivityToEventDescription             = "Manually pair one completed activity with one planned calendar event when intervals.icu auto-pairing misses (forum #97). This is a non-destructive write: it sets the activity's paired_event_id and does not delete activities or events."
	invalidLinkActivityToEventArgumentsMessage = "invalid link_activity_to_event arguments; provide non-empty activity_id and numeric event_id"
	linkActivityToEventMessage                 = "could not link activity to event; check intervals.icu credentials, activity ID, and event ID"
)

// ActivityEventLinkClient links activities to planned events.
type ActivityEventLinkClient interface {
	LinkActivityToEvent(context.Context, intervals.LinkActivityToEventParams) (intervals.Activity, error)
}

type linkActivityToEventRequest struct {
	ActivityID  string `json:"activity_id"`
	EventID     string `json:"event_id"`
	IncludeFull bool   `json:"include_full,omitempty"`
}

type linkActivityToEventResponse struct {
	ActivityID string                  `json:"activity_id"`
	EventID    string                  `json:"event_id"`
	Status     string                  `json:"status"`
	Full       map[string]any          `json:"full,omitempty"`
	Meta       linkActivityToEventMeta `json:"_meta"`
}

type linkActivityToEventMeta struct {
	Warnings    []linkActivityToEventWarning `json:"warnings,omitempty"`
	IncludeFull bool                         `json:"include_full"`
}

type linkActivityToEventWarning struct {
	Code         string `json:"code"`
	Message      string `json:"message"`
	ActivityDate string `json:"activity_date,omitempty"`
	EventDate    string `json:"event_date,omitempty"`
}

func newLinkActivityToEventTool(client ActivityEventLinkClient, activityClient ActivityDetailsClient, eventClient EventByIDClient, version string, debugMetadata bool) Tool {
	return coreTool(Tool{Name: linkActivityToEventName, Description: linkActivityToEventDescription, InputSchema: linkActivityToEventInputSchema(), OutputSchema: linkActivityToEventOutputSchema(), Requirement: RequirementWrite, Handler: linkActivityToEventHandler(client, activityClient, eventClient, version, debugMetadata)})
}

func linkActivityToEventHandler(client ActivityEventLinkClient, activityClient ActivityDetailsClient, eventClient EventByIDClient, version string, debugMetadata bool) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeLinkActivityToEventRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidLinkActivityToEventArgumentsMessage, err)
		}
		if client == nil {
			return Result{}, NewUserError(linkActivityToEventMessage, errors.New("missing activity/event link client"))
		}
		linked, err := client.LinkActivityToEvent(ctx, intervals.LinkActivityToEventParams{ActivityID: args.ActivityID, EventID: args.EventID})
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(linkActivityToEventMessage, err)
		}
		warnings, err := linkActivityToEventWarnings(ctx, activityClient, eventClient, args.ActivityID, args.EventID)
		if err != nil {
			return Result{}, err
		}
		payload := linkActivityToEventResponse{ActivityID: args.ActivityID, EventID: args.EventID, Status: "linked", Meta: linkActivityToEventMeta{Warnings: warnings, IncludeFull: args.IncludeFull}}
		if args.IncludeFull {
			payload.Full = linked.Raw
		}
		return encodeShaped(payload, args.IncludeFull, nil, version, debugMetadata, linkActivityToEventName, "")
	}
}

func decodeLinkActivityToEventRequest(raw json.RawMessage) (linkActivityToEventRequest, error) {
	var args linkActivityToEventRequest
	if err := decodeStrict(raw, &args); err != nil {
		return args, err
	}
	args.ActivityID = strings.TrimSpace(args.ActivityID)
	args.EventID = strings.TrimSpace(args.EventID)
	if args.ActivityID == "" {
		return args, errors.New("activity_id is required")
	}
	if args.EventID == "" {
		return args, errors.New("event_id is required")
	}
	if _, err := intervals.ParseEventID(args.EventID); err != nil {
		return args, err
	}
	return args, nil
}

func linkActivityToEventWarnings(ctx context.Context, activityClient ActivityDetailsClient, eventClient EventByIDClient, activityID string, eventID string) ([]linkActivityToEventWarning, error) {
	if activityClient == nil || eventClient == nil {
		return nil, nil
	}
	activity, activityErr := activityClient.GetActivity(ctx, activityID)
	if activityErr != nil {
		if errors.Is(activityErr, context.Canceled) || errors.Is(activityErr, context.DeadlineExceeded) {
			return nil, activityErr
		}
		return nil, nil
	}
	event, eventErr := eventClient.GetEvent(ctx, eventID)
	if eventErr != nil {
		if errors.Is(eventErr, context.Canceled) || errors.Is(eventErr, context.DeadlineExceeded) {
			return nil, eventErr
		}
		return nil, nil
	}
	activityDate := localDatePrefix(stringValue(activity.StartDateLocal))
	eventDate := localDatePrefix(stringValue(event.StartDateLocal))
	if activityDate == "" || eventDate == "" || activityDate == eventDate {
		return nil, nil
	}
	return []linkActivityToEventWarning{{Code: "date_mismatch", Message: "activity and event start dates differ; link was applied as requested", ActivityDate: activityDate, EventDate: eventDate}}, nil
}

func localDatePrefix(value string) string {
	value = strings.TrimSpace(value)
	if len(value) < len("2006-01-02") {
		return ""
	}
	prefix := value[:len("2006-01-02")]
	if !validDate(prefix) {
		return ""
	}
	return prefix
}

func linkActivityToEventInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"activity_id", "event_id"}, "properties": map[string]any{
		"activity_id":  map[string]any{"type": "string", "description": "Required intervals.icu activity ID to pair. Surrounding whitespace is trimmed; IDs are otherwise preserved exactly."},
		"event_id":     map[string]any{"type": "string", "description": "Required numeric intervals.icu planned event ID to set as the activity paired_event_id. Surrounding whitespace is trimmed."},
		"include_full": map[string]any{"type": "boolean", "default": false, "description": "When true, include the raw upstream updated activity payload under full; default returns a terse link confirmation."},
	}}
}

func linkActivityToEventOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Non-destructive activity/event link confirmation with activity_id, event_id, status, and _meta.warnings when the activity/event dates differ."}
}

package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ricardocabral/icuvisor/internal/safety"
)

// Registry registers the MCP tools exposed by icuvisor.
type Registry interface {
	Register(context.Context, Registrar) error
}

// RegistryOptions configures the default tool registry.
type RegistryOptions struct {
	Version          string
	TimezoneFallback string
	DebugMetadata    bool
	Capability       safety.Capability
}

// NewRegistry creates the default tool registry.
func NewRegistry(profileClient ProfileClient, version string, timezoneFallback ...string) Registry {
	return NewRegistryWithOptions(profileClient, RegistryOptions{
		Version:          version,
		TimezoneFallback: firstNonEmpty(timezoneFallback...),
	})
}

// NewRegistryWithOptions creates the default registry with explicit response-shaping options.
func NewRegistryWithOptions(profileClient ProfileClient, opts RegistryOptions) Registry {
	return &defaultRegistry{
		profileClient:    profileClient,
		version:          normalizeVersion(opts.Version),
		timezoneFallback: normalizeTimezoneFallback(opts.TimezoneFallback),
		debugMetadata:    opts.DebugMetadata,
		capability:       capabilityOrSafe(opts.Capability),
	}
}

type defaultRegistry struct {
	profileClient    ProfileClient
	version          string
	timezoneFallback string
	debugMetadata    bool
	capability       safety.Capability
}

func capabilityOrSafe(capability safety.Capability) safety.Capability {
	if capability != nil {
		return capability
	}
	return safety.NewCapability(safety.ModeSafe)
}

func (r *defaultRegistry) Register(ctx context.Context, registrar Registrar) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if r.profileClient == nil {
		return fmt.Errorf("registering %s: missing profile client", getAthleteProfileName)
	}
	if registrar == nil {
		return fmt.Errorf("registering %s: missing registrar", getAthleteProfileName)
	}
	if err := registrar.AddTool(newGetAthleteProfileTool(r.profileClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
		return err
	}
	if fitnessClient, ok := r.profileClient.(FitnessClient); ok {
		if err := registrar.AddTool(newGetFitnessTool(fitnessClient, r.profileClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
			return err
		}
		if err := registrar.AddTool(newGetTrainingSummaryTool(fitnessClient, r.profileClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
			return err
		}
	}
	if wellnessClient, ok := r.profileClient.(WellnessClient); ok {
		if err := registrar.AddTool(newGetWellnessDataTool(wellnessClient, r.profileClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
			return err
		}
	}
	if bestEffortsClient, ok := r.profileClient.(BestEffortsClient); ok {
		if err := registrar.AddTool(newGetBestEffortsTool(bestEffortsClient, r.version, r.debugMetadata)); err != nil {
			return err
		}
	}
	if powerCurvesClient, ok := r.profileClient.(PowerCurvesClient); ok {
		if err := registrar.AddTool(newGetPowerCurvesTool(powerCurvesClient, r.version, r.debugMetadata)); err != nil {
			return err
		}
	}
	if activityClient, ok := r.profileClient.(ActivitiesClient); ok {
		if err := registrar.AddTool(newGetActivitiesTool(activityClient, r.profileClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
			return err
		}
	}
	if eventsClient, ok := r.profileClient.(EventsClient); ok {
		if err := registrar.AddTool(newGetEventsTool(eventsClient, r.profileClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
			return err
		}
	}
	if eventByIDClient, ok := r.profileClient.(EventByIDClient); ok {
		if err := registrar.AddTool(newGetEventByIDTool(eventByIDClient, r.profileClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
			return err
		}
	}
	if eventWriterClient, ok := r.profileClient.(EventWriterClient); ok {
		if err := registrar.AddTool(newAddOrUpdateEventTool(eventWriterClient, r.profileClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
			return err
		}
	}
	if linkClient, ok := r.profileClient.(ActivityEventLinkClient); ok {
		activityClient, _ := r.profileClient.(ActivityDetailsClient)
		eventClient, _ := r.profileClient.(EventByIDClient)
		if err := registrar.AddTool(newLinkActivityToEventTool(linkClient, activityClient, eventClient, r.version, r.debugMetadata)); err != nil {
			return err
		}
	}
	if trainingPlanClient, ok := r.profileClient.(TrainingPlanClient); ok {
		if err := registrar.AddTool(newGetTrainingPlanTool(trainingPlanClient, r.profileClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
			return err
		}
	}
	if workoutLibraryClient, ok := r.profileClient.(WorkoutLibraryClient); ok {
		if err := registrar.AddTool(newGetWorkoutLibraryTool(workoutLibraryClient, r.profileClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
			return err
		}
		if err := registrar.AddTool(newGetWorkoutsInFolderTool(workoutLibraryClient, r.profileClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
			return err
		}
	}
	if customItemsClient, ok := r.profileClient.(CustomItemsClient); ok {
		if err := registrar.AddTool(newGetCustomItemsTool(customItemsClient, r.profileClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
			return err
		}
		if err := registrar.AddTool(newGetCustomItemByIDTool(customItemsClient, r.profileClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
			return err
		}
	}
	if detailsClient, ok := r.profileClient.(ActivityDetailsClient); ok {
		if err := registrar.AddTool(newGetActivityDetailsTool(detailsClient, r.profileClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
			return err
		}
	}
	var intervalsClient ActivityIntervalsClient
	if client, ok := r.profileClient.(ActivityIntervalsClient); ok {
		intervalsClient = client
		detailsClient, _ := r.profileClient.(ActivityDetailsClient)
		if err := registrar.AddTool(newGetActivityIntervalsTool(intervalsClient, detailsClient, r.version, r.debugMetadata)); err != nil {
			return err
		}
	}
	if streamsClient, ok := r.profileClient.(ActivityStreamsClient); ok {
		if err := registrar.AddTool(newGetActivityStreamsTool(streamsClient, r.version, r.debugMetadata)); err != nil {
			return err
		}
		if intervalsClient != nil {
			if err := registrar.AddTool(newGetActivitySplitsTool(streamsClient, intervalsClient, r.profileClient, r.version, r.debugMetadata)); err != nil {
				return err
			}
		}
	}
	if messagesClient, ok := r.profileClient.(ActivityMessagesClient); ok {
		detailsClient, _ := r.profileClient.(ActivityDetailsClient)
		if err := registrar.AddTool(newGetActivityMessagesTool(messagesClient, r.profileClient, detailsClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
			return err
		}
	}
	if extendedClient, ok := r.profileClient.(ExtendedMetricsClient); ok {
		if err := registrar.AddTool(newGetExtendedMetricsTool(extendedClient, r.profileClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
			return err
		}
	}
	return nil
}

// Registrar accepts tool definitions from a Registry.
type Registrar interface {
	AddTool(Tool) error
}

// Handler handles a tool call using raw JSON arguments.
type Handler func(context.Context, Request) (Result, error)

// Requirement describes the registration-time capability needed by a tool.
type Requirement string

const (
	// RequirementRead registers the tool in every mode.
	RequirementRead Requirement = "read"
	// RequirementWrite registers the tool only when write tools are allowed.
	RequirementWrite Requirement = "write"
	// RequirementDelete registers the tool only when delete tools are allowed.
	RequirementDelete Requirement = "delete"
)

// Tool describes one MCP tool without exposing SDK-specific types.
type Tool struct {
	Name         string
	Description  string
	InputSchema  any
	OutputSchema any
	Requirement  Requirement
	Handler      Handler
}

// RequiresWrite reports whether the tool needs write capability to be registered.
func (t Tool) RequiresWrite() bool {
	return t.Requirement == RequirementWrite || t.Requirement == RequirementDelete
}

// RequiresDelete reports whether the tool needs delete capability to be registered.
func (t Tool) RequiresDelete() bool {
	return t.Requirement == RequirementDelete
}

// Request carries an MCP tool call to a Handler.
type Request struct {
	Name      string
	Arguments json.RawMessage
}

// Result is returned from a Handler.
type Result struct {
	Content           []Content
	StructuredContent any
	IsError           bool
}

// Content is a user-visible MCP response content item.
type Content struct {
	Type ContentType
	Text string
}

// ContentType identifies supported response content kinds.
type ContentType string

const (
	// ContentTypeText is plain text response content.
	ContentTypeText ContentType = "text"
)

// UserError carries a short public message and an optional internal cause.
type UserError struct {
	Message string
	Err     error
}

// Error returns the short public message safe to show to an LLM.
func (e *UserError) Error() string {
	return e.Message
}

// Unwrap returns the internal cause, if any.
func (e *UserError) Unwrap() error {
	return e.Err
}

// NewUserError creates a user-facing tool error with an optional internal cause.
func NewUserError(message string, err error) *UserError {
	return &UserError{Message: message, Err: err}
}

// PublicErrorMessage reports the short public message for err, if it has one.
func PublicErrorMessage(err error) (string, bool) {
	var userErr *UserError
	if errors.As(err, &userErr) && userErr.Message != "" {
		return userErr.Message, true
	}
	return "", false
}

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
	Toolset          safety.Toolset
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
		toolset:          safety.ParseToolset(string(opts.Toolset)),
	}
}

type defaultRegistry struct {
	profileClient    ProfileClient
	version          string
	timezoneFallback string
	debugMetadata    bool
	capability       safety.Capability
	toolset          safety.Toolset
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
		return fmt.Errorf("registering tools: missing registrar")
	}
	collector := &catalogCollectingRegistrar{downstream: registrar}
	registrar = collector
	if err := registerTool(registrar, newGetAthleteProfileTool(r.profileClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
		return err
	}
	if fitnessClient, ok := r.profileClient.(FitnessClient); ok {
		if err := registerTool(registrar, newGetFitnessTool(fitnessClient, r.profileClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
			return err
		}
		if err := registerTool(registrar, newGetTrainingSummaryTool(fitnessClient, r.profileClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
			return err
		}
	}
	if wellnessClient, ok := r.profileClient.(WellnessClient); ok {
		if err := registerTool(registrar, newGetWellnessDataTool(wellnessClient, r.profileClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
			return err
		}
	}
	if wellnessWriterClient, ok := r.profileClient.(WellnessWriterClient); ok {
		if err := registerTool(registrar, newUpdateWellnessTool(wellnessWriterClient, r.profileClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
			return err
		}
	}
	if sportSettingsWriterClient, ok := r.profileClient.(SportSettingsWriterClient); ok {
		if err := registerTool(registrar, newUpdateSportSettingsTool(sportSettingsWriterClient, r.profileClient, r.version, r.timezoneFallback, r.debugMetadata, r.capability)); err != nil {
			return err
		}
	}
	if bestEffortsClient, ok := r.profileClient.(BestEffortsClient); ok {
		if err := registerTool(registrar, newGetBestEffortsTool(bestEffortsClient, r.version, r.debugMetadata)); err != nil {
			return err
		}
	}
	if powerCurvesClient, ok := r.profileClient.(PowerCurvesClient); ok {
		if err := registerTool(registrar, newGetPowerCurvesTool(powerCurvesClient, r.version, r.debugMetadata)); err != nil {
			return err
		}
	}
	if activityClient, ok := r.profileClient.(ActivitiesClient); ok {
		if err := registerTool(registrar, newGetActivitiesTool(activityClient, r.profileClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
			return err
		}
	}
	if eventsClient, ok := r.profileClient.(EventsClient); ok {
		if err := registerTool(registrar, newGetEventsTool(eventsClient, r.profileClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
			return err
		}
	}
	if eventByIDClient, ok := r.profileClient.(EventByIDClient); ok {
		if err := registerTool(registrar, newGetEventByIDTool(eventByIDClient, r.profileClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
			return err
		}
	}
	if eventWriterClient, ok := r.profileClient.(EventWriterClient); ok {
		if err := registerTool(registrar, newAddOrUpdateEventTool(eventWriterClient, r.profileClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
			return err
		}
	}
	if applyTrainingPlanClient, ok := r.profileClient.(ApplyTrainingPlanClient); ok {
		if err := registerTool(registrar, newApplyTrainingPlanTool(applyTrainingPlanClient, r.profileClient, r.version, r.timezoneFallback, r.debugMetadata, r.capability)); err != nil {
			return err
		}
	}
	if eventDeleterClient, ok := r.profileClient.(EventDeleterClient); ok {
		if err := registerTool(registrar, newDeleteEventTool(eventDeleterClient, r.profileClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
			return err
		}
	}
	if eventsByDateRangeDeleterClient, ok := r.profileClient.(EventsByDateRangeDeleterClient); ok {
		if err := registerTool(registrar, newDeleteEventsByDateRangeTool(eventsByDateRangeDeleterClient, r.profileClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
			return err
		}
	}
	if linkClient, ok := r.profileClient.(ActivityEventLinkClient); ok {
		activityClient, _ := r.profileClient.(ActivityDetailsClient)
		eventClient, _ := r.profileClient.(EventByIDClient)
		if err := registerTool(registrar, newLinkActivityToEventTool(linkClient, activityClient, eventClient, r.version, r.debugMetadata)); err != nil {
			return err
		}
	}
	if trainingPlanClient, ok := r.profileClient.(TrainingPlanClient); ok {
		if err := registerTool(registrar, newGetTrainingPlanTool(trainingPlanClient, r.profileClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
			return err
		}
	}
	if workoutLibraryClient, ok := r.profileClient.(WorkoutLibraryClient); ok {
		if err := registerTool(registrar, newGetWorkoutLibraryTool(workoutLibraryClient, r.profileClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
			return err
		}
		if err := registerTool(registrar, newGetWorkoutsInFolderTool(workoutLibraryClient, r.profileClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
			return err
		}
	}
	if workoutCreatorClient, ok := r.profileClient.(WorkoutCreatorClient); ok {
		if err := registerTool(registrar, newCreateWorkoutTool(workoutCreatorClient, r.profileClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
			return err
		}
	}
	if workoutUpdaterClient, ok := r.profileClient.(WorkoutUpdaterClient); ok {
		if err := registerTool(registrar, newUpdateWorkoutTool(workoutUpdaterClient, r.profileClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
			return err
		}
	}
	if workoutDeleterClient, ok := r.profileClient.(WorkoutDeleterClient); ok {
		if err := registerTool(registrar, newDeleteWorkoutTool(workoutDeleterClient, r.profileClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
			return err
		}
	}
	if sportSettingsDeleterClient, ok := r.profileClient.(SportSettingsDeleterClient); ok {
		if err := registerTool(registrar, newDeleteSportSettingsTool(sportSettingsDeleterClient, r.profileClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
			return err
		}
	}
	var customItemsClient CustomItemsClient
	if client, ok := r.profileClient.(CustomItemsClient); ok {
		customItemsClient = client
		if err := registerTool(registrar, newGetCustomItemsTool(customItemsClient, r.profileClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
			return err
		}
		if err := registerTool(registrar, newGetCustomItemByIDTool(customItemsClient, r.profileClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
			return err
		}
	}
	if customItemCreatorClient, ok := r.profileClient.(CustomItemCreatorClient); ok {
		if err := registerTool(registrar, newCreateCustomItemTool(customItemCreatorClient, customItemsClient, r.profileClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
			return err
		}
	}
	if customItemUpdaterClient, ok := r.profileClient.(CustomItemUpdaterClient); ok {
		if err := registerTool(registrar, newUpdateCustomItemTool(customItemUpdaterClient, customItemsClient, r.profileClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
			return err
		}
	}
	if customItemDeleterClient, ok := r.profileClient.(CustomItemDeleterClient); ok {
		if err := registerTool(registrar, newDeleteCustomItemTool(customItemDeleterClient, r.profileClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
			return err
		}
	}
	if detailsClient, ok := r.profileClient.(ActivityDetailsClient); ok {
		if err := registerTool(registrar, newGetActivityDetailsTool(detailsClient, r.profileClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
			return err
		}
	}
	if activityDeleterClient, ok := r.profileClient.(ActivityDeleterClient); ok {
		if err := registerTool(registrar, newDeleteActivityTool(activityDeleterClient, r.profileClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
			return err
		}
	}
	var intervalsClient ActivityIntervalsClient
	if client, ok := r.profileClient.(ActivityIntervalsClient); ok {
		intervalsClient = client
		detailsClient, _ := r.profileClient.(ActivityDetailsClient)
		if err := registerTool(registrar, newGetActivityIntervalsTool(intervalsClient, detailsClient, r.version, r.debugMetadata)); err != nil {
			return err
		}
	}
	if streamsClient, ok := r.profileClient.(ActivityStreamsClient); ok {
		if err := registerTool(registrar, newGetActivityStreamsTool(streamsClient, r.version, r.debugMetadata)); err != nil {
			return err
		}
		if intervalsClient != nil {
			if err := registerTool(registrar, newGetActivitySplitsTool(streamsClient, intervalsClient, r.profileClient, r.version, r.debugMetadata)); err != nil {
				return err
			}
		}
	}
	if messagesClient, ok := r.profileClient.(ActivityMessagesClient); ok {
		detailsClient, _ := r.profileClient.(ActivityDetailsClient)
		if err := registerTool(registrar, newGetActivityMessagesTool(messagesClient, r.profileClient, detailsClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
			return err
		}
	}
	if messageWriterClient, ok := r.profileClient.(ActivityMessageWriterClient); ok {
		if err := registerTool(registrar, newAddActivityMessageTool(messageWriterClient, r.profileClient, r.version, r.debugMetadata)); err != nil {
			return err
		}
	}
	if extendedClient, ok := r.profileClient.(ExtendedMetricsClient); ok {
		if err := registerTool(registrar, newGetExtendedMetricsTool(extendedClient, r.profileClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
			return err
		}
	}
	if gearDeleterClient, ok := r.profileClient.(GearDeleterClient); ok {
		if err := registerTool(registrar, newDeleteGearTool(gearDeleterClient, r.profileClient, r.version, r.timezoneFallback, r.debugMetadata)); err != nil {
			return err
		}
	}
	if err := registerTool(collector.downstream, newListAdvancedCapabilitiesTool(collector.tools, r.toolset)); err != nil {
		return err
	}
	return nil
}

func registerTool(registrar Registrar, tool Tool) error {
	if err := registrar.AddTool(tool); err != nil {
		return fmt.Errorf("registering %s: %w", tool.Name, err)
	}
	return nil
}

type catalogCollectingRegistrar struct {
	downstream Registrar
	tools      []Tool
}

func (r *catalogCollectingRegistrar) AddTool(tool Tool) error {
	r.tools = append(r.tools, tool)
	return r.downstream.AddTool(tool)
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
	Toolset      safety.Toolset
	Handler      Handler
}

func coreTool(tool Tool) Tool {
	tool.Toolset = safety.ToolsetCore
	return tool
}

func fullTool(tool Tool) Tool {
	tool.Toolset = safety.ToolsetFull
	return tool
}

// EffectiveToolset reports the registration tier declared by the tool. Empty
// values default to full so new/unmarked tools do not silently expand core.
func (t Tool) EffectiveToolset() safety.Toolset {
	switch t.Toolset {
	case safety.ToolsetCore, safety.ToolsetFull:
		return t.Toolset
	default:
		return safety.ToolsetFull
	}
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

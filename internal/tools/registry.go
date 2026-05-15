package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/response"
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
func NewRegistry(client *intervals.Client, version string, timezoneFallback ...string) Registry {
	return NewRegistryWithOptions(client, RegistryOptions{
		Version:          version,
		TimezoneFallback: firstNonEmpty(timezoneFallback...),
	})
}

// NewRegistryWithOptions creates the default registry with explicit response-shaping options.
func NewRegistryWithOptions(client *intervals.Client, opts RegistryOptions) Registry {
	capability := capabilityOrSafe(opts.Capability)
	toolset := safety.ParseToolset(opts.Toolset.String())
	return &defaultRegistry{
		client:           client,
		version:          normalizeVersion(opts.Version),
		timezoneFallback: normalizeTimezoneFallback(opts.TimezoneFallback),
		debugMetadata:    opts.DebugMetadata,
		capability:       capability,
		deleteMode:       safety.ParseMode(capability.Mode()),
		toolset:          toolset,
	}
}

type defaultRegistry struct {
	client           *intervals.Client
	version          string
	timezoneFallback string
	debugMetadata    bool
	capability       safety.Capability
	deleteMode       safety.Mode
	toolset          safety.Toolset
}

type responseShaping struct {
	deleteMode safety.Mode
	toolset    safety.Toolset
}

func responseShapingOrDefault(shaping []responseShaping) responseShaping {
	if len(shaping) > 0 {
		return responseShaping{deleteMode: safety.ParseMode(shaping[0].deleteMode.String()), toolset: safety.ParseToolset(shaping[0].toolset.String())}
	}
	return responseShaping{deleteMode: safety.ModeSafe, toolset: safety.ToolsetCore}
}

func (s responseShaping) options(includeFull bool, rowCollections []string, version string, debugMetadata bool, queryType string, unitSystem response.UnitSystem) response.Options {
	return response.Options{IncludeFull: includeFull, RowCollections: rowCollections, ServerVersion: version, DebugMetadata: debugMetadata, QueryType: queryType, UnitSystem: unitSystem, DeleteMode: s.deleteMode, Toolset: s.toolset}
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
	if r.client == nil {
		return errors.New("registering tools: missing intervals client")
	}
	if registrar == nil {
		return errors.New("registering tools: missing registrar")
	}
	collector := &catalogCollectingRegistrar{downstream: registrar}
	registrar = collector
	shaping := responseShaping{deleteMode: r.deleteMode, toolset: r.toolset}
	add := func(tool Tool) error {
		if err := registrar.AddTool(tool); err != nil {
			return fmt.Errorf("registering %s: %w", tool.Name, err)
		}
		return nil
	}
	if err := add(newGetAthleteProfileTool(r.client, r.version, r.timezoneFallback, r.debugMetadata, shaping)); err != nil {
		return err
	}
	if err := add(newGetFitnessTool(r.client, r.client, r.version, r.timezoneFallback, r.debugMetadata, shaping)); err != nil {
		return err
	}
	if err := add(newGetTrainingSummaryTool(r.client, r.client, r.version, r.timezoneFallback, r.debugMetadata, shaping)); err != nil {
		return err
	}
	if err := add(newGetWellnessDataTool(r.client, r.client, r.version, r.timezoneFallback, r.debugMetadata, shaping)); err != nil {
		return err
	}
	if err := add(newUpdateWellnessTool(r.client, r.client, r.version, r.timezoneFallback, r.debugMetadata, shaping)); err != nil {
		return err
	}
	if err := add(newUpdateSportSettingsTool(r.client, r.client, r.version, r.timezoneFallback, r.debugMetadata, r.capability)); err != nil {
		return err
	}
	if err := add(newGetBestEffortsTool(r.client, r.version, r.debugMetadata, shaping)); err != nil {
		return err
	}
	if err := add(newGetPowerCurvesTool(r.client, r.version, r.debugMetadata, shaping)); err != nil {
		return err
	}
	if err := add(newGetActivitiesTool(r.client, r.client, r.version, r.timezoneFallback, r.debugMetadata, shaping)); err != nil {
		return err
	}
	if err := add(newGetEventsTool(r.client, r.client, r.version, r.timezoneFallback, r.debugMetadata, shaping)); err != nil {
		return err
	}
	if err := add(newGetEventByIDTool(r.client, r.client, r.version, r.timezoneFallback, r.debugMetadata, shaping)); err != nil {
		return err
	}
	if err := add(newAddOrUpdateEventTool(r.client, r.client, r.version, r.timezoneFallback, r.debugMetadata, shaping)); err != nil {
		return err
	}
	if err := add(newApplyTrainingPlanTool(r.client, r.client, r.version, r.timezoneFallback, r.debugMetadata, r.capability, shaping)); err != nil {
		return err
	}
	if err := add(newDeleteEventTool(r.client, r.client, r.version, r.timezoneFallback, r.debugMetadata, shaping)); err != nil {
		return err
	}
	if err := add(newDeleteEventsByDateRangeTool(r.client, r.client, r.version, r.timezoneFallback, r.debugMetadata, shaping)); err != nil {
		return err
	}
	if err := add(newLinkActivityToEventTool(r.client, r.client, r.client, r.version, r.debugMetadata, shaping)); err != nil {
		return err
	}
	if err := add(newGetTrainingPlanTool(r.client, r.client, r.version, r.timezoneFallback, r.debugMetadata, shaping)); err != nil {
		return err
	}
	if err := add(newGetWorkoutLibraryTool(r.client, r.client, r.version, r.timezoneFallback, r.debugMetadata, shaping)); err != nil {
		return err
	}
	if err := add(newGetWorkoutsInFolderTool(r.client, r.client, r.version, r.timezoneFallback, r.debugMetadata, shaping)); err != nil {
		return err
	}
	if err := add(newCreateWorkoutTool(r.client, r.client, r.version, r.timezoneFallback, r.debugMetadata, shaping)); err != nil {
		return err
	}
	if err := add(newUpdateWorkoutTool(r.client, r.client, r.version, r.timezoneFallback, r.debugMetadata, shaping)); err != nil {
		return err
	}
	if err := add(newDeleteWorkoutTool(r.client, r.client, r.version, r.timezoneFallback, r.debugMetadata, shaping)); err != nil {
		return err
	}
	if err := add(newDeleteSportSettingsTool(r.client, r.client, r.version, r.timezoneFallback, r.debugMetadata, shaping)); err != nil {
		return err
	}
	if err := add(newGetCustomItemsTool(r.client, r.client, r.version, r.timezoneFallback, r.debugMetadata, shaping)); err != nil {
		return err
	}
	if err := add(newGetCustomItemByIDTool(r.client, r.client, r.version, r.timezoneFallback, r.debugMetadata, shaping)); err != nil {
		return err
	}
	if err := add(newCreateCustomItemTool(r.client, r.client, r.client, r.version, r.timezoneFallback, r.debugMetadata, shaping)); err != nil {
		return err
	}
	if err := add(newUpdateCustomItemTool(r.client, r.client, r.client, r.version, r.timezoneFallback, r.debugMetadata, shaping)); err != nil {
		return err
	}
	if err := add(newDeleteCustomItemTool(r.client, r.client, r.version, r.timezoneFallback, r.debugMetadata, shaping)); err != nil {
		return err
	}
	if err := add(newGetActivityDetailsTool(r.client, r.client, r.version, r.timezoneFallback, r.debugMetadata, shaping)); err != nil {
		return err
	}
	if err := add(newDeleteActivityTool(r.client, r.client, r.version, r.timezoneFallback, r.debugMetadata, shaping)); err != nil {
		return err
	}
	if err := add(newGetActivityIntervalsTool(r.client, r.client, r.version, r.debugMetadata, shaping)); err != nil {
		return err
	}
	if err := add(newGetActivityStreamsTool(r.client, r.version, r.debugMetadata, shaping)); err != nil {
		return err
	}
	if err := add(newGetActivitySplitsTool(r.client, r.client, r.client, r.version, r.debugMetadata, shaping)); err != nil {
		return err
	}
	if err := add(newGetActivityMessagesTool(r.client, r.client, r.client, r.version, r.timezoneFallback, r.debugMetadata, shaping)); err != nil {
		return err
	}
	if err := add(newAddActivityMessageTool(r.client, r.client, r.version, r.debugMetadata, shaping)); err != nil {
		return err
	}
	if err := add(newGetExtendedMetricsTool(r.client, r.client, r.version, r.timezoneFallback, r.debugMetadata, shaping)); err != nil {
		return err
	}
	if err := add(newDeleteGearTool(r.client, r.client, r.version, r.timezoneFallback, r.debugMetadata, shaping)); err != nil {
		return err
	}
	advancedTool := newListAdvancedCapabilitiesTool(collector.tools, r.toolset)
	if err := collector.downstream.AddTool(advancedTool); err != nil {
		return fmt.Errorf("registering %s: %w", advancedTool.Name, err)
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

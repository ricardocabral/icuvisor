package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ricardocabral/icuvisor/internal/coach"
	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/response"
	"github.com/ricardocabral/icuvisor/internal/safety"
	"github.com/ricardocabral/icuvisor/internal/toolcatalog"
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
	CoachModeEnabled bool
	CoachConfig      coach.Config
	CatalogFilter    func(Tool) bool
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
		coachModeEnabled: opts.CoachModeEnabled,
		coachConfig:      opts.CoachConfig,
		catalogFilter:    opts.CatalogFilter,
		gearCache:        newGearListCache(),
		customFieldCache: newCustomFieldCache(),
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
	coachModeEnabled bool
	coachConfig      coach.Config
	catalogFilter    func(Tool) bool
	gearCache        *gearListCache
	customFieldCache *customFieldCache
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
		if !toolcatalog.IsKnownTool(tool.Name) {
			return fmt.Errorf("registering %s: not present in shared tool catalog", tool.Name)
		}
		if err := registrar.AddTool(tool); err != nil {
			return fmt.Errorf("registering %s: %w", tool.Name, err)
		}
		return nil
	}
	for _, tool := range registryBaseTools(r.client, registryToolOptions{
		version:          r.version,
		timezoneFallback: r.timezoneFallback,
		debugMetadata:    r.debugMetadata,
		capability:       r.capability,
		shaping:          shaping,
		gearCache:        r.gearCache,
		customFieldCache: r.customFieldCache,
		coachModeEnabled: r.coachModeEnabled,
		coachConfig:      r.coachConfig,
	}) {
		if err := add(tool); err != nil {
			return err
		}
	}
	advancedTool := newListAdvancedCapabilitiesTool(filteredCatalog(collector.tools, r.catalogFilter), r.toolset, shaping)
	if !toolcatalog.IsKnownTool(advancedTool.Name) {
		return fmt.Errorf("registering %s: not present in shared tool catalog", advancedTool.Name)
	}
	if err := collector.downstream.AddTool(advancedTool); err != nil {
		return fmt.Errorf("registering %s: %w", advancedTool.Name, err)
	}
	return nil
}

func filteredCatalog(catalog []Tool, filter func(Tool) bool) []Tool {
	if filter == nil {
		return catalog
	}
	out := make([]Tool, 0, len(catalog))
	for _, tool := range catalog {
		if filter(tool) {
			out = append(out, tool)
		}
	}
	return out
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
	requirementDefault Requirement = ""

	// RequirementRead registers the tool in every mode.
	RequirementRead Requirement = "read"
	// RequirementWrite registers the tool only when write tools are allowed.
	RequirementWrite Requirement = "write"
	// RequirementDelete registers the tool only when delete tools are allowed.
	RequirementDelete Requirement = "delete"
)

func (r Requirement) effective() Requirement {
	if r == requirementDefault {
		return RequirementRead
	}
	return r
}

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
	requirement := t.Requirement.effective()
	return requirement == RequirementWrite || requirement == RequirementDelete
}

// RequiresDelete reports whether the tool needs delete capability to be registered.
func (t Tool) RequiresDelete() bool {
	return t.Requirement.effective() == RequirementDelete
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

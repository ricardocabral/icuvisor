// Package icuvisor exposes the stable public core library facade for host reuse.
package icuvisor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"slices"
	"strings"
	"time"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/intervals"
	internalmcp "github.com/ricardocabral/icuvisor/internal/mcp"
	"github.com/ricardocabral/icuvisor/internal/prompts"
	"github.com/ricardocabral/icuvisor/internal/resources"
	"github.com/ricardocabral/icuvisor/internal/safety"
	"github.com/ricardocabral/icuvisor/internal/toolcatalog"
	"github.com/ricardocabral/icuvisor/internal/tools"
)

// StreamableHTTPPath is the default MCP Streamable HTTP endpoint path.
const StreamableHTTPPath = internalmcp.StreamableHTTPPath

// DeleteMode controls write/delete tool registration.
type DeleteMode string

const (
	// DeleteModeSafe allows write tools but skips delete tools.
	DeleteModeSafe DeleteMode = "safe"
	// DeleteModeFull allows write and delete tools.
	DeleteModeFull DeleteMode = "full"
	// DeleteModeNone skips write and delete tools.
	DeleteModeNone DeleteMode = "none"
)

// Toolset controls the registered tool catalog tier.
type Toolset string

const (
	// ToolsetCompact exposes the smallest model-compatible tool catalog.
	ToolsetCompact Toolset = "compact"
	// ToolsetCore exposes the curated daily-use tool catalog.
	ToolsetCore Toolset = "core"
	// ToolsetFull exposes the full tool catalog.
	ToolsetFull Toolset = "full"
)

// Requirement describes the registration-time capability needed by a tool.
type Requirement string

const (
	// RequirementRead registers the tool in every mode.
	RequirementRead Requirement = "read"
	// RequirementWrite registers the tool only when writes are allowed.
	RequirementWrite Requirement = "write"
	// RequirementDelete registers the tool only when deletes are allowed.
	RequirementDelete Requirement = "delete"
)

// Config contains non-hosted runtime settings consumed by core registries and servers.
type Config struct {
	APIKey          string
	AthleteID       string
	Timezone        string
	APIBaseURL      string
	HTTPTimeout     time.Duration
	DeleteMode      DeleteMode
	Toolset         Toolset
	DebugMetadata   bool
	HTTPBindAddress string
}

// RetryConfig controls retry behavior for idempotent Intervals API requests.
type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	Jitter      float64
}

// Client is an opaque Intervals API client usable by public core registries.
type Client struct {
	inner *intervals.Client
}

// APIKeyClientOptions configures an API-key Intervals client.
type APIKeyClientOptions struct {
	Config     Config
	Version    string
	HTTPClient *http.Client
	Retry      RetryConfig
}

// BearerClientOptions configures an OAuth Bearer Intervals client.
type BearerClientOptions struct {
	AccessToken string
	AthleteID   string
	APIBaseURL  string
	Version     string
	HTTPClient  *http.Client
	HTTPTimeout time.Duration
	Retry       RetryConfig
}

// NewAPIKeyClient constructs an Intervals client using API-key Basic Auth.
func NewAPIKeyClient(opts APIKeyClientOptions) (*Client, error) {
	client, err := intervals.NewClient(intervals.Options{Config: opts.Config.toInternal(), Version: opts.Version, HTTPClient: opts.HTTPClient, Retry: opts.Retry.toInternal()})
	if err != nil {
		return nil, err
	}
	return &Client{inner: client}, nil
}

// NewBearerClient constructs an Intervals client using OAuth Bearer auth.
func NewBearerClient(opts BearerClientOptions) (*Client, error) {
	client, err := intervals.NewOAuthBearerClient(intervals.OAuthBearerOptions{AccessToken: opts.AccessToken, AthleteID: opts.AthleteID, APIBaseURL: opts.APIBaseURL, Version: opts.Version, HTTPClient: opts.HTTPClient, HTTPTimeout: opts.HTTPTimeout, Retry: opts.Retry.toInternal()})
	if err != nil {
		return nil, err
	}
	return &Client{inner: client}, nil
}

// Registry is an opaque MCP tool registry.
type Registry struct {
	inner tools.Registry
	core  *coreRegistrySpec
}

type coreRegistrySpec struct {
	client *Client
	opts   RegistryOptions
}

// RegistryOptions configures the default core tool registry.
type RegistryOptions struct {
	Config           Config
	Version          string
	TimezoneFallback string
	DebugMetadata    bool
	DeleteMode       DeleteMode
	Toolset          Toolset
	ToolFilter       func(ToolInfo) bool
	ExtraTools       []Tool
	CatalogHash      string
}

// NewCoreRegistry creates the default core tool registry with optional host policy extensions.
func NewCoreRegistry(client *Client, opts RegistryOptions) Registry {
	base := buildCoreRegistry(client, opts, Config{})
	return Registry{inner: base, core: &coreRegistrySpec{client: client, opts: opts}}
}

// NewResourceRegistry creates the default MCP resource registry.
func NewResourceRegistry(client *Client, opts ResourceRegistryOptions) ResourceRegistry {
	inner := buildResourceRegistry(client, opts, Config{})
	return ResourceRegistry{inner: inner, core: &resourceRegistrySpec{client: client, opts: opts}}
}

// ResourceRegistry is an opaque MCP resource registry.
type ResourceRegistry struct {
	inner resources.Registry
	core  *resourceRegistrySpec
}

type resourceRegistrySpec struct {
	client *Client
	opts   ResourceRegistryOptions
}

// ResourceRegistryOptions configures default MCP resources.
type ResourceRegistryOptions struct {
	Config                Config
	Version               string
	TimezoneFallback      string
	DebugMetadata         bool
	DeleteMode            DeleteMode
	Toolset               Toolset
	CatalogHash           string
	AthleteProfileTTL     time.Duration
	DisableAthleteProfile bool
	Now                   func() time.Time
}

// NewPromptRegistry creates the default MCP prompt registry.
func NewPromptRegistry() PromptRegistry {
	return PromptRegistry{inner: prompts.NewRegistry()}
}

// PromptRegistry is an opaque MCP prompt registry.
type PromptRegistry struct {
	inner prompts.Registry
}

// Server is an opaque MCP server.
type Server struct {
	inner *internalmcp.Server
}

// Invoker invokes registered core tools without an MCP transport.
type Invoker struct {
	tools map[string]tools.Tool
}

// InvokerOptions configures direct core-tool invocation for local CLI and agent views.
type InvokerOptions struct {
	Config     Config
	Registry   Registry
	DeleteMode DeleteMode
	Toolset    Toolset
}

// NewInvoker constructs a direct invoker from the same registry and safety policy used by MCP.
func NewInvoker(ctx context.Context, opts InvokerOptions) (*Invoker, error) {
	cfg, err := opts.Config.toInternalValidated()
	if err != nil {
		return nil, err
	}
	deleteMode, toolset, err := effectivePolicy(opts.DeleteMode, opts.Toolset, opts.Config)
	if err != nil {
		return nil, err
	}
	registry := registryForConfig(opts.Registry, opts.Config)
	if opts.Registry.core != nil {
		catalogHash, err := internalmcp.ComputeToolCatalogHash(ctx, internalmcp.CatalogHashOptions{
			Config:     cfg,
			Registry:   registry,
			Capability: safety.NewCapability(deleteMode.toInternal()),
			Toolset:    toolset.toInternal(),
		})
		if err != nil {
			return nil, err
		}
		registry = registryForConfigWithCatalogHash(opts.Registry, opts.Config, catalogHash)
	}
	registrar := directRegistrar{
		tools:      make(map[string]tools.Tool),
		capability: safety.NewCapability(deleteMode.toInternal()),
		toolset:    toolset.toInternal(),
	}
	if err := registry.Register(ctx, &registrar); err != nil {
		return nil, err
	}
	return &Invoker{tools: registrar.tools}, nil
}

// Tools returns the tools available to this direct invocation view.
func (i *Invoker) Tools() []ToolInfo {
	if i == nil {
		return nil
	}
	out := make([]ToolInfo, 0, len(i.tools))
	for _, tool := range i.tools {
		out = append(out, toolInfoFromInternal(tool))
	}
	slices.SortFunc(out, func(a, b ToolInfo) int { return strings.Compare(a.Name, b.Name) })
	return out
}

// Invoke calls one registered tool without an MCP transport.
func (i *Invoker) Invoke(ctx context.Context, name string, arguments json.RawMessage) (ToolResult, error) {
	if i == nil {
		return ToolResult{}, errors.New("invoking tool: nil invoker")
	}
	tool, ok := i.tools[name]
	if !ok {
		return ToolResult{}, fmt.Errorf("unknown or unavailable tool %q", name)
	}
	result, err := tool.Handler(ctx, tools.Request{Name: name, Arguments: arguments})
	if err != nil {
		return ToolResult{}, err
	}
	return toolResultFromInternal(result), nil
}

// directRegistrar applies the same registration-time safety and toolset policy as MCP.
type directRegistrar struct {
	tools      map[string]tools.Tool
	capability safety.Capability
	toolset    safety.Toolset
}

func (r *directRegistrar) AddTool(tool tools.Tool) error {
	if _, exists := r.tools[tool.Name]; exists {
		return fmt.Errorf("duplicate tool name %q", tool.Name)
	}
	if !directCapabilityAllows(r.capability, tool) || !directToolsetAllows(r.toolset, tool) {
		return nil
	}
	r.tools[tool.Name] = tool
	return nil
}

func directCapabilityAllows(capability safety.Capability, tool tools.Tool) bool {
	if tool.RequiresDelete() && !capability.CanDelete() {
		return false
	}
	return !tool.RequiresWrite() || capability.CanWrite()
}

func directToolsetAllows(active safety.Toolset, tool tools.Tool) bool {
	switch safety.ParseToolset(active.String()) {
	case safety.ToolsetFull:
		return true
	case safety.ToolsetCompact:
		return toolcatalog.IsCompactTool(tool.Name)
	default:
		return tool.EffectiveToolset() == safety.ToolsetCore
	}
}

// ServerOptions configures core MCP server construction.
type ServerOptions struct {
	Config                     Config
	Version                    string
	Logger                     *slog.Logger
	Registry                   Registry
	ResourceRegistry           ResourceRegistry
	PromptRegistry             PromptRegistry
	DeleteMode                 DeleteMode
	Toolset                    Toolset
	Transport                  sdkmcp.Transport
	RecentToolCallRecorder     RecentToolCallRecorder
	SkipRuntimeCatalogMetadata bool
}

// RecentToolCallRecorder records tool-call names for diagnostics without exposing arguments or payloads.
type RecentToolCallRecorder interface {
	RecordToolCall(context.Context, string, time.Time) error
}

// NewServer constructs an MCP server from public core facade inputs.
func NewServer(ctx context.Context, opts ServerOptions) (*Server, error) {
	cfg, err := opts.Config.toInternalValidated()
	if err != nil {
		return nil, err
	}
	deleteMode, toolset, err := effectivePolicy(opts.DeleteMode, opts.Toolset, opts.Config)
	if err != nil {
		return nil, err
	}
	registry := registryForConfig(opts.Registry, opts.Config)
	resourceRegistry := resourceRegistryForConfig(opts.ResourceRegistry, opts.Config)
	if opts.SkipRuntimeCatalogMetadata {
		catalogHash, err := internalmcp.ComputeToolCatalogHash(ctx, internalmcp.CatalogHashOptions{Config: cfg, Registry: registry, Capability: safety.NewCapability(deleteMode.toInternal()), Toolset: toolset.toInternal(), Logger: opts.Logger})
		if err != nil {
			return nil, err
		}
		registry = registryForConfigWithCatalogHash(opts.Registry, opts.Config, catalogHash)
		resourceRegistry = resourceRegistryForConfigWithCatalogHash(opts.ResourceRegistry, opts.Config, catalogHash)
	}
	server, err := internalmcp.NewServer(ctx, internalmcp.Options{Config: cfg, Version: opts.Version, Logger: opts.Logger, Registry: registry, ResourceRegistry: resourceRegistry, PromptRegistry: opts.PromptRegistry.inner, Capability: safety.NewCapability(deleteMode.toInternal()), Toolset: toolset.toInternal(), Transport: opts.Transport, RecentToolCallRecorder: opts.RecentToolCallRecorder, SkipRuntimeCatalogMetadata: opts.SkipRuntimeCatalogMetadata})
	if err != nil {
		return nil, err
	}
	return &Server{inner: server}, nil
}

// CatalogHash reports the deterministic hash of the exposed tool catalog.
func (s *Server) CatalogHash() string {
	if s == nil || s.inner == nil {
		return ""
	}
	return s.inner.CatalogHash()
}

// Run serves one MCP stdio session until the client disconnects or ctx is cancelled.
func (s *Server) Run(ctx context.Context) error {
	return s.inner.Run(ctx)
}

// ServeStreamableHTTP serves Streamable HTTP on listener until ctx is cancelled.
func (s *Server) ServeStreamableHTTP(ctx context.Context, listener net.Listener) error {
	return s.inner.ServeStreamableHTTP(ctx, listener)
}

// RunStreamableHTTP serves Streamable HTTP on address until ctx is cancelled.
func (s *Server) RunStreamableHTTP(ctx context.Context, address string) error {
	return s.inner.RunStreamableHTTP(ctx, address)
}

// CatalogOptions configures catalog collection/hash operations.
type CatalogOptions struct {
	Config   Config
	Registry Registry
	Mode     DeleteMode
	Toolset  Toolset
	Logger   *slog.Logger
}

// CollectToolCatalog returns exposed MCP tool definitions without starting a transport.
func CollectToolCatalog(ctx context.Context, opts CatalogOptions) ([]ToolInfo, error) {
	cfg, err := opts.Config.toInternalValidated()
	if err != nil {
		return nil, err
	}
	deleteMode, toolset, err := effectivePolicy(opts.Mode, opts.Toolset, opts.Config)
	if err != nil {
		return nil, err
	}
	catalog, err := internalmcp.CollectToolCatalog(ctx, internalmcp.CatalogHashOptions{Config: cfg, Registry: registryForConfig(opts.Registry, opts.Config), Capability: safety.NewCapability(deleteMode.toInternal()), Toolset: toolset.toInternal(), Logger: opts.Logger})
	if err != nil {
		return nil, err
	}
	out := make([]ToolInfo, 0, len(catalog))
	for _, tool := range catalog {
		out = append(out, toolInfoFromInternal(tool))
	}
	return out, nil
}

// ComputeToolCatalogHash returns the deterministic hash of the exposed tool catalog.
func ComputeToolCatalogHash(ctx context.Context, opts CatalogOptions) (string, error) {
	cfg, err := opts.Config.toInternalValidated()
	if err != nil {
		return "", err
	}
	deleteMode, toolset, err := effectivePolicy(opts.Mode, opts.Toolset, opts.Config)
	if err != nil {
		return "", err
	}
	return internalmcp.ComputeToolCatalogHash(ctx, internalmcp.CatalogHashOptions{Config: cfg, Registry: registryForConfig(opts.Registry, opts.Config), Capability: safety.NewCapability(deleteMode.toInternal()), Toolset: toolset.toInternal(), Logger: opts.Logger})
}

// ToolInfo is public non-secret metadata about a registered tool.
type ToolInfo struct {
	Name         string      `json:"name"`
	Description  string      `json:"description"`
	InputSchema  any         `json:"input_schema"`
	OutputSchema any         `json:"output_schema"`
	Requirement  Requirement `json:"requirement"`
	Toolset      Toolset     `json:"toolset"`
}

// Tool is a public custom tool definition for host diagnostics and policy extensions.
type Tool struct {
	Name         string
	Description  string
	InputSchema  any
	OutputSchema any
	Requirement  Requirement
	Toolset      Toolset
	Handler      Handler
}

// Handler handles a public tool call using raw JSON arguments.
type Handler func(context.Context, ToolRequest) (ToolResult, error)

// ToolRequest carries an MCP tool call to a Handler.
type ToolRequest struct {
	Name      string
	Arguments json.RawMessage
}

// ToolResult is returned from a Handler.
type ToolResult struct {
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

// TextResult returns a text MCP result with the same shaped structured content.
func TextResult(shaped any) ToolResult {
	text, _ := json.Marshal(shaped)
	return ToolResult{Content: []Content{{Type: ContentTypeText, Text: string(text)}}, StructuredContent: shaped}
}

// NewUserError creates a user-facing tool error with an optional internal cause.
func NewUserError(message string, err error) error {
	return tools.NewUserError(message, err)
}

// StreamableHTTPHandlerOptions configures a reusable Streamable HTTP handler.
type StreamableHTTPHandlerOptions struct {
	Logger              *slog.Logger
	Stateless           bool
	JSONResponse        bool
	FactoryErrorMessage string
}

// StreamableHTTPServerFactory builds or resolves an MCP server for one HTTP request.
type StreamableHTTPServerFactory func(*http.Request) (*Server, error)

// NewStreamableHTTPHandler adapts request-scoped Server construction to Streamable HTTP.
func NewStreamableHTTPHandler(factory StreamableHTTPServerFactory, opts StreamableHTTPHandlerOptions) http.Handler {
	return internalmcp.NewStreamableHTTPHandler(func(req *http.Request) (*internalmcp.Server, error) {
		server, err := factory(req)
		if err != nil || server == nil {
			return nil, err
		}
		return server.inner, nil
	}, internalmcp.StreamableHTTPHandlerOptions{Logger: opts.Logger, Stateless: opts.Stateless, JSONResponse: opts.JSONResponse, FactoryErrorMessage: opts.FactoryErrorMessage})
}

func internalTools(publicTools []Tool) []tools.Tool {
	out := make([]tools.Tool, 0, len(publicTools))
	for _, tool := range publicTools {
		out = append(out, internalTool(tool))
	}
	return out
}

func internalTool(tool Tool) tools.Tool {
	return tools.Tool{Name: tool.Name, Description: tool.Description, InputSchema: tool.InputSchema, OutputSchema: tool.OutputSchema, Requirement: tools.Requirement(tool.Requirement), Toolset: safety.Toolset(tool.Toolset), Handler: internalHandler(tool.Handler)}
}

func internalHandler(handler Handler) tools.Handler {
	if handler == nil {
		return nil
	}
	return func(ctx context.Context, req tools.Request) (tools.Result, error) {
		result, err := handler(ctx, ToolRequest{Name: req.Name, Arguments: req.Arguments})
		if err != nil {
			return tools.Result{}, err
		}
		return internalResult(result), nil
	}
}

func internalResult(result ToolResult) tools.Result {
	content := make([]tools.Content, 0, len(result.Content))
	for _, item := range result.Content {
		content = append(content, tools.Content{Type: tools.ContentType(item.Type), Text: item.Text})
	}
	return tools.Result{Content: content, StructuredContent: result.StructuredContent, IsError: result.IsError}
}

func toolResultFromInternal(result tools.Result) ToolResult {
	content := make([]Content, 0, len(result.Content))
	for _, item := range result.Content {
		content = append(content, Content{Type: ContentType(item.Type), Text: item.Text})
	}
	return ToolResult{Content: content, StructuredContent: result.StructuredContent, IsError: result.IsError}
}

func internalToolFilter(filter func(ToolInfo) bool) func(tools.Tool) bool {
	if filter == nil {
		return nil
	}
	return func(tool tools.Tool) bool {
		return filter(toolInfoFromInternal(tool))
	}
}

func toolInfoFromInternal(tool tools.Tool) ToolInfo {
	return ToolInfo{Name: tool.Name, Description: tool.Description, InputSchema: tool.InputSchema, OutputSchema: tool.OutputSchema, Requirement: requirementFromInternal(tool), Toolset: Toolset(tool.EffectiveToolset().String())}
}

func requirementFromInternal(tool tools.Tool) Requirement {
	if tool.RequiresDelete() {
		return RequirementDelete
	}
	if tool.RequiresWrite() {
		return RequirementWrite
	}
	return RequirementRead
}

type errorToolRegistry struct {
	err error
}

func (r errorToolRegistry) Register(context.Context, tools.Registrar) error {
	return r.err
}

type errorResourceRegistry struct {
	err error
}

func (r errorResourceRegistry) Register(context.Context, resources.Registrar) error {
	return r.err
}

func buildCoreRegistry(client *Client, opts RegistryOptions, boundary Config) tools.Registry {
	var innerClient *intervals.Client
	if client != nil {
		innerClient = client.inner
	}
	filter := internalToolFilter(opts.ToolFilter)
	deleteMode, toolset, err := effectivePolicy(opts.DeleteMode, opts.Toolset, opts.Config, boundary)
	if err != nil {
		return errorToolRegistry{err: err}
	}
	return tools.NewRegistryWithOptions(innerClient, tools.RegistryOptions{Version: opts.Version, TimezoneFallback: opts.TimezoneFallback, DebugMetadata: opts.DebugMetadata, Capability: safety.NewCapability(deleteMode.toInternal()), Toolset: toolset.toInternal(), CatalogFilter: filter, CatalogHash: opts.CatalogHash, ExtraTools: internalTools(opts.ExtraTools)})
}

func registryForConfig(reg Registry, cfg Config) tools.Registry {
	return registryForConfigWithCatalogHash(reg, cfg, "")
}

func registryForConfigWithCatalogHash(reg Registry, cfg Config, catalogHash string) tools.Registry {
	if reg.core != nil {
		opts := reg.core.opts
		if opts.CatalogHash == "" {
			opts.CatalogHash = catalogHash
		}
		return buildCoreRegistry(reg.core.client, opts, cfg)
	}
	return reg.inner
}

func buildResourceRegistry(client *Client, opts ResourceRegistryOptions, boundary Config) resources.Registry {
	var profileClient *intervals.Client
	if client != nil {
		profileClient = client.inner
	}
	deleteMode, toolset, err := effectivePolicy(opts.DeleteMode, opts.Toolset, opts.Config, boundary)
	if err != nil {
		return errorResourceRegistry{err: err}
	}
	return resources.NewRegistryWithOptions(profileClient, resources.ResourceOptions{Version: opts.Version, TimezoneFallback: opts.TimezoneFallback, DebugMetadata: opts.DebugMetadata, DeleteMode: deleteMode.toInternal(), Toolset: toolset.toInternal(), CatalogHash: opts.CatalogHash, AthleteProfileTTL: opts.AthleteProfileTTL, DisableAthleteProfile: opts.DisableAthleteProfile, Now: opts.Now})
}

func resourceRegistryForConfig(reg ResourceRegistry, cfg Config) resources.Registry {
	return resourceRegistryForConfigWithCatalogHash(reg, cfg, "")
}

func resourceRegistryForConfigWithCatalogHash(reg ResourceRegistry, cfg Config, catalogHash string) resources.Registry {
	if reg.core != nil {
		opts := reg.core.opts
		if opts.CatalogHash == "" {
			opts.CatalogHash = catalogHash
		}
		return buildResourceRegistry(reg.core.client, opts, cfg)
	}
	return reg.inner
}

func effectivePolicy(optionMode DeleteMode, optionToolset Toolset, configs ...Config) (DeleteMode, Toolset, error) {
	if err := validateDeleteMode(optionMode); err != nil {
		return "", "", err
	}
	if err := validateToolset(optionToolset); err != nil {
		return "", "", err
	}
	modes := make([]DeleteMode, 0, len(configs))
	toolsets := make([]Toolset, 0, len(configs))
	for _, cfg := range configs {
		if err := validateConfigPolicy(cfg); err != nil {
			return "", "", err
		}
		modes = append(modes, cfg.DeleteMode)
		toolsets = append(toolsets, cfg.Toolset)
	}
	return effectiveDeleteMode(optionMode, modes...), effectiveToolset(optionToolset, toolsets...), nil
}

func validateConfigPolicy(cfg Config) error {
	if err := validateDeleteMode(cfg.DeleteMode); err != nil {
		return err
	}
	return validateToolset(cfg.Toolset)
}

func validateDeleteMode(mode DeleteMode) error {
	switch mode {
	case "", DeleteModeSafe, DeleteModeFull, DeleteModeNone:
		return nil
	default:
		return fmt.Errorf("invalid delete mode %q", mode)
	}
}

func validateToolset(toolset Toolset) error {
	switch toolset {
	case "", ToolsetCompact, ToolsetCore, ToolsetFull:
		return nil
	default:
		return fmt.Errorf("invalid toolset %q", toolset)
	}
}

func effectiveDeleteMode(option DeleteMode, cfgs ...DeleteMode) DeleteMode {
	if option != "" {
		return option
	}
	for _, cfg := range cfgs {
		if cfg != "" {
			return cfg
		}
	}
	return DeleteModeSafe
}

func effectiveToolset(option Toolset, cfgs ...Toolset) Toolset {
	if option != "" {
		return option
	}
	for _, cfg := range cfgs {
		if cfg != "" {
			return cfg
		}
	}
	return ToolsetCore
}

func (cfg Config) toInternalValidated() (config.Config, error) {
	if err := validateConfigPolicy(cfg); err != nil {
		return config.Config{}, err
	}
	out := cfg.toInternal()
	normalizedAthleteID, err := config.NormalizeAthleteID(out.AthleteID)
	if err != nil {
		return config.Config{}, fmt.Errorf("normalizing athlete ID: %w", err)
	}
	out.AthleteID = normalizedAthleteID
	return out, nil
}

func (cfg Config) toInternal() config.Config {
	httpTimeout := cfg.HTTPTimeout
	if httpTimeout == 0 {
		httpTimeout = config.DefaultHTTPTimeout
	}
	timezone := cfg.Timezone
	if timezone == "" {
		timezone = config.DefaultTimezone
	}
	return config.Config{APIKey: cfg.APIKey, AthleteID: cfg.AthleteID, Timezone: timezone, APIBaseURL: cfg.APIBaseURL, HTTPTimeout: httpTimeout, HTTPBindAddress: cfg.HTTPBindAddress, DeleteMode: cfg.DeleteMode.toInternal(), Toolset: cfg.Toolset.toInternal(), DebugMetadata: cfg.DebugMetadata}
}

func (r RetryConfig) toInternal() intervals.RetryConfig {
	return intervals.RetryConfig{MaxAttempts: r.MaxAttempts, BaseDelay: r.BaseDelay, MaxDelay: r.MaxDelay, Jitter: r.Jitter}
}

func (m DeleteMode) toInternal() safety.Mode {
	return safety.ParseMode(string(m))
}

func (t Toolset) toInternal() safety.Toolset {
	return safety.ParseToolset(string(t))
}

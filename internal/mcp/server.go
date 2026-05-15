package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"

	"github.com/ricardocabral/icuvisor/internal/coach"
	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/prompts"
	"github.com/ricardocabral/icuvisor/internal/resources"
	"github.com/ricardocabral/icuvisor/internal/response"
	"github.com/ricardocabral/icuvisor/internal/safety"
	"github.com/ricardocabral/icuvisor/internal/toolcatalog"
	"github.com/ricardocabral/icuvisor/internal/tools"
)

const genericToolErrorMessage = "tool failed; try again or check icuvisor logs"
const genericResourceErrorMessage = "resource read failed; try again or check icuvisor logs"
const invalidTargetAthleteMessage = "invalid athlete_id; use a configured target athlete"
const athleteIDArgumentDescription = "Target athlete; defaults to selected athlete in coach mode, or the only athlete otherwise. Format: i12345 or 12345."

// StreamableHTTPPath is the local HTTP path serving the MCP Streamable HTTP endpoint.
const StreamableHTTPPath = "/mcp"

const streamableHTTPSessionTimeout = 30 * time.Minute
const streamableHTTPShutdownTimeout = 5 * time.Second

var snakeCaseToolName = regexp.MustCompile(`^[a-z][a-z0-9]*(?:_[a-z0-9]+)*$`)

// Options contains dependencies for constructing the MCP server.
type Options struct {
	Config           config.Config
	Version          string
	Logger           *slog.Logger
	Registry         tools.Registry
	ResourceRegistry resources.Registry
	PromptRegistry   prompts.Registry
	Capability       safety.Capability
	Toolset          safety.Toolset
	Transport        sdkmcp.Transport
	SelectionStore   *coach.SelectionStore
}

// Server wraps the SDK server and selected transport.
type Server struct {
	server      *sdkmcp.Server
	transport   sdkmcp.Transport
	logger      *slog.Logger
	version     string
	catalogHash string
}

// NewServer constructs an icuvisor MCP server.
func NewServer(ctx context.Context, opts Options) (*Server, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	version := opts.Version
	if version == "" {
		version = "dev"
	}
	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}

	transport := opts.Transport
	if transport == nil {
		transport = &sdkmcp.StdioTransport{}
	}

	sdkServer, err := newSDKServer(version, logger)
	if err != nil {
		return nil, err
	}
	catalogHash, err := hashToolCatalog(nil)
	if err != nil {
		return nil, fmt.Errorf("hashing empty tool catalog: %w", err)
	}
	selectionStore := opts.SelectionStore
	if selectionStore == nil {
		selectionStore = coach.NewSelectionStore(opts.Config.Coach.DefaultAthleteID)
	}
	if opts.Registry != nil {
		registrar := &safeRegistrar{server: sdkServer, logger: logger, config: opts.Config, coachEvaluator: coach.NewEvaluator(opts.Config.CoachModeEnabled(), opts.Config.Coach), selectionStore: selectionStore, capability: capabilityOrSafe(opts.Capability), toolset: toolsetOrCore(opts), names: make(map[string]struct{})}
		if err := opts.Registry.Register(ctx, registrar); err != nil {
			return nil, fmt.Errorf("registering tools: %w", err)
		}
		catalogHash, err = hashToolCatalog(registrar.registeredTools)
		if err != nil {
			return nil, fmt.Errorf("hashing tool catalog: %w", err)
		}
		if opts.Config.CoachModeEnabled() {
			sdkServer.AddReceivingMiddleware(registrar.visibilityMiddleware())
		}
		logger.Info("tool registration complete", "registered_count", registrar.registeredCount, "skipped_toolset_count", registrar.skippedToolsetCount, "skipped_capability_count", registrar.skippedCapabilityCount, "skipped_coach_count", registrar.skippedCoachCount)
	}
	if opts.ResourceRegistry != nil {
		registrar := &safeResourceRegistrar{server: sdkServer, logger: logger, uris: make(map[string]struct{})}
		if err := opts.ResourceRegistry.Register(ctx, registrar); err != nil {
			return nil, fmt.Errorf("registering resources: %w", err)
		}
		logger.Info("resource registration complete", "registered_count", registrar.registeredCount)
	}
	if opts.PromptRegistry != nil {
		registrar := &safePromptRegistrar{server: sdkServer, logger: logger, names: make(map[string]struct{})}
		if err := opts.PromptRegistry.Register(ctx, registrar); err != nil {
			return nil, fmt.Errorf("registering prompts: %w", err)
		}
		logger.Info("prompt registration complete", "registered_count", registrar.registeredCount)
	}
	response.SetRuntimeCatalogMetadata(version, catalogHash)

	return &Server{server: sdkServer, transport: transport, logger: logger, version: version, catalogHash: catalogHash}, nil
}

// CatalogHash reports the deterministic hash of the exposed tool catalog.
func (s *Server) CatalogHash() string {
	if s == nil {
		return ""
	}
	return s.catalogHash
}

// Run serves one MCP session until the client disconnects or ctx is cancelled.
func (s *Server) Run(ctx context.Context) error {
	if s == nil || s.server == nil || s.transport == nil {
		return errors.New("mcp server is not initialized")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	logger := s.logger
	if logger == nil {
		logger = slog.Default()
	}
	version := s.version
	if version == "" {
		version = "dev"
	}
	serverSession, err := s.server.Connect(ctx, s.transport, nil)
	if err != nil {
		logger.Error("server startup failed", "version", version, "transport", transportName(s.transport), "error", err)
		return err
	}
	logger.Info("server started listening", "version", version, "transport", transportName(s.transport))

	closed := make(chan error)
	go func() {
		closed <- serverSession.Wait()
	}()

	select {
	case <-ctx.Done():
		serverSession.Close()
		<-closed
		logger.Error("server run cancelled", "version", version, "transport", transportName(s.transport), "error", ctx.Err())
		return ctx.Err()
	case err := <-closed:
		if err != nil {
			logger.Error("server session ended with error", "version", version, "transport", transportName(s.transport), "error", err)
		} else {
			logger.Info("server session ended", "version", version, "transport", transportName(s.transport))
		}
		return err
	}
}

// RunStreamableHTTP serves the shared MCP server over Streamable HTTP.
func (s *Server) RunStreamableHTTP(ctx context.Context, address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("listening for streamable HTTP on %s: %w", address, err)
	}
	return s.ServeStreamableHTTP(ctx, listener)
}

// ServeStreamableHTTP serves Streamable HTTP on listener until ctx is cancelled.
func (s *Server) ServeStreamableHTTP(ctx context.Context, listener net.Listener) error {
	if s == nil || s.server == nil {
		return errors.New("mcp server is not initialized")
	}
	if listener == nil {
		return errors.New("streamable HTTP listener is nil")
	}
	if err := ctx.Err(); err != nil {
		_ = listener.Close()
		return err
	}
	logger := s.logger
	if logger == nil {
		logger = slog.Default()
	}
	version := s.version
	if version == "" {
		version = "dev"
	}

	streamableHandler := sdkmcp.NewStreamableHTTPHandler(func(*http.Request) *sdkmcp.Server {
		return s.server
	}, &sdkmcp.StreamableHTTPOptions{
		Stateless:                  false,
		JSONResponse:               false,
		Logger:                     logger,
		SessionTimeout:             streamableHTTPSessionTimeout,
		DisableLocalhostProtection: false,
		CrossOriginProtection:      nil,
	})
	mux := http.NewServeMux()
	mux.Handle(StreamableHTTPPath, streamableHandler)
	httpServer := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	serveDone := make(chan error, 1)
	go func() {
		serveDone <- httpServer.Serve(listener)
	}()
	logger.Info("server started listening", "version", version, "transport", "streamable_http", "address", listener.Addr().String(), "path", StreamableHTTPPath)

	select {
	case err := <-serveDone:
		return normalizeHTTPServerError(err)
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), streamableHTTPShutdownTimeout)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			logger.Error("streamable HTTP shutdown failed; closing listener", "version", version, "error", err)
			_ = httpServer.Close()
		}
		if err := normalizeHTTPServerError(<-serveDone); err != nil {
			return err
		}
		logger.Info("server session ended", "version", version, "transport", "streamable_http")
		return ctx.Err()
	}
}

func normalizeHTTPServerError(err error) error {
	if err == nil || errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func transportName(transport sdkmcp.Transport) string {
	switch transport.(type) {
	case *sdkmcp.StdioTransport:
		return "stdio"
	case *sdkmcp.IOTransport:
		return "io"
	case *sdkmcp.InMemoryTransport:
		return "in_memory"
	case *sdkmcp.StreamableServerTransport:
		return "streamable_http"
	default:
		return fmt.Sprintf("%T", transport)
	}
}

func capabilityOrSafe(capability safety.Capability) safety.Capability {
	if capability != nil {
		return capability
	}
	return safety.NewCapability(safety.ModeSafe)
}

func toolsetOrCore(opts Options) safety.Toolset {
	toolset := opts.Toolset
	if toolset == "" {
		toolset = opts.Config.Toolset
	}
	return safety.ParseToolset(string(toolset))
}

// withPanicRecovery converts SDK protocol-boundary panics into errors so callers can fail safely.
func withPanicRecovery(name string, fn func() error) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("%s: %v", name, recovered)
		}
	}()
	return fn()
}

func newSDKServer(version string, logger *slog.Logger) (server *sdkmcp.Server, err error) {
	err = withPanicRecovery("constructing MCP server", func() error {
		server = sdkmcp.NewServer(&sdkmcp.Implementation{Name: "icuvisor", Version: version}, &sdkmcp.ServerOptions{Logger: logger})
		return nil
	})
	return server, err
}

type safeRegistrar struct {
	server                 *sdkmcp.Server
	logger                 *slog.Logger
	config                 config.Config
	coachEvaluator         coach.Evaluator
	selectionStore         *coach.SelectionStore
	capability             safety.Capability
	toolset                safety.Toolset
	names                  map[string]struct{}
	registeredTools        []tools.Tool
	coachVisibleCatalog    []tools.Tool
	registeredCount        int
	skippedToolsetCount    int
	skippedCapabilityCount int
	skippedCoachCount      int
}

func (r *safeRegistrar) AddTool(tool tools.Tool) error {
	tool = r.prepareTool(tool)
	if err := r.validateTool(tool); err != nil {
		return err
	}
	r.names[tool.Name] = struct{}{}
	if tool.Name == toolcatalog.ICUvisorListAdvancedCapabilities && r.config.CoachModeEnabled() {
		tool.Handler = r.coachFilteredAdvancedCapabilitiesHandler()
	}
	if !r.capabilityAllows(tool) {
		r.skippedCapabilityCount++
		return nil
	}
	if tool.Name != toolcatalog.ICUvisorListAdvancedCapabilities && r.coachAllows(tool) {
		r.coachVisibleCatalog = append(r.coachVisibleCatalog, tool)
	}
	if !r.toolsetAllows(tool) {
		r.skippedToolsetCount++
		return nil
	}
	if !r.coachAllows(tool) {
		r.skippedCoachCount++
		return nil
	}

	return withPanicRecovery(fmt.Sprintf("registering tool %q", tool.Name), func() error {
		r.server.AddTool(&sdkmcp.Tool{
			Name:         tool.Name,
			Description:  tool.Description,
			InputSchema:  tool.InputSchema,
			OutputSchema: tool.OutputSchema,
		}, func(ctx context.Context, req *sdkmcp.CallToolRequest) (*sdkmcp.CallToolResult, error) {
			callCtx := r.withSelection(ctx, req.Session)
			callCtx, arguments, err := r.resolveToolTarget(callCtx, tool.Name, req.Params.Arguments)
			if err != nil {
				return toolErrorResult(publicToolErrorMessage(err)), nil
			}
			result, err := tool.Handler(callCtx, tools.Request{
				Name:      req.Params.Name,
				Arguments: arguments,
			})
			if err != nil {
				r.logger.Error("tool handler failed", "tool", tool.Name, "error", err)
				return toolErrorResult(publicToolErrorMessage(err)), nil
			}
			converted, err := convertResult(result)
			if err != nil {
				r.logger.Error("tool result conversion failed", "tool", tool.Name, "error", err)
				return toolErrorResult(genericToolErrorMessage), nil
			}
			return converted, nil
		})
		r.registeredTools = append(r.registeredTools, tool)
		r.registeredCount++
		return nil
	})
}

func (r *safeRegistrar) visibilityMiddleware() sdkmcp.Middleware {
	return func(next sdkmcp.MethodHandler) sdkmcp.MethodHandler {
		return func(ctx context.Context, method string, req sdkmcp.Request) (sdkmcp.Result, error) {
			result, err := next(ctx, method, req)
			if err != nil || method != "tools/list" {
				return result, err
			}
			toolsResult, ok := result.(*sdkmcp.ListToolsResult)
			if !ok {
				return result, err
			}
			selectionCtx := r.withSelection(ctx, req.GetSession().(*sdkmcp.ServerSession))
			athleteID := r.config.Coach.DefaultAthleteID
			if selection, ok := coach.SelectionContextFromContext(selectionCtx); ok && selection.Store != nil {
				athleteID = selection.Store.Selected(selection.Key)
			}
			filtered := toolsResult.Tools[:0]
			for _, tool := range toolsResult.Tools {
				if r.visibleForAthlete(athleteID, tool.Name) {
					filtered = append(filtered, tool)
				}
			}
			toolsResult.Tools = filtered
			return toolsResult, nil
		}
	}
}

func (r *safeRegistrar) visibleToolNamesForAthlete(athleteID string) []string {
	out := make([]string, 0, len(r.registeredTools))
	for _, tool := range r.registeredTools {
		if r.visibleForAthlete(athleteID, tool.Name) {
			out = append(out, tool.Name)
		}
	}
	sort.Strings(out)
	return out
}

func (r *safeRegistrar) visibleForAthlete(athleteID string, toolName string) bool {
	if toolName == toolcatalog.ListAthletes || toolName == toolcatalog.SelectAthlete || toolName == toolcatalog.ICUvisorListAdvancedCapabilities {
		return true
	}
	allowed, _ := r.coachEvaluator.Evaluate(athleteID, toolName)
	return allowed
}

func (r *safeRegistrar) prepareTool(tool tools.Tool) tools.Tool {
	if !r.config.CoachModeEnabled() || !toolcatalog.IsAthleteScopedTool(tool.Name) {
		return tool
	}
	tool.InputSchema = schemaWithAthleteID(tool.InputSchema)
	return tool
}

func (r *safeRegistrar) coachAllows(tool tools.Tool) bool {
	return r.coachEvaluator.AllowedForAny(tool.Name)
}

func (r *safeRegistrar) withSelection(ctx context.Context, session *sdkmcp.ServerSession) context.Context {
	if r.selectionStore == nil {
		return ctx
	}
	sessionID := ""
	if session != nil {
		sessionID = session.ID()
	}
	key, scope := r.selectionStore.Key(sessionID)
	return coach.WithSelectionContext(ctx, coach.SelectionContext{Store: r.selectionStore, Key: key, Scope: scope, VisibleTools: r.visibleToolNamesForAthlete})
}

func (r *safeRegistrar) resolveToolTarget(ctx context.Context, toolName string, raw json.RawMessage) (context.Context, json.RawMessage, error) {
	if !r.config.CoachModeEnabled() || !toolcatalog.IsAthleteScopedTool(toolName) {
		return ctx, raw, nil
	}
	arguments, suppliedAthleteID, err := stripAthleteID(raw)
	if err != nil {
		return ctx, nil, tools.NewUserError(invalidTargetAthleteMessage, err)
	}
	targetAthleteID, err := r.resolveAthleteID(ctx, suppliedAthleteID)
	if err != nil {
		return ctx, nil, tools.NewUserError(invalidTargetAthleteMessage, err)
	}
	if err := r.coachEvaluator.MustEvaluate(targetAthleteID, toolName); err != nil {
		return ctx, nil, tools.NewUserError(invalidTargetAthleteMessage, err)
	}
	return intervals.WithTargetAthleteID(ctx, targetAthleteID), arguments, nil
}

func (r *safeRegistrar) resolveAthleteID(ctx context.Context, suppliedAthleteID string) (string, error) {
	if r.config.CoachModeEnabled() {
		targetAthleteID := strings.TrimSpace(suppliedAthleteID)
		if targetAthleteID == "" {
			targetAthleteID = r.config.Coach.DefaultAthleteID
			if selection, ok := coach.SelectionContextFromContext(ctx); ok && selection.Store != nil {
				targetAthleteID = selection.Store.Selected(selection.Key)
			}
		}
		normalized, err := config.NormalizeAthleteID(targetAthleteID)
		if err != nil || !r.coachEvaluator.HasAthlete(normalized) {
			return "", errors.New("invalid target athlete")
		}
		return normalized, nil
	}
	configured := r.config.AthleteID
	targetAthleteID := strings.TrimSpace(suppliedAthleteID)
	if targetAthleteID == "" {
		return configured, nil
	}
	normalized, err := config.NormalizeAthleteID(targetAthleteID)
	if err != nil || normalized != configured {
		return "", errors.New("invalid target athlete")
	}
	return normalized, nil
}

func schemaWithAthleteID(schema any) any {
	asMap, ok := schema.(map[string]any)
	if !ok {
		return schema
	}
	out := make(map[string]any, len(asMap))
	for key, value := range asMap {
		out[key] = value
	}
	properties, _ := asMap["properties"].(map[string]any)
	copiedProperties := make(map[string]any, len(properties)+1)
	for key, value := range properties {
		copiedProperties[key] = value
	}
	copiedProperties["athlete_id"] = map[string]any{"type": "string", "description": athleteIDArgumentDescription}
	out["properties"] = copiedProperties
	return out
}

func stripAthleteID(raw json.RawMessage) (json.RawMessage, string, error) {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return raw, "", nil
	}
	if !strings.HasPrefix(trimmed, "{") {
		return nil, "", errors.New("arguments must be an object")
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, "", err
	}
	athleteRaw, ok := fields["athlete_id"]
	if !ok {
		return raw, "", nil
	}
	delete(fields, "athlete_id")
	var athleteID string
	if len(athleteRaw) > 0 && strings.TrimSpace(string(athleteRaw)) != "null" {
		if err := json.Unmarshal(athleteRaw, &athleteID); err != nil {
			return nil, "", err
		}
	}
	cleaned, err := json.Marshal(fields)
	if err != nil {
		return nil, "", err
	}
	return cleaned, athleteID, nil
}

func (r *safeRegistrar) coachFilteredAdvancedCapabilitiesHandler() tools.Handler {
	catalog := append([]tools.Tool(nil), r.coachVisibleCatalog...)
	toolset := safety.ParseToolset(string(r.toolset))
	return func(ctx context.Context, req tools.Request) (tools.Result, error) {
		if err := ctx.Err(); err != nil {
			return tools.Result{}, err
		}
		trimmed := strings.TrimSpace(string(req.Arguments))
		if trimmed != "" && trimmed != "{}" && trimmed != "null" {
			return tools.Result{}, tools.NewUserError("invalid icuvisor_list_advanced_capabilities arguments; no arguments are supported", nil)
		}
		athleteID := r.config.Coach.DefaultAthleteID
		if selection, ok := coach.SelectionContextFromContext(ctx); ok && selection.Store != nil {
			athleteID = selection.Store.Selected(selection.Key)
		}
		rows := make([]map[string]any, 0, len(catalog))
		for _, tool := range catalog {
			if tool.Name == toolcatalog.ICUvisorListAdvancedCapabilities || tool.EffectiveToolset() != safety.ToolsetFull || !r.visibleForAthlete(athleteID, tool.Name) {
				continue
			}
			rows = append(rows, map[string]any{"name": tool.Name, "summary": firstSentence(tool.Description), "requirement": toolRequirement(tool)})
		}
		sort.Slice(rows, func(i, j int) bool { return rows[i]["name"].(string) < rows[j]["name"].(string) })
		status := "The default core toolset is active; full-only tools are hidden from tools/list."
		if toolset == safety.ToolsetFull {
			status = "The full toolset is already enabled; these full-only tools should already be visible when delete-mode also allows them."
		}
		return tools.TextResult(map[string]any{
			"current_toolset":       toolset.String(),
			"status":                status,
			"enable_instruction":    "Set ICUVISOR_TOOLSET=full in the MCP client/server environment and restart icuvisor to enable the full icuvisor toolset.",
			"advanced_capabilities": rows,
			"_meta": map[string]any{
				"count":            len(rows),
				"source":           "registered catalog metadata",
				"delete_mode_note": "Tools with requirement=delete also require ICUVISOR_DELETE_MODE=full; write tools require delete mode safe or full.",
				"toolset":          toolset.String(),
			},
		}), nil
	}
}

func firstSentence(description string) string {
	description = strings.Join(strings.Fields(description), " ")
	if idx := strings.Index(description, "."); idx >= 0 {
		return strings.TrimSpace(description[:idx+1])
	}
	return description
}

func toolRequirement(tool tools.Tool) string {
	if tool.RequiresDelete() {
		return string(tools.RequirementDelete)
	}
	if tool.RequiresWrite() {
		return string(tools.RequirementWrite)
	}
	return string(tools.RequirementRead)
}

func (r *safeRegistrar) toolsetAllows(tool tools.Tool) bool {
	active := safety.ParseToolset(string(r.toolset))
	if active == safety.ToolsetFull {
		return true
	}
	return tool.EffectiveToolset() == safety.ToolsetCore
}

func (r *safeRegistrar) capabilityAllows(tool tools.Tool) bool {
	capability := r.capability
	if capability == nil {
		capability = safety.NewCapability(safety.ModeSafe)
	}
	if tool.RequiresDelete() && !capability.CanDelete() {
		return false
	}
	return !tool.RequiresWrite() || capability.CanWrite()
}

func (r *safeRegistrar) validateTool(tool tools.Tool) error {
	if !snakeCaseToolName.MatchString(tool.Name) {
		return fmt.Errorf("invalid tool name %q; use snake_case", tool.Name)
	}
	if _, exists := r.names[tool.Name]; exists {
		return fmt.Errorf("duplicate tool name %q", tool.Name)
	}
	if tool.Description == "" {
		return fmt.Errorf("tool %q is missing a description", tool.Name)
	}
	if err := validateToolset(tool); err != nil {
		return err
	}
	if err := validateObjectSchema("input", tool.Name, tool.InputSchema, true); err != nil {
		return err
	}
	if err := validateObjectSchema("output", tool.Name, tool.OutputSchema, false); err != nil {
		return err
	}
	if tool.Handler == nil {
		return fmt.Errorf("tool %q is missing a handler", tool.Name)
	}
	return nil
}

type safeResourceRegistrar struct {
	server          *sdkmcp.Server
	logger          *slog.Logger
	uris            map[string]struct{}
	registeredCount int
}

func (r *safeResourceRegistrar) AddResource(resource resources.Resource) error {
	if err := r.validateResource(resource); err != nil {
		return err
	}
	r.uris[resource.URI] = struct{}{}

	return withPanicRecovery(fmt.Sprintf("registering resource %q", resource.URI), func() error {
		r.server.AddResource(&sdkmcp.Resource{
			URI:         resource.URI,
			Name:        resource.Name,
			Title:       resource.Title,
			Description: resource.Description,
			MIMEType:    resource.MIMEType,
		}, func(ctx context.Context, req *sdkmcp.ReadResourceRequest) (*sdkmcp.ReadResourceResult, error) {
			result, err := resource.Handler(ctx, resources.Request{URI: req.Params.URI})
			if err != nil {
				if isResourceNotFound(err) {
					return nil, err
				}
				r.logger.Error("resource handler failed", "resource_uri", resource.URI, "error", err)
				return nil, errors.New(genericResourceErrorMessage)
			}
			return &sdkmcp.ReadResourceResult{Contents: []*sdkmcp.ResourceContents{{
				URI:      stringOrDefault(result.URI, req.Params.URI),
				MIMEType: stringOrDefault(result.MIMEType, resource.MIMEType),
				Text:     result.Text,
			}}}, nil
		})
		r.registeredCount++
		return nil
	})
}

func (r *safeResourceRegistrar) validateResource(resource resources.Resource) error {
	if resource.URI == "" {
		return errors.New("resource is missing a URI")
	}
	parsed, err := url.Parse(resource.URI)
	if err != nil || !parsed.IsAbs() || parsed.Scheme != "icuvisor" {
		return fmt.Errorf("invalid resource URI %q; use absolute icuvisor:// URI", resource.URI)
	}
	if _, exists := r.uris[resource.URI]; exists {
		return fmt.Errorf("duplicate resource URI %q", resource.URI)
	}
	if resource.Name == "" {
		return fmt.Errorf("resource %q is missing a name", resource.URI)
	}
	if resource.Title == "" {
		return fmt.Errorf("resource %q is missing a title", resource.URI)
	}
	if resource.Description == "" {
		return fmt.Errorf("resource %q is missing a description", resource.URI)
	}
	if resource.MIMEType == "" {
		return fmt.Errorf("resource %q is missing a MIME type", resource.URI)
	}
	if resource.Handler == nil {
		return fmt.Errorf("resource %q is missing a handler", resource.URI)
	}
	return nil
}

func isResourceNotFound(err error) bool {
	var rpcErr *jsonrpc.Error
	return errors.As(err, &rpcErr) && rpcErr.Code == sdkmcp.CodeResourceNotFound
}

func stringOrDefault(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

func validateToolset(tool tools.Tool) error {
	switch tool.Toolset {
	case "", safety.ToolsetCore, safety.ToolsetFull:
		return nil
	default:
		return fmt.Errorf("tool %q has invalid toolset %q", tool.Name, tool.Toolset)
	}
}

func validateObjectSchema(kind, name string, schema any, required bool) error {
	if schema == nil {
		if required {
			return fmt.Errorf("tool %q is missing an %s schema", name, kind)
		}
		return nil
	}
	if asMap, ok := schema.(map[string]any); ok {
		if asMap["type"] == "object" {
			return nil
		}
		return fmt.Errorf("tool %q %s schema must have type object", name, kind)
	}
	return nil
}

func convertResult(result tools.Result) (*sdkmcp.CallToolResult, error) {
	content, err := convertContent(result.Content)
	if err != nil {
		return nil, err
	}
	return &sdkmcp.CallToolResult{
		Content:           content,
		StructuredContent: result.StructuredContent,
		IsError:           result.IsError,
	}, nil
}

func convertContent(content []tools.Content) ([]sdkmcp.Content, error) {
	if len(content) == 0 {
		return nil, nil
	}
	converted := make([]sdkmcp.Content, 0, len(content))
	for _, item := range content {
		switch item.Type {
		case "", tools.ContentTypeText:
			converted = append(converted, &sdkmcp.TextContent{Text: item.Text})
		default:
			return nil, fmt.Errorf("unsupported content type %q", item.Type)
		}
	}
	return converted, nil
}

func publicToolErrorMessage(err error) string {
	if message, ok := tools.PublicErrorMessage(err); ok {
		return message
	}
	return genericToolErrorMessage
}

func toolErrorResult(message string) *sdkmcp.CallToolResult {
	return &sdkmcp.CallToolResult{
		Content: []sdkmcp.Content{
			&sdkmcp.TextContent{Text: message},
		},
		IsError: true,
	}
}

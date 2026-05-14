package mcp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"time"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"

	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/prompts"
	"github.com/ricardocabral/icuvisor/internal/resources"
	"github.com/ricardocabral/icuvisor/internal/safety"
	"github.com/ricardocabral/icuvisor/internal/tools"
)

const genericToolErrorMessage = "tool failed; try again or check icuvisor logs"
const genericResourceErrorMessage = "resource read failed; try again or check icuvisor logs"

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
}

// Server wraps the SDK server and selected transport.
type Server struct {
	server    *sdkmcp.Server
	transport sdkmcp.Transport
	logger    *slog.Logger
	version   string
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
	if opts.Registry != nil {
		registrar := &safeRegistrar{server: sdkServer, logger: logger, capability: capabilityOrSafe(opts.Capability), toolset: toolsetOrCore(opts), names: make(map[string]struct{})}
		if err := opts.Registry.Register(ctx, registrar); err != nil {
			return nil, fmt.Errorf("registering tools: %w", err)
		}
		logger.Info("tool registration complete", "registered_count", registrar.registeredCount, "skipped_toolset_count", registrar.skippedToolsetCount, "skipped_capability_count", registrar.skippedCapabilityCount)
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

	return &Server{server: sdkServer, transport: transport, logger: logger, version: version}, nil
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

func newSDKServer(version string, logger *slog.Logger) (server *sdkmcp.Server, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("constructing MCP server: %v", recovered)
		}
	}()
	return sdkmcp.NewServer(&sdkmcp.Implementation{Name: "icuvisor", Version: version}, &sdkmcp.ServerOptions{Logger: logger}), nil
}

type safeRegistrar struct {
	server                 *sdkmcp.Server
	logger                 *slog.Logger
	capability             safety.Capability
	toolset                safety.Toolset
	names                  map[string]struct{}
	registeredCount        int
	skippedToolsetCount    int
	skippedCapabilityCount int
}

func (r *safeRegistrar) AddTool(tool tools.Tool) (err error) {
	if err := r.validateTool(tool); err != nil {
		return err
	}
	r.names[tool.Name] = struct{}{}
	skippedByToolset := !r.toolsetAllows(tool)
	skippedByCapability := !r.capabilityAllows(tool)
	if skippedByToolset {
		r.skippedToolsetCount++
	}
	if skippedByCapability {
		r.skippedCapabilityCount++
	}
	if skippedByToolset || skippedByCapability {
		return nil
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("registering tool %q: %v", tool.Name, recovered)
		}
	}()

	r.server.AddTool(&sdkmcp.Tool{
		Name:         tool.Name,
		Description:  tool.Description,
		InputSchema:  tool.InputSchema,
		OutputSchema: tool.OutputSchema,
	}, func(ctx context.Context, req *sdkmcp.CallToolRequest) (*sdkmcp.CallToolResult, error) {
		result, err := tool.Handler(ctx, tools.Request{
			Name:      req.Params.Name,
			Arguments: req.Params.Arguments,
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
	r.registeredCount++
	return nil
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

func (r *safeResourceRegistrar) AddResource(resource resources.Resource) (err error) {
	if err := r.validateResource(resource); err != nil {
		return err
	}
	r.uris[resource.URI] = struct{}{}
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("registering resource %q: %v", resource.URI, recovered)
		}
	}()

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

package mcp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/safety"
	"github.com/ricardocabral/icuvisor/internal/tools"
)

const genericToolErrorMessage = "tool failed; try again or check icuvisor logs"

var snakeCaseToolName = regexp.MustCompile(`^[a-z][a-z0-9]*(?:_[a-z0-9]+)*$`)

// Options contains dependencies for constructing the MCP server.
type Options struct {
	Config     config.Config
	Version    string
	Logger     *slog.Logger
	Registry   tools.Registry
	Capability safety.Capability
	Transport  sdkmcp.Transport
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
		registrar := &safeRegistrar{server: sdkServer, logger: logger, capability: capabilityOrSafe(opts.Capability), names: make(map[string]struct{})}
		if err := opts.Registry.Register(ctx, registrar); err != nil {
			return nil, fmt.Errorf("registering tools: %w", err)
		}
		logger.Info("tool registration complete", "registered_count", registrar.registeredCount, "skipped_count", registrar.skippedCount)
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

func newSDKServer(version string, logger *slog.Logger) (server *sdkmcp.Server, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("constructing MCP server: %v", recovered)
		}
	}()
	return sdkmcp.NewServer(&sdkmcp.Implementation{Name: "icuvisor", Version: version}, &sdkmcp.ServerOptions{Logger: logger}), nil
}

type safeRegistrar struct {
	server          *sdkmcp.Server
	logger          *slog.Logger
	capability      safety.Capability
	names           map[string]struct{}
	registeredCount int
	skippedCount    int
}

func (r *safeRegistrar) AddTool(tool tools.Tool) (err error) {
	if err := r.validateTool(tool); err != nil {
		return err
	}
	if r.shouldSkip(tool) {
		r.skippedCount++
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
	r.names[tool.Name] = struct{}{}
	r.registeredCount++
	return nil
}

func (r *safeRegistrar) shouldSkip(tool tools.Tool) bool {
	capability := r.capability
	if capability == nil {
		capability = safety.NewCapability(safety.ModeSafe)
	}
	if tool.RequiresDelete() && !capability.CanDelete() {
		return true
	}
	return tool.RequiresWrite() && !capability.CanWrite()
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

package mcp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/tools"
)

var snakeCaseToolName = regexp.MustCompile(`^[a-z][a-z0-9]*(?:_[a-z0-9]+)*$`)

// Options contains dependencies for constructing the MCP server.
type Options struct {
	Config    config.Config
	Version   string
	Logger    *slog.Logger
	Registry  tools.Registry
	Transport sdkmcp.Transport
}

// Server wraps the SDK server and selected transport.
type Server struct {
	server    *sdkmcp.Server
	transport sdkmcp.Transport
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
		registrar := &safeRegistrar{server: sdkServer}
		if err := opts.Registry.Register(ctx, registrar); err != nil {
			return nil, fmt.Errorf("registering tools: %w", err)
		}
	}

	return &Server{server: sdkServer, transport: transport}, nil
}

// Run serves one MCP session until the client disconnects or ctx is cancelled.
func (s *Server) Run(ctx context.Context) error {
	if s == nil || s.server == nil || s.transport == nil {
		return errors.New("mcp server is not initialized")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return s.server.Run(ctx, s.transport)
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
	server *sdkmcp.Server
}

func (r *safeRegistrar) AddTool(tool tools.Tool) (err error) {
	if err := validateTool(tool); err != nil {
		return err
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("registering tool %q: %v", tool.Name, recovered)
		}
	}()

	r.server.AddTool(&sdkmcp.Tool{
		Name:        tool.Name,
		Description: tool.Description,
		InputSchema: tool.InputSchema,
	}, func(ctx context.Context, req *sdkmcp.CallToolRequest) (*sdkmcp.CallToolResult, error) {
		result, err := tool.Handler(ctx, tools.Request{
			Name:      req.Params.Name,
			Arguments: req.Params.Arguments,
		})
		if err != nil {
			return toolErrorResult(err.Error()), nil
		}
		return convertResult(result), nil
	})
	return nil
}

func validateTool(tool tools.Tool) error {
	if !snakeCaseToolName.MatchString(tool.Name) {
		return fmt.Errorf("invalid tool name %q; use snake_case", tool.Name)
	}
	if tool.Description == "" {
		return fmt.Errorf("tool %q is missing a description", tool.Name)
	}
	if tool.InputSchema == nil {
		return fmt.Errorf("tool %q is missing an input schema", tool.Name)
	}
	if tool.Handler == nil {
		return fmt.Errorf("tool %q is missing a handler", tool.Name)
	}
	return nil
}

func convertResult(result tools.Result) *sdkmcp.CallToolResult {
	return &sdkmcp.CallToolResult{
		Content:           convertContent(result.Content),
		StructuredContent: result.StructuredContent,
		IsError:           result.IsError,
	}
}

func convertContent(content []tools.Content) []sdkmcp.Content {
	if len(content) == 0 {
		return nil
	}
	converted := make([]sdkmcp.Content, 0, len(content))
	for _, item := range content {
		switch item.Type {
		case "", tools.ContentTypeText:
			converted = append(converted, &sdkmcp.TextContent{Text: item.Text})
		}
	}
	return converted
}

func toolErrorResult(message string) *sdkmcp.CallToolResult {
	return &sdkmcp.CallToolResult{
		Content: []sdkmcp.Content{
			&sdkmcp.TextContent{Text: message},
		},
		IsError: true,
	}
}

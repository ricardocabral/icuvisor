// Package mcp wires icuvisor registries into the MCP SDK server.
package mcp

import (
	"context"
	"fmt"
	"log/slog"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ricardocabral/icuvisor/internal/coach"
	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/diagnostics"
	"github.com/ricardocabral/icuvisor/internal/prompts"
	"github.com/ricardocabral/icuvisor/internal/resources"
	"github.com/ricardocabral/icuvisor/internal/response"
	"github.com/ricardocabral/icuvisor/internal/safety"
	"github.com/ricardocabral/icuvisor/internal/tools"
)

// Options contains dependencies for constructing the MCP server.
type Options struct {
	Config                     config.Config
	Version                    string
	Logger                     *slog.Logger
	Registry                   tools.Registry
	ResourceRegistry           resources.Registry
	PromptRegistry             prompts.Registry
	Capability                 safety.Capability
	Toolset                    safety.Toolset
	Transport                  sdkmcp.Transport
	SelectionStore             *coach.SelectionStore
	RecentToolCallRecorder     diagnostics.RecentToolCallRecorder
	SkipRuntimeCatalogMetadata bool
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
		coachEvaluator := coach.NewEvaluator(opts.Config.CoachModeEnabled(), opts.Config.Coach)
		registrar := &safeRegistrar{server: sdkServer, logger: logger, config: opts.Config, coachFilter: coach.NewToolFilter(coachEvaluator), selectionStore: selectionStore, capability: capabilityOrSafe(opts.Capability), toolset: toolsetOrCore(opts), names: make(map[string]struct{}), recentToolCalls: opts.RecentToolCallRecorder}
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
	if !opts.SkipRuntimeCatalogMetadata {
		response.SetRuntimeCatalogMetadata(version, catalogHash)
	}

	return &Server{server: sdkServer, transport: transport, logger: logger, version: version, catalogHash: catalogHash}, nil
}

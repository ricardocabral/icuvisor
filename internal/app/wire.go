package app

import (
	"context"
	"log/slog"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/clients"
	"github.com/ricardocabral/icuvisor/internal/coach"
	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/diagnostics"
	"github.com/ricardocabral/icuvisor/internal/intervals"
	mcpserver "github.com/ricardocabral/icuvisor/internal/mcp"
	"github.com/ricardocabral/icuvisor/internal/prompts"
	"github.com/ricardocabral/icuvisor/internal/resources"
	"github.com/ricardocabral/icuvisor/internal/safety"
	"github.com/ricardocabral/icuvisor/internal/tools"
	core "github.com/ricardocabral/icuvisor/pkg/icuvisor"
)

type deps struct {
	Logger                 func() *slog.Logger
	NewClient              func(intervals.Options) (*intervals.Client, error)
	NewSelectionStore      func(string) *coach.SelectionStore
	RecentToolCallRecorder func() (diagnostics.RecentToolCallRecorder, error)
	NewPromptRegistry      func() prompts.Registry
	NewResourceRegistry    func(clients.ProfileClient, resources.ResourceOptions) resources.Registry
	NewToolRegistry        func(*intervals.Client, tools.RegistryOptions) tools.Registry
	NewServer              func(context.Context, mcpserver.Options) (*mcpserver.Server, error)
}

func (d deps) withDefaults() deps {
	if d.Logger == nil {
		d.Logger = slog.Default
	}
	if d.NewClient == nil {
		d.NewClient = intervals.NewClient
	}
	if d.NewSelectionStore == nil {
		d.NewSelectionStore = coach.NewSelectionStore
	}
	if d.RecentToolCallRecorder == nil {
		d.RecentToolCallRecorder = defaultRecentToolCallRecorder
	}
	if d.NewPromptRegistry == nil {
		d.NewPromptRegistry = prompts.NewRegistry
	}
	if d.NewResourceRegistry == nil {
		d.NewResourceRegistry = resources.NewRegistryWithOptions
	}
	if d.NewToolRegistry == nil {
		d.NewToolRegistry = tools.NewRegistryWithOptions
	}
	if d.NewServer == nil {
		d.NewServer = mcpserver.NewServer
	}
	return d
}

func defaultStartServer(ctx context.Context, info ServerInfo) error {
	if info.Config.CoachModeEnabled() {
		server, cleanup, err := wireServer(ctx, info, deps{})
		if err != nil {
			return err
		}
		defer cleanup()
		if info.Config.Transport == config.TransportHTTP {
			return server.RunStreamableHTTP(ctx, info.Config.HTTPBindAddress)
		}
		return server.Run(ctx)
	}

	server, err := wireCoreServer(ctx, info)
	if err != nil {
		return err
	}
	if info.Config.Transport == config.TransportHTTP {
		return server.RunStreamableHTTP(ctx, info.Config.HTTPBindAddress)
	}
	return server.Run(ctx)
}

func wireCoreServer(ctx context.Context, info ServerInfo) (*core.Server, error) {
	logger := slog.Default()
	version := strings.TrimSpace(info.Version)
	if version == "" {
		version = "dev"
	}
	info.Version = version
	logger.Info("server starting", "version", version, "athlete_id", info.Config.AthleteID, "api_key_source", info.Config.APIKeySource)
	if info.Config.Transport == config.TransportHTTP && !config.HTTPBindAddressIsLoopback(info.Config.HTTPBindAddress) {
		logger.Warn("http transport non-loopback bind active", "transport", info.Config.Transport, "http_bind", info.Config.HTTPBindAddress, "security", "any host that can reach this address can connect")
	}

	capability := info.Capability
	if capability == nil {
		capability = safety.NewCapability(info.DeleteMode)
	}
	deleteMode := safety.ParseMode(capability.Mode())
	toolset := safety.ParseToolset(info.Toolset.String())
	safety.LogResolvedMode(logger, deleteMode)
	safety.LogResolvedToolset(logger, toolset)

	coreConfig := coreConfigFromInternal(info.Config)
	client, err := core.NewAPIKeyClient(core.APIKeyClientOptions{Config: coreConfig, Version: info.Version})
	if err != nil {
		return nil, err
	}
	recentToolCalls, recentErr := defaultRecentToolCallRecorder()
	if recentErr != nil {
		logger.Warn("diagnostics recent-tool-call recorder unavailable", "error", recentErr)
	}
	registry := core.NewCoreRegistry(client, core.RegistryOptions{
		Version:          info.Version,
		TimezoneFallback: info.Config.Timezone,
		DebugMetadata:    info.DebugMetadata,
		DeleteMode:       core.DeleteMode(deleteMode.String()),
		Toolset:          core.Toolset(toolset.String()),
	})
	resourceRegistry := core.NewResourceRegistry(client, core.ResourceRegistryOptions{
		Version:          info.Version,
		TimezoneFallback: info.Config.Timezone,
		DebugMetadata:    info.DebugMetadata,
		DeleteMode:       core.DeleteMode(deleteMode.String()),
		Toolset:          core.Toolset(toolset.String()),
	})
	return core.NewServer(ctx, core.ServerOptions{
		Config:                 coreConfig,
		Version:                info.Version,
		Logger:                 logger,
		Registry:               registry,
		ResourceRegistry:       resourceRegistry,
		PromptRegistry:         core.NewPromptRegistry(),
		DeleteMode:             core.DeleteMode(deleteMode.String()),
		Toolset:                core.Toolset(toolset.String()),
		RecentToolCallRecorder: recentToolCalls,
	})
}

func coreConfigFromInternal(cfg config.Config) core.Config {
	return core.Config{
		APIKey:          cfg.APIKey,
		AthleteID:       cfg.AthleteID,
		Timezone:        cfg.Timezone,
		APIBaseURL:      cfg.APIBaseURL,
		HTTPTimeout:     cfg.HTTPTimeout,
		DeleteMode:      core.DeleteMode(cfg.DeleteMode.String()),
		Toolset:         core.Toolset(cfg.Toolset.String()),
		DebugMetadata:   cfg.DebugMetadata,
		HTTPBindAddress: cfg.HTTPBindAddress,
	}
}

func wireServer(ctx context.Context, info ServerInfo, d deps) (*mcpserver.Server, func(), error) {
	d = d.withDefaults()
	cleanup := func() {}

	logger := d.Logger()
	version := strings.TrimSpace(info.Version)
	if version == "" {
		version = "dev"
	}
	info.Version = version
	logger.Info("server starting", "version", version, "athlete_id", info.Config.AthleteID, "api_key_source", info.Config.APIKeySource)
	if info.Config.Transport == config.TransportHTTP && !config.HTTPBindAddressIsLoopback(info.Config.HTTPBindAddress) {
		logger.Warn("http transport non-loopback bind active", "transport", info.Config.Transport, "http_bind", info.Config.HTTPBindAddress, "security", "any host that can reach this address can connect")
	}

	capability := info.Capability
	if capability == nil {
		capability = safety.NewCapability(info.DeleteMode)
	}
	deleteMode := safety.ParseMode(capability.Mode())
	toolset := safety.ParseToolset(info.Toolset.String())
	safety.LogResolvedMode(logger, deleteMode)
	safety.LogResolvedToolset(logger, toolset)
	client, err := d.NewClient(intervals.Options{Config: info.Config, Version: info.Version})
	if err != nil {
		return nil, cleanup, err
	}
	selectionStore := d.NewSelectionStore(info.Config.Coach.DefaultAthleteID)
	recentToolCalls, recentErr := d.RecentToolCallRecorder()
	if recentErr != nil {
		logger.Warn("diagnostics recent-tool-call recorder unavailable", "error", recentErr)
	}
	server, err := d.NewServer(ctx, mcpserver.Options{
		Config:                 info.Config,
		Version:                info.Version,
		Logger:                 logger,
		Capability:             capability,
		Toolset:                toolset,
		SelectionStore:         selectionStore,
		RecentToolCallRecorder: recentToolCalls,
		PromptRegistry:         d.NewPromptRegistry(),
		ResourceRegistry: d.NewResourceRegistry(client, resources.ResourceOptions{
			Version:               info.Version,
			TimezoneFallback:      info.Config.Timezone,
			DebugMetadata:         info.DebugMetadata,
			DeleteMode:            deleteMode,
			Toolset:               toolset,
			DisableAthleteProfile: info.Config.CoachModeEnabled(),
		}),
		Registry: d.NewToolRegistry(client, tools.RegistryOptions{
			Version:          info.Version,
			TimezoneFallback: info.Config.Timezone,
			DebugMetadata:    info.DebugMetadata,
			Capability:       capability,
			Toolset:          toolset,
			CoachModeEnabled: info.Config.CoachModeEnabled(),
			CoachConfig:      info.Config.Coach,
		}),
	})
	if err != nil {
		return nil, cleanup, err
	}
	return server, cleanup, nil
}

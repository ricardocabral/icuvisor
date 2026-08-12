package icuvisor

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	internaltools "github.com/ricardocabral/icuvisor/internal/tools"
)

func TestToolMiddlewareWrapsBaseExtraAndGeneratedToolsWithoutChangingMetadata(t *testing.T) {
	t.Parallel()

	wrapped := make(map[string]int)
	seenInfo := make(map[string]ToolInfo)
	middleware := ToolMiddleware(func(info ToolInfo, next Handler) Handler {
		wrapped[info.Name]++
		seenInfo[info.Name] = info
		return func(ctx context.Context, req ToolRequest) (ToolResult, error) {
			return next(ctx, req)
		}
	})
	extra := facadeExtraTool("Report hosted setup status.", ToolsetCore)
	registry := NewCoreRegistry(newFacadeTestClient(t), RegistryOptions{
		Version:        "v-public",
		DeleteMode:     DeleteModeFull,
		Toolset:        ToolsetFull,
		ExtraTools:     []Tool{extra},
		ToolMiddleware: middleware,
	})
	registrar := &collectingToolRegistrar{}
	if err := registry.inner.Register(context.Background(), registrar); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	tests := []struct {
		name string
		args json.RawMessage
	}{
		{name: "validate_workout", args: json.RawMessage(`{"description":"Easy aerobic run"}`)},
		{name: "hosted_setup_status", args: json.RawMessage(`{}`)},
		{name: "icuvisor_list_advanced_capabilities", args: json.RawMessage(`{}`)},
		{name: "icuvisor_check_server_version", args: json.RawMessage(`{}`)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tool := findInternalTool(t, registrar.tools, tc.name)
			if got := wrapped[tc.name]; got != 1 {
				t.Fatalf("middleware applications = %d, want 1", got)
			}
			if got, want := seenInfo[tc.name], toolInfoFromInternal(tool); !reflect.DeepEqual(got, want) {
				t.Fatalf("middleware ToolInfo = %#v, want registered metadata %#v", got, want)
			}
			if _, err := tool.Handler(context.Background(), internaltools.Request{Name: tc.name, Arguments: tc.args}); err != nil {
				t.Fatalf("wrapped Handler() error = %v", err)
			}
		})
	}
}

func TestNilToolMiddlewareLeavesHandlerBehaviorUnchanged(t *testing.T) {
	t.Parallel()

	extra := facadeExtraTool("Report hosted setup status.", ToolsetCore)
	registry := NewCoreRegistry(newFacadeTestClient(t), RegistryOptions{
		Version:        "v-public",
		DeleteMode:     DeleteModeFull,
		Toolset:        ToolsetFull,
		ExtraTools:     []Tool{extra},
		ToolMiddleware: nil,
	})
	registrar := &collectingToolRegistrar{}
	if err := registry.inner.Register(context.Background(), registrar); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	result, err := findInternalTool(t, registrar.tools, extra.Name).Handler(context.Background(), internaltools.Request{Name: extra.Name, Arguments: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if got := result.Content[0].Text; !strings.Contains(got, `"status":"ok"`) {
		t.Fatalf("Handler() text = %q, want original result", got)
	}
}

func TestToolMiddlewarePropagatesWrappedHandlerErrors(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("sentinel handler failure")
	extra := facadeExtraTool("Return a test error.", ToolsetCore)
	extra.Handler = func(context.Context, ToolRequest) (ToolResult, error) {
		return ToolResult{}, wantErr
	}
	registry := NewCoreRegistry(newFacadeTestClient(t), RegistryOptions{
		Version:    "v-public",
		Toolset:    ToolsetFull,
		ExtraTools: []Tool{extra},
		ToolMiddleware: func(_ ToolInfo, next Handler) Handler {
			return func(ctx context.Context, req ToolRequest) (ToolResult, error) {
				return next(ctx, req)
			}
		},
	})
	registrar := &collectingToolRegistrar{}
	if err := registry.inner.Register(context.Background(), registrar); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	_, err := findInternalTool(t, registrar.tools, extra.Name).Handler(context.Background(), internaltools.Request{Name: extra.Name, Arguments: json.RawMessage(`{}`)})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Handler() error = %v, want sentinel", err)
	}
}

func TestToolMiddlewareDenialDoesNotCallOriginalWriteHandler(t *testing.T) {
	t.Parallel()

	called := false
	extra := facadeExtraTool("Perform a hosted write.", ToolsetCore)
	extra.Name = "hosted_write"
	extra.Requirement = RequirementWrite
	extra.Handler = func(context.Context, ToolRequest) (ToolResult, error) {
		called = true
		return TextResult(map[string]any{"status": "written"}), nil
	}
	registry := NewCoreRegistry(newFacadeTestClient(t), RegistryOptions{
		Version:    "v-public",
		Toolset:    ToolsetFull,
		ExtraTools: []Tool{extra},
		ToolMiddleware: func(info ToolInfo, next Handler) Handler {
			if info.Name != extra.Name {
				return next
			}
			return func(context.Context, ToolRequest) (ToolResult, error) {
				return ToolResult{}, NewUserError("hosted write is disabled", nil)
			}
		},
	})
	registrar := &collectingToolRegistrar{}
	if err := registry.inner.Register(context.Background(), registrar); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	_, err := findInternalTool(t, registrar.tools, extra.Name).Handler(context.Background(), internaltools.Request{Name: extra.Name, Arguments: json.RawMessage(`{}`)})
	if err == nil || !strings.Contains(err.Error(), "hosted write is disabled") {
		t.Fatalf("Handler() error = %v, want middleware denial", err)
	}
	if called {
		t.Fatal("original write handler was called after middleware denial")
	}
}

func TestToolMiddlewareRunsOnlyForToolsThatPassCatalogFilter(t *testing.T) {
	t.Parallel()

	applications := make(map[string]int)
	registry := NewCoreRegistry(newFacadeTestClient(t), RegistryOptions{
		Version:    "v-public",
		DeleteMode: DeleteModeFull,
		Toolset:    ToolsetFull,
		ToolFilter: func(info ToolInfo) bool { return info.Name != "get_fitness" },
		ToolMiddleware: func(info ToolInfo, next Handler) Handler {
			applications[info.Name]++
			return next
		},
	})
	registrar := &collectingToolRegistrar{}
	if err := registry.inner.Register(context.Background(), registrar); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if applications["get_fitness"] != 0 {
		t.Fatalf("filtered tool middleware applications = %d, want 0", applications["get_fitness"])
	}
	for _, tool := range registrar.tools {
		if got := applications[tool.Name]; got != 1 {
			t.Fatalf("middleware applications for %s = %d, want 1", tool.Name, got)
		}
	}
}

func TestServerCatalogPolicyIsStableAcrossExecutionPolicyChanges(t *testing.T) {
	t.Parallel()

	cfg := facadeTestConfig()
	client := newFacadeTestClient(t)
	newServer := func(t *testing.T, executionMode DeleteMode, executionToolset Toolset, catalogMode DeleteMode, catalogToolset Toolset) *Server {
		t.Helper()
		registry := NewCoreRegistry(client, RegistryOptions{Version: "v-public", DeleteMode: executionMode, Toolset: executionToolset})
		server, err := NewServer(context.Background(), ServerOptions{
			Config:                     cfg,
			Version:                    "v-public",
			Registry:                   registry,
			DeleteMode:                 executionMode,
			Toolset:                    executionToolset,
			CatalogMode:                catalogMode,
			CatalogToolset:             catalogToolset,
			SkipRuntimeCatalogMetadata: true,
		})
		if err != nil {
			t.Fatalf("NewServer() error = %v", err)
		}
		return server
	}

	safeExecution := newServer(t, DeleteModeSafe, ToolsetCore, DeleteModeFull, ToolsetFull)
	fullExecution := newServer(t, DeleteModeFull, ToolsetFull, DeleteModeFull, ToolsetFull)
	if safeExecution.CatalogHash() != fullExecution.CatalogHash() {
		t.Fatalf("full catalog hashes differ by execution policy: safe/core=%q full/full=%q", safeExecution.CatalogHash(), fullExecution.CatalogHash())
	}
	defaultCatalog := newServer(t, DeleteModeSafe, ToolsetCore, "", "")
	if defaultCatalog.CatalogHash() == safeExecution.CatalogHash() {
		t.Fatalf("empty catalog policy did not retain safe/core execution defaults: hash=%q", defaultCatalog.CatalogHash())
	}
}

type collectingToolRegistrar struct {
	tools []internaltools.Tool
}

func (r *collectingToolRegistrar) AddTool(tool internaltools.Tool) error {
	r.tools = append(r.tools, tool)
	return nil
}

func findInternalTool(t *testing.T, catalog []internaltools.Tool, name string) internaltools.Tool {
	t.Helper()
	for _, tool := range catalog {
		if tool.Name == name {
			return tool
		}
	}
	t.Fatalf("catalog missing %s", name)
	return internaltools.Tool{}
}

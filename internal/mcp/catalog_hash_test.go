package mcp

import (
	"context"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/safety"
	"github.com/ricardocabral/icuvisor/internal/tools"
)

func TestHashToolCatalogGoldenFixture(t *testing.T) {
	t.Parallel()

	got, err := hashToolCatalog(catalogHashFixtureTools())
	if err != nil {
		t.Fatalf("hashToolCatalog() error = %v", err)
	}
	const want = "04fa73eccf1a365dd37653b8e5d7eac273e3b6fa37e427ad8d85583617fcf75b"
	if got != want {
		t.Fatalf("hashToolCatalog() = %s, want %s", got, want)
	}
}

func TestHashToolCatalogDeterministic(t *testing.T) {
	t.Parallel()

	base := catalogHashFixtureTools()
	baseHash, err := hashToolCatalog(base)
	if err != nil {
		t.Fatalf("hashToolCatalog(base) error = %v", err)
	}
	reordered := []tools.Tool{base[1], base[0]}
	for i := range 10 {
		got, err := hashToolCatalog(reordered)
		if err != nil {
			t.Fatalf("hashToolCatalog(reordered) run %d error = %v", i, err)
		}
		if got != baseHash {
			t.Fatalf("hashToolCatalog(reordered) run %d = %s, want %s", i, got, baseHash)
		}
	}
	remapped := []tools.Tool{
		catalogHashTestTool("catalog_alpha", "Alpha tool.", schemaWithOrderedKeys([]string{"description", "type"}), map[string]any{"type": "object"}, safety.ToolsetCore, tools.RequirementRead),
		catalogHashTestTool("catalog_beta", "Beta tool.", schemaWithOrderedKeys([]string{"type", "description"}), map[string]any{"type": "object"}, safety.ToolsetCore, tools.RequirementRead),
	}
	got, err := hashToolCatalog(remapped)
	if err != nil {
		t.Fatalf("hashToolCatalog(remapped) error = %v", err)
	}
	if got != baseHash {
		t.Fatalf("hashToolCatalog(remapped nested map order) = %s, want %s", got, baseHash)
	}
}

func TestHashToolCatalogSensitivity(t *testing.T) {
	t.Parallel()

	base := catalogHashFixtureTools()
	baseHash, err := hashToolCatalog(base)
	if err != nil {
		t.Fatalf("hashToolCatalog(base) error = %v", err)
	}
	tests := []struct {
		name  string
		tools []tools.Tool
	}{
		{name: "tool added", tools: append(cloneTools(base), catalogHashTestTool("catalog_gamma", "Gamma tool.", schemaWithOrderedKeys([]string{"type", "description"}), nil, safety.ToolsetCore, tools.RequirementRead))},
		{name: "tool removed", tools: cloneTools(base[:1])},
		{name: "argument renamed", tools: mutateTool(base, 0, func(tool *tools.Tool) { tool.InputSchema = schemaWithArgument("renamed", "alpha input") })},
		{name: "argument description edited", tools: mutateTool(base, 0, func(tool *tools.Tool) { tool.InputSchema = schemaWithArgument("value", "updated alpha input") })},
		{name: "tool description edited", tools: mutateTool(base, 0, func(tool *tools.Tool) { tool.Description = "Updated alpha tool." })},
		{name: "output schema edited", tools: mutateTool(base, 0, func(tool *tools.Tool) {
			tool.OutputSchema = map[string]any{"type": "object", "description": "updated output"}
		})},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := hashToolCatalog(tc.tools)
			if err != nil {
				t.Fatalf("hashToolCatalog() error = %v", err)
			}
			if got == baseHash {
				t.Fatalf("hashToolCatalog() = base hash %s, want changed hash", got)
			}
		})
	}
}

func TestHashToolCatalogRejectsNonMarshalableSchema(t *testing.T) {
	t.Parallel()

	_, err := hashToolCatalog([]tools.Tool{catalogHashTestTool("catalog_bad", "Bad schema.", map[string]any{"type": func() {}}, nil, safety.ToolsetCore, tools.RequirementRead)})
	if err == nil {
		t.Fatal("hashToolCatalog() error = nil, want marshal error")
	}
	if !strings.Contains(err.Error(), "marshalling input schema for catalog_bad") {
		t.Fatalf("hashToolCatalog() error = %q, want input schema context", err.Error())
	}
}

func TestNewServerCatalogHashUsesExposedCatalog(t *testing.T) {
	t.Parallel()

	coreRead := catalogHashTestTool("catalog_core", "Core tool.", schemaWithArgument("value", "core input"), map[string]any{"type": "object"}, safety.ToolsetCore, tools.RequirementRead)
	fullRead := catalogHashTestTool("catalog_full", "Full tool.", schemaWithArgument("value", "full input"), map[string]any{"type": "object"}, safety.ToolsetFull, tools.RequirementRead)
	coreDelete := catalogHashTestTool("catalog_delete", "Delete tool.", schemaWithArgument("value", "delete input"), map[string]any{"type": "object"}, safety.ToolsetCore, tools.RequirementDelete)
	registry := registryFunc(func(_ context.Context, registrar tools.Registrar) error {
		for _, tool := range []tools.Tool{coreRead, fullRead, coreDelete} {
			if err := registrar.AddTool(tool); err != nil {
				return err
			}
		}
		return nil
	})

	server, err := NewServer(context.Background(), Options{Registry: registry, Capability: safety.NewCapability(safety.ModeSafe), Toolset: safety.ToolsetCore})
	if err != nil {
		t.Fatalf("NewServer(core/safe) error = %v", err)
	}
	wantCore, err := hashToolCatalog([]tools.Tool{coreRead})
	if err != nil {
		t.Fatalf("hashToolCatalog(coreRead) error = %v", err)
	}
	if server.CatalogHash() != wantCore {
		t.Fatalf("NewServer(core/safe).CatalogHash() = %s, want %s", server.CatalogHash(), wantCore)
	}
	fullServer, err := NewServer(context.Background(), Options{Registry: registry, Capability: safety.NewCapability(safety.ModeFull), Toolset: safety.ToolsetFull})
	if err != nil {
		t.Fatalf("NewServer(full/full) error = %v", err)
	}
	wantFull, err := hashToolCatalog([]tools.Tool{coreRead, fullRead, coreDelete})
	if err != nil {
		t.Fatalf("hashToolCatalog(all tools) error = %v", err)
	}
	if fullServer.CatalogHash() != wantFull {
		t.Fatalf("NewServer(full/full).CatalogHash() = %s, want %s", fullServer.CatalogHash(), wantFull)
	}
	if fullServer.CatalogHash() == server.CatalogHash() {
		t.Fatalf("filtered catalog hash = full catalog hash %s, want filtering to affect hash", server.CatalogHash())
	}
}

func catalogHashFixtureTools() []tools.Tool {
	return []tools.Tool{
		catalogHashTestTool("catalog_alpha", "Alpha tool.", schemaWithOrderedKeys([]string{"type", "description"}), map[string]any{"type": "object"}, safety.ToolsetCore, tools.RequirementRead),
		catalogHashTestTool("catalog_beta", "Beta tool.", schemaWithOrderedKeys([]string{"description", "type"}), map[string]any{"type": "object"}, safety.ToolsetCore, tools.RequirementRead),
	}
}

func catalogHashTestTool(name, description string, inputSchema, outputSchema any, toolset safety.Toolset, requirement tools.Requirement) tools.Tool {
	return tools.Tool{
		Name:         name,
		Description:  description,
		InputSchema:  inputSchema,
		OutputSchema: outputSchema,
		Toolset:      toolset,
		Requirement:  requirement,
		Handler: func(context.Context, tools.Request) (tools.Result, error) {
			return tools.Result{}, nil
		},
	}
}

func schemaWithOrderedKeys(order []string) map[string]any {
	schema := make(map[string]any, len(order)+2)
	for _, key := range order {
		switch key {
		case "type":
			schema[key] = "object"
		case "description":
			schema[key] = "fixture schema"
		}
	}
	schema["properties"] = map[string]any{
		"value": map[string]any{
			"type":        "string",
			"description": "alpha input",
		},
	}
	schema["additionalProperties"] = false
	return schema
}

func schemaWithArgument(name, description string) map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			name: map[string]any{
				"type":        "string",
				"description": description,
			},
		},
	}
}

func cloneTools(in []tools.Tool) []tools.Tool {
	out := make([]tools.Tool, len(in))
	copy(out, in)
	return out
}

func mutateTool(in []tools.Tool, index int, mutate func(*tools.Tool)) []tools.Tool {
	out := cloneTools(in)
	mutate(&out[index])
	return out
}

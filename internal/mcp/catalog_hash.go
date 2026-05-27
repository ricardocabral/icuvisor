package mcp

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"

	"github.com/ricardocabral/icuvisor/internal/coach"
	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/safety"
	"github.com/ricardocabral/icuvisor/internal/tools"
)

type catalogHashTool struct {
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	InputSchema  json.RawMessage `json:"input_schema"`
	OutputSchema json.RawMessage `json:"output_schema,omitempty"`
}

// CatalogHashOptions configures deterministic active tool-catalog hash calculation.
type CatalogHashOptions struct {
	Config     config.Config
	Registry   tools.Registry
	Capability safety.Capability
	Toolset    safety.Toolset
	Logger     *slog.Logger
}

// ComputeToolCatalogHash returns the same tool-catalog hash NewServer exposes without starting a transport.
func ComputeToolCatalogHash(ctx context.Context, opts CatalogHashOptions) (string, error) {
	toolCatalog, err := CollectToolCatalog(ctx, opts)
	if err != nil {
		return "", err
	}
	catalogHash, err := hashToolCatalog(toolCatalog)
	if err != nil {
		return "", fmt.Errorf("hashing tool catalog: %w", err)
	}
	return catalogHash, nil
}

// CollectToolCatalog returns the exposed MCP tool definitions without starting a transport.
func CollectToolCatalog(ctx context.Context, opts CatalogHashOptions) ([]tools.Tool, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if opts.Registry == nil {
		return nil, nil
	}
	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}
	coachEvaluator := coach.NewEvaluator(opts.Config.CoachModeEnabled(), opts.Config.Coach)
	registrar := &safeRegistrar{
		logger:         logger,
		config:         opts.Config,
		coachFilter:    coach.NewToolFilter(coachEvaluator),
		selectionStore: coach.NewSelectionStore(opts.Config.Coach.DefaultAthleteID),
		capability:     capabilityOrSafe(opts.Capability),
		toolset:        safety.ParseToolset(opts.Toolset.String()),
		names:          make(map[string]struct{}),
	}
	if err := opts.Registry.Register(ctx, registrar); err != nil {
		return nil, fmt.Errorf("registering tools for catalog collection: %w", err)
	}
	return append([]tools.Tool(nil), registrar.registeredTools...), nil
}

func hashToolCatalog(toolCatalog []tools.Tool) (string, error) {
	records := make([]catalogHashTool, 0, len(toolCatalog))
	for _, tool := range toolCatalog {
		inputSchema, err := marshalCatalogSchema(tool.Name, "input", tool.InputSchema)
		if err != nil {
			return "", err
		}
		outputSchema, err := marshalCatalogSchema(tool.Name, "output", tool.OutputSchema)
		if err != nil {
			return "", err
		}
		records = append(records, catalogHashTool{
			Name:         tool.Name,
			Description:  tool.Description,
			InputSchema:  inputSchema,
			OutputSchema: outputSchema,
		})
	}
	sort.Slice(records, func(i, j int) bool {
		return records[i].Name < records[j].Name
	})
	payload, err := json.Marshal(records)
	if err != nil {
		return "", fmt.Errorf("marshalling catalog hash records: %w", err)
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:]), nil
}

func marshalCatalogSchema(toolName, schemaName string, schema any) (json.RawMessage, error) {
	if schema == nil {
		return nil, nil
	}
	payload, err := json.Marshal(schema)
	if err != nil {
		return nil, fmt.Errorf("marshalling %s schema for %s: %w", schemaName, toolName, err)
	}
	return json.RawMessage(payload), nil
}

// CatalogHash reports the deterministic hash of the exposed tool catalog.
func (s *Server) CatalogHash() string {
	if s == nil {
		return ""
	}
	return s.catalogHash
}

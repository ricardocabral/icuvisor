package mcp

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/ricardocabral/icuvisor/internal/tools"
)

type catalogHashTool struct {
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	InputSchema  json.RawMessage `json:"input_schema"`
	OutputSchema json.RawMessage `json:"output_schema,omitempty"`
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

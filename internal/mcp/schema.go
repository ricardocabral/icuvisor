package mcp

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ricardocabral/icuvisor/internal/safety"
	"github.com/ricardocabral/icuvisor/internal/tools"
)

const athleteIDArgumentDescription = "Target athlete; defaults to selected athlete in coach mode, or the only athlete otherwise. Format: i12345 or 12345."

var snakeCaseToolName = regexp.MustCompile(`^[a-z][a-z0-9]*(?:_[a-z0-9]+)*$`)

// SchemaWithAthleteID returns schema with the coach-mode athlete_id selector added.
func SchemaWithAthleteID(schema any) any {
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

func stripAthleteID(raw json.RawMessage) (json.RawMessage, string, bool, error) {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return raw, "", false, nil
	}
	if !strings.HasPrefix(trimmed, "{") {
		return nil, "", false, errors.New("arguments must be an object")
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, "", false, err
	}
	athleteRaw, ok := fields["athlete_id"]
	if !ok {
		return raw, "", false, nil
	}
	delete(fields, "athlete_id")
	var athleteID string
	if len(athleteRaw) > 0 && strings.TrimSpace(string(athleteRaw)) != "null" {
		if err := json.Unmarshal(athleteRaw, &athleteID); err != nil {
			return nil, "", true, err
		}
	}
	cleaned, err := json.Marshal(fields)
	if err != nil {
		return nil, "", true, err
	}
	return cleaned, athleteID, true, nil
}

func validateToolset(tool tools.Tool) error {
	switch tool.Toolset {
	case "", safety.ToolsetCore, safety.ToolsetFull:
		return nil
	default:
		return fmt.Errorf("tool %q has invalid toolset %q", tool.Name, tool.Toolset)
	}
}

func validateRequirement(tool tools.Tool) error {
	switch tool.Requirement {
	case "", tools.RequirementRead, tools.RequirementWrite, tools.RequirementDelete:
		return nil
	default:
		return fmt.Errorf("tool %q has invalid requirement %q", tool.Name, tool.Requirement)
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

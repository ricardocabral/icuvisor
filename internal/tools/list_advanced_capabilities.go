package tools

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/ricardocabral/icuvisor/internal/safety"
)

const (
	listAdvancedCapabilitiesName        = "icuvisor_list_advanced_capabilities"
	listAdvancedCapabilitiesDescription = "Use only when the user asks what other icuvisor tools are available or when a requested advanced/delete capability is not visible in the current tool catalog. If a specific visible tool can satisfy the request, call that tool instead. Explains server env gates such as ICUVISOR_TOOLSET=full and ICUVISOR_DELETE_MODE=full; the LLM cannot bypass them at call time. This tool makes no intervals.icu API calls."
	listAdvancedCapabilitiesInstruction = "Set ICUVISOR_TOOLSET=full in the MCP client/server environment to expose the full icuvisor toolset, and ICUVISOR_DELETE_MODE=full to additionally enable delete tools (default safe registers writes only; none disables both). restart icuvisor after changing either."
)

type advancedCapabilityRow struct {
	Name        string `json:"name"`
	Summary     string `json:"summary"`
	Requirement string `json:"requirement"`
}

type listAdvancedCapabilitiesResponse struct {
	CurrentToolset       string                   `json:"current_toolset"`
	Status               string                   `json:"status"`
	EnableInstruction    string                   `json:"enable_instruction"`
	AdvancedCapabilities []advancedCapabilityRow  `json:"advanced_capabilities"`
	Meta                 advancedCapabilitiesMeta `json:"_meta"`
}

type advancedCapabilitiesMeta struct {
	Count          int    `json:"count"`
	Source         string `json:"source"`
	DeleteModeNote string `json:"delete_mode_note"`
	Toolset        string `json:"toolset"`
}

func newListAdvancedCapabilitiesTool(catalog []Tool, activeToolset safety.Toolset, shaping ...responseShaping) Tool {
	capabilities := fullOnlyCapabilities(catalog)
	shapeCfg := responseShapingOrDefault(shaping)
	return coreTool(Tool{
		Name:         listAdvancedCapabilitiesName,
		Description:  listAdvancedCapabilitiesDescription,
		InputSchema:  listAdvancedCapabilitiesInputSchema(),
		OutputSchema: genericOutputSchema("Full-only icuvisor tools and instructions for enabling them."),
		Requirement:  RequirementRead,
		Handler:      listAdvancedCapabilitiesHandler(capabilities, activeToolset, shapeCfg),
	})
}

func listAdvancedCapabilitiesInputSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties":           map[string]any{},
	}
}

func fullOnlyCapabilities(catalog []Tool) []advancedCapabilityRow {
	return filteredFullOnlyCapabilities(catalog, nil)
}

func filteredFullOnlyCapabilities(catalog []Tool, include func(Tool) bool) []advancedCapabilityRow {
	rows := make([]advancedCapabilityRow, 0, len(catalog))
	for _, tool := range catalog {
		if tool.Name == listAdvancedCapabilitiesName || tool.EffectiveToolset() != safety.ToolsetFull {
			continue
		}
		if include != nil && !include(tool) {
			continue
		}
		rows = append(rows, advancedCapabilityRow{Name: tool.Name, Summary: firstDescriptionSentence(tool.Description), Requirement: toolRequirement(tool)})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Name < rows[j].Name })
	return rows
}

func listAdvancedCapabilitiesHandler(capabilities []advancedCapabilityRow, activeToolset safety.Toolset, shapeCfg responseShaping) Handler {
	toolset := safety.ParseToolset(string(activeToolset))
	return func(ctx context.Context, req Request) (Result, error) {
		if err := ctx.Err(); err != nil {
			return Result{}, err
		}
		if err := validateListAdvancedCapabilitiesArguments(req.Arguments); err != nil {
			return Result{}, err
		}
		return encodeAdvancedCapabilitiesResult(capabilities, toolset, shapeCfg)
	}
}

// NewFilteredAdvancedCapabilitiesHandler builds an advanced-capabilities handler with per-call catalog filtering.
func NewFilteredAdvancedCapabilitiesHandler(catalog []Tool, activeToolset safety.Toolset, include func(context.Context, Tool) bool) Handler {
	catalog = append([]Tool(nil), catalog...)
	toolset := safety.ParseToolset(string(activeToolset))
	return func(ctx context.Context, req Request) (Result, error) {
		if err := ctx.Err(); err != nil {
			return Result{}, err
		}
		if err := validateListAdvancedCapabilitiesArguments(req.Arguments); err != nil {
			return Result{}, err
		}
		capabilities := filteredFullOnlyCapabilities(catalog, func(tool Tool) bool {
			return include == nil || include(ctx, tool)
		})
		return filteredAdvancedCapabilitiesResult(capabilities, toolset), nil
	}
}

func validateListAdvancedCapabilitiesArguments(arguments json.RawMessage) error {
	if strings.TrimSpace(string(arguments)) != "" && strings.TrimSpace(string(arguments)) != "{}" && strings.TrimSpace(string(arguments)) != "null" {
		return NewUserError("invalid icuvisor_list_advanced_capabilities arguments; no arguments are supported", nil)
	}
	return nil
}

func filteredAdvancedCapabilitiesResult(capabilities []advancedCapabilityRow, toolset safety.Toolset) Result {
	rows := make([]map[string]any, 0, len(capabilities))
	for _, capability := range capabilities {
		rows = append(rows, map[string]any{"name": capability.Name, "summary": capability.Summary, "requirement": capability.Requirement})
	}
	status := "The default core toolset is active; full-only tools are hidden from tools/list."
	if toolset == safety.ToolsetFull {
		status = "The full toolset is already enabled; these full-only tools should already be visible when delete-mode also allows them."
	}
	return TextResult(map[string]any{
		"current_toolset":       toolset.String(),
		"status":                status,
		"enable_instruction":    "Set ICUVISOR_TOOLSET=full in the MCP client/server environment to expose the full icuvisor toolset, and ICUVISOR_DELETE_MODE=full to additionally enable delete tools. restart icuvisor after changing either.",
		"advanced_capabilities": rows,
		"_meta": map[string]any{
			"count":            len(rows),
			"source":           "registered catalog metadata",
			"delete_mode_note": "Tools with requirement=delete also require ICUVISOR_DELETE_MODE=full; write tools require delete mode safe or full.",
			"toolset":          toolset.String(),
		},
	})
}

func encodeAdvancedCapabilitiesResult(capabilities []advancedCapabilityRow, toolset safety.Toolset, shapeCfg responseShaping) (Result, error) {
	status := "The default core toolset is active; full-only tools are hidden from tools/list."
	if toolset == safety.ToolsetFull {
		status = "The full toolset is already enabled; these full-only tools should already be visible when delete-mode also allows them."
	}
	response := listAdvancedCapabilitiesResponse{
		CurrentToolset:       toolset.String(),
		Status:               status,
		EnableInstruction:    listAdvancedCapabilitiesInstruction,
		AdvancedCapabilities: append([]advancedCapabilityRow(nil), capabilities...),
		Meta: advancedCapabilitiesMeta{
			Count:          len(capabilities),
			Source:         "registered catalog metadata",
			DeleteModeNote: "Tools with requirement=delete also require ICUVISOR_DELETE_MODE=full; write tools require delete mode safe or full.",
			Toolset:        toolset.String(),
		},
	}
	return encodeShaped(response, false, nil, "", false, listAdvancedCapabilitiesName, "", shapeCfg)
}

func firstDescriptionSentence(description string) string {
	description = strings.Join(strings.Fields(description), " ")
	for index, r := range description {
		if r != '.' {
			continue
		}
		if index == len(description)-1 || nextSentenceStartsUpper(description[index+1:]) {
			return strings.TrimSpace(description[:index+1])
		}
	}
	return description
}

func nextSentenceStartsUpper(value string) bool {
	value = strings.TrimLeftFunc(value, unicode.IsSpace)
	if value == "" {
		return true
	}
	r, _ := utf8.DecodeRuneInString(value)
	return unicode.IsUpper(r)
}

func toolRequirement(tool Tool) string {
	return string(tool.Requirement.effective())
}

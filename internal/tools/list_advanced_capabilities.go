package tools

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/ricardocabral/icuvisor/internal/response"
	"github.com/ricardocabral/icuvisor/internal/safety"
)

const (
	listAdvancedCapabilitiesName        = "icuvisor_list_advanced_capabilities"
	listAdvancedCapabilitiesDescription = "Discover tools hidden from the default core catalog and explain how to enable the full icuvisor toolset. This tool makes no intervals.icu API calls."
	listAdvancedCapabilitiesInstruction = "Set ICUVISOR_TOOLSET=full in the MCP client/server environment and restart icuvisor to enable the full tool catalog."
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

func newListAdvancedCapabilitiesTool(catalog []Tool, activeToolset safety.Toolset) Tool {
	capabilities := fullOnlyCapabilities(catalog)
	return coreTool(Tool{
		Name:         listAdvancedCapabilitiesName,
		Description:  listAdvancedCapabilitiesDescription,
		InputSchema:  listAdvancedCapabilitiesInputSchema(),
		OutputSchema: genericOutputSchema("Full-only icuvisor tools and instructions for enabling them."),
		Requirement:  RequirementRead,
		Handler:      listAdvancedCapabilitiesHandler(capabilities, activeToolset),
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
	rows := make([]advancedCapabilityRow, 0, len(catalog))
	for _, tool := range catalog {
		if tool.Name == listAdvancedCapabilitiesName || tool.EffectiveToolset() != safety.ToolsetFull {
			continue
		}
		rows = append(rows, advancedCapabilityRow{Name: tool.Name, Summary: firstDescriptionSentence(tool.Description), Requirement: toolRequirement(tool)})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Name < rows[j].Name })
	return rows
}

func listAdvancedCapabilitiesHandler(capabilities []advancedCapabilityRow, activeToolset safety.Toolset) Handler {
	toolset := safety.ParseToolset(string(activeToolset))
	return func(ctx context.Context, req Request) (Result, error) {
		if err := ctx.Err(); err != nil {
			return Result{}, err
		}
		if strings.TrimSpace(string(req.Arguments)) != "" && strings.TrimSpace(string(req.Arguments)) != "{}" && strings.TrimSpace(string(req.Arguments)) != "null" {
			return Result{}, NewUserError("invalid icuvisor_list_advanced_capabilities arguments; no arguments are supported", nil)
		}
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
				Toolset:        response.Toolset(),
			},
		}
		if _, err := json.Marshal(response); err != nil {
			return Result{}, err
		}
		return TextResult(response), nil
	}
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

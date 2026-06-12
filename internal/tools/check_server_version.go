package tools

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/response"
	"github.com/ricardocabral/icuvisor/internal/safety"
)

const (
	checkServerVersionName                          = "icuvisor_check_server_version"
	checkServerVersionStatus                        = "compare_visible_description"
	checkServerVersionFingerprintField              = "description_catalog_fingerprint="
	checkServerVersionFingerprintSentinel           = "<description_catalog_fingerprint>"
	checkServerVersionFingerprintPlaceholder        = "pending-description-catalog-fingerprint"
	checkServerVersionDescriptionFingerprintScope   = "catalog-mode tool records after delete-mode/toolset gates; dynamic coach per-athlete ACL visibility is excluded"
	checkServerVersionNoNetworkSource               = "runtime catalog metadata and registered tool descriptions"
	checkServerVersionCompareVisibleDescriptionText = "Compare the visible tool description fields with this response. If description_server_version, description_catalog_fingerprint, description_toolset, or description_delete_mode differ, reconnect the MCP client or start a new conversation; if _meta.schema_changed is visible, start a new conversation."
)

type checkServerVersionResponse struct {
	ServerVersion                 string                 `json:"server_version"`
	CatalogHash                   string                 `json:"catalog_hash"`
	DescriptionServerVersion      string                 `json:"description_server_version"`
	DescriptionCatalogFingerprint string                 `json:"description_catalog_fingerprint"`
	Toolset                       string                 `json:"toolset"`
	DeleteMode                    string                 `json:"delete_mode"`
	Status                        string                 `json:"status"`
	Action                        string                 `json:"action"`
	Meta                          checkServerVersionMeta `json:"_meta"`
}

type checkServerVersionMeta struct {
	Source           string `json:"source"`
	FingerprintScope string `json:"fingerprint_scope"`
	NoNetwork        bool   `json:"no_network"`
	Toolset          string `json:"toolset"`
	DeleteMode       string `json:"delete_mode"`
}

type catalogFingerprintTool struct {
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	InputSchema  json.RawMessage `json:"input_schema"`
	OutputSchema json.RawMessage `json:"output_schema,omitempty"`
}

func newCheckServerVersionTool(version string, catalog []Tool, deleteMode safety.Mode, toolset safety.Toolset, shaping ...responseShaping) (Tool, error) {
	version = normalizeVersion(version)
	deleteMode = safety.ParseMode(deleteMode.String())
	toolset = safety.ParseToolset(toolset.String())
	shapeCfg := responseShapingOrDefault(shaping)

	placeholder := checkServerVersionTool(version, checkServerVersionFingerprintPlaceholder, deleteMode, toolset, shapeCfg)
	fingerprintCatalog := append([]Tool(nil), catalog...)
	fingerprintCatalog = append(fingerprintCatalog, placeholder)
	fingerprint, err := descriptionCatalogFingerprint(fingerprintCatalog)
	if err != nil {
		return Tool{}, err
	}
	return checkServerVersionTool(version, fingerprint, deleteMode, toolset, shapeCfg), nil
}

func checkServerVersionTool(version string, descriptionFingerprint string, deleteMode safety.Mode, toolset safety.Toolset, shapeCfg responseShaping) Tool {
	version = normalizeVersion(version)
	deleteMode = safety.ParseMode(deleteMode.String())
	toolset = safety.ParseToolset(toolset.String())
	return coreTool(Tool{
		Name:         checkServerVersionName,
		Description:  checkServerVersionDescription(version, descriptionFingerprint, deleteMode, toolset),
		InputSchema:  noArgsSchema(),
		OutputSchema: genericOutputSchema("Server version, live catalog hash, and visible description fingerprint for stale MCP catalog diagnosis."),
		Requirement:  RequirementRead,
		Handler:      checkServerVersionHandler(version, descriptionFingerprint, deleteMode, toolset, shapeCfg),
	})
}

func checkServerVersionDescription(version string, descriptionFingerprint string, deleteMode safety.Mode, toolset safety.Toolset) string {
	return fmt.Sprintf("Check whether the MCP client is using stale icuvisor tool descriptions after an upgrade. Visible baseline: description_server_version=%s; description_catalog_fingerprint=%s; description_toolset=%s; description_delete_mode=%s. Call with no arguments and compare these visible description fields with the response fields; if they differ, reconnect the MCP client or start a new conversation. This tool makes no intervals.icu API calls and returns no athlete data, API keys, filesystem paths, usernames, or raw environment values.", normalizeVersion(version), strings.TrimSpace(descriptionFingerprint), safety.ParseToolset(toolset.String()).String(), safety.ParseMode(deleteMode.String()).String())
}

func checkServerVersionHandler(descriptionVersion string, descriptionFingerprint string, deleteMode safety.Mode, toolset safety.Toolset, shapeCfg responseShaping) Handler {
	descriptionVersion = normalizeVersion(descriptionVersion)
	descriptionFingerprint = strings.TrimSpace(descriptionFingerprint)
	deleteMode = safety.ParseMode(deleteMode.String())
	toolset = safety.ParseToolset(toolset.String())
	return func(ctx context.Context, req Request) (Result, error) {
		if err := ctx.Err(); err != nil {
			return Result{}, err
		}
		if err := noArgs(req.Arguments, checkServerVersionName); err != nil {
			return Result{}, err
		}
		runtime := response.RuntimeCatalogMetadata()
		if strings.TrimSpace(shapeCfg.catalogHash) != "" {
			runtime = response.RuntimeCatalogMetadataSnapshot{Version: descriptionVersion, CatalogHash: shapeCfg.catalogHash}
		}
		payload := checkServerVersionResponse{
			ServerVersion:                 runtime.Version,
			CatalogHash:                   runtime.CatalogHash,
			DescriptionServerVersion:      descriptionVersion,
			DescriptionCatalogFingerprint: descriptionFingerprint,
			Toolset:                       toolset.String(),
			DeleteMode:                    deleteMode.String(),
			Status:                        checkServerVersionStatus,
			Action:                        checkServerVersionCompareVisibleDescriptionText,
			Meta: checkServerVersionMeta{
				Source:           checkServerVersionNoNetworkSource,
				FingerprintScope: checkServerVersionDescriptionFingerprintScope,
				NoNetwork:        true,
				Toolset:          toolset.String(),
				DeleteMode:       deleteMode.String(),
			},
		}
		return encodeShaped(payload, false, nil, runtime.Version, false, checkServerVersionName, "", shapeCfg)
	}
}

func effectiveDiagnosticCatalog(catalog []Tool, capability safety.Capability, toolset safety.Toolset) []Tool {
	capability = capabilityOrSafe(capability)
	toolset = safety.ParseToolset(toolset.String())
	out := make([]Tool, 0, len(catalog))
	for _, tool := range catalog {
		if !diagnosticCapabilityAllows(tool, capability) || !diagnosticToolsetAllows(tool, toolset) {
			continue
		}
		out = append(out, tool)
	}
	return out
}

func diagnosticCapabilityAllows(tool Tool, capability safety.Capability) bool {
	switch tool.Requirement.effective() {
	case RequirementDelete:
		return capability.CanDelete()
	case RequirementWrite:
		return capability.CanWrite()
	default:
		return true
	}
}

func diagnosticToolsetAllows(tool Tool, active safety.Toolset) bool {
	if active == safety.ToolsetFull {
		return true
	}
	return tool.EffectiveToolset() == safety.ToolsetCore
}

func descriptionCatalogFingerprint(catalog []Tool) (string, error) {
	records := make([]catalogFingerprintTool, 0, len(catalog))
	for _, tool := range catalog {
		inputSchema, err := marshalCatalogFingerprintSchema(tool.Name, "input", tool.InputSchema)
		if err != nil {
			return "", err
		}
		outputSchema, err := marshalCatalogFingerprintSchema(tool.Name, "output", tool.OutputSchema)
		if err != nil {
			return "", err
		}
		description := tool.Description
		if tool.Name == checkServerVersionName {
			description = normalizeCheckServerVersionFingerprint(description)
		}
		records = append(records, catalogFingerprintTool{Name: tool.Name, Description: description, InputSchema: inputSchema, OutputSchema: outputSchema})
	}
	sort.Slice(records, func(i, j int) bool { return records[i].Name < records[j].Name })
	payload, err := json.Marshal(records)
	if err != nil {
		return "", fmt.Errorf("marshalling description catalog fingerprint records: %w", err)
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:]), nil
}

func normalizeCheckServerVersionFingerprint(description string) string {
	start := strings.Index(description, checkServerVersionFingerprintField)
	if start < 0 {
		return description
	}
	valueStart := start + len(checkServerVersionFingerprintField)
	valueEnd := valueStart + strings.Index(description[valueStart:], ";")
	if valueEnd < valueStart {
		valueEnd = len(description)
	}
	return description[:valueStart] + checkServerVersionFingerprintSentinel + description[valueEnd:]
}

func marshalCatalogFingerprintSchema(toolName, schemaName string, schema any) (json.RawMessage, error) {
	if schema == nil {
		return nil, nil
	}
	payload, err := json.Marshal(schema)
	if err != nil {
		return nil, fmt.Errorf("marshalling %s schema for %s description fingerprint: %w", schemaName, toolName, err)
	}
	return json.RawMessage(payload), nil
}

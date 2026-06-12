package response

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

const defaultCatalogHash = "dev-catalog-hash"

var catalogRuntime = struct {
	sync.Mutex
	current   catalogSnapshot
	firstSeen *catalogSnapshot
}{current: catalogSnapshot{CatalogHash: defaultCatalogHash}}

// RuntimeCatalogMetadataSnapshot is the current non-secret server/catalog metadata.
type RuntimeCatalogMetadataSnapshot struct {
	Version     string
	CatalogHash string
}

type catalogSnapshot struct {
	Version     string
	CatalogHash string
}

var responseOwnedMetaKeys = map[string]struct{}{
	"server_version":        {},
	"catalog_hash":          {},
	"schema_changed":        {},
	"schema_change_message": {},
	"previous_version":      {},
	"current_version":       {},
	"previous_catalog_hash": {},
	"delete_mode":           {},
	"toolset":               {},
	"units":                 {},
}

// SetRuntimeCatalogMetadata stores the process-global catalog metadata reported in response metadata.
func SetRuntimeCatalogMetadata(version string, catalogHash string) {
	catalogRuntime.Lock()
	defer catalogRuntime.Unlock()
	catalogRuntime.current = catalogSnapshot{Version: normalizeVersion(version), CatalogHash: normalizeCatalogHash(catalogHash)}
}

// RuntimeCatalogMetadata returns the current non-secret server/catalog metadata.
func RuntimeCatalogMetadata() RuntimeCatalogMetadataSnapshot {
	catalogRuntime.Lock()
	defer catalogRuntime.Unlock()
	current := catalogRuntime.current
	return RuntimeCatalogMetadataSnapshot{Version: normalizeVersion(current.Version), CatalogHash: normalizeCatalogHash(current.CatalogHash)}
}

func resetRuntimeCatalogMetadataForTest() {
	catalogRuntime.Lock()
	defer catalogRuntime.Unlock()
	catalogRuntime.current = catalogSnapshot{CatalogHash: defaultCatalogHash}
	catalogRuntime.firstSeen = nil
}

func setRuntimeCatalogMetadataForTest(version string, catalogHash string) {
	SetRuntimeCatalogMetadata(version, catalogHash)
}

func addDebugMetadata(row map[string]any, opts Options) {
	if queryType := strings.TrimSpace(opts.QueryType); queryType != "" {
		row["query_type"] = queryType
	}
	fetchedAt := opts.FetchedAt
	if fetchedAt.IsZero() {
		fetchedAt = time.Now()
	}
	row["fetched_at"] = fetchedAt.UTC().Format(time.RFC3339)
}

func addStripMeta(row map[string]any, missing []string) {
	meta := map[string]any{}
	if existing, ok := row["_meta"].(map[string]any); ok {
		for key, value := range existing {
			meta[key] = value
		}
	}
	meta["fields_present"] = presentFields(row)
	meta["missing_fields"] = sortedStrings(missing)
	row["_meta"] = meta
}

func addScaleMeta(row map[string]any) {
	scales := scalesForRow(row)
	meta := map[string]any{}
	if existing, ok := row["_meta"].(map[string]any); ok {
		for key, value := range existing {
			if key != "scales" && key != "units" {
				meta[key] = value
			}
		}
	} else if len(scales) == 0 {
		return
	}
	if len(scales) > 0 {
		meta["scales"] = scales
	}
	if len(meta) == 0 {
		delete(row, "_meta")
		return
	}
	row["_meta"] = meta
}

func scalesForRow(row map[string]any) map[string]string {
	scales := map[string]string{}
	collectScaleLabels(row, scales)
	return scales
}

func collectScaleLabels(value any, scales map[string]string) {
	walkJSON(value, "", walkRoot, func(path string, item any, _ jsonWalkContainer) jsonWalkDecision {
		if isMetaPath(path) {
			return jsonWalkDecision{Stop: true}
		}
		row, ok := item.(map[string]any)
		if !ok {
			return jsonWalkDecision{}
		}
		for key, field := range row {
			if label, ok := defaultScaleLabels[key]; ok && field != nil {
				scales[key] = label
			}
		}
		return jsonWalkDecision{}
	})
}

func addCommonMeta(row map[string]any, opts Options) {
	meta := map[string]any{}
	if existing, ok := row["_meta"].(map[string]any); ok {
		for key, value := range existing {
			if _, owned := responseOwnedMetaKeys[key]; !owned {
				meta[key] = value
			}
		}
	}
	serverVersion := normalizeVersion(opts.ServerVersion)
	meta["server_version"] = serverVersion
	for key, value := range schemaCatalogMetaForOptions(serverVersion, opts.CatalogHash) {
		meta[key] = value
	}
	meta["delete_mode"] = opts.DeleteMode.String()
	meta["toolset"] = opts.Toolset.String()
	if opts.UnitSystem != "" {
		meta["units"] = opts.UnitSystem.Metadata()
	}
	row["_meta"] = meta
}

func schemaCatalogMetaForOptions(serverVersion string, catalogHash string) map[string]any {
	if strings.TrimSpace(catalogHash) != "" {
		return map[string]any{"catalog_hash": normalizeCatalogHash(catalogHash)}
	}
	return schemaCatalogMeta(serverVersion)
}

func schemaCatalogMeta(serverVersion string) map[string]any {
	catalogRuntime.Lock()
	defer catalogRuntime.Unlock()
	current := catalogRuntime.current
	if current.Version == "" {
		current.Version = normalizeVersion(serverVersion)
	}
	current.CatalogHash = normalizeCatalogHash(current.CatalogHash)
	meta := map[string]any{"catalog_hash": current.CatalogHash}
	if catalogRuntime.firstSeen == nil {
		firstSeen := current
		catalogRuntime.firstSeen = &firstSeen
		return meta
	}
	firstSeen := *catalogRuntime.firstSeen
	firstSeen.CatalogHash = normalizeCatalogHash(firstSeen.CatalogHash)
	if firstSeen.CatalogHash != current.CatalogHash {
		meta["schema_changed"] = true
		meta["schema_change_message"] = schemaChangeMessage(firstSeen.Version, current.Version)
		meta["previous_version"] = firstSeen.Version
		meta["current_version"] = current.Version
		meta["previous_catalog_hash"] = firstSeen.CatalogHash
	}
	return meta
}

func schemaChangeMessage(previousVersion, currentVersion string) string {
	return fmt.Sprintf("icuvisor was upgraded from %s to %s since this conversation started; tool schemas may have changed. Open a new conversation to use the latest tools.", normalizeVersion(previousVersion), normalizeVersion(currentVersion))
}

func filterDebugMissing(missing []string) []string {
	filtered := make([]string, 0, len(missing))
	for _, path := range missing {
		if !isDebugPath(path) {
			filtered = append(filtered, path)
		}
	}
	return filtered
}

func isDebugPath(path string) bool {
	if isProvenanceFetchedAtPath(path) {
		return false
	}
	for _, part := range strings.Split(path, ".") {
		if part == "fetched_at" || part == "query_type" || strings.HasPrefix(part, "fetched_at[") || strings.HasPrefix(part, "query_type[") {
			return true
		}
	}
	return false
}

func dropDebugMetadata(value any, path string) any {
	dropped, _, _ := walkJSON(value, path, walkRoot, dropDebugVisitor)
	return dropped
}

func dropDebugVisitor(path string, _ any, _ jsonWalkContainer) jsonWalkDecision {
	if path != "" && isDebugPath(path) && !isProvenancePath(path) {
		return jsonWalkDecision{Drop: true}
	}
	return jsonWalkDecision{}
}

func isProvenancePath(path string) bool {
	return path == "_meta.provenance" || strings.HasPrefix(path, "_meta.provenance.") || strings.Contains(path, "._meta.provenance")
}

func isProvenanceFetchedAtPath(path string) bool {
	return strings.Contains(path, "_meta.provenance.") && strings.HasSuffix(path, ".fetched_at")
}

func normalizeVersion(version string) string {
	version = strings.TrimSpace(version)
	if version == "" {
		return "dev"
	}
	return version
}

func normalizeCatalogHash(catalogHash string) string {
	catalogHash = strings.TrimSpace(catalogHash)
	if catalogHash == "" {
		return defaultCatalogHash
	}
	return catalogHash
}

func isMetaPath(path string) bool {
	for _, part := range strings.Split(path, ".") {
		if part == "_meta" || strings.HasPrefix(part, "_meta[") {
			return true
		}
	}
	return false
}

func presentFields(row map[string]any) []string {
	fields := make([]string, 0, len(row))
	for key := range row {
		if key != "_meta" {
			fields = append(fields, key)
		}
	}
	return sortedStrings(fields)
}

func sortedStrings(values []string) []string {
	out := append([]string(nil), values...)
	sort.Strings(out)
	return out
}

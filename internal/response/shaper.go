package response

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"github.com/ricardocabral/icuvisor/internal/safety"
)

const EnvDebugMetadata = "ICUVISOR_DEBUG_METADATA"

var processDeleteMode atomic.Value

func init() {
	processDeleteMode.Store(safety.ModeSafe.String())
}

var defaultScaleLabels = map[string]string{
	"feel":         "1-5 (athlete-reported feel)",
	"fatigue":      "1-5 (athlete-reported fatigue)",
	"mood":         "1-5 (athlete-reported mood)",
	"motivation":   "1-5 (athlete-reported motivation)",
	"rpe":          "1-10 (rating of perceived exertion)",
	"session_rpe":  "1-10 (session rating of perceived exertion)",
	"sleepQuality": "1-4 (athlete-entered, 1=poor 4=great)",
	"sleepScore":   "0-100 (device-imported nightly score)",
	"soreness":     "1-5 (athlete-reported soreness)",
	"stress":       "1-5 (athlete-reported stress)",
}

// Options controls response shaping at the MCP response boundary.
type Options struct {
	IncludeFull    bool
	RowCollections []string
	ServerVersion  string
	DebugMetadata  bool
	QueryType      string
	FetchedAt      time.Time
	UnitSystem     UnitSystem
}

// SetDeleteMode stores the process-global delete mode reported in response metadata.
func SetDeleteMode(mode string) {
	processDeleteMode.Store(safety.ParseMode(mode).String())
}

// DeleteMode returns the process-global delete mode reported in response metadata.
func DeleteMode() string {
	mode, ok := processDeleteMode.Load().(string)
	if !ok || strings.TrimSpace(mode) == "" {
		return safety.ModeSafe.String()
	}
	return mode
}

// DebugMetadataFromEnv reads the debug metadata toggle for startup configuration.
func DebugMetadataFromEnv() bool {
	return ParseDebugMetadata(os.Getenv(EnvDebugMetadata))
}

// ParseDebugMetadata reports whether a raw debug metadata value enables debug output.
func ParseDebugMetadata(value string) bool {
	return strings.EqualFold(strings.TrimSpace(value), "true")
}

// RegisteredScaleLabels returns the central field-name to scale-label registry.
func RegisteredScaleLabels() map[string]string {
	out := make(map[string]string, len(defaultScaleLabels))
	for field, label := range defaultScaleLabels {
		out[field] = label
	}
	return out
}

// Shape converts value through JSON tags and applies response-boundary shaping.
func Shape(value any, opts Options) (any, error) {
	jsonValue, err := marshalToJSONValue(value)
	if err != nil {
		return nil, err
	}
	root, ok := jsonValue.(map[string]any)
	if !ok {
		return nil, errors.New("response shape must be a JSON object wrapper")
	}
	return shapeRoot(root, opts), nil
}

func marshalToJSONValue(value any) (any, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("marshaling response value: %w", err)
	}
	var out any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("unmarshaling response value: %w", err)
	}
	return out, nil
}

func shapeRoot(root map[string]any, opts Options) map[string]any {
	if len(opts.RowCollections) > 0 {
		return shapeWrapperRow(root, opts)
	}
	return shapeRow(root, opts, true)
}

func shapeRows(values []any, opts Options, includeCommonMeta bool) []any {
	rows := make([]any, 0, len(values))
	for _, item := range values {
		if row, ok := item.(map[string]any); ok {
			rows = append(rows, shapeRow(row, opts, includeCommonMeta))
			continue
		}
		if opts.IncludeFull {
			rows = append(rows, item)
			continue
		}
		stripped, _ := stripNulls(item, "")
		rows = append(rows, stripped)
	}
	return rows
}

func shapeWrapperRow(row map[string]any, opts Options) map[string]any {
	collections := rowCollectionSet(opts.RowCollections)
	out := make(map[string]any, len(row))
	var missing []string
	for key, item := range row {
		itemPath := key
		if item == nil {
			if opts.IncludeFull {
				out[key] = nil
			} else {
				missing = append(missing, itemPath)
			}
			continue
		}
		if collections[key] {
			if values, ok := item.([]any); ok {
				out[key] = shapeRows(values, opts, false)
			} else {
				out[key] = item
			}
			continue
		}
		if opts.IncludeFull {
			out[key] = item
			continue
		}
		stripped, nestedMissing := stripNulls(item, itemPath)
		out[key] = stripped
		missing = append(missing, nestedMissing...)
	}
	if opts.DebugMetadata {
		addDebugMetadata(out, opts)
	} else {
		dropDebugMetadata(out, "")
		missing = filterDebugMissing(missing)
	}
	if !opts.IncludeFull && len(missing) > 0 {
		addStripMeta(out, missing)
	}
	addScaleMeta(out)
	addCommonMeta(out, opts)
	return out
}

func shapeRow(row map[string]any, opts Options, includeCommonMeta bool) map[string]any {
	var shaped map[string]any
	var missing []string
	if opts.IncludeFull {
		shaped = cloneMap(row)
	} else {
		stripped, strippedMissing := stripNulls(row, "")
		var ok bool
		shaped, ok = stripped.(map[string]any)
		if !ok {
			shaped = cloneMap(row)
		}
		missing = strippedMissing
	}
	if opts.DebugMetadata {
		if includeCommonMeta {
			addDebugMetadata(shaped, opts)
		}
	} else {
		dropDebugMetadata(shaped, "")
		missing = filterDebugMissing(missing)
	}
	if !opts.IncludeFull && len(missing) > 0 {
		addStripMeta(shaped, missing)
	}
	addScaleMeta(shaped)
	if includeCommonMeta {
		addCommonMeta(shaped, opts)
	}
	return shaped
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
	switch typed := value.(type) {
	case map[string]any:
		for key, item := range typed {
			if label, ok := defaultScaleLabels[key]; ok && item != nil {
				scales[key] = label
			}
			if key != "_meta" {
				collectScaleLabels(item, scales)
			}
		}
	case []any:
		for _, item := range typed {
			collectScaleLabels(item, scales)
		}
	}
}

func addCommonMeta(row map[string]any, opts Options) {
	meta := map[string]any{}
	if existing, ok := row["_meta"].(map[string]any); ok {
		for key, value := range existing {
			if key != "units" {
				meta[key] = value
			}
		}
	}
	meta["server_version"] = normalizeVersion(opts.ServerVersion)
	meta["delete_mode"] = DeleteMode()
	if opts.UnitSystem != "" {
		meta["units"] = opts.UnitSystem.Metadata()
	}
	row["_meta"] = meta
}

func stripNulls(value any, path string) (any, []string) {
	switch typed := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(typed))
		var missing []string
		for key, item := range typed {
			itemPath := joinPath(path, key)
			if item == nil {
				missing = append(missing, itemPath)
				continue
			}
			stripped, nestedMissing := stripNulls(item, itemPath)
			out[key] = stripped
			missing = append(missing, nestedMissing...)
		}
		return out, missing
	case []any:
		out := make([]any, 0, len(typed))
		var missing []string
		for i, item := range typed {
			itemPath := fmt.Sprintf("%s[%d]", path, i)
			if path == "" {
				itemPath = fmt.Sprintf("[%d]", i)
			}
			if item == nil {
				out = append(out, nil)
				continue
			}
			stripped, nestedMissing := stripNulls(item, itemPath)
			out = append(out, stripped)
			missing = append(missing, nestedMissing...)
		}
		return out, missing
	default:
		return value, nil
	}
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

func dropDebugMetadata(value any, path string) {
	switch typed := value.(type) {
	case map[string]any:
		if !isProvenancePath(path) {
			delete(typed, "fetched_at")
			delete(typed, "query_type")
		}
		for key, item := range typed {
			dropDebugMetadata(item, joinPath(path, key))
		}
	case []any:
		for i, item := range typed {
			itemPath := fmt.Sprintf("%s[%d]", path, i)
			if path == "" {
				itemPath = fmt.Sprintf("[%d]", i)
			}
			dropDebugMetadata(item, itemPath)
		}
	}
}

func isProvenancePath(path string) bool {
	return path == "_meta.provenance" || strings.HasPrefix(path, "_meta.provenance.") || strings.Contains(path, "._meta.provenance")
}

func isProvenanceFetchedAtPath(path string) bool {
	return strings.Contains(path, "_meta.provenance.") && strings.HasSuffix(path, ".fetched_at")
}

func cloneMap(row map[string]any) map[string]any {
	out := make(map[string]any, len(row))
	for key, value := range row {
		out[key] = value
	}
	return out
}

func normalizeVersion(version string) string {
	version = strings.TrimSpace(version)
	if version == "" {
		return "dev"
	}
	return version
}

func joinPath(base string, key string) string {
	if base == "" {
		return key
	}
	return base + "." + key
}

func rowCollectionSet(rowCollections []string) map[string]bool {
	collections := make(map[string]bool, len(rowCollections))
	for _, key := range rowCollections {
		if key != "" {
			collections[key] = true
		}
	}
	return collections
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

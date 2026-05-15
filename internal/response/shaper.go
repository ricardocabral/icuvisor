package response

import (
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ricardocabral/icuvisor/internal/safety"
)

const defaultCatalogHash = "dev-catalog-hash"

var catalogRuntime = struct {
	sync.Mutex
	current   catalogSnapshot
	firstSeen *catalogSnapshot
}{current: catalogSnapshot{CatalogHash: defaultCatalogHash}}

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
	DeleteMode     safety.Mode
	Toolset        safety.Toolset
}

// SetRuntimeCatalogMetadata stores the process-global catalog metadata reported in response metadata.
func SetRuntimeCatalogMetadata(version string, catalogHash string) {
	catalogRuntime.Lock()
	defer catalogRuntime.Unlock()
	catalogRuntime.current = catalogSnapshot{Version: normalizeVersion(version), CatalogHash: normalizeCatalogHash(catalogHash)}
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
	out, err := toJSONValue(reflect.ValueOf(value))
	if err != nil {
		return nil, err
	}
	return out, nil
}

func toJSONValue(value reflect.Value) (any, error) {
	if !value.IsValid() {
		return nil, nil
	}
	if value.Kind() == reflect.Interface {
		if value.IsNil() {
			return nil, nil
		}
		return toJSONValue(value.Elem())
	}
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil, nil
		}
		if canInterface(value) {
			if marshaled, ok, err := marshalSpecialValue(value.Interface()); ok || err != nil {
				return marshaled, err
			}
		}
		return toJSONValue(value.Elem())
	}
	if canInterface(value) {
		if marshaled, ok, err := marshalSpecialValue(value.Interface()); ok || err != nil {
			return marshaled, err
		}
	}
	if value.Kind() == reflect.Slice && value.Type().Elem().Kind() == reflect.Uint8 {
		if canInterface(value) {
			return marshalJSONValue(value.Interface())
		}
	}
	return reflectJSONValue(value)
}

func reflectJSONValue(value reflect.Value) (any, error) {
	switch value.Kind() {
	case reflect.Bool:
		return value.Bool(), nil
	case reflect.String:
		return value.String(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(value.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return float64(value.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return value.Convert(reflect.TypeOf(float64(0))).Float(), nil
	case reflect.Map:
		return mapToJSONValue(value)
	case reflect.Slice, reflect.Array:
		return sliceToJSONValue(value)
	case reflect.Struct:
		return structToJSONValue(value)
	case reflect.Invalid:
		return nil, nil
	default:
		return nil, fmt.Errorf("marshaling response value: unsupported JSON value %s", value.Kind())
	}
}

func mapToJSONValue(value reflect.Value) (any, error) {
	if value.IsNil() {
		return nil, nil
	}
	if value.Type().Key().Kind() != reflect.String {
		if canInterface(value) {
			return marshalJSONValue(value.Interface())
		}
		return nil, fmt.Errorf("marshaling response value: unsupported map key type %s", value.Type().Key())
	}
	out := make(map[string]any, value.Len())
	iter := value.MapRange()
	for iter.Next() {
		item, err := toJSONValue(iter.Value())
		if err != nil {
			return nil, err
		}
		out[iter.Key().String()] = item
	}
	return out, nil
}

func sliceToJSONValue(value reflect.Value) (any, error) {
	if value.Kind() == reflect.Slice && value.IsNil() {
		return nil, nil
	}
	out := make([]any, value.Len())
	for i := range value.Len() {
		item, err := toJSONValue(value.Index(i))
		if err != nil {
			return nil, err
		}
		out[i] = item
	}
	return out, nil
}

func structToJSONValue(value reflect.Value) (any, error) {
	out := make(map[string]any, value.NumField())
	valueType := value.Type()
	for i := range value.NumField() {
		field := valueType.Field(i)
		if field.PkgPath != "" {
			continue
		}
		name, omitEmpty, skip, fallback := jsonField(field)
		if skip {
			continue
		}
		if fallback {
			return marshalJSONValue(value.Interface())
		}
		fieldValue := value.Field(i)
		if omitEmpty && isEmptyJSONValue(fieldValue) {
			continue
		}
		item, err := toJSONValue(fieldValue)
		if err != nil {
			return nil, err
		}
		out[name] = item
	}
	return out, nil
}

func jsonField(field reflect.StructField) (name string, omitEmpty bool, skip bool, fallback bool) {
	name = field.Name
	tag := field.Tag.Get("json")
	if tag == "-" {
		return "", false, true, false
	}
	parts := strings.Split(tag, ",")
	if parts[0] != "" {
		name = parts[0]
	} else if field.Anonymous {
		return "", false, false, true
	}
	for _, option := range parts[1:] {
		if option == "omitempty" {
			omitEmpty = true
		}
	}
	return name, omitEmpty, false, false
}

func isEmptyJSONValue(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return value.Len() == 0
	case reflect.Bool:
		return !value.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return value.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return value.Float() == 0
	case reflect.Interface, reflect.Pointer:
		return value.IsNil()
	default:
		return false
	}
}

func marshalSpecialValue(value any) (any, bool, error) {
	if marshaler, ok := value.(json.Marshaler); ok {
		out, err := marshalJSONValue(marshaler)
		return out, true, err
	}
	if marshaler, ok := value.(encoding.TextMarshaler); ok {
		text, err := marshaler.MarshalText()
		if err != nil {
			return nil, true, fmt.Errorf("marshaling response value: %w", err)
		}
		return string(text), true, nil
	}
	return nil, false, nil
}

func marshalJSONValue(value any) (any, error) {
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

func canInterface(value reflect.Value) bool {
	return value.IsValid() && value.CanInterface()
}

type jsonWalkContainer int

const (
	walkRoot jsonWalkContainer = iota
	walkMapValue
	walkSliceValue
)

type jsonWalkDecision struct {
	Drop    bool
	Stop    bool
	Missing []string
}

type jsonWalkVisitor func(path string, value any, container jsonWalkContainer) jsonWalkDecision

func walkJSON(value any, path string, container jsonWalkContainer, visitor jsonWalkVisitor) (any, []string, bool) {
	decision := visitor(path, value, container)
	if decision.Drop {
		return nil, decision.Missing, true
	}
	if decision.Stop {
		return value, decision.Missing, false
	}
	switch typed := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(typed))
		missing := append([]string(nil), decision.Missing...)
		for key, item := range typed {
			itemPath := joinPath(path, key)
			walked, nestedMissing, dropped := walkJSON(item, itemPath, walkMapValue, visitor)
			missing = append(missing, nestedMissing...)
			if dropped {
				continue
			}
			out[key] = walked
		}
		return out, missing, false
	case []any:
		out := make([]any, 0, len(typed))
		missing := append([]string(nil), decision.Missing...)
		for i, item := range typed {
			walked, nestedMissing, dropped := walkJSON(item, indexPath(path, i), walkSliceValue, visitor)
			missing = append(missing, nestedMissing...)
			if dropped {
				out = append(out, nil)
				continue
			}
			out = append(out, walked)
		}
		return out, missing, false
	default:
		return value, decision.Missing, false
	}
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
		if dropped, ok := dropDebugMetadata(out, "").(map[string]any); ok {
			out = dropped
		}
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
		if dropped, ok := dropDebugMetadata(shaped, "").(map[string]any); ok {
			shaped = dropped
		}
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
	for key, value := range schemaCatalogMeta(serverVersion) {
		meta[key] = value
	}
	meta["delete_mode"] = opts.DeleteMode.String()
	meta["toolset"] = opts.Toolset.String()
	if opts.UnitSystem != "" {
		meta["units"] = opts.UnitSystem.Metadata()
	}
	row["_meta"] = meta
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

func stripNulls(value any, path string) (any, []string) {
	stripped, missing, _ := walkJSON(value, path, walkRoot, stripNullVisitor)
	return stripped, missing
}

func stripNullVisitor(path string, value any, container jsonWalkContainer) jsonWalkDecision {
	if value != nil {
		return jsonWalkDecision{}
	}
	if container == walkMapValue {
		return jsonWalkDecision{Drop: true, Missing: []string{path}}
	}
	return jsonWalkDecision{Stop: true}
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
	if path != "" && isDebugPath(path) {
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

func normalizeCatalogHash(catalogHash string) string {
	catalogHash = strings.TrimSpace(catalogHash)
	if catalogHash == "" {
		return defaultCatalogHash
	}
	return catalogHash
}

func joinPath(base string, key string) string {
	if base == "" {
		return key
	}
	return base + "." + key
}

func indexPath(base string, index int) string {
	if base == "" {
		return fmt.Sprintf("[%d]", index)
	}
	return fmt.Sprintf("%s[%d]", base, index)
}

func isMetaPath(path string) bool {
	return path == "_meta" || strings.Contains(path, "._meta")
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

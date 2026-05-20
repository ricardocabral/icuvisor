package tools

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

const (
	getWellnessDataName                    = "get_wellness_data"
	getWellnessDataDescription             = "Get daily wellness rows for a local date range with distinct sleepQuality, sleepScore, sleepSecs, custom fields, and native provider sidecars. Dates are athlete-local YYYY-MM-DD values."
	invalidGetWellnessDataArgumentsMessage = "invalid get_wellness_data arguments; provide oldest/newest dates as YYYY-MM-DD and optional include_full"
	fetchWellnessDataMessage               = "could not fetch wellness data; check intervals.icu credentials, athlete ID, and date range"
)

var wellnessSleepScoreNativeScales = map[string]string{
	"garmin": "0-100 Garmin sleep score",
	"oura":   "0-100 Oura sleep score",
	"polar":  "1-100 Polar sleep_score",
	"whoop":  "0-100 WHOOP sleep performance percentage",
}

var wellnessReadinessNativeScales = map[string]string{
	"garmin": "0-100 Garmin Body Battery",
	"oura":   "0-100 Oura readiness score",
	"whoop":  "0-100 WHOOP recovery score",
}

// WellnessClient retrieves athlete wellness rows for tools.
type WellnessClient interface {
	ListWellness(context.Context, intervals.WellnessParams) ([]intervals.Wellness, error)
}

type getWellnessDataRequest struct {
	Oldest      string   `json:"oldest"`
	Newest      string   `json:"newest"`
	Fields      []string `json:"fields,omitempty"`
	IncludeFull bool     `json:"include_full,omitempty"`
}

type getWellnessDataResponse struct {
	Wellness []map[string]any    `json:"wellness"`
	Meta     getWellnessDataMeta `json:"_meta"`
}

type getWellnessDataMeta struct {
	ServerVersion string   `json:"server_version"`
	Oldest        string   `json:"oldest"`
	Newest        string   `json:"newest"`
	Fields        []string `json:"fields,omitempty"`
	IncludeFull   bool     `json:"include_full"`
}

func newGetWellnessDataTool(client WellnessClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return coreTool(Tool{Name: getWellnessDataName, Description: getWellnessDataDescription, InputSchema: wellnessDataInputSchema(), OutputSchema: getWellnessDataOutputSchema(), Handler: getWellnessDataHandler(client, profileClient, version, timezoneFallback, debugMetadata, shapeCfg)})
}

func getWellnessDataHandler(client WellnessClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeGetWellnessDataRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidGetWellnessDataArgumentsMessage, err)
		}
		unitSystem, _, err := toolProfile(ctx, profileClient, timezoneFallback)
		if err != nil {
			return Result{}, NewUserError(fetchWellnessDataMessage, err)
		}
		rows, err := client.ListWellness(ctx, intervals.WellnessParams{Oldest: args.Oldest, Newest: args.Newest, Fields: args.Fields})
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(fetchWellnessDataMessage, err)
		}
		payload := getWellnessDataResponse{Wellness: wellnessRows(rows, args.IncludeFull), Meta: getWellnessDataMeta{ServerVersion: normalizeVersion(version), Oldest: args.Oldest, Newest: args.Newest, Fields: args.Fields, IncludeFull: args.IncludeFull}}
		return encodeShaped(payload, args.IncludeFull, []string{"wellness"}, version, debugMetadata, getWellnessDataName, unitSystem, shapeCfg)
	}
}

func decodeGetWellnessDataRequest(raw json.RawMessage) (getWellnessDataRequest, error) {
	var args getWellnessDataRequest
	if strings.TrimSpace(string(raw)) == "" {
		return args, errors.New("arguments must be a JSON object")
	}
	decoded, err := DecodeStrict[getWellnessDataRequest](raw)
	if err != nil {
		return args, err
	}
	args = decoded
	args.Oldest = strings.TrimSpace(args.Oldest)
	args.Newest = strings.TrimSpace(args.Newest)
	if !validDate(args.Oldest) || !validDate(args.Newest) {
		return args, errors.New("oldest and newest must be YYYY-MM-DD")
	}
	if args.Newest < args.Oldest {
		return args, errors.New("newest must be on or after oldest")
	}
	args.Fields = compactToolStrings(args.Fields)
	return args, nil
}

func wellnessRows(rows []intervals.Wellness, includeFull bool) []map[string]any {
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		out = append(out, wellnessRow(row, includeFull))
	}
	sort.SliceStable(out, func(i, j int) bool {
		left, _ := out[i]["date"].(string)
		right, _ := out[j]["date"].(string)
		return left < right
	})
	return out
}

func wellnessRow(row intervals.Wellness, includeFull bool) map[string]any {
	out := cloneJSONMap(row.Raw)
	for _, key := range row.NativeClaimedKeys {
		delete(out, key)
	}
	setWellnessField(out, "date", row.ID)
	setWellnessField(out, "id", row.ID)
	setWellnessField(out, "ctl", row.CTL)
	setWellnessField(out, "atl", row.ATL)
	setWellnessField(out, "rampRate", row.RampRate)
	setWellnessField(out, "ctlLoad", row.CTLLoad)
	setWellnessField(out, "atlLoad", row.ATLLoad)
	if row.SportInfo != nil || hasKey(row.Raw, "sportInfo") {
		out["sportInfo"] = row.SportInfo
	}
	setWellnessField(out, "updated", row.Updated)
	setWellnessField(out, "weight", row.Weight)
	setWellnessField(out, "restingHR", row.RestingHR)
	setWellnessField(out, "hrv", row.HRV)
	setWellnessField(out, "hrvSDNN", row.HRVSDNN)
	setWellnessField(out, "menstrualPhase", row.MenstrualPhase)
	setWellnessField(out, "menstrualPhasePredicted", row.MenstrualPhasePredicted)
	setWellnessField(out, "kcalConsumed", row.KcalConsumed)
	setWellnessField(out, "sleepSecs", row.SleepSecs)
	setWellnessField(out, "sleepScore", row.SleepScore)
	setWellnessField(out, "sleepQuality", row.SleepQuality)
	setWellnessField(out, "avgSleepingHR", row.AvgSleepingHR)
	setWellnessField(out, "feel", row.Feel)
	setWellnessField(out, "soreness", row.Soreness)
	setWellnessField(out, "fatigue", row.Fatigue)
	setWellnessField(out, "stress", row.Stress)
	setWellnessField(out, "mood", row.Mood)
	setWellnessField(out, "motivation", row.Motivation)
	if row.Injury != nil || hasKey(row.Raw, "injury") {
		out["injury"] = row.Injury
	}
	setWellnessField(out, "spO2", row.SpO2)
	setWellnessField(out, "systolic", row.Systolic)
	setWellnessField(out, "diastolic", row.Diastolic)
	setWellnessField(out, "hydration", row.Hydration)
	setWellnessField(out, "hydrationVolume", row.HydrationVolume)
	setWellnessField(out, "readiness", row.Readiness)
	setWellnessField(out, "baevskySI", row.BaevskySI)
	setWellnessField(out, "bloodGlucose", row.BloodGlucose)
	setWellnessField(out, "lactate", row.Lactate)
	setWellnessField(out, "bodyFat", row.BodyFat)
	setWellnessField(out, "abdomen", row.Abdomen)
	setWellnessField(out, "vo2max", row.VO2Max)
	setWellnessField(out, "comments", row.Comments)
	setWellnessField(out, "steps", row.Steps)
	setWellnessField(out, "respiration", row.Respiration)
	setWellnessField(out, "carbohydrates", row.Carbohydrates)
	setWellnessField(out, "protein", row.Protein)
	setWellnessField(out, "fatTotal", row.FatTotal)
	setWellnessField(out, "locked", row.Locked)
	setWellnessField(out, "tempWeight", row.TempWeight)
	setWellnessField(out, "tempRestingHR", row.TempRestingHR)
	if len(row.Native) > 0 {
		out["_native"] = cloneNestedJSONMap(row.Native)
	}
	addWellnessMeta(out, row)
	if includeFull {
		out["full"] = cloneJSONMap(row.Raw)
	}
	return out
}

func setWellnessField[T any](out map[string]any, key string, value *T) {
	if value != nil {
		out[key] = *value
	}
}

func addWellnessMeta(out map[string]any, row intervals.Wellness) {
	provenance, staleSource := wellnessProvenance(out, row)
	if len(provenance) == 0 {
		return
	}
	meta := map[string]any{}
	if existing, ok := out["_meta"].(map[string]any); ok {
		for key, value := range existing {
			meta[key] = value
		}
	}
	meta["provenance"] = provenance
	if staleSource != "" {
		meta["stale"] = true
		if staleSource == "unknown" {
			meta["stale_reason"] = "wellness bridge data is older than 24h for this wellness date"
		} else {
			meta["stale_reason"] = staleSource + " bridge data is older than 24h for this wellness date"
		}
	}
	out["_meta"] = meta
}

func wellnessProvenance(out map[string]any, row intervals.Wellness) (map[string]any, string) {
	provenance := map[string]any{}
	staleSource := ""
	for _, field := range []string{"sleepScore", "sleepSecs", "readiness", "avgSleepingHR"} {
		if shouldAddWellnessProvenance(out, field) {
			entry, stale := wellnessProvenanceEntry(row, field)
			provenance[field] = entry
			if stale && staleSource == "" {
				staleSource, _ = entry["source"].(string)
			}
		}
	}
	if wellnessProviderEvidence(row) != "" || len(row.Native) > 0 {
		for _, field := range []string{"restingHR", "hrv", "hrvSDNN", "spO2", "respiration", "steps", "vo2max", "baevskySI"} {
			if shouldAddWellnessProvenance(out, field) {
				entry, stale := wellnessProvenanceEntry(row, field)
				provenance[field] = entry
				if stale && staleSource == "" {
					staleSource, _ = entry["source"].(string)
				}
			}
		}
	}
	return provenance, staleSource
}

func shouldAddWellnessProvenance(out map[string]any, field string) bool {
	value, ok := out[field]
	return ok && value != nil
}

func wellnessProvenanceEntry(row intervals.Wellness, field string) (map[string]any, bool) {
	source := wellnessFieldSource(row, field)
	if source == "" {
		source = "unknown"
	}
	fetchedAt, fetchedTime, hasFetchedTime := wellnessFetchedAt(row.Raw, source)
	entry := map[string]any{
		"source":       source,
		"native_scale": wellnessNativeScale(row, field, source),
		"fetched_at":   fetchedAt,
	}
	return entry, hasFetchedTime && wellnessFetchedAtIsStale(row, fetchedTime)
}

func wellnessFieldSource(row intervals.Wellness, field string) string {
	switch field {
	case "sleepScore":
		if hasNativeField(row, "polar", "sleep_score") {
			return "polar"
		}
		if hasNativeField(row, "oura", "sleep_score") {
			return "oura"
		}
	case "readiness":
		if hasNativeField(row, "polar", "nightly_recharge_status") || hasNativeField(row, "polar", "ans_charge") {
			return "polar"
		}
		if hasNativeField(row, "garmin", "body_battery_min") || hasNativeField(row, "garmin", "body_battery_max") {
			return "garmin"
		}
	}
	return wellnessProviderEvidence(row)
}

func wellnessProviderEvidence(row intervals.Wellness) string {
	for _, source := range []string{"polar", "garmin", "oura"} {
		if len(row.Native[source]) > 0 {
			return source
		}
	}
	for _, key := range []string{"source", "provider", "device", "wellnessSource", "wellness_source", "integration"} {
		if source := normalizeWellnessSource(row.Raw[key]); source != "" {
			return source
		}
	}
	return ""
}

func normalizeWellnessSource(value any) string {
	text, ok := value.(string)
	if !ok {
		return ""
	}
	lower := strings.ToLower(text)
	for _, source := range []string{"polar", "garmin", "oura", "whoop"} {
		if strings.Contains(lower, source) {
			return source
		}
	}
	if strings.Contains(lower, "apple") || strings.Contains(lower, "health") {
		return "apple_health"
	}
	return ""
}

func hasNativeField(row intervals.Wellness, source string, field string) bool {
	fields := row.Native[source]
	if fields == nil {
		return false
	}
	_, ok := fields[field]
	return ok
}

func wellnessNativeScale(row intervals.Wellness, field string, source string) string {
	switch field {
	case "sleepScore":
		if label, ok := wellnessSleepScoreNativeScales[source]; ok {
			return label
		}
		return "unknown"
	case "sleepSecs":
		return "seconds"
	case "avgSleepingHR", "restingHR":
		return "bpm"
	case "readiness":
		return wellnessReadinessNativeScale(row, source)
	case "hrv":
		return "ms rMSSD"
	case "hrvSDNN":
		return "ms SDNN"
	case "spO2":
		return "percent"
	case "respiration":
		return "breaths/min"
	case "steps":
		return "count"
	case "vo2max":
		return "ml/kg/min"
	case "baevskySI":
		return "Baevsky stress index"
	default:
		return "unknown"
	}
}

func wellnessReadinessNativeScale(row intervals.Wellness, source string) string {
	if source == "polar" {
		if hasNativeField(row, "polar", "nightly_recharge_status") {
			return "1-6 Polar nightly_recharge_status"
		}
		if hasNativeField(row, "polar", "ans_charge") {
			return "-10 to +10 Polar ans_charge"
		}
		return "unknown"
	}
	if label, ok := wellnessReadinessNativeScales[source]; ok {
		return label
	}
	return "unknown"
}

func wellnessFetchedAt(raw map[string]any, source string) (string, time.Time, bool) {
	for _, key := range wellnessFetchedAtKeys() {
		if fetchedAt, parsed, ok := parseWellnessTimestamp(raw[key]); ok {
			return fetchedAt, parsed, !parsed.IsZero()
		}
	}
	if provider, ok := raw[source].(map[string]any); ok {
		for _, key := range wellnessFetchedAtKeys() {
			if fetchedAt, parsed, ok := parseWellnessTimestamp(provider[key]); ok {
				return fetchedAt, parsed, !parsed.IsZero()
			}
		}
	}
	return "unknown", time.Time{}, false
}

func wellnessFetchedAtKeys() []string {
	return []string{"bridge_fetched_at", "bridgeFetchedAt", "provider_fetched_at", "providerFetchedAt", "imported_at", "importedAt", "importedAtUtc", "imported_at_utc", "fetched_at", "fetchedAt"}
}

func parseWellnessTimestamp(value any) (string, time.Time, bool) {
	text, ok := value.(string)
	if !ok || strings.TrimSpace(text) == "" {
		return "", time.Time{}, false
	}
	text = strings.TrimSpace(text)
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, time.DateOnly} {
		parsed, err := time.Parse(layout, text)
		if err == nil {
			return text, parsed.UTC(), true
		}
	}
	return text, time.Time{}, true
}

func wellnessFetchedAtIsStale(row intervals.Wellness, fetchedAt time.Time) bool {
	if fetchedAt.IsZero() || row.ID == nil {
		return false
	}
	reference, err := time.Parse(time.DateOnly, strings.TrimSpace(*row.ID))
	if err != nil {
		return false
	}
	return reference.UTC().Sub(fetchedAt.UTC()) > 24*time.Hour
}

func cloneJSONMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func cloneNestedJSONMap(in map[string]map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for source, fields := range in {
		out[source] = cloneJSONMap(fields)
	}
	return out
}

func hasKey(in map[string]any, key string) bool {
	_, ok := in[key]
	return ok
}

func compactToolStrings(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" && !seen[trimmed] {
			seen[trimmed] = true
			out = append(out, trimmed)
		}
	}
	return out
}

func wellnessDataInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"oldest", "newest"}, "properties": map[string]any{"oldest": map[string]any{"type": "string", "description": "Local oldest wellness date YYYY-MM-DD in the athlete timezone."}, "newest": map[string]any{"type": "string", "description": "Local newest wellness date YYYY-MM-DD in the athlete timezone, inclusive."}, "fields": map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Optional upstream wellness field names to request; custom fields are preserved when returned."}, "include_full": map[string]any{"type": "boolean", "default": false, "description": "When true, include the raw upstream wellness row under full and keep null fields."}}}
}

func getWellnessDataOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Daily wellness rows with distinct sleepQuality (1-4), sleepScore (0-100), sleepSecs, custom fields, and _native provider fields."}
}

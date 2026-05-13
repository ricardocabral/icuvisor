package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/response"
	"github.com/ricardocabral/icuvisor/internal/safety"
)

const (
	updateSportSettingsName                    = "update_sport_settings"
	updateSportSettingsDescription             = "Update one sport's FTP, threshold heart rate, threshold pace, or zone definitions. Threshold-only writes are allowed in safe/full modes; supplying zones overwrites prior zone definitions and requires ICUVISOR_DELETE_MODE=full."
	invalidUpdateSportSettingsArgumentsMessage = "invalid update_sport_settings arguments; provide sport, effective_date, and at least one documented threshold or gated zones field"
	writeSportSettingsMessage                  = "could not update sport settings; check intervals.icu credentials, athlete ID, sport, effective date, and writable fields"
	zoneOverwriteGateMessage                   = "zones overwrite prior sport-setting zone definitions; set ICUVISOR_DELETE_MODE=full to allow this destructive argument"
)

var supportedSportSettingsSports = []string{"Ride", "Run", "Swim", "VirtualRide", "Walk", "Hike", "Rowing", "WeightTraining", "AlpineSki", "NordicSki", "Other"}

type SportSettingsWriterClient interface {
	UpdateSportSettings(context.Context, intervals.WriteSportSettingsParams) (intervals.SportSettings, error)
}

type updateSportSettingsRequest struct {
	Sport         string                           `json:"sport"`
	EffectiveDate string                           `json:"effective_date"`
	FTP           *int                             `json:"ftp,omitempty"`
	ThresholdHR   *int                             `json:"threshold_hr,omitempty"`
	ThresholdPace *updateSportSettingsPaceRequest  `json:"threshold_pace,omitempty"`
	Zones         []updateSportSettingsZoneRequest `json:"zones,omitempty"`
	zonesProvided bool
}

type updateSportSettingsPaceRequest struct {
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

type updateSportSettingsZoneRequest struct {
	Kind       string    `json:"kind"`
	Boundaries []float64 `json:"boundaries"`
	Names      []string  `json:"names,omitempty"`
}

type updateSportSettingsResponse struct {
	SportSettings updateSportSettingsEcho `json:"sport_settings"`
	Meta          updateSportSettingsMeta `json:"_meta"`
}

type updateSportSettingsEcho struct {
	Sport                       string                        `json:"sport"`
	SportSettingID              int                           `json:"sport_setting_id,omitempty"`
	FTPWatts                    *int                          `json:"ftp_watts,omitempty"`
	ThresholdHRBPM              *int                          `json:"threshold_hr_bpm,omitempty"`
	ThresholdPaceSecondsPerKM   *float64                      `json:"threshold_pace_seconds_per_km,omitempty"`
	ThresholdPaceSecondsPerMile *float64                      `json:"threshold_pace_seconds_per_mile,omitempty"`
	ThresholdPaceSecondsPer100M *float64                      `json:"threshold_pace_seconds_per_100m,omitempty"`
	ThresholdPaceSecondsPer500M *float64                      `json:"threshold_pace_seconds_per_500m,omitempty"`
	ThresholdPaceValue          *float64                      `json:"threshold_pace_value,omitempty"`
	PaceUnitsSource             string                        `json:"pace_units_source,omitempty"`
	Zones                       []updateSportSettingsZoneEcho `json:"zones,omitempty"`
	ZoneDefinitionsOverwritten  bool                          `json:"zone_definitions_overwritten,omitempty"`
	Upstream                    map[string]any                `json:"upstream,omitempty"`
}

type updateSportSettingsZoneEcho struct {
	Kind       string    `json:"kind"`
	Boundaries []float64 `json:"boundaries"`
	Names      []string  `json:"names,omitempty"`
}

type updateSportSettingsMeta struct {
	ServerVersion    string            `json:"server_version"`
	DeleteMode       string            `json:"delete_mode"`
	EffectiveDate    string            `json:"effective_date"`
	Timezone         string            `json:"timezone,omitempty"`
	FieldsUpdated    []string          `json:"fields_updated"`
	RecomputePending bool              `json:"recompute_pending"`
	ZonesProvided    bool              `json:"zones_provided"`
	PaceInputUnit    string            `json:"pace_input_unit,omitempty"`
	PaceUpstreamUnit string            `json:"pace_upstream_unit,omitempty"`
	Units            map[string]string `json:"units,omitempty"`
}

func newUpdateSportSettingsTool(client SportSettingsWriterClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, capability safety.Capability) Tool {
	return Tool{Name: updateSportSettingsName, Description: updateSportSettingsDescription, InputSchema: updateSportSettingsInputSchema(), OutputSchema: updateSportSettingsOutputSchema(), Requirement: RequirementWrite, Handler: updateSportSettingsHandler(client, profileClient, version, timezoneFallback, debugMetadata, capabilityOrSafe(capability))}
}

func updateSportSettingsHandler(client SportSettingsWriterClient, profileClient ProfileClient, version string, timezoneFallback string, _ bool, capability safety.Capability) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeUpdateSportSettingsRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidUpdateSportSettingsArgumentsMessage, err)
		}
		if args.zonesProvided && !capability.CanDelete() {
			return Result{}, NewUserError(zoneOverwriteGateMessage, errors.New("zone overwrite requires delete capability"))
		}
		if profileClient == nil || client == nil {
			return Result{}, NewUserError(writeSportSettingsMessage, errors.New("missing sport settings writer client"))
		}
		profile, err := profileClient.GetAthleteProfile(ctx)
		if err != nil {
			return Result{}, NewUserError(writeSportSettingsMessage, err)
		}
		setting, ok := findSportSetting(profile.SportSettings, args.Sport)
		if !ok || setting.ID <= 0 {
			return Result{}, NewUserError(writeSportSettingsMessage, fmt.Errorf("sport %q settings not found", args.Sport))
		}
		params, meta, err := sportSettingsWriteParams(args, setting, profile, timezoneFallback, version, capability.Mode())
		if err != nil {
			return Result{}, NewUserError(invalidUpdateSportSettingsArgumentsMessage, err)
		}
		updated, err := client.UpdateSportSettings(ctx, params)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(writeSportSettingsMessage, err)
		}
		payload := shapeUpdateSportSettingsResponse(args, params, updated, meta)
		text, err := json.Marshal(payload)
		if err != nil {
			return Result{}, fmt.Errorf("encoding update_sport_settings response: %w", err)
		}
		return Result{Content: []Content{{Type: ContentTypeText, Text: string(text)}}, StructuredContent: payload}, nil
	}
}

func decodeUpdateSportSettingsRequest(raw json.RawMessage) (updateSportSettingsRequest, error) {
	zonesProvided, err := rawObjectHasField(raw, "zones")
	if err != nil {
		return updateSportSettingsRequest{}, err
	}
	var args updateSportSettingsRequest
	if err := decodeStrict(raw, &args); err != nil {
		return args, err
	}
	args.zonesProvided = zonesProvided
	args.Sport = canonicalSport(args.Sport)
	args.EffectiveDate = strings.TrimSpace(args.EffectiveDate)
	if args.Sport == "" || !validSport(args.Sport) {
		return args, errors.New("sport must be one of the documented enum values")
	}
	if !validDate(args.EffectiveDate) {
		return args, errors.New("effective_date must be athlete-local YYYY-MM-DD")
	}
	if err := validateSportSettingsThresholds(args); err != nil {
		return args, err
	}
	if zonesProvided {
		if err := validateSportSettingsZones(args.Zones); err != nil {
			return args, err
		}
	}
	if len(updateSportSettingsFieldsUpdated(args)) == 0 {
		return args, errors.New("at least one writable sport-settings field is required")
	}
	return args, nil
}

func rawObjectHasField(raw json.RawMessage, field string) (bool, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return false, nil
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &fields); err != nil {
		return false, err
	}
	_, ok := fields[field]
	return ok, nil
}

func validateSportSettingsThresholds(args updateSportSettingsRequest) error {
	if args.FTP != nil && *args.FTP <= 0 {
		return errors.New("ftp must be > 0 watts")
	}
	if args.ThresholdHR != nil && *args.ThresholdHR <= 0 {
		return errors.New("threshold_hr must be > 0 bpm")
	}
	if args.ThresholdPace != nil {
		if args.ThresholdPace.Value <= 0 || !isSupportedPaceInputUnit(args.ThresholdPace.Unit) {
			return errors.New("threshold_pace must include value > 0 and a supported unit")
		}
	}
	return nil
}

func validateSportSettingsZones(zones []updateSportSettingsZoneRequest) error {
	if len(zones) == 0 {
		return errors.New("zones must contain at least one zone definition when supplied")
	}
	seen := map[string]bool{}
	for _, zone := range zones {
		kind := normalizeZoneKind(zone.Kind)
		if kind == "" {
			return errors.New("zone kind must be power, hr, or pace")
		}
		if seen[kind] {
			return fmt.Errorf("duplicate %s zone definition", kind)
		}
		seen[kind] = true
		if len(zone.Boundaries) == 0 {
			return fmt.Errorf("%s zone boundaries are required", kind)
		}
		if len(zone.Names) > 0 && len(zone.Names) != len(zone.Boundaries) {
			return fmt.Errorf("%s zone names must match boundaries length", kind)
		}
		for _, boundary := range zone.Boundaries {
			if boundary < 0 {
				return fmt.Errorf("%s zone boundaries must be >= 0", kind)
			}
		}
	}
	return nil
}

func sportSettingsWriteParams(args updateSportSettingsRequest, setting intervals.SportSettings, profile intervals.AthleteWithSportSettings, timezoneFallback string, version string, deleteMode string) (intervals.WriteSportSettingsParams, updateSportSettingsMeta, error) {
	params := intervals.WriteSportSettingsParams{SportSettingID: setting.ID, EffectiveDate: args.EffectiveDate, FTP: args.FTP, ThresholdHR: args.ThresholdHR, ZonesProvided: args.zonesProvided}
	meta := updateSportSettingsMeta{ServerVersion: normalizeVersion(version), DeleteMode: deleteMode, EffectiveDate: args.EffectiveDate, Timezone: profileTimezone(profile.Timezone, timezoneFallback), FieldsUpdated: updateSportSettingsFieldsUpdated(args), RecomputePending: true, ZonesProvided: args.zonesProvided, Units: profileUnitSystem(profile).Metadata()}
	if args.ThresholdPace != nil {
		pace, upstreamUnit, err := convertThresholdPaceForUpstream(*args.ThresholdPace, setting, profileUnitSystem(profile))
		if err != nil {
			return params, meta, err
		}
		params.ThresholdPace = &intervals.SportSettingsPace{Value: pace, Unit: upstreamUnit}
		meta.PaceInputUnit = normalizePaceInputUnit(args.ThresholdPace.Unit)
		meta.PaceUpstreamUnit = upstreamUnit
	}
	if args.zonesProvided {
		params.Zones = make([]intervals.SportSettingsZoneDefinition, 0, len(args.Zones))
		for _, zone := range args.Zones {
			params.Zones = append(params.Zones, intervals.SportSettingsZoneDefinition{Kind: normalizeZoneKind(zone.Kind), Boundaries: append([]float64(nil), zone.Boundaries...), Names: append([]string(nil), zone.Names...)})
		}
	}
	return params, meta, nil
}

func updateSportSettingsFieldsUpdated(args updateSportSettingsRequest) []string {
	fields := []string{}
	if args.FTP != nil {
		fields = append(fields, "ftp")
	}
	if args.ThresholdHR != nil {
		fields = append(fields, "threshold_hr")
	}
	if args.ThresholdPace != nil {
		fields = append(fields, "threshold_pace")
	}
	if args.zonesProvided {
		fields = append(fields, "zones")
	}
	sort.Strings(fields)
	return fields
}

func findSportSetting(settings []intervals.SportSettings, sport string) (intervals.SportSettings, bool) {
	canonical := canonicalSport(sport)
	for _, setting := range settings {
		if strings.EqualFold(setting.Type, canonical) {
			return setting, true
		}
		for _, value := range setting.Types {
			if strings.EqualFold(value, canonical) {
				return setting, true
			}
		}
	}
	return intervals.SportSettings{}, false
}

func shapeUpdateSportSettingsResponse(args updateSportSettingsRequest, params intervals.WriteSportSettingsParams, updated intervals.SportSettings, meta updateSportSettingsMeta) updateSportSettingsResponse {
	echo := updateSportSettingsEcho{Sport: args.Sport, SportSettingID: params.SportSettingID, ZoneDefinitionsOverwritten: args.zonesProvided}
	if updated.ID > 0 {
		echo.SportSettingID = updated.ID
	}
	if value := firstPositiveInt(updated.FTP, params.FTP); value != nil {
		echo.FTPWatts = value
	}
	if value := firstPositiveInt(firstNonZero(updated.LTHR, updated.FTHR), params.ThresholdHR); value != nil {
		echo.ThresholdHRBPM = value
	}
	paceValue := firstPositiveFloat(firstNonZeroFloat(updated.ThresholdPace, updated.PaceThreshold), nil)
	if params.ThresholdPace != nil {
		paceValue = &params.ThresholdPace.Value
		echo.PaceUnitsSource = params.ThresholdPace.Unit
	}
	if updated.PaceUnits != "" && echo.PaceUnitsSource == "" {
		echo.PaceUnitsSource = updated.PaceUnits
	}
	assignUpdateSportSettingsPace(&echo, paceValue, echo.PaceUnitsSource)
	if args.zonesProvided {
		echo.Zones = make([]updateSportSettingsZoneEcho, 0, len(params.Zones))
		for _, zone := range params.Zones {
			echo.Zones = append(echo.Zones, updateSportSettingsZoneEcho{Kind: zone.Kind, Boundaries: append([]float64(nil), zone.Boundaries...), Names: append([]string(nil), zone.Names...)})
		}
	}
	return updateSportSettingsResponse{SportSettings: echo, Meta: meta}
}

func assignUpdateSportSettingsPace(echo *updateSportSettingsEcho, value *float64, unit string) {
	if value == nil {
		return
	}
	switch strings.TrimSpace(unit) {
	case "MINS_KM":
		echo.ThresholdPaceSecondsPerKM = value
	case "MINS_MILE":
		echo.ThresholdPaceSecondsPerMile = value
	case "SECS_100M":
		echo.ThresholdPaceSecondsPer100M = value
	case "SECS_500M":
		echo.ThresholdPaceSecondsPer500M = value
	default:
		echo.ThresholdPaceValue = value
	}
}

func convertThresholdPaceForUpstream(input updateSportSettingsPaceRequest, setting intervals.SportSettings, unitSystem response.UnitSystem) (float64, string, error) {
	inputUnit := normalizePaceInputUnit(input.Unit)
	secondsPerMeter, err := inputPaceSecondsPerMeter(input.Value, inputUnit)
	if err != nil {
		return 0, "", err
	}
	upstreamUnit := strings.TrimSpace(setting.PaceUnits)
	if upstreamUnit == "" {
		upstreamUnit = upstreamPaceUnitFromInput(inputUnit, unitSystem)
	}
	switch upstreamUnit {
	case "MINS_KM":
		return secondsPerMeter * 1000, upstreamUnit, nil
	case "MINS_MILE":
		return secondsPerMeter * 1609.344, upstreamUnit, nil
	case "SECS_100M":
		return secondsPerMeter * 100, upstreamUnit, nil
	case "SECS_500M":
		return secondsPerMeter * 500, upstreamUnit, nil
	default:
		return input.Value, upstreamUnit, nil
	}
}

func inputPaceSecondsPerMeter(value float64, unit string) (float64, error) {
	if value <= 0 || math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, errors.New("threshold_pace value must be finite and > 0")
	}
	switch unit {
	case "seconds_per_km":
		return value / 1000, nil
	case "seconds_per_mile":
		return value / 1609.344, nil
	case "seconds_per_100m":
		return value / 100, nil
	case "seconds_per_500m":
		return value / 500, nil
	case "minutes_per_km":
		return value * 60 / 1000, nil
	case "minutes_per_mile":
		return value * 60 / 1609.344, nil
	default:
		return 0, fmt.Errorf("unsupported threshold_pace unit %q", unit)
	}
}

func upstreamPaceUnitFromInput(inputUnit string, unitSystem response.UnitSystem) string {
	switch inputUnit {
	case "seconds_per_100m":
		return "SECS_100M"
	case "seconds_per_500m":
		return "SECS_500M"
	case "seconds_per_mile", "minutes_per_mile":
		return "MINS_MILE"
	case "seconds_per_km", "minutes_per_km":
		return "MINS_KM"
	default:
		if unitSystem == response.UnitSystemImperial {
			return "MINS_MILE"
		}
		return "MINS_KM"
	}
}

func firstPositiveInt(updated int, fallback *int) *int {
	if updated > 0 {
		value := updated
		return &value
	}
	if fallback != nil {
		value := *fallback
		return &value
	}
	return nil
}

func firstPositiveFloat(updated float64, fallback *float64) *float64 {
	if updated > 0 {
		value := updated
		return &value
	}
	if fallback != nil {
		value := *fallback
		return &value
	}
	return nil
}

func firstNonZero(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}

func firstNonZeroFloat(values ...float64) float64 {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}

func validSport(value string) bool {
	for _, sport := range supportedSportSettingsSports {
		if value == sport {
			return true
		}
	}
	return false
}

func canonicalSport(value string) string {
	trimmed := strings.TrimSpace(value)
	for _, sport := range supportedSportSettingsSports {
		if strings.EqualFold(trimmed, sport) {
			return sport
		}
	}
	return trimmed
}

func normalizeZoneKind(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "power":
		return "power"
	case "hr", "heart_rate", "heart-rate":
		return "hr"
	case "pace":
		return "pace"
	default:
		return ""
	}
}

func normalizePaceInputUnit(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func isSupportedPaceInputUnit(value string) bool {
	switch normalizePaceInputUnit(value) {
	case "seconds_per_km", "seconds_per_mile", "seconds_per_100m", "seconds_per_500m", "minutes_per_km", "minutes_per_mile":
		return true
	default:
		return false
	}
}

func updateSportSettingsInputSchema() map[string]any {
	examples := updateSportSettingsInputExamples()
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"sport", "effective_date"}, "examples": examples, "input_examples": examples, "properties": map[string]any{
		"sport":          map[string]any{"type": "string", "enum": supportedSportSettingsSports, "description": "Sport setting to update, matching intervals.icu sport type (for example Ride, Run, Swim)."},
		"effective_date": map[string]any{"type": "string", "description": "Required athlete-local effective date as YYYY-MM-DD; used as the oldest date for upstream sport-setting recompute."},
		"ftp":            map[string]any{"type": "integer", "minimum": 1, "description": "Functional Threshold Power in watts for the selected sport."},
		"threshold_hr":   map[string]any{"type": "integer", "minimum": 1, "description": "Threshold heart rate in bpm for the selected sport."},
		"threshold_pace": map[string]any{"type": "object", "additionalProperties": false, "required": []string{"value", "unit"}, "description": "Threshold pace with an explicit unit; seconds_per_km is 4:15/km as 255.", "properties": map[string]any{
			"value": map[string]any{"type": "number", "exclusiveMinimum": 0, "description": "Threshold pace numeric value in the provided unit."},
			"unit":  map[string]any{"type": "string", "enum": []string{"seconds_per_km", "seconds_per_mile", "seconds_per_100m", "seconds_per_500m", "minutes_per_km", "minutes_per_mile"}, "description": "Unit for threshold_pace value."},
		}},
		"zones": map[string]any{"type": "array", "description": "Optional destructive replacement zone definitions. Supplying zones overwrites prior power/hr/pace zone definitions for this sport and is rejected unless ICUVISOR_DELETE_MODE=full.", "items": map[string]any{"type": "object", "additionalProperties": false, "required": []string{"kind", "boundaries"}, "properties": map[string]any{
			"kind":       map[string]any{"type": "string", "enum": []string{"power", "hr", "pace"}, "description": "Zone family to overwrite."},
			"boundaries": map[string]any{"type": "array", "minItems": 1, "items": map[string]any{"type": "number", "minimum": 0}, "description": "Ordered zone boundary values: watts for power, bpm for hr, seconds in the sport pace unit for pace."},
			"names":      map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Optional zone names; when supplied, length must match boundaries."},
		}}},
	}}
}

func updateSportSettingsInputExamples() []map[string]any {
	return []map[string]any{
		{
			"sport":          "Ride",
			"effective_date": "2026-06-01",
			"ftp":            285,
		},
		{
			"sport":          "Run",
			"effective_date": "2026-06-01",
			"threshold_hr":   172,
			"threshold_pace": map[string]any{"value": 255, "unit": "seconds_per_km"},
		},
		{
			"sport":          "Ride",
			"effective_date": "2026-07-01",
			"ftp":            290,
			"threshold_hr":   168,
			"zones": []any{
				map[string]any{"kind": "power", "boundaries": []any{150, 200, 250, 300}, "names": []any{"Endurance", "Tempo", "Threshold", "VO2"}},
			},
		},
	}
}

func updateSportSettingsOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Echoes updated sport settings with recompute metadata, delete-mode metadata, and unit metadata."}
}

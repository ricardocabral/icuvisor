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
	"github.com/ricardocabral/icuvisor/internal/units"
)

const (
	updateSportSettingsName                    = "update_sport_settings"
	updateSportSettingsDescription             = "Update one sport's FTP, indoor FTP, threshold heart rate, threshold pace, or zone definitions referenced by get_athlete_profile _meta.warnings. Threshold-only writes are allowed in safe/full modes; supplying zones overwrites prior zone definitions and requires ICUVISOR_DELETE_MODE=full."
	invalidUpdateSportSettingsArgumentsMessage = "invalid update_sport_settings arguments; provide sport and at least one documented threshold or gated zones field"
	writeSportSettingsMessage                  = "could not update sport settings; check intervals.icu credentials, athlete ID, sport, and writable fields"
	zoneOverwriteGateMessage                   = "zones overwrite prior sport-setting zone definitions; set ICUVISOR_DELETE_MODE=full to allow this destructive argument"
	metersPer100Yards                          = 91.44
)

var supportedSportSettingsSports = []string{"Ride", "Run", "Swim", "VirtualRide", "Walk", "Hike", "Rowing", "WeightTraining", "AlpineSki", "NordicSki", "Other"}

type SportSettingsWriterClient interface {
	UpdateSportSettings(context.Context, intervals.WriteSportSettingsParams) (intervals.SportSettings, error)
}

type updateSportSettingsRequest struct {
	Sport         string                           `json:"sport"`
	RecalcHRZones *bool                            `json:"recalc_hr_zones,omitempty"`
	FTP           *int                             `json:"ftp,omitempty"`
	IndoorFTP     *int                             `json:"indoor_ftp,omitempty"`
	ThresholdHR   *int                             `json:"threshold_hr,omitempty"`
	ThresholdPace *updateSportSettingsPaceRequest  `json:"threshold_pace,omitempty"`
	Zones         []updateSportSettingsZoneRequest `json:"zones,omitempty"`
	zonesProvided bool
}

type updateSportSettingsPaceRequest struct {
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

type updateSportSettingsResponse struct {
	SportSettings updateSportSettingsEcho `json:"sport_settings"`
	Meta          updateSportSettingsMeta `json:"_meta"`
}

type updateSportSettingsEcho struct {
	Sport                        string                        `json:"sport"`
	SportSettingID               int                           `json:"sport_setting_id,omitempty"`
	FTPWatts                     *int                          `json:"ftp_watts,omitempty"`
	IndoorFTPWatts               *int                          `json:"indoor_ftp_watts,omitempty"`
	ThresholdHRBPM               *int                          `json:"threshold_hr_bpm,omitempty"`
	ThresholdPaceSecondsPerKM    *float64                      `json:"threshold_pace_seconds_per_km,omitempty"`
	ThresholdPaceSecondsPerMile  *float64                      `json:"threshold_pace_seconds_per_mile,omitempty"`
	ThresholdPaceSecondsPer100M  *float64                      `json:"threshold_pace_seconds_per_100m,omitempty"`
	ThresholdPaceSecondsPer100Y  *float64                      `json:"threshold_pace_seconds_per_100y,omitempty"`
	ThresholdPaceSecondsPer500M  *float64                      `json:"threshold_pace_seconds_per_500m,omitempty"`
	ThresholdPaceSecondsPer400M  *float64                      `json:"threshold_pace_seconds_per_400m,omitempty"`
	ThresholdPaceSecondsPer250M  *float64                      `json:"threshold_pace_seconds_per_250m,omitempty"`
	ThresholdPaceMetersPerSecond *float64                      `json:"threshold_pace_meters_per_second,omitempty"`
	PaceUnitsSource              string                        `json:"pace_units_source,omitempty"`
	PaceLoadType                 string                        `json:"pace_load_type,omitempty"`
	Zones                        []updateSportSettingsZoneEcho `json:"zones,omitempty"`
	ZoneDefinitionsOverwritten   bool                          `json:"zone_definitions_overwritten,omitempty"`
	Upstream                     map[string]any                `json:"upstream,omitempty"`
}

type updateSportSettingsMeta struct {
	ServerVersion                string            `json:"server_version"`
	DeleteMode                   string            `json:"delete_mode"`
	Timezone                     string            `json:"timezone,omitempty"`
	FieldsUpdated                []string          `json:"fields_updated"`
	HRZoneRecalculationRequested bool              `json:"hr_zone_recalculation_requested"`
	ZonesProvided                bool              `json:"zones_provided"`
	PaceInputUnit                string            `json:"pace_input_unit,omitempty"`
	PaceDisplayUnit              string            `json:"pace_display_unit,omitempty"`
	PaceLoadType                 string            `json:"pace_load_type,omitempty"`
	Units                        map[string]string `json:"units,omitempty"`
}

func newUpdateSportSettingsTool(client SportSettingsWriterClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, capability safety.Capability, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return fullTool(Tool{Name: updateSportSettingsName, Description: updateSportSettingsDescription, InputSchema: updateSportSettingsInputSchema(), OutputSchema: updateSportSettingsOutputSchema(), Requirement: RequirementWrite, Handler: updateSportSettingsHandler(client, profileClient, version, timezoneFallback, debugMetadata, capabilityOrSafe(capability), shapeCfg)})
}

func updateSportSettingsHandler(client SportSettingsWriterClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, capability safety.Capability, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeUpdateSportSettingsRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidUpdateSportSettingsArgumentsMessage, err)
		}
		if err := ensureSportSettingsZonesAllowed(args.zonesProvided, capability); err != nil {
			return Result{}, err
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
			if msg, ok := validationErrorMessage(err); ok {
				return Result{}, NewUserError(msg, err)
			}
			return Result{}, NewUserError(writeSportSettingsMessage, err)
		}
		payload := shapeUpdateSportSettingsResponse(args, params, updated, meta)
		return encodeShaped(payload, false, nil, version, debugMetadata, updateSportSettingsName, profileUnitSystem(profile), shapeCfg)
	}
}

func decodeUpdateSportSettingsRequest(raw json.RawMessage) (updateSportSettingsRequest, error) {
	zonesProvided, err := rawObjectHasField(raw, "zones")
	if err != nil {
		return updateSportSettingsRequest{}, err
	}
	recalcHRZonesNull, err := rawObjectFieldIsNull(raw, "recalc_hr_zones")
	if err != nil {
		return updateSportSettingsRequest{}, err
	}
	if recalcHRZonesNull {
		return updateSportSettingsRequest{}, errors.New("recalc_hr_zones must be a boolean when supplied")
	}
	var args updateSportSettingsRequest
	if strings.TrimSpace(string(raw)) == "" {
		return args, errors.New("arguments must be a JSON object")
	}
	decoded, err := DecodeStrict[updateSportSettingsRequest](raw)
	if err != nil {
		return args, err
	}
	args = decoded
	args.zonesProvided = zonesProvided
	args.Sport = canonicalSport(args.Sport)
	if args.Sport == "" || !validSport(args.Sport) {
		return args, errors.New("sport must be one of the documented enum values")
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
	fields, err := rawSportSettingsObjectFields(raw)
	if err != nil {
		return false, err
	}
	_, ok := fields[field]
	return ok, nil
}

func rawObjectFieldIsNull(raw json.RawMessage, field string) (bool, error) {
	fields, err := rawSportSettingsObjectFields(raw)
	if err != nil {
		return false, err
	}
	value, ok := fields[field]
	return ok && bytes.Equal(bytes.TrimSpace(value), []byte("null")), nil
}

func rawSportSettingsObjectFields(raw json.RawMessage) (map[string]json.RawMessage, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return nil, nil
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &fields); err != nil {
		return nil, err
	}
	return fields, nil
}

func validateSportSettingsThresholds(args updateSportSettingsRequest) error {
	if args.FTP != nil && *args.FTP <= 0 {
		return errors.New("ftp must be > 0 watts")
	}
	if args.IndoorFTP != nil && *args.IndoorFTP <= 0 {
		return errors.New("indoor_ftp must be > 0 watts")
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

func sportSettingsWriteParams(args updateSportSettingsRequest, setting intervals.SportSettings, profile intervals.AthleteWithSportSettings, timezoneFallback string, version string, deleteMode string) (intervals.WriteSportSettingsParams, updateSportSettingsMeta, error) {
	recalcHRZones := true
	if args.RecalcHRZones != nil {
		recalcHRZones = *args.RecalcHRZones
	}
	params := intervals.WriteSportSettingsParams{SportSettingID: setting.ID, RecalcHRZones: recalcHRZones, FTP: args.FTP, IndoorFTP: args.IndoorFTP, ThresholdHR: args.ThresholdHR, ZonesProvided: args.zonesProvided}
	meta := updateSportSettingsMeta{ServerVersion: normalizeVersion(version), DeleteMode: deleteMode, Timezone: profileTimezone(profile.Timezone, timezoneFallback), FieldsUpdated: updateSportSettingsFieldsUpdated(args), HRZoneRecalculationRequested: recalcHRZones, ZonesProvided: args.zonesProvided, Units: profileUnitSystem(profile).Metadata()}
	if args.ThresholdPace != nil {
		metersPerSecond, paceUnits, paceLoadType, err := convertThresholdPaceForUpstream(*args.ThresholdPace, setting, args.Sport)
		if err != nil {
			return params, meta, err
		}
		params.ThresholdPace = &intervals.SportSettingsPace{Value: metersPerSecond, PaceUnits: paceUnits, PaceLoadType: paceLoadType}
		meta.PaceInputUnit = normalizePaceInputUnit(args.ThresholdPace.Unit)
		meta.PaceDisplayUnit = paceUnits
		meta.PaceLoadType = paceLoadType
	}
	if args.zonesProvided {
		params.Zones = sportSettingsZoneDefinitions(args.Zones)
	}
	return params, meta, nil
}

func updateSportSettingsFieldsUpdated(args updateSportSettingsRequest) []string {
	fields := []string{}
	if args.FTP != nil {
		fields = append(fields, "ftp")
	}
	if args.IndoorFTP != nil {
		fields = append(fields, "indoor_ftp")
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
	if value := firstPositiveInt(updated.IndoorFTP, params.IndoorFTP); value != nil {
		echo.IndoorFTPWatts = value
	}
	if value := firstPositiveInt(firstNonZero(updated.LTHR, updated.FTHR), params.ThresholdHR); value != nil {
		echo.ThresholdHRBPM = value
	}
	paceValue := firstPositiveFloat(firstNonZeroFloat(updated.ThresholdPace, updated.PaceThreshold), nil)
	if params.ThresholdPace != nil && paceValue == nil {
		paceValue = &params.ThresholdPace.Value
	}
	if params.ThresholdPace != nil {
		echo.PaceUnitsSource = params.ThresholdPace.PaceUnits
		echo.PaceLoadType = params.ThresholdPace.PaceLoadType
	}
	if strings.TrimSpace(updated.PaceUnits) != "" {
		echo.PaceUnitsSource = strings.TrimSpace(updated.PaceUnits)
	}
	if strings.TrimSpace(updated.PaceLoadType) != "" {
		echo.PaceLoadType = strings.TrimSpace(updated.PaceLoadType)
	}
	assignUpdateSportSettingsPace(&echo, paceValue, echo.PaceUnitsSource)
	if args.zonesProvided {
		echo.Zones = sportSettingsZoneEchoes(params.Zones)
	}
	return updateSportSettingsResponse{SportSettings: echo, Meta: meta}
}

func assignUpdateSportSettingsPace(echo *updateSportSettingsEcho, value *float64, paceUnits string) {
	if value == nil {
		return
	}
	paceUnit, _ := units.ParseUnit(paceUnits)
	seconds, ok := response.PaceSecondsFromMetersPerSecond(*value, paceUnit)
	if !ok {
		metersPerSecond := *value
		echo.ThresholdPaceMetersPerSecond = &metersPerSecond
		return
	}
	switch paceUnit {
	case units.UnitMinsKM:
		echo.ThresholdPaceSecondsPerKM = &seconds
	case units.UnitMinsMile:
		echo.ThresholdPaceSecondsPerMile = &seconds
	case units.UnitSecs100M:
		echo.ThresholdPaceSecondsPer100M = &seconds
	case units.UnitSecs100Y:
		echo.ThresholdPaceSecondsPer100Y = &seconds
	case units.UnitSecs500M:
		echo.ThresholdPaceSecondsPer500M = &seconds
	case units.UnitSecs400M:
		echo.ThresholdPaceSecondsPer400M = &seconds
	case units.UnitSecs250M:
		echo.ThresholdPaceSecondsPer250M = &seconds
	}
}

func convertThresholdPaceForUpstream(input updateSportSettingsPaceRequest, setting intervals.SportSettings, sport string) (float64, string, string, error) {
	seconds, inputPaceUnit, err := inputPaceSeconds(input.Value, normalizePaceInputUnit(input.Unit))
	if err != nil {
		return 0, "", "", err
	}
	metersPerSecond, ok := response.PaceMetersPerSecondFromSeconds(seconds, inputPaceUnit)
	if !ok {
		return 0, "", "", errors.New("threshold_pace must resolve to a finite m/s value")
	}
	return metersPerSecond, paceDisplayUnit(setting.PaceUnits, inputPaceUnit), paceLoadTypeForWrite(setting.PaceLoadType, sport), nil
}

func inputPaceSeconds(value float64, inputUnit string) (float64, units.Unit, error) {
	if value <= 0 || math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, units.UnitUnknown, errors.New("threshold_pace value must be finite and > 0")
	}
	var seconds float64
	var paceUnit units.Unit
	switch inputUnit {
	case "seconds_per_km":
		seconds, paceUnit = value, units.UnitMinsKM
	case "seconds_per_mile":
		seconds, paceUnit = value, units.UnitMinsMile
	case "seconds_per_100m":
		seconds, paceUnit = value, units.UnitSecs100M
	case "seconds_per_100y":
		seconds, paceUnit = value, units.UnitSecs100Y
	case "seconds_per_500m":
		seconds, paceUnit = value, units.UnitSecs500M
	case "minutes_per_km":
		seconds, paceUnit = value*60, units.UnitMinsKM
	case "minutes_per_mile":
		seconds, paceUnit = value*60, units.UnitMinsMile
	default:
		return 0, units.UnitUnknown, fmt.Errorf("unsupported threshold_pace unit %q", inputUnit)
	}
	if seconds <= 0 || math.IsNaN(seconds) || math.IsInf(seconds, 0) {
		return 0, units.UnitUnknown, errors.New("threshold_pace must resolve to finite seconds")
	}
	return seconds, paceUnit, nil
}

func paceDisplayUnit(existing string, fallback units.Unit) string {
	if paceUnit, _ := units.ParseUnit(existing); responsePaceUnit(paceUnit) {
		return string(paceUnit)
	}
	return string(fallback)
}

func responsePaceUnit(paceUnit units.Unit) bool {
	_, ok := response.PaceDistanceMeters(paceUnit)
	return ok
}

func paceLoadTypeForWrite(existing string, sport string) string {
	switch strings.ToUpper(strings.TrimSpace(existing)) {
	case "RUN", "SWIM":
		return strings.ToUpper(strings.TrimSpace(existing))
	}
	switch canonicalSport(sport) {
	case "Run":
		return "RUN"
	case "Swim":
		return "SWIM"
	default:
		return ""
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

func normalizePaceInputUnit(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func isSupportedPaceInputUnit(value string) bool {
	switch normalizePaceInputUnit(value) {
	case "seconds_per_km", "seconds_per_mile", "seconds_per_100m", "seconds_per_100y", "seconds_per_500m", "minutes_per_km", "minutes_per_mile":
		return true
	default:
		return false
	}
}

func updateSportSettingsInputSchema() map[string]any {
	examples := updateSportSettingsInputExamples()
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"sport"}, "examples": examples, "input_examples": examples, "properties": map[string]any{
		"sport":           map[string]any{"type": "string", "enum": supportedSportSettingsSports, "description": "Sport setting to update, matching intervals.icu sport type (for example Ride, Run, Swim)."},
		"recalc_hr_zones": map[string]any{"type": "boolean", "default": true, "description": "Whether intervals.icu should recalculate heart-rate zones for the updated sport; defaults to true."},
		"ftp":             map[string]any{"type": "integer", "minimum": 1, "description": "Functional Threshold Power in watts for the selected sport."},
		"indoor_ftp":      map[string]any{"type": "integer", "minimum": 1, "description": "Indoor Functional Threshold Power in watts for the selected sport."},
		"threshold_hr":    map[string]any{"type": "integer", "minimum": 1, "description": "Threshold heart rate in bpm for the selected sport."},
		"threshold_pace": map[string]any{"type": "object", "additionalProperties": false, "required": []string{"value", "unit"}, "description": "Threshold pace with an explicit pace-duration unit; seconds_per_km is 4:15/km as 255, seconds_per_mile is 8:00/mi as 480, and seconds_per_100y is 1:30/100y as 90.", "properties": map[string]any{
			"value": map[string]any{"type": "number", "exclusiveMinimum": 0, "description": "Threshold pace duration in the provided unit, not speed."},
			"unit":  map[string]any{"type": "string", "enum": []string{"seconds_per_km", "seconds_per_mile", "seconds_per_100m", "seconds_per_100y", "seconds_per_500m", "minutes_per_km", "minutes_per_mile"}, "description": "Pace-duration unit for threshold_pace value."},
		}},
		"zones": map[string]any{"type": "array", "description": "Optional destructive replacement zone definitions. Supplying zones overwrites prior power/hr/pace zone definitions for this sport and is rejected unless ICUVISOR_DELETE_MODE=full.", "items": map[string]any{"type": "object", "additionalProperties": false, "required": []string{"kind", "boundaries"}, "properties": map[string]any{
			"kind":       map[string]any{"type": "string", "enum": []string{"power", "hr", "pace"}, "description": "Zone family to overwrite."},
			"boundaries": map[string]any{"type": "array", "minItems": 1, "items": map[string]any{"type": "number", "minimum": 0}, "description": "Ordered zone boundary values: watts for power, bpm for hr, and strictly increasing percent-of-threshold-pace values in (0, 200] for pace. For pace, 100 means threshold pace; values such as 77.5 and 100 are percentages, never durations."},
			"names":      map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Optional zone names; when supplied, length must match boundaries."},
		}}},
	}}
}

func updateSportSettingsInputExamples() []map[string]any {
	return []map[string]any{
		{
			"sport":      "Ride",
			"ftp":        285,
			"indoor_ftp": 265,
		},
		{
			"sport":          "Run",
			"threshold_hr":   172,
			"threshold_pace": map[string]any{"value": 255, "unit": "seconds_per_km"},
		},
		{
			"sport":          "Swim",
			"threshold_pace": map[string]any{"value": 90, "unit": "seconds_per_100y"},
		},
		{
			"sport": "Run",
			"zones": []any{
				map[string]any{"kind": "pace", "boundaries": []any{77.5, 100}, "names": []any{"Easy", "Threshold"}},
			},
		},
		{
			"sport":           "Ride",
			"recalc_hr_zones": false,
			"ftp":             290,
			"threshold_hr":    168,
			"zones": []any{
				map[string]any{"kind": "power", "boundaries": []any{150, 200, 250, 300}, "names": []any{"Endurance", "Tempo", "Threshold", "VO2"}},
			},
		},
	}
}

func updateSportSettingsOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Echoes updated sport settings with HR-zone recalculation request, delete-mode, and unit metadata."}
}

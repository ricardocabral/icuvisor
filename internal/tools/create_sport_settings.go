package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

const (
	createSportSettingsName                    = "create_sport_settings"
	createSportSettingsDescription             = "Create threshold-only sport settings for a sport that does not yet have settings. Provide sport and at least one FTP, indoor FTP, threshold heart rate, or threshold pace field. This creation path cannot replace zones, recalculate settings, or apply settings to historical activities."
	invalidCreateSportSettingsArgumentsMessage = "invalid create_sport_settings arguments; provide a documented sport and at least one positive threshold field"
	createSportSettingsMessage                 = "could not create sport settings; check intervals.icu credentials, athlete ID, sport, and writable threshold fields"
	sportSettingsAlreadyExistMessage           = "sport settings already exist for this sport; use update_sport_settings"
)

// SportSettingsCreatorClient creates threshold-only sport settings for tools.
type SportSettingsCreatorClient interface {
	CreateSportSettings(context.Context, intervals.CreateSportSettingsParams) (intervals.SportSettings, error)
}

type createSportSettingsRequest struct {
	Sport         string                          `json:"sport"`
	FTP           *int                            `json:"ftp,omitempty"`
	IndoorFTP     *int                            `json:"indoor_ftp,omitempty"`
	ThresholdHR   *int                            `json:"threshold_hr,omitempty"`
	ThresholdPace *updateSportSettingsPaceRequest `json:"threshold_pace,omitempty"`
}

type createSportSettingsResponse struct {
	SportSettings createSportSettingsEcho `json:"sport_settings"`
	Meta          createSportSettingsMeta `json:"_meta"`
}

type createSportSettingsEcho struct {
	Sport                        string   `json:"sport"`
	SportSettingID               int      `json:"sport_setting_id,omitempty"`
	FTPWatts                     *int     `json:"ftp_watts,omitempty"`
	IndoorFTPWatts               *int     `json:"indoor_ftp_watts,omitempty"`
	ThresholdHRBPM               *int     `json:"threshold_hr_bpm,omitempty"`
	ThresholdPaceSecondsPerKM    *float64 `json:"threshold_pace_seconds_per_km,omitempty"`
	ThresholdPaceSecondsPerMile  *float64 `json:"threshold_pace_seconds_per_mile,omitempty"`
	ThresholdPaceSecondsPer100M  *float64 `json:"threshold_pace_seconds_per_100m,omitempty"`
	ThresholdPaceSecondsPer100Y  *float64 `json:"threshold_pace_seconds_per_100y,omitempty"`
	ThresholdPaceSecondsPer500M  *float64 `json:"threshold_pace_seconds_per_500m,omitempty"`
	ThresholdPaceSecondsPer400M  *float64 `json:"threshold_pace_seconds_per_400m,omitempty"`
	ThresholdPaceSecondsPer250M  *float64 `json:"threshold_pace_seconds_per_250m,omitempty"`
	ThresholdPaceMetersPerSecond *float64 `json:"threshold_pace_meters_per_second,omitempty"`
	PaceUnitsSource              string   `json:"pace_units_source,omitempty"`
	PaceLoadType                 string   `json:"pace_load_type,omitempty"`
}

type createSportSettingsMeta struct {
	Operation       string            `json:"operation"`
	ServerVersion   string            `json:"server_version"`
	Timezone        string            `json:"timezone,omitempty"`
	FieldsCreated   []string          `json:"fields_created"`
	PaceInputUnit   string            `json:"pace_input_unit,omitempty"`
	PaceDisplayUnit string            `json:"pace_display_unit,omitempty"`
	PaceLoadType    string            `json:"pace_load_type,omitempty"`
	Units           map[string]string `json:"units,omitempty"`
}

func newCreateSportSettingsTool(client SportSettingsCreatorClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return fullTool(Tool{Name: createSportSettingsName, Description: createSportSettingsDescription, InputSchema: createSportSettingsInputSchema(), OutputSchema: createSportSettingsOutputSchema(), Requirement: RequirementWrite, Handler: createSportSettingsHandler(client, profileClient, version, timezoneFallback, debugMetadata, shapeCfg)})
}

func createSportSettingsHandler(client SportSettingsCreatorClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeCreateSportSettingsRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidCreateSportSettingsArgumentsMessage, err)
		}
		if profileClient == nil || client == nil {
			return Result{}, NewUserError(createSportSettingsMessage, errors.New("missing sport settings creator or profile client"))
		}
		profile, err := profileClient.GetAthleteProfile(ctx)
		if err != nil {
			return Result{}, NewUserError(createSportSettingsMessage, err)
		}
		if _, exists := findSportSetting(profile.SportSettings, args.Sport); exists {
			return Result{}, NewUserError(sportSettingsAlreadyExistMessage, fmt.Errorf("sport %q already has settings", args.Sport))
		}
		params, meta, err := createSportSettingsParams(args, profile, timezoneFallback, version)
		if err != nil {
			return Result{}, NewUserError(invalidCreateSportSettingsArgumentsMessage, err)
		}
		created, err := client.CreateSportSettings(ctx, params)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			if msg, ok := validationErrorMessage(err); ok {
				return Result{}, NewUserError(msg, err)
			}
			return Result{}, NewUserError(createSportSettingsMessage, err)
		}
		payload := shapeCreateSportSettingsResponse(args, params, created, meta)
		return encodeShaped(payload, false, nil, version, debugMetadata, createSportSettingsName, profileUnitSystem(profile), shapeCfg)
	}
}

func decodeCreateSportSettingsRequest(raw json.RawMessage) (createSportSettingsRequest, error) {
	var args createSportSettingsRequest
	if strings.TrimSpace(string(raw)) == "" {
		return args, errors.New("arguments must be a JSON object")
	}
	decoded, err := DecodeStrict[createSportSettingsRequest](raw)
	if err != nil {
		return args, err
	}
	args = decoded
	args.Sport = canonicalSport(args.Sport)
	if args.Sport == "" || !validSport(args.Sport) {
		return args, errors.New("sport must be one of the documented enum values")
	}
	if err := validateSportSettingsThresholds(updateSportSettingsRequest{FTP: args.FTP, IndoorFTP: args.IndoorFTP, ThresholdHR: args.ThresholdHR, ThresholdPace: args.ThresholdPace}); err != nil {
		return args, err
	}
	if len(createSportSettingsFieldsCreated(args)) == 0 {
		return args, errors.New("at least one threshold field is required")
	}
	return args, nil
}

func createSportSettingsParams(args createSportSettingsRequest, profile intervals.AthleteWithSportSettings, timezoneFallback string, version string) (intervals.CreateSportSettingsParams, createSportSettingsMeta, error) {
	params := intervals.CreateSportSettingsParams{Sport: args.Sport, FTP: args.FTP, IndoorFTP: args.IndoorFTP, ThresholdHR: args.ThresholdHR}
	meta := createSportSettingsMeta{Operation: "create", ServerVersion: normalizeVersion(version), Timezone: profileTimezone(profile.Timezone, timezoneFallback), FieldsCreated: createSportSettingsFieldsCreated(args), Units: profileUnitSystem(profile).Metadata()}
	if args.ThresholdPace != nil {
		metersPerSecond, paceUnits, paceLoadType, err := convertThresholdPaceForUpstream(*args.ThresholdPace, intervals.SportSettings{}, args.Sport)
		if err != nil {
			return params, meta, err
		}
		params.ThresholdPace = &intervals.SportSettingsPace{Value: metersPerSecond, PaceUnits: paceUnits, PaceLoadType: paceLoadType}
		meta.PaceInputUnit = normalizePaceInputUnit(args.ThresholdPace.Unit)
		meta.PaceDisplayUnit = paceUnits
		meta.PaceLoadType = paceLoadType
	}
	return params, meta, nil
}

func createSportSettingsFieldsCreated(args createSportSettingsRequest) []string {
	return updateSportSettingsFieldsUpdated(updateSportSettingsRequest{FTP: args.FTP, IndoorFTP: args.IndoorFTP, ThresholdHR: args.ThresholdHR, ThresholdPace: args.ThresholdPace})
}

func shapeCreateSportSettingsResponse(args createSportSettingsRequest, params intervals.CreateSportSettingsParams, created intervals.SportSettings, meta createSportSettingsMeta) createSportSettingsResponse {
	echo := createSportSettingsEcho{Sport: args.Sport, SportSettingID: created.ID}
	if args.FTP != nil {
		echo.FTPWatts = firstPositiveInt(created.FTP, params.FTP)
	}
	if args.IndoorFTP != nil {
		echo.IndoorFTPWatts = firstPositiveInt(created.IndoorFTP, params.IndoorFTP)
	}
	if args.ThresholdHR != nil {
		echo.ThresholdHRBPM = firstPositiveInt(firstNonZero(created.LTHR, created.FTHR), params.ThresholdHR)
	}
	if args.ThresholdPace != nil {
		paceValue := firstPositiveFloat(firstNonZeroFloat(created.ThresholdPace, created.PaceThreshold), &params.ThresholdPace.Value)
		echo.PaceUnitsSource = params.ThresholdPace.PaceUnits
		echo.PaceLoadType = params.ThresholdPace.PaceLoadType
		if strings.TrimSpace(created.PaceUnits) != "" {
			echo.PaceUnitsSource = strings.TrimSpace(created.PaceUnits)
		}
		if strings.TrimSpace(created.PaceLoadType) != "" {
			echo.PaceLoadType = strings.TrimSpace(created.PaceLoadType)
		}
		assignCreateSportSettingsPace(&echo, paceValue, echo.PaceUnitsSource)
	}
	return createSportSettingsResponse{SportSettings: echo, Meta: meta}
}

func assignCreateSportSettingsPace(echo *createSportSettingsEcho, value *float64, paceUnits string) {
	if value == nil {
		return
	}
	updateEcho := updateSportSettingsEcho{}
	assignUpdateSportSettingsPace(&updateEcho, value, paceUnits)
	echo.ThresholdPaceSecondsPerKM = updateEcho.ThresholdPaceSecondsPerKM
	echo.ThresholdPaceSecondsPerMile = updateEcho.ThresholdPaceSecondsPerMile
	echo.ThresholdPaceSecondsPer100M = updateEcho.ThresholdPaceSecondsPer100M
	echo.ThresholdPaceSecondsPer100Y = updateEcho.ThresholdPaceSecondsPer100Y
	echo.ThresholdPaceSecondsPer500M = updateEcho.ThresholdPaceSecondsPer500M
	echo.ThresholdPaceSecondsPer400M = updateEcho.ThresholdPaceSecondsPer400M
	echo.ThresholdPaceSecondsPer250M = updateEcho.ThresholdPaceSecondsPer250M
	echo.ThresholdPaceMetersPerSecond = updateEcho.ThresholdPaceMetersPerSecond
}

func createSportSettingsInputSchema() map[string]any {
	examples := createSportSettingsInputExamples()
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"sport"}, "examples": examples, "input_examples": examples, "properties": map[string]any{
		"sport":        map[string]any{"type": "string", "enum": supportedSportSettingsSports, "description": "Sport setting to create, matching an intervals.icu sport type that has no existing setting (for example Ride, Run, Swim)."},
		"ftp":          map[string]any{"type": "integer", "minimum": 1, "description": "Optional Functional Threshold Power in watts for the new sport setting."},
		"indoor_ftp":   map[string]any{"type": "integer", "minimum": 1, "description": "Optional indoor Functional Threshold Power in watts for the new sport setting."},
		"threshold_hr": map[string]any{"type": "integer", "minimum": 1, "description": "Optional threshold heart rate in bpm for the new sport setting."},
		"threshold_pace": map[string]any{"type": "object", "additionalProperties": false, "required": []string{"value", "unit"}, "description": "Optional threshold pace with an explicit pace-duration unit; seconds_per_km is 4:15/km as 255, seconds_per_mile is 8:00/mi as 480, and seconds_per_100y is 1:30/100y as 90.", "properties": map[string]any{
			"value": map[string]any{"type": "number", "exclusiveMinimum": 0, "description": "Threshold pace duration in the provided unit, not speed."},
			"unit":  map[string]any{"type": "string", "enum": []string{"seconds_per_km", "seconds_per_mile", "seconds_per_100m", "seconds_per_100y", "seconds_per_500m", "minutes_per_km", "minutes_per_mile"}, "description": "Pace-duration unit for threshold_pace value."},
		}},
	}}
}

func createSportSettingsInputExamples() []map[string]any {
	return []map[string]any{
		{"sport": "Ride", "ftp": 285, "indoor_ftp": 265},
		{"sport": "Run", "threshold_hr": 172, "threshold_pace": map[string]any{"value": 255, "unit": "seconds_per_km"}},
		{"sport": "Swim", "threshold_pace": map[string]any{"value": 90, "unit": "seconds_per_100y"}},
	}
}

func createSportSettingsOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Create confirmation containing the new sport, created sport-setting ID, requested threshold echoes, pace rendering in selected units, and creation metadata."}
}

package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/athleteprofile"
	"github.com/ricardocabral/icuvisor/internal/clients"
	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/response"
)

const (
	getAthleteProfileName                    = "get_athlete_profile"
	getAthleteProfileDescription             = "Get the configured athlete profile, FTP/thresholds, zones, and sport settings from intervals.icu. Use this for athlete identity, units, timezone, FTP, heart-rate thresholds, pace thresholds, and zone configuration; do not use it for activities, wellness, fitness trends, events, or workouts."
	invalidGetAthleteProfileArgumentsMessage = "invalid get_athlete_profile arguments; only include_full is supported"
	fetchAthleteProfileMessage               = "could not fetch athlete profile; check intervals.icu credentials and athlete ID"
)

// ProfileClient is the shared athlete profile client interface used by tools.
type ProfileClient = clients.ProfileClient

// GetAthleteProfileRequest contains the get_athlete_profile tool arguments.
type GetAthleteProfileRequest struct {
	IncludeFull bool `json:"include_full,omitempty"`
}

// GetAthleteProfileResponse is the structured get_athlete_profile response.
type GetAthleteProfileResponse = athleteprofile.Response

// GetAthleteProfileUnits describes athlete unit preferences.
type GetAthleteProfileUnits = athleteprofile.Units

// GetAthleteProfileSport contains thresholds and zones for one sport setting.
type GetAthleteProfileSport = athleteprofile.Sport

// GetAthleteProfileMeta contains response-shaping metadata.
type GetAthleteProfileMeta = athleteprofile.Meta

func newGetAthleteProfileTool(client ProfileClient, version string, timezoneFallback string, debugMetadata bool) Tool {
	return coreTool(Tool{
		Name:         getAthleteProfileName,
		Description:  getAthleteProfileDescription,
		InputSchema:  getAthleteProfileInputSchema(),
		OutputSchema: getAthleteProfileOutputSchema(),
		Handler:      getAthleteProfileHandler(client, version, timezoneFallback, debugMetadata),
	})
}

func getAthleteProfileHandler(client ProfileClient, version string, timezoneFallback string, debugMetadata bool) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		if err := ctx.Err(); err != nil {
			return Result{}, err
		}
		args, err := decodeGetAthleteProfileRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidGetAthleteProfileArgumentsMessage, err)
		}
		if client == nil {
			return Result{}, NewUserError(fetchAthleteProfileMessage, errors.New("missing profile client"))
		}
		profile, err := client.GetAthleteProfile(ctx)
		if err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return Result{}, ctxErr
			}
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(fetchAthleteProfileMessage, err)
		}
		shaped, err := shapeGetAthleteProfileResponse(profile, version, timezoneFallback, args.IncludeFull, debugMetadata)
		if err != nil {
			return Result{}, fmt.Errorf("shaping get_athlete_profile response: %w", err)
		}
		text, err := json.Marshal(shaped)
		if err != nil {
			return Result{}, fmt.Errorf("encoding get_athlete_profile response: %w", err)
		}
		return Result{
			Content:           []Content{{Type: ContentTypeText, Text: string(text)}},
			StructuredContent: shaped,
		}, nil
	}
}

func decodeGetAthleteProfileRequest(raw json.RawMessage) (GetAthleteProfileRequest, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return GetAthleteProfileRequest{}, nil
	}
	if trimmed[0] != '{' {
		return GetAthleteProfileRequest{}, errors.New("arguments must be a JSON object")
	}
	decoder := json.NewDecoder(bytes.NewReader(trimmed))
	decoder.DisallowUnknownFields()
	var args GetAthleteProfileRequest
	if err := decoder.Decode(&args); err != nil {
		return GetAthleteProfileRequest{}, err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return GetAthleteProfileRequest{}, errors.New("unexpected trailing JSON")
	}
	return args, nil
}

func shapeGetAthleteProfileResponse(profile intervals.AthleteWithSportSettings, version string, timezoneFallback string, includeFull bool, debugMetadata bool) (any, error) {
	return athleteprofile.Shape(profile, version, timezoneFallback, includeFull, debugMetadata)
}

func newGetAthleteProfileResponse(profile intervals.AthleteWithSportSettings, version string, timezoneFallback string, includeFull bool) GetAthleteProfileResponse {
	return athleteprofile.NewResponse(profile, version, timezoneFallback, includeFull)
}

func getAthleteProfileInputSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"include_full": map[string]any{
				"type":        "boolean",
				"default":     false,
				"description": "When true, include additional typed, non-secret profile and sport-setting identifiers. Defaults to false; raw upstream payloads and credentials are never returned.",
			},
		},
	}
}

func getAthleteProfileOutputSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": true,
		"description":          "Terse athlete profile with normalized athlete_id, units, timezone, sport thresholds/zones, and _meta.server_version.",
	}
}

func profileTimezone(profileTimezone string, fallback string) string {
	if timezone := strings.TrimSpace(profileTimezone); timezone != "" {
		return timezone
	}
	return strings.TrimSpace(fallback)
}

func normalizeTimezoneFallback(values ...string) string {
	if fallback := firstNonEmpty(values...); fallback != "" {
		return fallback
	}
	return config.DefaultTimezone
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func normalizeVersion(version string) string {
	version = strings.TrimSpace(version)
	if version == "" {
		return "dev"
	}
	return version
}

func profileUnitSystem(profile intervals.AthleteWithSportSettings) response.UnitSystem {
	if unitSystem, ok := response.UnitSystemFromProfile(profile.PreferredUnits, profile.MeasurementPreference, profile.WeightPrefLB); ok {
		return unitSystem
	}
	return response.UnitSystemMetric
}

package config

import (
	"fmt"
	"log/slog"
)

func (c Config) String() string {
	apiKey := "<unset>"
	if c.APIKey != "" {
		apiKey = "<redacted>"
	}
	apiKeySource := string(c.APIKeySource)
	if apiKeySource == "" {
		apiKeySource = "<unset>"
	}
	athleteID := "<unset>"
	if c.AthleteID != "" {
		athleteID = "<set>"
	}
	return fmt.Sprintf("api_key=%s api_key_source=%s athlete_id=%s timezone=%q api_base_url=%q http_timeout=%s transport=%s http_bind=%q http_stateless=%t http_json_response=%t delete_mode=%s toolset=%s coach_mode=%s coach_enabled=%t coach_athletes=%d", apiKey, apiKeySource, athleteID, c.Timezone, c.APIBaseURL, c.HTTPTimeout, c.Transport, c.HTTPBindAddress, c.HTTPStateless, c.HTTPJSONResponse, c.DeleteMode, c.Toolset, c.CoachMode, c.CoachModeEnabled(), len(c.Coach.Athletes))
}

// LogValue returns a structured, redacted representation for slog.
func (c Config) LogValue() slog.Value {
	athleteID := "<unset>"
	if c.AthleteID != "" {
		athleteID = "<set>"
	}
	return slog.GroupValue(
		slog.String("api_base_url", c.APIBaseURL),
		slog.String("default_athlete_id", athleteID),
		slog.String("http_bind", c.HTTPBindAddress),
		slog.Bool("http_stateless", c.HTTPStateless),
		slog.Bool("http_json_response", c.HTTPJSONResponse),
		slog.Int("coach_athletes_count", len(c.Coach.Athletes)),
		slog.String("delete_mode", c.DeleteMode.String()),
		slog.String("toolset", c.Toolset.String()),
	)
}

package config

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/ricardocabral/icuvisor/internal/coach"
	"github.com/ricardocabral/icuvisor/internal/credstore"
	"github.com/ricardocabral/icuvisor/internal/safety"
)

func validate(raw rawConfig) (Config, error) {
	apiKey, err := validateAPIKey(raw)
	if err != nil {
		return Config{}, err
	}
	credentialRef, err := validateCredentialReference(raw.credentialRef)
	if err != nil {
		return Config{}, err
	}
	coachMode, err := validateCoachMode(raw)
	if err != nil {
		return Config{}, err
	}
	coachConfig, err := validateCoachConfig(raw, coachMode)
	if err != nil {
		return Config{}, err
	}
	athleteID, err := validateAthleteID(raw, coachMode, coachConfig)
	if err != nil {
		return Config{}, err
	}
	loc, err := validateTimezone(raw)
	if err != nil {
		return Config{}, err
	}
	baseURL, err := validateAPIBaseURL(raw)
	if err != nil {
		return Config{}, err
	}
	timeout, err := validateHTTPTimeout(raw)
	if err != nil {
		return Config{}, err
	}
	transport, err := validateTransport(raw)
	if err != nil {
		return Config{}, err
	}
	httpBindAddress, err := validateHTTPBind(raw)
	if err != nil {
		return Config{}, err
	}
	return buildConfig(raw, apiKey, credentialRef, athleteID, loc, baseURL, timeout, transport, httpBindAddress, coachMode, coachConfig), nil
}

func validateAPIKey(raw rawConfig) (string, error) {
	apiKey := strings.TrimSpace(raw.apiKey)
	if apiKey == "" {
		return "", fmt.Errorf("missing intervals.icu API key; set %s, store it in OS keychain service %q account %q, or set legacy api_key in config JSON/.env", EnvAPIKey, credstore.ServiceName, credstore.IntervalsAPIKeyAccount)
	}
	return apiKey, nil
}

func validateCredentialReference(ref CredentialReference) (CredentialReference, error) {
	if ref.empty() {
		return CredentialReference{}, nil
	}
	expected := DefaultCredentialReference()
	ref.Type = strings.TrimSpace(ref.Type)
	ref.Service = strings.TrimSpace(ref.Service)
	ref.Account = strings.TrimSpace(ref.Account)
	if ref != expected {
		return CredentialReference{}, fmt.Errorf("unsupported credential_ref; use type %q with OS keychain service %q account %q", expected.Type, expected.Service, expected.Account)
	}
	return ref, nil
}

func validateCoachMode(raw rawConfig) (coach.Mode, error) {
	return coach.ParseMode(raw.coachMode)
}

func validateCoachConfig(raw rawConfig, coachMode coach.Mode) (coach.Config, error) {
	var rawCoach coach.Config
	if raw.coach != nil {
		rawCoach = *raw.coach
	}
	return coach.ValidateConfig(rawCoach, coachMode, NormalizeAthleteID)
}

func validateAthleteID(raw rawConfig, coachMode coach.Mode, coachConfig coach.Config) (string, error) {
	if coach.EffectiveMode(coachMode, coachConfig) == coach.ModeOn {
		return coachConfig.DefaultAthleteID, nil
	}
	return NormalizeAthleteID(raw.athleteID)
}

func validateTimezone(raw rawConfig) (*time.Location, error) {
	timezone := raw.timezone
	if timezone == "" {
		timezone = DefaultTimezone
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("invalid timezone %q; use an IANA timezone like UTC or America/Sao_Paulo", timezone)
	}
	return loc, nil
}

func validateAPIBaseURL(raw rawConfig) (string, error) {
	baseURL := raw.apiBaseURL
	if baseURL == "" {
		baseURL = DefaultAPIBaseURL
	}
	parsedURL, err := url.Parse(baseURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		return "", errors.New("invalid API base URL; use an absolute http or https URL")
	}
	return baseURL, nil
}

func validateHTTPTimeout(raw rawConfig) (time.Duration, error) {
	timeout := DefaultHTTPTimeout
	if raw.httpTimeout == "" {
		return timeout, nil
	}
	timeout, err := time.ParseDuration(raw.httpTimeout)
	if err != nil || timeout <= 0 {
		return 0, errors.New("invalid HTTP timeout; use a positive duration like 30s")
	}
	return timeout, nil
}

func validateTransport(raw rawConfig) (Transport, error) {
	transport := TransportStdio
	if raw.transport != "" {
		transport = Transport(strings.ToLower(raw.transport))
	}
	if transport != TransportStdio && transport != TransportHTTP {
		return "", errors.New("invalid MCP transport; use stdio or http")
	}
	return transport, nil
}

func validateHTTPBind(raw rawConfig) (string, error) {
	httpBindAddress := raw.httpBindAddress
	if httpBindAddress == "" {
		httpBindAddress = DefaultHTTPBindAddress
	}
	return NormalizeHTTPBindAddress(httpBindAddress)
}

func buildConfig(raw rawConfig, apiKey string, credentialRef CredentialReference, athleteID string, loc *time.Location, baseURL string, timeout time.Duration, transport Transport, httpBindAddress string, coachMode coach.Mode, coachConfig coach.Config) Config {
	return Config{
		APIKey:          apiKey,
		APIKeySource:    raw.apiKeySource,
		CredentialRef:   credentialRef,
		AthleteID:       athleteID,
		Timezone:        loc.String(),
		APIBaseURL:      strings.TrimRight(baseURL, "/"),
		HTTPTimeout:     timeout,
		Transport:       transport,
		HTTPBindAddress: httpBindAddress,
		DeleteMode:      safety.ParseMode(raw.deleteMode),
		Toolset:         safety.ParseToolset(raw.toolset),
		DebugMetadata:   ParseDebugMetadata(raw.debugMetadata),
		CoachMode:       coachMode,
		Coach:           coachConfig,
	}
}

// ParseDebugMetadata reports whether a raw debug metadata value enables debug output.
func ParseDebugMetadata(value string) bool {
	return strings.EqualFold(strings.TrimSpace(value), "true")
}

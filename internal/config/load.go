package config

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/coach"
	"github.com/ricardocabral/icuvisor/internal/credstore"
	"github.com/ricardocabral/icuvisor/internal/safety"
)

type fileConfig struct {
	APIKey          string              `json:"api_key"`
	CredentialRef   CredentialReference `json:"credential_ref"`
	AthleteID       string              `json:"athlete_id"`
	Timezone        string              `json:"timezone"`
	APIBaseURL      string              `json:"api_base_url"`
	HTTPTimeout     string              `json:"http_timeout"`
	Transport       string              `json:"transport"`
	HTTPBindAddress string              `json:"http_bind"`
	Coach           *coach.Config       `json:"coach"`
}

type rawConfig struct {
	apiKey          string
	apiKeySource    APIKeySource
	apiKeyLocation  string
	credentialRef   CredentialReference
	athleteID       string
	timezone        string
	apiBaseURL      string
	httpTimeout     string
	transport       string
	httpBindAddress string
	deleteMode      string
	toolset         string
	debugMetadata   string
	coachMode       string
	coach           *coach.Config
}

func load(ctx context.Context, opts Options) (Config, error) {
	if err := ctx.Err(); err != nil {
		return Config{}, err
	}

	env := opts.Env
	if env == nil {
		env = processEnv()
	}

	path := strings.TrimSpace(opts.Path)
	if path == "" {
		path = strings.TrimSpace(env[EnvConfigPath])
	}

	var raw rawConfig
	if path != "" {
		fileRaw, err := readJSONConfig(ctx, path)
		if err != nil {
			return Config{}, err
		}
		raw.merge(fileRaw, false)
		slog.Default().Info("config file loaded", "path", path)
	} else {
		slog.Default().Info("config file not used", "hint", "set --config or "+EnvConfigPath)
	}

	dotEnvPath := strings.TrimSpace(opts.DotEnvPath)
	explicitDotEnv := opts.DotEnvExplicit
	if dotEnvPath == "" {
		if envPath := strings.TrimSpace(env[EnvDotEnvPath]); envPath != "" {
			dotEnvPath = envPath
			explicitDotEnv = true
		}
	}
	if dotEnvPath == "" {
		dotEnvPath = ".env"
	}
	if dotEnv, err := readDotEnv(ctx, dotEnvPath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return Config{}, err
		}
		if explicitDotEnv {
			return Config{}, fmt.Errorf("env file %q not found; check --env-file path or %s", dotEnvPath, EnvDotEnvPath)
		}
		// Missing default .env is normal when credentials live in the keychain or config file; staying quiet here.
	} else {
		raw.merge(rawFromEnv(dotEnv, APIKeySourceFile, "env_file"), true)
		slog.Default().Info("env file loaded", "path", dotEnvPath)
	}

	processRaw := rawFromEnv(env, APIKeySourceEnv, "process_env")
	if processRaw.apiKey != "" {
		raw.merge(processRaw, false)
	} else {
		if opts.CredentialStore != nil {
			apiKey, err := opts.CredentialStore.Get(ctx, credstore.IntervalsAPIKeyAccount)
			if err != nil {
				if !errors.Is(err, credstore.ErrNotFound) {
					return Config{}, fmt.Errorf("read intervals.icu API key from OS keychain service %q account %q: %w", credstore.ServiceName, credstore.IntervalsAPIKeyAccount, err)
				}
			} else {
				raw.merge(rawConfig{apiKey: strings.TrimSpace(apiKey), apiKeySource: APIKeySourceKeychain, apiKeyLocation: "os_keychain"}, false)
				slog.Default().Info("intervals.icu API key loaded from OS keychain", "service", credstore.ServiceName, "account", credstore.IntervalsAPIKeyAccount)
			}
		}
		raw.merge(processRaw, false)
	}
	raw.merge(rawConfig{transport: strings.TrimSpace(opts.Transport), httpBindAddress: strings.TrimSpace(opts.HTTPBindAddress)}, false)
	cfg, err := validate(raw)
	if err != nil {
		return Config{}, err
	}
	warnLegacyAPIKey(cfg, raw)
	return cfg, nil
}

func readJSONConfig(ctx context.Context, path string) (rawConfig, error) {
	if err := ctx.Err(); err != nil {
		return rawConfig{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return rawConfig{}, fmt.Errorf("config file %q not found; check --config path or ICUVISOR_CONFIG", path)
		}
		return rawConfig{}, fmt.Errorf("read config file %q: %w", path, err)
	}
	if err := ctx.Err(); err != nil {
		return rawConfig{}, err
	}

	var file fileConfig
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&file); err != nil {
		return rawConfig{}, fmt.Errorf("invalid config JSON in %q; expected fields api_key, credential_ref, athlete_id, timezone, api_base_url, http_timeout, transport, http_bind, coach: %w", path, err)
	}

	apiKey := strings.TrimSpace(file.APIKey)
	apiKeySource := APIKeySourceFile
	sourceLocation := "config_json"
	if apiKey == "" {
		apiKeySource = ""
		sourceLocation = ""
	}

	return rawConfig{
		apiKey:          apiKey,
		apiKeySource:    apiKeySource,
		apiKeyLocation:  sourceLocation,
		credentialRef:   file.CredentialRef,
		athleteID:       strings.TrimSpace(file.AthleteID),
		timezone:        strings.TrimSpace(file.Timezone),
		apiBaseURL:      strings.TrimSpace(file.APIBaseURL),
		httpTimeout:     strings.TrimSpace(file.HTTPTimeout),
		transport:       strings.TrimSpace(file.Transport),
		httpBindAddress: strings.TrimSpace(file.HTTPBindAddress),
		coach:           file.Coach,
	}, nil
}
func processEnv() map[string]string {
	values := make(map[string]string)
	for _, entry := range os.Environ() {
		key, value, ok := strings.Cut(entry, "=")
		if ok {
			values[key] = value
		}
	}
	return values
}
func rawFromEnv(env map[string]string, apiKeySource APIKeySource, apiKeyLocation string) rawConfig {
	apiKey := strings.TrimSpace(env[EnvAPIKey])
	if apiKey == "" {
		apiKeySource = ""
		apiKeyLocation = ""
	}
	return rawConfig{
		apiKey:          apiKey,
		apiKeySource:    apiKeySource,
		apiKeyLocation:  apiKeyLocation,
		athleteID:       strings.TrimSpace(env[EnvAthleteID]),
		timezone:        strings.TrimSpace(env[EnvTimezone]),
		apiBaseURL:      strings.TrimSpace(env[EnvAPIBaseURL]),
		httpTimeout:     strings.TrimSpace(env[EnvHTTPTimeout]),
		transport:       strings.TrimSpace(env[EnvTransport]),
		httpBindAddress: strings.TrimSpace(env[EnvHTTPBind]),
		deleteMode:      strings.TrimSpace(env[safety.EnvDeleteMode]),
		toolset:         strings.TrimSpace(env[safety.EnvToolset]),
		debugMetadata:   strings.TrimSpace(env[EnvDebugMetadata]),
		coachMode:       strings.TrimSpace(env[EnvCoachMode]),
	}
}

func (r *rawConfig) merge(next rawConfig, absentOnly bool) {
	if !next.credentialRef.empty() && (!absentOnly || r.credentialRef.empty()) {
		r.credentialRef = next.credentialRef
	}
	if shouldSet(r.apiKey, next.apiKey, absentOnly) {
		r.apiKey = next.apiKey
		r.apiKeySource = next.apiKeySource
		r.apiKeyLocation = next.apiKeyLocation
	}
	if shouldSet(r.athleteID, next.athleteID, absentOnly) {
		r.athleteID = next.athleteID
	}
	if shouldSet(r.timezone, next.timezone, absentOnly) {
		r.timezone = next.timezone
	}
	if shouldSet(r.apiBaseURL, next.apiBaseURL, absentOnly) {
		r.apiBaseURL = next.apiBaseURL
	}
	if shouldSet(r.httpTimeout, next.httpTimeout, absentOnly) {
		r.httpTimeout = next.httpTimeout
	}
	if shouldSet(r.transport, next.transport, absentOnly) {
		r.transport = next.transport
	}
	if shouldSet(r.httpBindAddress, next.httpBindAddress, absentOnly) {
		r.httpBindAddress = next.httpBindAddress
	}
	if shouldSet(r.deleteMode, next.deleteMode, absentOnly) {
		r.deleteMode = next.deleteMode
	}
	if shouldSet(r.toolset, next.toolset, absentOnly) {
		r.toolset = next.toolset
	}
	if shouldSet(r.debugMetadata, next.debugMetadata, absentOnly) {
		r.debugMetadata = next.debugMetadata
	}
	if shouldSet(r.coachMode, next.coachMode, absentOnly) {
		r.coachMode = next.coachMode
	}
	if next.coach != nil && (!absentOnly || r.coach == nil) {
		r.coach = next.coach
	}
}
func shouldSet(current, next string, absentOnly bool) bool {
	if next == "" {
		return false
	}
	return !absentOnly || current == ""
}
func warnLegacyAPIKey(cfg Config, raw rawConfig) {
	if cfg.APIKeySource != APIKeySourceFile {
		return
	}
	slog.Default().Warn("api_key found in plaintext config; consider migrating to OS keychain", "source", raw.apiKeyLocation, "migration", "README Getting an API key")
}

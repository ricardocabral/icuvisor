package config

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"
	"unicode"
)

const (
	EnvAPIKey      = "INTERVALS_ICU_API_KEY" // #nosec G101 -- environment variable name, not a credential.
	EnvAthleteID   = "INTERVALS_ICU_ATHLETE_ID"
	EnvConfigPath  = "ICUVISOR_CONFIG"
	EnvTimezone    = "ICUVISOR_TIMEZONE"
	EnvAPIBaseURL  = "ICUVISOR_API_BASE_URL"
	EnvHTTPTimeout = "ICUVISOR_HTTP_TIMEOUT"

	DefaultAPIBaseURL  = "https://intervals.icu/api/v1"
	DefaultTimezone    = "UTC"
	DefaultHTTPTimeout = 30 * time.Second
)

// Config contains the v0.1 runtime configuration consumed by lower layers.
type Config struct {
	APIKey      string        `json:"api_key"`
	AthleteID   string        `json:"athlete_id"`
	Timezone    string        `json:"timezone"`
	APIBaseURL  string        `json:"api_base_url"`
	HTTPTimeout time.Duration `json:"-"`
}

// Options controls config loading inputs.
type Options struct {
	Path       string
	DotEnvPath string
	Env        map[string]string
}

type fileConfig struct {
	APIKey      string `json:"api_key"`
	AthleteID   string `json:"athlete_id"`
	Timezone    string `json:"timezone"`
	APIBaseURL  string `json:"api_base_url"`
	HTTPTimeout string `json:"http_timeout"`
}

type rawConfig struct {
	apiKey      string
	athleteID   string
	timezone    string
	apiBaseURL  string
	httpTimeout string
}

// Load reads v0.1 config from JSON, .env, and process environment.
func Load(ctx context.Context, opts Options) (Config, error) {
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
	}

	dotEnvPath := opts.DotEnvPath
	if dotEnvPath == "" {
		dotEnvPath = ".env"
	}
	if dotEnv, err := readDotEnv(ctx, dotEnvPath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return Config{}, err
		}
	} else {
		raw.merge(rawFromEnv(dotEnv), true)
	}

	raw.merge(rawFromEnv(env), false)
	return validate(raw)
}

// NormalizeAthleteID accepts intervals.icu athlete IDs with or without the i prefix.
func NormalizeAthleteID(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", errors.New("missing athlete ID; set INTERVALS_ICU_ATHLETE_ID or athlete_id")
	}

	if trimmed[0] == 'i' || trimmed[0] == 'I' {
		trimmed = trimmed[1:]
	}
	if trimmed == "" {
		return "", errors.New("invalid athlete ID; use 12345 or i12345")
	}
	for _, r := range trimmed {
		if !unicode.IsDigit(r) {
			return "", errors.New("invalid athlete ID; use 12345 or i12345")
		}
	}
	return "i" + trimmed, nil
}

// NormalizeAthleteIDForDisplay returns the canonical public athlete ID when possible.
func NormalizeAthleteIDForDisplay(value string) string {
	normalized, err := NormalizeAthleteID(value)
	if err != nil {
		return strings.TrimSpace(value)
	}
	return normalized
}

func (c Config) String() string {
	apiKey := "<unset>"
	if c.APIKey != "" {
		apiKey = "<redacted>"
	}
	athleteID := "<unset>"
	if c.AthleteID != "" {
		athleteID = "<set>"
	}
	return fmt.Sprintf("api_key=%s athlete_id=%s timezone=%q api_base_url=%q http_timeout=%s", apiKey, athleteID, c.Timezone, c.APIBaseURL, c.HTTPTimeout)
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
		return rawConfig{}, fmt.Errorf("invalid config JSON in %q; expected fields api_key, athlete_id, timezone, api_base_url, http_timeout: %w", path, err)
	}

	return rawConfig{
		apiKey:      strings.TrimSpace(file.APIKey),
		athleteID:   strings.TrimSpace(file.AthleteID),
		timezone:    strings.TrimSpace(file.Timezone),
		apiBaseURL:  strings.TrimSpace(file.APIBaseURL),
		httpTimeout: strings.TrimSpace(file.HTTPTimeout),
	}, nil
}

func readDotEnv(ctx context.Context, path string) (map[string]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("read .env: %w", err)
	}
	defer file.Close()

	values := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		if !recognizedEnvKey(key) {
			continue
		}
		values[key] = cleanEnvValue(value)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read .env: %w", err)
	}
	return values, nil
}

func cleanEnvValue(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 {
		quote := value[0]
		if (quote == '\'' || quote == '"') && value[len(value)-1] == quote {
			return value[1 : len(value)-1]
		}
	}
	return value
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

func rawFromEnv(env map[string]string) rawConfig {
	return rawConfig{
		apiKey:      strings.TrimSpace(env[EnvAPIKey]),
		athleteID:   strings.TrimSpace(env[EnvAthleteID]),
		timezone:    strings.TrimSpace(env[EnvTimezone]),
		apiBaseURL:  strings.TrimSpace(env[EnvAPIBaseURL]),
		httpTimeout: strings.TrimSpace(env[EnvHTTPTimeout]),
	}
}

func (r *rawConfig) merge(next rawConfig, absentOnly bool) {
	if shouldSet(r.apiKey, next.apiKey, absentOnly) {
		r.apiKey = next.apiKey
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
}

func shouldSet(current, next string, absentOnly bool) bool {
	if next == "" {
		return false
	}
	return !absentOnly || current == ""
}

func validate(raw rawConfig) (Config, error) {
	apiKey := strings.TrimSpace(raw.apiKey)
	if apiKey == "" {
		return Config{}, errors.New("missing intervals.icu API key; set INTERVALS_ICU_API_KEY or api_key")
	}

	athleteID, err := NormalizeAthleteID(raw.athleteID)
	if err != nil {
		return Config{}, err
	}

	timezone := raw.timezone
	if timezone == "" {
		timezone = DefaultTimezone
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return Config{}, fmt.Errorf("invalid timezone %q; use an IANA timezone like UTC or America/Sao_Paulo", timezone)
	}

	baseURL := raw.apiBaseURL
	if baseURL == "" {
		baseURL = DefaultAPIBaseURL
	}
	parsedURL, err := url.Parse(baseURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		return Config{}, errors.New("invalid API base URL; use an absolute http or https URL")
	}

	timeout := DefaultHTTPTimeout
	if raw.httpTimeout != "" {
		timeout, err = time.ParseDuration(raw.httpTimeout)
		if err != nil || timeout <= 0 {
			return Config{}, errors.New("invalid HTTP timeout; use a positive duration like 30s")
		}
	}

	return Config{
		APIKey:      apiKey,
		AthleteID:   athleteID,
		Timezone:    loc.String(),
		APIBaseURL:  strings.TrimRight(baseURL, "/"),
		HTTPTimeout: timeout,
	}, nil
}

func recognizedEnvKey(key string) bool {
	switch key {
	case EnvAPIKey, EnvAthleteID, EnvConfigPath, EnvTimezone, EnvAPIBaseURL, EnvHTTPTimeout:
		return true
	default:
		return false
	}
}

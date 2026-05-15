package config

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/ricardocabral/icuvisor/internal/coach"
	"github.com/ricardocabral/icuvisor/internal/credstore"
	"github.com/ricardocabral/icuvisor/internal/safety"
)

const (
	EnvAPIKey        = "INTERVALS_ICU_API_KEY" // #nosec G101 -- environment variable name, not a credential.
	EnvAthleteID     = "INTERVALS_ICU_ATHLETE_ID"
	EnvConfigPath    = "ICUVISOR_CONFIG"
	EnvTimezone      = "ICUVISOR_TIMEZONE"
	EnvAPIBaseURL    = "ICUVISOR_API_BASE_URL"
	EnvHTTPTimeout   = "ICUVISOR_HTTP_TIMEOUT"
	EnvTransport     = "ICUVISOR_TRANSPORT"
	EnvHTTPBind      = "ICUVISOR_HTTP_BIND"
	EnvDotEnvPath    = "ICUVISOR_ENV_FILE"
	EnvDebugMetadata = "ICUVISOR_DEBUG_METADATA"
	EnvCoachMode     = "ICUVISOR_COACH_MODE"

	DefaultAPIBaseURL      = "https://intervals.icu/api/v1"
	DefaultTimezone        = "UTC"
	DefaultHTTPTimeout     = 30 * time.Second
	DefaultHTTPBindAddress = "127.0.0.1:8765"
)

// Transport identifies the selected MCP transport.
type Transport string

const (
	TransportStdio Transport = "stdio"
	TransportHTTP  Transport = "http"
)

// String returns the configured transport name.
func (t Transport) String() string {
	if t == "" {
		return string(TransportStdio)
	}
	return string(t)
}

// Config contains the v0.1 runtime configuration consumed by lower layers.
type Config struct {
	APIKey          string         `json:"api_key"`
	APIKeySource    APIKeySource   `json:"-"`
	AthleteID       string         `json:"athlete_id"`
	Timezone        string         `json:"timezone"`
	APIBaseURL      string         `json:"api_base_url"`
	HTTPTimeout     time.Duration  `json:"-"`
	Transport       Transport      `json:"-"`
	HTTPBindAddress string         `json:"-"`
	DeleteMode      safety.Mode    `json:"-"`
	Toolset         safety.Toolset `json:"-"`
	DebugMetadata   bool           `json:"-"`
	CoachMode       coach.Mode     `json:"-"`
	Coach           coach.Config   `json:"coach,omitempty"`
}

// APIKeySource identifies where the loaded API key came from.
type APIKeySource string

const (
	APIKeySourceEnv      APIKeySource = "env"
	APIKeySourceKeychain APIKeySource = "keychain"
	APIKeySourceFile     APIKeySource = "file"
)

// Options controls config loading inputs.
type Options struct {
	Path            string
	DotEnvPath      string
	DotEnvExplicit  bool
	Env             map[string]string
	CredentialStore credstore.Store
	Transport       string
	HTTPBindAddress string
}

type fileConfig struct {
	APIKey          string        `json:"api_key"`
	AthleteID       string        `json:"athlete_id"`
	Timezone        string        `json:"timezone"`
	APIBaseURL      string        `json:"api_base_url"`
	HTTPTimeout     string        `json:"http_timeout"`
	Transport       string        `json:"transport"`
	HTTPBindAddress string        `json:"http_bind"`
	Coach           *coach.Config `json:"coach"`
}

type writeFileConfig struct {
	AthleteID  string `json:"athlete_id"`
	Timezone   string `json:"timezone"`
	APIBaseURL string `json:"api_base_url,omitempty"`
}

// WriteOptions controls config file writes.
type WriteOptions struct {
	AllowOverwrite bool
}

type rawConfig struct {
	apiKey          string
	apiKeySource    APIKeySource
	apiKeyLocation  string
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

// DefaultPath returns the platform default icuvisor config path.
func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("locating user config directory: %w", err)
	}
	return filepath.Join(dir, "icuvisor", "config.json"), nil
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
		slog.Default().Info("env file not found", "path", dotEnvPath)
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

// Write stores non-secret config fields as JSON.
func Write(ctx context.Context, path string, cfg Config, opts WriteOptions) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	trimmedPath := strings.TrimSpace(path)
	if trimmedPath == "" {
		var err error
		trimmedPath, err = DefaultPath()
		if err != nil {
			return err
		}
	}
	athleteID, err := NormalizeAthleteID(cfg.AthleteID)
	if err != nil {
		return fmt.Errorf("normalizing athlete ID for config write: %w", err)
	}
	timezoneName := strings.TrimSpace(cfg.Timezone)
	if timezoneName == "" {
		timezoneName = DefaultTimezone
	}
	if _, err := time.LoadLocation(timezoneName); err != nil {
		return fmt.Errorf("validating timezone for config write: %w", err)
	}
	apiBaseURL := strings.TrimSpace(cfg.APIBaseURL)
	if apiBaseURL == DefaultAPIBaseURL {
		apiBaseURL = ""
	}
	payload, err := json.MarshalIndent(writeFileConfig{AthleteID: athleteID, Timezone: timezoneName, APIBaseURL: apiBaseURL}, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding config file: %w", err)
	}
	payload = append(payload, '\n')
	if err := ctx.Err(); err != nil {
		return err
	}
	return writeConfigFile(trimmedPath, payload, opts.AllowOverwrite)
}

func writeConfigFile(path string, payload []byte, allowOverwrite bool) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create config directory %q: %w", dir, err)
	}
	tmp, err := os.CreateTemp(dir, ".icuvisor-config-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary config file: %w", err)
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("set temporary config permissions: %w", err)
	}
	if _, err := tmp.Write(payload); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temporary config file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temporary config file: %w", err)
	}
	if allowOverwrite {
		if err := os.Rename(tmpPath, path); err != nil {
			return fmt.Errorf("replace config file %q: %w", path, err)
		}
		return nil
	}
	if err := os.Link(tmpPath, path); err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("config file %q already exists; rerun setup with --force or approve overwrite prompt", path)
		}
		return fmt.Errorf("create config file %q: %w", path, err)
	}
	return nil
}

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
	return fmt.Sprintf("api_key=%s api_key_source=%s athlete_id=%s timezone=%q api_base_url=%q http_timeout=%s transport=%s http_bind=%q delete_mode=%s toolset=%s coach_mode=%s coach_enabled=%t coach_athletes=%d", apiKey, apiKeySource, athleteID, c.Timezone, c.APIBaseURL, c.HTTPTimeout, c.Transport, c.HTTPBindAddress, c.DeleteMode, c.Toolset, c.CoachMode, c.CoachModeEnabled(), len(c.Coach.Athletes))
}

// EffectiveCoachMode resolves auto against the parsed coach roster.
func (c Config) EffectiveCoachMode() coach.Mode {
	return coach.EffectiveMode(c.CoachMode, c.Coach)
}

// CoachModeEnabled reports whether coach mode is effectively on.
func (c Config) CoachModeEnabled() bool {
	return c.EffectiveCoachMode() == coach.ModeOn
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
		return rawConfig{}, fmt.Errorf("invalid config JSON in %q; expected fields api_key, athlete_id, timezone, api_base_url, http_timeout, transport, http_bind, coach: %w", path, err)
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
		athleteID:       strings.TrimSpace(file.AthleteID),
		timezone:        strings.TrimSpace(file.Timezone),
		apiBaseURL:      strings.TrimSpace(file.APIBaseURL),
		httpTimeout:     strings.TrimSpace(file.HTTPTimeout),
		transport:       strings.TrimSpace(file.Transport),
		httpBindAddress: strings.TrimSpace(file.HTTPBindAddress),
		coach:           file.Coach,
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

func validate(raw rawConfig) (Config, error) {
	apiKey := strings.TrimSpace(raw.apiKey)
	if apiKey == "" {
		return Config{}, fmt.Errorf("missing intervals.icu API key; set %s, store it in OS keychain service %q account %q, or set legacy api_key in config JSON/.env", EnvAPIKey, credstore.ServiceName, credstore.IntervalsAPIKeyAccount)
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

	transport := TransportStdio
	if raw.transport != "" {
		transport = Transport(strings.ToLower(raw.transport))
	}
	if transport != TransportStdio && transport != TransportHTTP {
		return Config{}, errors.New("invalid MCP transport; use stdio or http")
	}

	httpBindAddress := raw.httpBindAddress
	if httpBindAddress == "" {
		httpBindAddress = DefaultHTTPBindAddress
	}
	httpBindAddress, err = NormalizeHTTPBindAddress(httpBindAddress)
	if err != nil {
		return Config{}, err
	}

	coachMode, err := coach.ParseMode(raw.coachMode)
	if err != nil {
		return Config{}, err
	}
	var rawCoach coach.Config
	if raw.coach != nil {
		rawCoach = *raw.coach
	}
	coachConfig, err := coach.ValidateConfig(rawCoach, coachMode, NormalizeAthleteID)
	if err != nil {
		return Config{}, err
	}

	return Config{
		APIKey:          apiKey,
		APIKeySource:    raw.apiKeySource,
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
	}, nil
}

// ParseDebugMetadata reports whether a raw debug metadata value enables debug output.
func ParseDebugMetadata(value string) bool {
	return strings.EqualFold(strings.TrimSpace(value), "true")
}

// ValidateHTTPBindAddress rejects accidental wildcard binds and malformed ports.
func ValidateHTTPBindAddress(value string) error {
	_, err := NormalizeHTTPBindAddress(value)
	return err
}

// NormalizeHTTPBindAddress validates and returns value as a canonical netip.AddrPort string.
func NormalizeHTTPBindAddress(value string) (string, error) {
	host, port, err := splitHTTPBindAddress(value)
	if err != nil {
		return "", err
	}
	if host == "" {
		return "", errors.New("invalid HTTP bind address; use an explicit IP host and port like 127.0.0.1:8765")
	}
	portNumber, err := strconv.ParseUint(port, 10, 16)
	if err != nil || portNumber == 0 {
		return "", errors.New("invalid HTTP bind address; port must be between 1 and 65535")
	}
	addr, err := netip.ParseAddr(host)
	if err != nil {
		return "", errors.New("invalid HTTP bind address; host must be an IP address like 127.0.0.1 or 192.168.1.10")
	}
	if !addr.IsValid() {
		return "", errors.New("invalid HTTP bind address; host must be a valid IP address")
	}
	return netip.AddrPortFrom(addr, uint16(portNumber)).String(), nil
}

// HTTPBindAddressIsLoopback reports whether value binds only to a loopback IP.
func HTTPBindAddressIsLoopback(value string) bool {
	normalized, err := NormalizeHTTPBindAddress(value)
	if err != nil {
		return false
	}
	addrPort, err := netip.ParseAddrPort(normalized)
	return err == nil && addrPort.Addr().IsLoopback()
}

func splitHTTPBindAddress(value string) (string, string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", "", errors.New("invalid HTTP bind address; use host:port like 127.0.0.1:8765")
	}
	host, port, ok := strings.Cut(trimmed, ":")
	if !ok || strings.Contains(port, ":") || strings.Contains(host, ":") {
		addrPort, err := netip.ParseAddrPort(trimmed)
		if err != nil {
			return "", "", errors.New("invalid HTTP bind address; use host:port like 127.0.0.1:8765 or [::1]:8765")
		}
		return addrPort.Addr().String(), strconv.Itoa(int(addrPort.Port())), nil
	}
	return strings.TrimSpace(host), strings.TrimSpace(port), nil
}

func recognizedEnvKey(key string) bool {
	switch key {
	case EnvAPIKey, EnvAthleteID, EnvConfigPath, EnvTimezone, EnvAPIBaseURL, EnvHTTPTimeout, EnvTransport, EnvHTTPBind, EnvCoachMode, safety.EnvDeleteMode, safety.EnvToolset:
		return true
	default:
		return false
	}
}

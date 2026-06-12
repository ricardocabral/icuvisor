// Package intervals implements the intervals.icu HTTP API client.
package intervals

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ricardocabral/icuvisor/internal/config"
)

const (
	basicAuthUsername  = "API_KEY"
	defaultMaxAttempts = 3
	defaultBaseDelay   = 200 * time.Millisecond
	defaultMaxDelay    = 2 * time.Second
	defaultJitter      = 0.2

	// maxResponseBodyBytes keeps successful JSON responses bounded to avoid unbounded upstream allocations.
	maxResponseBodyBytes = 32 << 20
)

// Options configures a Client.
type Options struct {
	Config     config.Config
	Auth       AuthMode
	Version    string
	HTTPClient *http.Client
	Retry      RetryConfig
}

// OAuthBearerOptions configures an OAuth Bearer client.
type OAuthBearerOptions struct {
	AccessToken string
	APIBaseURL  string
	Version     string
	HTTPClient  *http.Client
	HTTPTimeout time.Duration
	Retry       RetryConfig
}

// AuthMode applies an intervals.icu authorization strategy to outbound requests.
type AuthMode interface {
	apply(*http.Request)
	validate() error
	kind() string
}

// AuthModeAPIKey configures intervals.icu API-key Basic Auth.
type AuthModeAPIKey struct {
	apiKey string
}

// NewAuthModeAPIKey returns an API-key Basic Auth mode with token material kept unexported.
func NewAuthModeAPIKey(apiKey string) AuthModeAPIKey {
	return AuthModeAPIKey{apiKey: strings.TrimSpace(apiKey)}
}

func (m AuthModeAPIKey) apply(req *http.Request) {
	req.SetBasicAuth(basicAuthUsername, m.apiKey)
}

func (m AuthModeAPIKey) validate() error {
	if m.apiKey == "" {
		return errors.New("missing intervals.icu API key")
	}
	return nil
}

func (m AuthModeAPIKey) kind() string { return "api_key_basic" }

func (m AuthModeAPIKey) String() string { return "intervals.AuthModeAPIKey(redacted)" }

func (m AuthModeAPIKey) GoString() string { return m.String() }

// AuthModeBearer configures intervals.icu OAuth Bearer authorization.
type AuthModeBearer struct {
	accessToken string
}

// NewAuthModeBearer returns an OAuth Bearer Auth mode with token material kept unexported.
func NewAuthModeBearer(accessToken string) AuthModeBearer {
	return AuthModeBearer{accessToken: strings.TrimSpace(accessToken)}
}

func (m AuthModeBearer) apply(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+m.accessToken)
}

func (m AuthModeBearer) validate() error {
	if m.accessToken == "" {
		return errors.New("missing intervals.icu OAuth access token")
	}
	return nil
}

func (m AuthModeBearer) kind() string { return "oauth_bearer" }

func (m AuthModeBearer) String() string { return "intervals.AuthModeBearer(redacted)" }

func (m AuthModeBearer) GoString() string { return m.String() }

// RetryConfig controls retry behavior for idempotent requests.
type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	Jitter      float64
}

// WithDefaults returns a retry config with unset fields filled from client defaults.
func (cfg RetryConfig) WithDefaults() RetryConfig {
	allFieldsUnset := cfg.MaxAttempts == 0 && cfg.BaseDelay == 0 && cfg.MaxDelay == 0 && cfg.Jitter == 0
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = defaultMaxAttempts
	}
	if cfg.BaseDelay <= 0 {
		cfg.BaseDelay = defaultBaseDelay
	}
	if cfg.MaxDelay <= 0 {
		cfg.MaxDelay = defaultMaxDelay
	}
	if cfg.Jitter < 0 {
		cfg.Jitter = 0
	}
	if allFieldsUnset {
		cfg.Jitter = defaultJitter
	}
	return cfg
}

// Client is a typed intervals.icu API client.
type Client struct {
	baseURL    *url.URL
	auth       AuthMode
	athleteID  string
	userAgent  string
	httpClient *http.Client
	retry      RetryConfig
}

// NewOAuthBearerClient builds an OAuth Bearer client routed through token-owner athlete ID 0.
func NewOAuthBearerClient(opts OAuthBearerOptions) (*Client, error) {
	httpTimeout := opts.HTTPTimeout
	if httpTimeout == 0 {
		httpTimeout = config.DefaultHTTPTimeout
	}
	return NewClient(Options{
		Config: config.Config{
			AthleteID:   "0",
			APIBaseURL:  opts.APIBaseURL,
			HTTPTimeout: httpTimeout,
		},
		Auth:       NewAuthModeBearer(opts.AccessToken),
		Version:    opts.Version,
		HTTPClient: opts.HTTPClient,
		Retry:      opts.Retry,
	})
}

// NewClient builds a Client from validated runtime configuration.
func NewClient(opts Options) (*Client, error) {
	cfg := opts.Config
	auth := opts.Auth
	if auth == nil {
		auth = NewAuthModeAPIKey(cfg.APIKey)
	}
	if err := auth.validate(); err != nil {
		return nil, err
	}

	athleteID, err := config.NormalizeAthleteID(cfg.AthleteID)
	if err != nil {
		return nil, fmt.Errorf("normalizing athlete ID: %w", err)
	}

	baseURL := strings.TrimRight(strings.TrimSpace(cfg.APIBaseURL), "/")
	if baseURL == "" {
		baseURL = config.DefaultAPIBaseURL
	}
	parsedBaseURL, err := url.Parse(baseURL)
	if err != nil || parsedBaseURL.Scheme == "" || parsedBaseURL.Host == "" || (parsedBaseURL.Scheme != "http" && parsedBaseURL.Scheme != "https") {
		return nil, errors.New("invalid intervals.icu API base URL")
	}

	version := strings.TrimSpace(opts.Version)
	if version == "" {
		version = "dev"
	}

	httpClient := opts.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: cfg.HTTPTimeout}
	}

	return &Client{
		baseURL:    parsedBaseURL,
		auth:       auth,
		athleteID:  athleteID,
		userAgent:  fmt.Sprintf("icuvisor/%s", version),
		httpClient: httpClient,
		retry:      opts.Retry.WithDefaults(),
	}, nil
}

// GetAthleteProfile retrieves the configured athlete profile with sport settings.
func (c *Client) GetAthleteProfile(ctx context.Context) (AthleteWithSportSettings, error) {
	var profile AthleteWithSportSettings
	if err := c.doJSON(ctx, &profile, "athlete", c.athleteID); err != nil {
		return AthleteWithSportSettings{}, fmt.Errorf("getting athlete profile: %w", err)
	}
	return profile, nil
}

func (c *Client) newRequest(ctx context.Context, method string, pathParts ...string) (*http.Request, error) {
	pathParts = c.resolvePathAthleteID(ctx, pathParts)
	requestURL := c.baseURL.JoinPath(pathParts...)
	req, err := http.NewRequestWithContext(ctx, method, requestURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("building intervals.icu request: %w", err)
	}
	c.auth.apply(req)
	req.Header.Set("User-Agent", c.userAgent)
	return req, nil
}

func (c *Client) ensureActivityTarget(ctx context.Context, activity Activity) error {
	targetAthleteID, ok := targetAthleteIDFromContext(ctx)
	if !ok {
		return nil
	}
	if activity.ICUAthleteID == nil {
		return ErrTargetAthleteMismatch
	}
	activityAthleteID, err := config.NormalizeAthleteID(*activity.ICUAthleteID)
	if err != nil || activityAthleteID != targetAthleteID {
		return ErrTargetAthleteMismatch
	}
	return nil
}

func (c *Client) ensureActivityIDTarget(ctx context.Context, activityID string) error {
	if _, ok := targetAthleteIDFromContext(ctx); !ok {
		return nil
	}
	activity, err := c.GetActivity(ctx, activityID)
	if err != nil {
		return err
	}
	return c.ensureActivityTarget(ctx, activity)
}

func (c *Client) resolvePathAthleteID(ctx context.Context, pathParts []string) []string {
	targetAthleteID, ok := targetAthleteIDFromContext(ctx)
	if !ok {
		return pathParts
	}
	for i := 0; i+1 < len(pathParts); i++ {
		if pathParts[i] == "athlete" && pathParts[i+1] == c.athleteID {
			out := append([]string(nil), pathParts...)
			out[i+1] = targetAthleteID
			return out
		}
	}
	return pathParts
}

func (c *Client) doJSON(ctx context.Context, out any, pathParts ...string) error {
	return c.doJSONQuery(ctx, out, nil, pathParts...)
}

func (c *Client) do(ctx context.Context, query url.Values, pathParts ...string) (*http.Response, error) {
	req, err := c.newRequest(ctx, http.MethodGet, pathParts...)
	if err != nil {
		return nil, err
	}
	if len(query) > 0 {
		req.URL.RawQuery = query.Encode()
	}
	return c.httpClient.Do(req)
}

func (c *Client) doJSONQuery(ctx context.Context, out any, query url.Values, pathParts ...string) error {
	for attempt := 1; ; attempt++ {
		resp, err := c.do(ctx, query, pathParts...)
		if err != nil {
			if retry, wait := c.decideRetry(ctx, http.MethodGet, nil, err, attempt); retry {
				if sleepErr := c.sleepBeforeRetry(ctx, wait); sleepErr != nil {
					return sleepErr
				}
				continue
			}
			return fmt.Errorf("calling intervals.icu: %w", err)
		}

		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"), time.Now())
			apiErr := errorForStatus(resp.StatusCode, retryAfter)
			retry, wait := c.decideRetry(ctx, http.MethodGet, resp, nil, attempt)
			_, _ = io.Copy(io.Discard, resp.Body)
			closeErr := resp.Body.Close()
			if retry {
				if sleepErr := c.sleepBeforeRetry(ctx, wait); sleepErr != nil {
					return sleepErr
				}
				continue
			}
			if closeErr != nil {
				return fmt.Errorf("closing intervals.icu response: %w", closeErr)
			}
			return fmt.Errorf("calling intervals.icu: %w", apiErr)
		}

		body, readErr := readBody(resp.Body)
		closeErr := resp.Body.Close()
		if readErr != nil {
			return fmt.Errorf("reading intervals.icu response: %w", readErr)
		}
		if closeErr != nil {
			return fmt.Errorf("closing intervals.icu response: %w", closeErr)
		}
		if err := json.NewDecoder(bytes.NewReader(body)).Decode(out); err != nil {
			return fmt.Errorf("decoding intervals.icu response: %w", err)
		}
		return nil
	}
}

func readBody(r io.Reader) ([]byte, error) {
	body, err := io.ReadAll(io.LimitReader(r, maxResponseBodyBytes+1))
	if err != nil {
		return nil, fmt.Errorf("reading intervals.icu response body: %w", err)
	}
	if len(body) > maxResponseBodyBytes {
		return nil, ErrResponseTooLarge
	}
	return body, nil
}

func (c *Client) decideRetry(ctx context.Context, method string, resp *http.Response, err error, attempt int) (bool, time.Duration) {
	if method == http.MethodPost || attempt >= c.retry.MaxAttempts {
		return false, 0
	}
	if err != nil {
		if ctx.Err() != nil {
			return false, 0
		}
		return true, c.retryDelay(attempt, 0)
	}
	if resp == nil {
		return false, 0
	}
	if method == http.MethodGet && ctx.Err() != nil {
		return false, 0
	}
	if resp.StatusCode != http.StatusTooManyRequests && resp.StatusCode < http.StatusInternalServerError {
		return false, 0
	}
	retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"), time.Now())
	return true, c.retryDelay(attempt, retryAfter)
}

func (c *Client) sleepBeforeRetry(ctx context.Context, wait time.Duration) error {
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return fmt.Errorf("waiting to retry intervals.icu: %w", ctx.Err())
	case <-timer.C:
		return nil
	}
}

func (c *Client) retryDelay(attempt int, retryAfter time.Duration) time.Duration {
	if retryAfter > 0 {
		if retryAfter > c.retry.MaxDelay {
			return c.retry.MaxDelay
		}
		return retryAfter
	}
	delay := c.retry.BaseDelay << max(attempt-1, 0)
	if delay > c.retry.MaxDelay {
		delay = c.retry.MaxDelay
	}
	return addJitter(delay, c.retry.Jitter)
}

func addJitter(delay time.Duration, ratio float64) time.Duration {
	if delay <= 0 || ratio <= 0 {
		return delay
	}
	span := int64(float64(delay) * ratio)
	if span <= 0 {
		return delay
	}
	offset := rand.Int64N(2*span+1) - span //nolint:gosec // Retry jitter only needs non-cryptographic randomness to decorrelate retries.
	jittered := delay + time.Duration(offset)
	if jittered <= 0 {
		return delay
	}
	return jittered
}

func parseRetryAfter(value string, now time.Time) time.Duration {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	if seconds, err := strconv.Atoi(value); err == nil {
		if seconds <= 0 {
			return 0
		}
		return time.Duration(seconds) * time.Second
	}
	when, err := http.ParseTime(value)
	if err != nil {
		return 0
	}
	if !when.After(now) {
		return 0
	}
	return when.Sub(now)
}

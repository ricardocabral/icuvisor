// Package intervals implements the intervals.icu HTTP API client.
package intervals

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
)

// Options configures a Client.
type Options struct {
	Config     config.Config
	Version    string
	HTTPClient *http.Client
	Retry      RetryConfig
}

// RetryConfig controls retry behavior for idempotent requests.
type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	Jitter      float64
}

// Client is a typed intervals.icu API client.
type Client struct {
	baseURL    *url.URL
	apiKey     string
	athleteID  string
	userAgent  string
	httpClient *http.Client
	retry      RetryConfig
}

// NewClient builds a Client from validated runtime configuration.
func NewClient(opts Options) (*Client, error) {
	cfg := opts.Config
	apiKey := strings.TrimSpace(cfg.APIKey)
	if apiKey == "" {
		return nil, errors.New("missing intervals.icu API key")
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
		apiKey:     apiKey,
		athleteID:  athleteID,
		userAgent:  fmt.Sprintf("icuvisor/%s", version),
		httpClient: httpClient,
		retry:      normalizeRetryConfig(opts.Retry),
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
	requestURL := c.baseURL.JoinPath(pathParts...)
	req, err := http.NewRequestWithContext(ctx, method, requestURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("building intervals.icu request: %w", err)
	}
	req.SetBasicAuth(basicAuthUsername, c.apiKey)
	req.Header.Set("User-Agent", c.userAgent)
	return req, nil
}

func (c *Client) doJSON(ctx context.Context, out any, pathParts ...string) error {
	return c.doJSONQuery(ctx, out, nil, pathParts...)
}

func (c *Client) doJSONQuery(ctx context.Context, out any, query url.Values, pathParts ...string) error {
	for attempt := 1; ; attempt++ {
		req, err := c.newRequest(ctx, http.MethodGet, pathParts...)
		if err == nil && len(query) > 0 {
			req.URL.RawQuery = query.Encode()
		}
		if err != nil {
			return err
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			if c.shouldRetryTransport(ctx, attempt) {
				if sleepErr := c.sleepBeforeRetry(ctx, attempt, 0); sleepErr != nil {
					return sleepErr
				}
				continue
			}
			return fmt.Errorf("calling intervals.icu: %w", err)
		}

		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"), time.Now())
			apiErr := errorForStatus(resp.StatusCode, retryAfter)
			_, _ = io.Copy(io.Discard, resp.Body)
			closeErr := resp.Body.Close()
			if c.shouldRetryStatus(resp.StatusCode, attempt) {
				if sleepErr := c.sleepBeforeRetry(ctx, attempt, retryAfter); sleepErr != nil {
					return sleepErr
				}
				continue
			}
			if closeErr != nil {
				return fmt.Errorf("closing intervals.icu response: %w", closeErr)
			}
			return fmt.Errorf("calling intervals.icu: %w", apiErr)
		}

		defer resp.Body.Close()
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("decoding intervals.icu response: %w", err)
		}
		return nil
	}
}

func normalizeRetryConfig(cfg RetryConfig) RetryConfig {
	useDefaultJitter := cfg == (RetryConfig{})
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
	if useDefaultJitter {
		cfg.Jitter = defaultJitter
	}
	return cfg
}

func (c *Client) shouldRetryTransport(ctx context.Context, attempt int) bool {
	return ctx.Err() == nil && attempt < c.retry.MaxAttempts
}

func (c *Client) shouldRetryStatus(statusCode int, attempt int) bool {
	if attempt >= c.retry.MaxAttempts {
		return false
	}
	return statusCode == http.StatusTooManyRequests || statusCode >= http.StatusInternalServerError
}

func (c *Client) sleepBeforeRetry(ctx context.Context, attempt int, retryAfter time.Duration) error {
	delay := c.retryDelay(attempt, retryAfter)
	timer := time.NewTimer(delay)
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
	offset := time.Now().UnixNano()%(2*span+1) - span
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

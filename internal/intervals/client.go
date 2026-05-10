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
	"strings"

	"github.com/ricardocabral/icuvisor/internal/config"
)

const basicAuthUsername = "API_KEY"

// Options configures a Client.
type Options struct {
	Config     config.Config
	Version    string
	HTTPClient *http.Client
}

// Client is a typed intervals.icu API client.
type Client struct {
	baseURL    *url.URL
	apiKey     string
	athleteID  string
	userAgent  string
	httpClient *http.Client
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
		return nil, err
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
	}, nil
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

func (c *Client) doJSON(ctx context.Context, method string, out any, pathParts ...string) error {
	req, err := c.newRequest(ctx, method, pathParts...)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("calling intervals.icu: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, resp.Body)
		return fmt.Errorf("intervals.icu returned HTTP %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decoding intervals.icu response: %w", err)
	}
	return nil
}

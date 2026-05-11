package intervals

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// ActivityStreamsParams contains stream query parameters.
type ActivityStreamsParams struct {
	ActivityID      string
	Types           []string
	IncludeDefaults bool
}

// ActivityStream contains one intervals.icu activity stream and preserves raw fields.
type ActivityStream struct {
	Raw map[string]any `json:"-"`

	Type             string    `json:"type"`
	Name             string    `json:"name"`
	Data             []float64 `json:"data"`
	Data2            []float64 `json:"data2"`
	ValueTypeIsArray bool      `json:"valueTypeIsArray"`
	Anomalies        []int     `json:"anomalies"`
	Custom           bool      `json:"custom"`
	AllNull          bool      `json:"allNull"`
}

// UnmarshalJSON decodes ActivityStream while retaining the original object for full responses.
func (s *ActivityStream) UnmarshalJSON(data []byte) error {
	type streamAlias ActivityStream
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	var decoded streamAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*s = ActivityStream(decoded)
	s.Raw = raw
	return nil
}

// GetActivityStreams retrieves activity streams for one activity.
func (c *Client) GetActivityStreams(ctx context.Context, params ActivityStreamsParams) ([]ActivityStream, error) {
	activityID := strings.TrimSpace(params.ActivityID)
	if activityID == "" {
		return nil, fmt.Errorf("getting activity streams: activity ID is required")
	}
	query := url.Values{}
	if len(params.Types) > 0 {
		query.Set("types", strings.Join(compactStrings(params.Types), ","))
	}
	if params.IncludeDefaults {
		query.Set("includeDefaults", "true")
	}
	var streams []ActivityStream
	if err := c.doJSONQuery(ctx, &streams, query, "activity", activityID, "streams"); err != nil {
		return nil, fmt.Errorf("getting activity %s streams: %w", activityID, err)
	}
	return streams, nil
}

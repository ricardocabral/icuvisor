package intervals

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// GetActivity retrieves one activity without embedding interval rows.
func (c *Client) GetActivity(ctx context.Context, activityID string) (Activity, error) {
	activityID = strings.TrimSpace(activityID)
	if activityID == "" {
		return Activity{}, fmt.Errorf("getting activity: activity ID is required")
	}
	query := url.Values{}
	query.Set("intervals", "false")
	var activity Activity
	if err := c.doJSONQuery(ctx, &activity, query, "activity", activityID); err != nil {
		return Activity{}, fmt.Errorf("getting activity %s: %w", activityID, err)
	}
	return activity, nil
}

// IntervalsDTO contains intervals.icu analyzed interval data for an activity.
type IntervalsDTO struct {
	Raw map[string]any `json:"-"`

	ID           string             `json:"id"`
	Analyzed     bool               `json:"analyzed"`
	ICUIntervals []ActivityInterval `json:"icu_intervals"`
	ICUGroups    []IntervalGroup    `json:"icu_groups"`
}

// UnmarshalJSON decodes IntervalsDTO while retaining the original object for full responses.
func (d *IntervalsDTO) UnmarshalJSON(data []byte) error {
	type dtoAlias IntervalsDTO
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	var decoded dtoAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*d = IntervalsDTO(decoded)
	d.Raw = raw
	return nil
}

// ActivityInterval contains stable interval fields and preserves raw upstream fields.
type ActivityInterval struct {
	Raw map[string]any `json:"-"`

	ID            string   `json:"id"`
	Name          *string  `json:"name"`
	Type          *string  `json:"type"`
	Unit          *string  `json:"unit"`
	StartIndex    *int     `json:"start_index"`
	EndIndex      *int     `json:"end_index"`
	StartTime     *string  `json:"start_time"`
	EndTime       *string  `json:"end_time"`
	StartDistance *float64 `json:"start_distance"`
	EndDistance   *float64 `json:"end_distance"`
	Distance      *float64 `json:"distance"`
	Duration      *float64 `json:"duration"`
	AveragePower  *float64 `json:"average_power"`
	AverageHR     *float64 `json:"average_hr"`
	Pace          *float64 `json:"pace"`
}

// UnmarshalJSON decodes ActivityInterval while retaining the original object for full responses.
func (i *ActivityInterval) UnmarshalJSON(data []byte) error {
	type intervalAlias ActivityInterval
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	var decoded intervalAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*i = ActivityInterval(decoded)
	i.Raw = raw
	return nil
}

// IntervalGroup contains stable interval group fields and preserves raw upstream fields.
type IntervalGroup struct {
	Raw map[string]any `json:"-"`

	ID         string  `json:"id"`
	Name       *string `json:"name"`
	Type       *string `json:"type"`
	StartIndex *int    `json:"start_index"`
	EndIndex   *int    `json:"end_index"`
}

// UnmarshalJSON decodes IntervalGroup while retaining the original object for full responses.
func (g *IntervalGroup) UnmarshalJSON(data []byte) error {
	type groupAlias IntervalGroup
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	var decoded groupAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*g = IntervalGroup(decoded)
	g.Raw = raw
	return nil
}

// GetActivityIntervals retrieves analyzed intervals for one activity.
func (c *Client) GetActivityIntervals(ctx context.Context, activityID string) (IntervalsDTO, error) {
	activityID = strings.TrimSpace(activityID)
	if activityID == "" {
		return IntervalsDTO{}, fmt.Errorf("getting activity intervals: activity ID is required")
	}
	var intervals IntervalsDTO
	if err := c.doJSON(ctx, &intervals, "activity", activityID, "intervals"); err != nil {
		return IntervalsDTO{}, fmt.Errorf("getting activity %s intervals: %w", activityID, err)
	}
	return intervals, nil
}

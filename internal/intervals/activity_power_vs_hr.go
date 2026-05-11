package intervals

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// PowerVsHR contains power-vs-heart-rate curve summary fields.
type PowerVsHR struct {
	Raw map[string]any `json:"-"`

	PowerHR       *float64 `json:"powerHr"`
	PowerHRFirst  *float64 `json:"powerHrFirst"`
	PowerHRSecond *float64 `json:"powerHrSecond"`
	Decoupling    *float64 `json:"decoupling"`
	PowerHRZ2     *float64 `json:"powerHrZ2"`
}

// UnmarshalJSON decodes PowerVsHR while retaining the original object for full responses.
func (p *PowerVsHR) UnmarshalJSON(data []byte) error {
	type powerVsHRAlias PowerVsHR
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	var decoded powerVsHRAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*p = PowerVsHR(decoded)
	p.Raw = raw
	return nil
}

// GetActivityPowerVsHR retrieves power-vs-heart-rate data for one activity.
func (c *Client) GetActivityPowerVsHR(ctx context.Context, activityID string) (PowerVsHR, error) {
	activityID = strings.TrimSpace(activityID)
	if activityID == "" {
		return PowerVsHR{}, fmt.Errorf("getting activity power-vs-hr: activity ID is required")
	}
	var payload PowerVsHR
	if err := c.doJSON(ctx, &payload, "activity", activityID, "power-vs-hr.json"); err != nil {
		return PowerVsHR{}, fmt.Errorf("getting activity %s power-vs-hr: %w", activityID, err)
	}
	return payload, nil
}

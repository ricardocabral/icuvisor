package intervals

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// CustomItem contains an intervals.icu custom chart/field/zones item and preserves raw upstream fields.
type CustomItem struct {
	Raw map[string]any `json:"-"`

	ID          string  `json:"-"`
	AthleteID   *string `json:"athlete_id"`
	Type        *string `json:"type"`
	Visibility  *string `json:"visibility"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Image       *string `json:"image"`
	Content     any     `json:"content"`
	UsageCount  *int    `json:"usage_count"`
	Index       *int    `json:"index"`
	HideScript  *bool   `json:"hide_script"`
	HiddenByID  *string `json:"hidden_by_id"`
	Updated     *string `json:"updated"`
	FromID      *int    `json:"from_id"`
	FromAthlete any     `json:"from_athlete"`
}

// UnmarshalJSON decodes CustomItem while retaining the original object for full responses.
func (i *CustomItem) UnmarshalJSON(data []byte) error {
	type customItemAlias CustomItem
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	var decoded customItemAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*i = CustomItem(decoded)
	i.Raw = raw
	i.ID = rawIDString(raw["id"])
	return nil
}

// ListCustomItems lists custom charts, fields, streams, panels, and zones for the configured athlete.
func (c *Client) ListCustomItems(ctx context.Context) ([]CustomItem, error) {
	var items []CustomItem
	if err := c.doJSON(ctx, &items, "athlete", c.athleteID, "custom-item"); err != nil {
		return nil, fmt.Errorf("listing custom items: %w", err)
	}
	return items, nil
}

// GetCustomItem retrieves one custom item for the configured athlete.
func (c *Client) GetCustomItem(ctx context.Context, itemID string) (CustomItem, error) {
	itemID = strings.TrimSpace(itemID)
	if itemID == "" {
		return CustomItem{}, fmt.Errorf("getting custom item: item ID is required")
	}
	var item CustomItem
	if err := c.doJSON(ctx, &item, "athlete", c.athleteID, "custom-item", itemID); err != nil {
		return CustomItem{}, fmt.Errorf("getting custom item %s: %w", itemID, err)
	}
	return item, nil
}

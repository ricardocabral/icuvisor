package intervals

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Gear contains stable gear fields and preserves raw upstream fields.
type Gear struct {
	Raw map[string]any `json:"-"`

	ID          string  `json:"-"`
	Name        *string `json:"name"`
	Type        *string `json:"type"`
	Brand       *string `json:"brand"`
	Model       *string `json:"model"`
	Description *string `json:"description"`
	Retired     *bool   `json:"retired"`
}

// UnmarshalJSON decodes Gear while retaining the original object for terse delete echoes.
func (g *Gear) UnmarshalJSON(data []byte) error {
	type gearAlias Gear
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	var decoded gearAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*g = Gear(decoded)
	g.Raw = raw
	g.ID = rawIDString(raw["id"])
	return nil
}

// DeleteEvent deletes one calendar event for the configured athlete.
func (c *Client) DeleteEvent(ctx context.Context, eventID string) error {
	eventID = strings.TrimSpace(eventID)
	if eventID == "" {
		return fmt.Errorf("deleting event: event ID is required")
	}
	if err := c.doNoJSON(ctx, http.MethodDelete, "athlete", c.athleteID, "events", eventID); err != nil {
		return fmt.Errorf("deleting event %s: %w", eventID, err)
	}
	return nil
}

// DeleteActivity deletes one activity for the configured athlete.
func (c *Client) DeleteActivity(ctx context.Context, activityID string) error {
	activityID = strings.TrimSpace(activityID)
	if activityID == "" {
		return fmt.Errorf("deleting activity: activity ID is required")
	}
	if err := c.doNoJSON(ctx, http.MethodDelete, "activity", activityID); err != nil {
		return fmt.Errorf("deleting activity %s: %w", activityID, err)
	}
	return nil
}

// DeleteCustomItem deletes one custom item for the configured athlete.
func (c *Client) DeleteCustomItem(ctx context.Context, itemID string) error {
	itemID = strings.TrimSpace(itemID)
	if itemID == "" {
		return fmt.Errorf("deleting custom item: item ID is required")
	}
	if err := c.doNoJSON(ctx, http.MethodDelete, "athlete", c.athleteID, "custom-item", itemID); err != nil {
		return fmt.Errorf("deleting custom item %s: %w", itemID, err)
	}
	return nil
}

// DeleteSportSettings deletes one sport-settings definition for the configured athlete.
func (c *Client) DeleteSportSettings(ctx context.Context, sportSettingsID string) error {
	sportSettingsID = strings.TrimSpace(sportSettingsID)
	if sportSettingsID == "" {
		return fmt.Errorf("deleting sport settings: sport-settings ID is required")
	}
	if err := c.doNoJSON(ctx, http.MethodDelete, "athlete", c.athleteID, "sport-settings", sportSettingsID); err != nil {
		return fmt.Errorf("deleting sport settings %s: %w", sportSettingsID, err)
	}
	return nil
}

// GetGear retrieves one gear item for the configured athlete.
func (c *Client) GetGear(ctx context.Context, gearID string) (Gear, error) {
	gearID = strings.TrimSpace(gearID)
	if gearID == "" {
		return Gear{}, fmt.Errorf("getting gear: gear ID is required")
	}
	var gear Gear
	if err := c.doJSON(ctx, &gear, "athlete", c.athleteID, "gear", gearID); err != nil {
		return Gear{}, fmt.Errorf("getting gear %s: %w", gearID, err)
	}
	return gear, nil
}

// DeleteGear deletes one gear item for the configured athlete.
func (c *Client) DeleteGear(ctx context.Context, gearID string) error {
	gearID = strings.TrimSpace(gearID)
	if gearID == "" {
		return fmt.Errorf("deleting gear: gear ID is required")
	}
	if err := c.doNoJSON(ctx, http.MethodDelete, "athlete", c.athleteID, "gear", gearID); err != nil {
		return fmt.Errorf("deleting gear %s: %w", gearID, err)
	}
	return nil
}

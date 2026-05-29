package intervals

import (
	"context"
	"fmt"
	"strings"
)

// DeleteEvent deletes one calendar event for the configured athlete.
func (c *Client) DeleteEvent(ctx context.Context, eventID string) error {
	eventID = strings.TrimSpace(eventID)
	if eventID == "" {
		return fmt.Errorf("deleting event: event ID is required")
	}
	if err := c.doNoJSON(ctx, "athlete", c.athleteID, "events", eventID); err != nil {
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
	if err := c.ensureActivityIDTarget(ctx, activityID); err != nil {
		return fmt.Errorf("deleting activity %s: %w", activityID, err)
	}
	if err := c.doNoJSON(ctx, "activity", activityID, "tombstone"); err != nil {
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
	if err := c.doNoJSON(ctx, "athlete", c.athleteID, "custom-item", itemID); err != nil {
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
	if err := c.doNoJSON(ctx, "athlete", c.athleteID, "sport-settings", sportSettingsID); err != nil {
		return fmt.Errorf("deleting sport settings %s: %w", sportSettingsID, err)
	}
	return nil
}

// DeleteGear deletes one gear item for the configured athlete.
func (c *Client) DeleteGear(ctx context.Context, gearID string) error {
	gearID = strings.TrimSpace(gearID)
	if gearID == "" {
		return fmt.Errorf("deleting gear: gear ID is required")
	}
	if err := c.doNoJSON(ctx, "athlete", c.athleteID, "gear", gearID); err != nil {
		return fmt.Errorf("deleting gear %s: %w", gearID, err)
	}
	return nil
}

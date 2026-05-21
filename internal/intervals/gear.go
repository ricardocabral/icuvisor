package intervals

import (
	"context"
	"encoding/json"
	"fmt"
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

// UnmarshalJSON decodes Gear while retaining the original object for read tools and terse delete echoes.
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

// ListGear retrieves all gear items for the configured athlete.
func (c *Client) ListGear(ctx context.Context) ([]Gear, error) {
	var gear []Gear
	if err := c.doJSON(ctx, &gear, "athlete", c.athleteID, "gear"); err != nil {
		return nil, fmt.Errorf("listing gear: %w", err)
	}
	return gear, nil
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

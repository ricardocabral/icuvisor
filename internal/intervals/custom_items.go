package intervals

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// WriteCustomItemParams contains writable custom-item fields.
type WriteCustomItemParams struct {
	ItemID         string
	ItemType       string
	Name           string
	NameSet        bool
	Visibility     *string
	VisibilitySet  bool
	Description    *string
	DescriptionSet bool
	Image          *string
	ImageSet       bool
	Index          *int
	IndexSet       bool
	HideScript     *bool
	HideScriptSet  bool
	Content        map[string]any
	ContentSet     bool
}

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

// CreateCustomItem creates a custom item for the configured athlete.
func (c *Client) CreateCustomItem(ctx context.Context, params WriteCustomItemParams) (CustomItem, error) {
	body, err := writeCustomItemBody(params, false)
	if err != nil {
		return CustomItem{}, err
	}
	var item CustomItem
	if err := c.doJSONBody(ctx, http.MethodPost, body, &item, "athlete", c.athleteID, "custom-item"); err != nil {
		return CustomItem{}, fmt.Errorf("creating custom item: %w", err)
	}
	return item, nil
}

// UpdateCustomItem sparsely updates a custom item for the configured athlete.
func (c *Client) UpdateCustomItem(ctx context.Context, params WriteCustomItemParams) (CustomItem, error) {
	itemID := strings.TrimSpace(params.ItemID)
	if itemID == "" {
		return CustomItem{}, fmt.Errorf("updating custom item: item ID is required")
	}
	body, err := writeCustomItemBody(params, true)
	if err != nil {
		return CustomItem{}, err
	}
	var item CustomItem
	if err := c.doJSONBody(ctx, http.MethodPut, body, &item, "athlete", c.athleteID, "custom-item", itemID); err != nil {
		return CustomItem{}, fmt.Errorf("updating custom item %s: %w", itemID, err)
	}
	return item, nil
}

func writeCustomItemBody(params WriteCustomItemParams, allowSparse bool) (map[string]any, error) {
	body := map[string]any{}
	itemType := strings.TrimSpace(params.ItemType)
	name := strings.TrimSpace(params.Name)
	if !allowSparse {
		if itemType == "" {
			return nil, fmt.Errorf("writing custom item: item_type is required")
		}
		if name == "" {
			return nil, fmt.Errorf("writing custom item: name is required")
		}
		if params.Content == nil {
			return nil, fmt.Errorf("writing custom item: content is required")
		}
		body["type"] = itemType
		body["name"] = name
		body["content"] = cloneWriteCustomItemMap(params.Content)
	} else {
		if params.NameSet {
			if name == "" {
				return nil, fmt.Errorf("writing custom item: name cannot be empty")
			}
			body["name"] = name
		}
		if params.ContentSet {
			if params.Content == nil {
				return nil, fmt.Errorf("writing custom item: content cannot be null")
			}
			body["content"] = cloneWriteCustomItemMap(params.Content)
		}
	}
	if params.VisibilitySet {
		if params.Visibility == nil {
			return nil, fmt.Errorf("writing custom item: visibility cannot be null")
		}
		body["visibility"] = strings.TrimSpace(*params.Visibility)
	}
	if params.DescriptionSet {
		if params.Description == nil {
			return nil, fmt.Errorf("writing custom item: description cannot be null")
		}
		body["description"] = *params.Description
	}
	if params.ImageSet {
		if params.Image == nil {
			return nil, fmt.Errorf("writing custom item: image cannot be null")
		}
		body["image"] = strings.TrimSpace(*params.Image)
	}
	if params.IndexSet {
		if params.Index == nil {
			return nil, fmt.Errorf("writing custom item: index cannot be null")
		}
		body["index"] = *params.Index
	}
	if params.HideScriptSet {
		if params.HideScript == nil {
			return nil, fmt.Errorf("writing custom item: hide_script cannot be null")
		}
		body["hide_script"] = *params.HideScript
	}
	if len(body) == 0 {
		return nil, fmt.Errorf("writing custom item: at least one field is required")
	}
	return body, nil
}

func cloneWriteCustomItemMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

package tools

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

const (
	getCustomItemsName                    = "get_custom_items"
	getCustomItemsDescription             = "List custom item definitions such as charts, fields, streams, panels, histograms, maps, and zones. Returns terse rows with id, name, and item_type; use get_custom_item_by_id for the full content payload."
	invalidGetCustomItemsArgumentsMessage = "invalid get_custom_items arguments; only item_type is supported"
	fetchCustomItemsMessage               = "could not fetch custom items; check intervals.icu credentials and athlete ID"
	customItemsEndpoint                   = "/athlete/{id}/custom-item"
)

// CustomItemsClient retrieves custom items for tools.
type CustomItemsClient interface {
	ListCustomItems(context.Context) ([]intervals.CustomItem, error)
	GetCustomItem(context.Context, string) (intervals.CustomItem, error)
}

type getCustomItemsRequest struct {
	ItemType string `json:"item_type,omitempty"`
}

type getCustomItemsResponse struct {
	CustomItems []customItemRow    `json:"custom_items"`
	Meta        getCustomItemsMeta `json:"_meta"`
}

type customItemRow struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	ItemType    string `json:"item_type,omitempty"`
	Visibility  string `json:"visibility,omitempty"`
	Description string `json:"description,omitempty"`
	UsageCount  int    `json:"usage_count,omitempty"`
	Index       int    `json:"index,omitempty"`
	Updated     string `json:"updated,omitempty"`
}

type getCustomItemsMeta struct {
	SourceEndpoint      string         `json:"source_endpoint"`
	Count               int            `json:"count"`
	ItemTypeFilter      string         `json:"item_type_filter,omitempty"`
	CountsByItemType    map[string]int `json:"counts_by_item_type,omitempty"`
	DefaultPayloadScope string         `json:"default_payload_scope"`
}

func newGetCustomItemsTool(client CustomItemsClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool) Tool {
	return Tool{Name: getCustomItemsName, Description: getCustomItemsDescription, InputSchema: getCustomItemsInputSchema(), OutputSchema: getCustomItemsOutputSchema(), Handler: getCustomItemsHandler(client, profileClient, version, timezoneFallback, debugMetadata)}
}

func getCustomItemsHandler(client CustomItemsClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeGetCustomItemsRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidGetCustomItemsArgumentsMessage, err)
		}
		unitSystem, _, err := toolProfile(ctx, profileClient, timezoneFallback)
		if err != nil {
			return Result{}, NewUserError(fetchCustomItemsMessage, err)
		}
		if client == nil {
			return Result{}, NewUserError(fetchCustomItemsMessage, errors.New("missing custom items client"))
		}
		items, err := client.ListCustomItems(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(fetchCustomItemsMessage, err)
		}
		payload := shapeGetCustomItemsResponse(items, args)
		return encodeShaped(payload, false, []string{"custom_items"}, version, debugMetadata, getCustomItemsName, unitSystem)
	}
}

func decodeGetCustomItemsRequest(raw json.RawMessage) (getCustomItemsRequest, error) {
	var args getCustomItemsRequest
	if len(strings.TrimSpace(string(raw))) == 0 {
		return args, nil
	}
	if err := decodeStrict(raw, &args); err != nil {
		return args, err
	}
	args.ItemType = strings.TrimSpace(args.ItemType)
	return args, nil
}

func shapeGetCustomItemsResponse(items []intervals.CustomItem, args getCustomItemsRequest) getCustomItemsResponse {
	rows := make([]customItemRow, 0, len(items))
	counts := map[string]int{}
	for _, item := range items {
		itemType := customItemType(item)
		counts[itemType]++
		if args.ItemType != "" && itemType != args.ItemType {
			continue
		}
		rows = append(rows, customItemToRow(item))
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].ItemType != rows[j].ItemType {
			return rows[i].ItemType < rows[j].ItemType
		}
		if rows[i].Index != rows[j].Index {
			return rows[i].Index < rows[j].Index
		}
		if rows[i].Name != rows[j].Name {
			return rows[i].Name < rows[j].Name
		}
		return rows[i].ID < rows[j].ID
	})
	return getCustomItemsResponse{CustomItems: rows, Meta: getCustomItemsMeta{SourceEndpoint: customItemsEndpoint, Count: len(rows), ItemTypeFilter: args.ItemType, CountsByItemType: counts, DefaultPayloadScope: "terse custom-item index rows only; full per-item content requires get_custom_item_by_id"}}
}

func customItemToRow(item intervals.CustomItem) customItemRow {
	return customItemRow{ID: item.ID, Name: stringValue(item.Name), ItemType: customItemType(item), Visibility: stringValue(item.Visibility), Description: stringValue(item.Description), UsageCount: intValue(item.UsageCount), Index: intValue(item.Index), Updated: stringValue(item.Updated)}
}

func customItemType(item intervals.CustomItem) string {
	return firstNonEmpty(stringValue(item.Type), anyString(item.Raw["type"]), anyString(item.Raw["item_type"]))
}

func getCustomItemsInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "properties": map[string]any{
		"item_type": map[string]any{"type": "string", "description": "Optional upstream custom-item type filter, for example FITNESS_CHART, INPUT_FIELD, ACTIVITY_FIELD, ACTIVITY_STREAM, ACTIVITY_PANEL, ACTIVITY_HISTOGRAM, ACTIVITY_MAP, ACTIVITY_HEATMAP, TRACE_CHART, FITNESS_TABLE, or ZONES."},
	}}
}

func getCustomItemsOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Terse custom-item rows with id, name, item_type, and metadata counts by item type."}
}

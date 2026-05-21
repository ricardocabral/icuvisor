package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

const (
	getGearListName                    = "get_gear_list"
	getGearListDescription             = "List the athlete's intervals.icu gear with IDs and human-readable names for resolving activity gear_id fields. This read-only tool is available in the full toolset; use refresh=true to bypass the in-process per-athlete cache."
	invalidGetGearListArgumentsMessage = "invalid get_gear_list arguments; provide optional refresh and include_full booleans only"
	fetchGearListMessage               = "could not fetch gear list; check intervals.icu credentials and athlete ID"
	gearListEndpoint                   = "/athlete/{id}/gear"
)

// GearListClient retrieves gear for read tools.
type GearListClient interface {
	ListGear(context.Context) ([]intervals.Gear, error)
}

type getGearListRequest struct {
	Refresh     bool `json:"refresh,omitempty"`
	IncludeFull bool `json:"include_full,omitempty"`
}

type getGearListResponse struct {
	Gear []gearListRow `json:"gear"`
	Meta gearListMeta  `json:"_meta"`
}

type gearListRow struct {
	GearID      string         `json:"gear_id,omitempty"`
	Name        string         `json:"name,omitempty"`
	NameMissing bool           `json:"name_missing,omitempty"`
	Type        string         `json:"type,omitempty"`
	Brand       string         `json:"brand,omitempty"`
	Model       string         `json:"model,omitempty"`
	Retired     *bool          `json:"retired,omitempty"`
	Full        map[string]any `json:"full,omitempty"`
}

type gearListMeta struct {
	SourceEndpoint string `json:"source_endpoint"`
	Count          int    `json:"count"`
	UnnamedCount   int    `json:"unnamed_count"`
	Cached         bool   `json:"cached"`
	Refreshed      bool   `json:"refreshed"`
	CachePolicy    string `json:"cache_policy"`
	IncludeFull    bool   `json:"include_full"`
}

func newGetGearListTool(client GearListClient, cache *gearListCache, version string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return fullTool(Tool{Name: getGearListName, Description: getGearListDescription, InputSchema: getGearListInputSchema(), OutputSchema: getGearListOutputSchema(), Requirement: RequirementRead, Handler: getGearListHandler(client, cache, version, debugMetadata, shapeCfg)})
}

func getGearListHandler(client GearListClient, cache *gearListCache, version string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeGetGearListRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidGetGearListArgumentsMessage, err)
		}
		if client == nil {
			return Result{}, NewUserError(fetchGearListMessage, errors.New("missing gear list client"))
		}
		cacheKey, err := gearCacheKey(ctx)
		if err != nil {
			return Result{}, NewUserError(fetchGearListMessage, err)
		}
		result, err := cache.get(ctx, cacheKey, args.Refresh, client.ListGear)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(fetchGearListMessage, err)
		}
		payload := shapeGetGearListResponse(result.gear, args.IncludeFull, result.cached, result.refreshed)
		return encodeShaped(payload, args.IncludeFull, []string{"gear"}, version, debugMetadata, getGearListName, "", shapeCfg)
	}
}

func decodeGetGearListRequest(raw json.RawMessage) (getGearListRequest, error) {
	if strings.TrimSpace(string(raw)) == "" {
		return getGearListRequest{}, nil
	}
	return DecodeStrict[getGearListRequest](raw)
}

func shapeGetGearListResponse(gear []intervals.Gear, includeFull bool, cached bool, refreshed bool) getGearListResponse {
	rows := make([]gearListRow, 0, len(gear))
	unnamed := 0
	for _, item := range gear {
		row := gearListToRow(item, includeFull)
		if row.NameMissing {
			unnamed++
		}
		rows = append(rows, row)
	}
	return getGearListResponse{Gear: rows, Meta: gearListMeta{SourceEndpoint: gearListEndpoint, Count: len(rows), UnnamedCount: unnamed, Cached: cached, Refreshed: refreshed, CachePolicy: "manual_refresh_only", IncludeFull: includeFull}}
}

func gearListToRow(gear intervals.Gear, includeFull bool) gearListRow {
	row := gearListRow{GearID: gear.ID, Name: strings.TrimSpace(stringValue(gear.Name)), Type: stringValue(gear.Type), Brand: stringValue(gear.Brand), Model: stringValue(gear.Model), Retired: gear.Retired}
	if row.Name == "" {
		row.NameMissing = true
	}
	if includeFull {
		row.Full = gear.Raw
	}
	return row
}

func getGearListInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "properties": map[string]any{
		"refresh":      map[string]any{"type": "boolean", "default": false, "description": "When true, bypass the manual in-process per-athlete gear cache and fetch a fresh list from intervals.icu. Failed refreshes do not replace the last successful cache entry."},
		"include_full": map[string]any{"type": "boolean", "default": false, "description": "When true, include raw upstream gear fields under each row's full object; default terse rows include gear_id, name/name_missing, type, brand, model, and retired."},
	}}
}

func getGearListOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Full-toolset gear list with terse gear rows, explicit name_missing signals, and _meta cache/count fields."}
}

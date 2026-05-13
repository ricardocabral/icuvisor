package tools

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/response"
)

type deleteResourceResponse struct {
	DeletedID string             `json:"deleted_id"`
	Status    string             `json:"status"`
	Meta      deleteResourceMeta `json:"_meta"`
}

type deleteResourceMeta struct {
	Operation      string         `json:"operation"`
	ResourceType   string         `json:"resource_type"`
	SourceEndpoint string         `json:"source_endpoint"`
	Deleted        map[string]any `json:"deleted"`
}

func newDeleteResourceResponse(deletedID string, resourceType string, sourceEndpoint string, before map[string]any) deleteResourceResponse {
	return deleteResourceResponse{DeletedID: deletedID, Status: "deleted", Meta: deleteResourceMeta{Operation: "delete", ResourceType: resourceType, SourceEndpoint: sourceEndpoint, Deleted: before}}
}

func structToJSONMap(value any) (map[string]any, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func eventDeleteEcho(event intervals.Event, eventID string, timezoneName string) (map[string]any, error) {
	if event.ID == "" {
		event.ID = eventID
	}
	row, err := eventRow(event, false, timezoneName)
	if err != nil {
		return nil, err
	}
	out, err := structToJSONMap(row)
	if err != nil {
		return nil, fmt.Errorf("encoding deleted event echo: %w", err)
	}
	return out, nil
}

func activityDeleteEcho(activity intervals.Activity, activityID string, timezoneName string, unitSystem response.UnitSystem) (map[string]any, error) {
	if strings.TrimSpace(activity.ID) == "" {
		activity.ID = activityID
	}
	row := activityRow(activity, false, timezoneName, unitSystem)
	out, err := structToJSONMap(row)
	if err != nil {
		return nil, fmt.Errorf("encoding deleted activity echo: %w", err)
	}
	return out, nil
}

func customItemDeleteEcho(item intervals.CustomItem, itemID string) (map[string]any, error) {
	if item.ID == "" {
		item.ID = itemID
	}
	out, err := structToJSONMap(customItemToRow(item))
	if err != nil {
		return nil, fmt.Errorf("encoding deleted custom item echo: %w", err)
	}
	return out, nil
}

func sportSettingsDeleteEcho(setting intervals.SportSettings, sportSettingsID string) map[string]any {
	out := map[string]any{"sport_settings_id": sportSettingsID}
	if setting.ID > 0 {
		out["sport_settings_id"] = fmt.Sprint(setting.ID)
	}
	if strings.TrimSpace(setting.Type) != "" {
		out["sport"] = setting.Type
	}
	if setting.FTP > 0 {
		out["ftp_watts"] = setting.FTP
	}
	if setting.LTHR > 0 {
		out["threshold_hr_bpm"] = setting.LTHR
	} else if setting.FTHR > 0 {
		out["threshold_hr_bpm"] = setting.FTHR
	}
	if setting.ThresholdPace > 0 {
		out["threshold_pace"] = setting.ThresholdPace
	}
	if strings.TrimSpace(setting.PaceUnits) != "" {
		out["pace_units"] = setting.PaceUnits
	}
	return out
}

func gearDeleteEcho(gear intervals.Gear, gearID string) map[string]any {
	id := firstNonEmpty(gear.ID, gearID)
	out := map[string]any{"gear_id": id}
	if value := stringValue(gear.Name); value != "" {
		out["name"] = value
	}
	if value := stringValue(gear.Type); value != "" {
		out["type"] = value
	}
	if value := stringValue(gear.Brand); value != "" {
		out["brand"] = value
	}
	if value := stringValue(gear.Model); value != "" {
		out["model"] = value
	}
	if gear.Retired != nil {
		out["retired"] = *gear.Retired
	}
	return out
}

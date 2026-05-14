package customitemschemas

// FamilyDescriptor describes one custom-item content-shape family.
type FamilyDescriptor struct {
	Key         string
	Title       string
	Description string
	Items       []ItemTypeDescriptor
}

// ItemTypeDescriptor describes one concrete upstream custom item type.
type ItemTypeDescriptor struct {
	ItemType         string
	Description      string
	Sample           map[string]any
	SharesSchemaWith string
}

// Families returns static custom-item content samples used for documentation.
func Families() []FamilyDescriptor {
	families := []FamilyDescriptor{
		{
			Key:         "charts_tables_traces",
			Title:       "Charts, tables, traces, maps, histograms, and heatmaps",
			Description: "Display-oriented custom items define series/traces, formulas or fields, and layout/display options.",
			Items: []ItemTypeDescriptor{
				{ItemType: "FITNESS_CHART", Description: "Fitness page chart with series, axes, filters, and layout.", Sample: chartSample("ctl")},
				{ItemType: "FITNESS_TABLE", Description: "Fitness page table with columns and filters.", Sample: tableSample()},
				{ItemType: "TRACE_CHART", Description: "Trace chart with stream or field series.", Sample: chartSample("heart_rate")},
				{ItemType: "ACTIVITY_CHART", Description: "Activity detail chart with per-activity series.", Sample: chartSample("power")},
				{ItemType: "ACTIVITY_HISTOGRAM", Description: "Activity histogram with bucketed metric bins.", Sample: histogramSample()},
				{ItemType: "ACTIVITY_HEATMAP", Description: "Activity heatmap with x/y fields and color scale.", Sample: heatmapSample()},
				{ItemType: "ACTIVITY_MAP", Description: "Activity map overlay with path and metric coloring.", Sample: mapSample()},
			},
		},
		{
			Key:         "fields_streams",
			Title:       "Input fields, activity fields, interval fields, and streams",
			Description: "Field and stream items describe custom values, scripts/formulas, units, formats, and visibility.",
			Items: []ItemTypeDescriptor{
				{ItemType: "INPUT_FIELD", Description: "Athlete-entered custom field definition.", Sample: fieldSample("travel_fatigue", "number")},
				{ItemType: "ACTIVITY_FIELD", Description: "Computed or entered activity-level custom field.", Sample: fieldSample("fueling_score", "number")},
				{ItemType: "INTERVAL_FIELD", Description: "Interval-level custom field definition.", Sample: fieldSample("interval_note", "text")},
				{ItemType: "ACTIVITY_STREAM", Description: "Custom activity stream definition with samples or script metadata.", Sample: streamSample()},
			},
		},
		{
			Key:         "panels",
			Title:       "Activity panels",
			Description: "Panel items group metrics, labels, and display widgets for activity detail pages.",
			Items:       []ItemTypeDescriptor{{ItemType: "ACTIVITY_PANEL", Description: "Activity detail panel with widgets and layout.", Sample: panelSample()}},
		},
		{
			Key:         "zones",
			Title:       "Zones",
			Description: "Zone items define named ranges and display colors for a metric.",
			Items:       []ItemTypeDescriptor{{ItemType: "ZONES", Description: "Named metric zone ranges and colors.", Sample: zonesSample()}},
		},
	}
	out := make([]FamilyDescriptor, len(families))
	copy(out, families)
	return out
}

func chartSample(field string) map[string]any {
	return map[string]any{
		"series": []any{map[string]any{"field": field, "label": "Primary", "color": "blue", "formula": field}},
		"layout": map[string]any{"height": 240, "width": 600},
		"axes":   map[string]any{"left": map[string]any{"label": "Value"}},
		"filters": map[string]any{
			"sport": "Ride",
		},
	}
}

func tableSample() map[string]any {
	return map[string]any{
		"columns": []any{map[string]any{"field": "ctl", "label": "Fitness", "format": "0"}},
		"sort":    map[string]any{"field": "date", "direction": "desc"},
		"filters": map[string]any{"sport": "Ride"},
	}
}

func histogramSample() map[string]any {
	return map[string]any{
		"field":       "power",
		"bucket_size": 25,
		"units":       "W",
		"colors":      []any{"blue", "green", "orange"},
	}
}

func heatmapSample() map[string]any {
	return map[string]any{
		"x_field": "cadence",
		"y_field": "power",
		"color":   map[string]any{"field": "time", "scale": "viridis"},
	}
}

func mapSample() map[string]any {
	return map[string]any{
		"path":    map[string]any{"source": "latlng"},
		"overlay": map[string]any{"field": "power", "palette": "thermal"},
		"markers": []any{map[string]any{"field": "lap", "label": "Lap"}},
	}
}

func fieldSample(field string, valueType string) map[string]any {
	return map[string]any{
		"field":      field,
		"label":      "Custom field",
		"type":       valueType,
		"units":      "score",
		"format":     "0.0",
		"script":     "return input",
		"visibility": "PRIVATE",
	}
}

func streamSample() map[string]any {
	return map[string]any{
		"stream":  "custom_temperature",
		"units":   "C",
		"samples": []any{20.1, 20.4, 20.8},
		"script":  "return samples",
	}
}

func panelSample() map[string]any {
	return map[string]any{
		"widgets":    []any{map[string]any{"label": "FTP", "field": "ftp", "display": "number"}},
		"layout":     map[string]any{"columns": 2},
		"visibility": "PRIVATE",
	}
}

func zonesSample() map[string]any {
	return map[string]any{
		"metric": "power",
		"zones": []any{
			map[string]any{"name": "Z1", "min": 0, "max": 55, "color": "gray"},
			map[string]any{"name": "Z2", "min": 56, "max": 75, "color": "blue"},
		},
	}
}

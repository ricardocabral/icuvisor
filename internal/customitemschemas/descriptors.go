package customitemschemas

import "encoding/json"

// FamilyDescriptor describes one custom-item content-shape family.
type FamilyDescriptor struct {
	Key         string
	Title       string
	Description string
	ItemTypes   []string
	Sample      map[string]any
}

// Families returns static custom-item content samples used for documentation.
func Families() []FamilyDescriptor {
	families := []FamilyDescriptor{
		{
			Key:         "charts_tables_traces",
			Title:       "Charts, tables, traces, maps, histograms, and heatmaps",
			Description: "Display-oriented custom items define series/traces, formulas or fields, and layout/display options.",
			ItemTypes:   []string{"FITNESS_CHART", "FITNESS_TABLE", "TRACE_CHART", "ACTIVITY_CHART", "ACTIVITY_HISTOGRAM", "ACTIVITY_HEATMAP", "ACTIVITY_MAP"},
			Sample: mustSample(`{
				"series":[{"field":"ctl","label":"Fitness","color":"blue","formula":"ctl"}],
				"layout":{"height":240,"width":600},
				"axes":{"left":{"label":"Load"}},
				"filters":{"sport":"Ride"}
			}`),
		},
		{
			Key:         "fields_streams",
			Title:       "Input fields, activity fields, interval fields, and streams",
			Description: "Field and stream items describe custom values, scripts/formulas, units, formats, and visibility.",
			ItemTypes:   []string{"INPUT_FIELD", "ACTIVITY_FIELD", "INTERVAL_FIELD", "ACTIVITY_STREAM"},
			Sample: mustSample(`{
				"field":"travel_fatigue",
				"label":"Travel fatigue",
				"type":"number",
				"units":"score",
				"format":"0.0",
				"script":"return input",
				"visibility":"PRIVATE"
			}`),
		},
		{
			Key:         "panels",
			Title:       "Activity panels",
			Description: "Panel items group metrics, labels, and display widgets for activity detail pages.",
			ItemTypes:   []string{"ACTIVITY_PANEL"},
			Sample: mustSample(`{
				"widgets":[{"label":"FTP","field":"ftp","display":"number"}],
				"layout":{"columns":2},
				"visibility":"PRIVATE"
			}`),
		},
		{
			Key:         "zones",
			Title:       "Zones",
			Description: "Zone items define named ranges and display colors for a metric.",
			ItemTypes:   []string{"ZONES"},
			Sample: mustSample(`{
				"metric":"power",
				"zones":[{"name":"Z1","min":0,"max":55,"color":"gray"},{"name":"Z2","min":56,"max":75,"color":"blue"}]
			}`),
		},
	}
	out := make([]FamilyDescriptor, len(families))
	copy(out, families)
	return out
}

func mustSample(raw string) map[string]any {
	var sample map[string]any
	if err := json.Unmarshal([]byte(raw), &sample); err != nil {
		panic(err)
	}
	return sample
}

package athleteprofile

import (
	"encoding/json"
	"math"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

func TestSportSettingsPaceFieldsDecode(t *testing.T) {
	var profile intervals.AthleteWithSportSettings
	if err := json.Unmarshal([]byte(`{"sportSettings":[{"threshold_pace":3.5714285,"pace_units":"MINS_KM","pace_load_type":"RUN","pace_zones":[77.5,100],"pace_zone_names":["Easy","Threshold"]}]}`), &profile); err != nil {
		t.Fatalf("unmarshal athlete profile: %v", err)
	}
	if len(profile.SportSettings) != 1 {
		t.Fatalf("sport settings = %#v, want one decoded setting", profile.SportSettings)
	}
	setting := profile.SportSettings[0]
	if setting.ThresholdPace != 3.5714285 || setting.PaceUnits != "MINS_KM" || setting.PaceLoadType != "RUN" || len(setting.PaceZones) != 2 || setting.PaceZones[0] != 77.5 || len(setting.PaceZoneNames) != 2 || setting.PaceZoneNames[1] != "Threshold" {
		t.Fatalf("decoded pace fields = %+v", setting)
	}
}

func TestNewResponseShapesThresholdPaceFromMPSByPaceUnits(t *testing.T) {
	tests := []struct {
		name            string
		types           []string
		metersPerSecond float64
		paceUnits       string
		field           string
		wantSeconds     float64
	}{
		{name: "run kilometer", types: []string{"Run"}, metersPerSecond: 3.5714285, paceUnits: "MINS_KM", field: "km", wantSeconds: 280},
		{name: "run mile", types: []string{"Run"}, metersPerSecond: 3.5714285, paceUnits: "MINS_MILE", field: "mile", wantSeconds: 450.616329012327},
		{name: "swim 100 meters", types: []string{"Swim"}, metersPerSecond: 2, paceUnits: "SECS_100M", field: "100m", wantSeconds: 50},
		{name: "swim 100 yards", types: []string{"Swim"}, metersPerSecond: 2, paceUnits: "SECS_100Y", field: "100y", wantSeconds: 45.72},
		{name: "row 500 meters", types: []string{"Rowing"}, metersPerSecond: 4, paceUnits: "SECS_500M", field: "500m", wantSeconds: 125},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := NewResponse(intervals.AthleteWithSportSettings{ID: "i12345", SportSettings: []intervals.SportSettings{{Types: tt.types, ThresholdPace: tt.metersPerSecond, PaceUnits: tt.paceUnits}}}, "test", "UTC", false)
			sport := response.SportSettings[0]
			got := profileThresholdPaceField(sport, tt.field)
			if got == nil || math.Abs(*got-tt.wantSeconds) > 0.0001 || sport.ThresholdPaceMetersPerSecond != nil {
				t.Fatalf("pace shaping = %+v, want %s %.6f seconds", sport, tt.field, tt.wantSeconds)
			}
		})
	}
}

func profileThresholdPaceField(sport Sport, field string) *float64 {
	switch field {
	case "km":
		return sport.ThresholdPaceSecondsPerKM
	case "mile":
		return sport.ThresholdPaceSecondsPerMile
	case "100m":
		return sport.ThresholdPaceSecondsPer100M
	case "100y":
		return sport.ThresholdPaceSecondsPer100Y
	case "500m":
		return sport.ThresholdPaceSecondsPer500M
	default:
		return nil
	}
}

func TestNewResponsePreservesPaceZonePercentagesAndNames(t *testing.T) {
	response := NewResponse(intervals.AthleteWithSportSettings{
		ID: "i12345",
		SportSettings: []intervals.SportSettings{{
			ThresholdPace: 3.5714285,
			PaceUnits:     "MINS_KM",
			PaceZones:     []float64{77.5, 100},
			PaceZoneNames: []string{"Easy", "Threshold"},
			PaceLoadType:  "RUN",
		}},
	}, "test", "UTC", false)

	sport := response.SportSettings[0]
	if sport.PaceLoadType != "RUN" || len(sport.PaceZonesPercentOfThreshold) != 2 || sport.PaceZonesPercentOfThreshold[0] != 77.5 || sport.PaceZonesPercentOfThreshold[1] != 100 || len(sport.PaceZoneNames) != 2 || sport.PaceZoneNames[1] != "Threshold" {
		t.Fatalf("pace zones = %+v, want unchanged percentages and matching names", sport)
	}
	if !strings.Contains(response.Meta.ZoneBoundaryConvention, "percentage-of-threshold") {
		t.Fatalf("zone convention = %q, want percentage legend", response.Meta.ZoneBoundaryConvention)
	}
	encoded, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	if !strings.Contains(string(encoded), "pace_zones_percent_of_threshold") || strings.Contains(string(encoded), "pace_zones_seconds_per_") {
		t.Fatalf("serialized zones = %s, want percentage field and no legacy duration fields", encoded)
	}
}

func TestNewResponsePreservesUnknownPaceUnitFallback(t *testing.T) {
	response := NewResponse(intervals.AthleteWithSportSettings{
		ID: "i12345",
		SportSettings: []intervals.SportSettings{{
			ThresholdPace: 3.5714285,
			PaceUnits:     " FEET ",
			PaceZones:     []float64{77.5, 100},
		}},
	}, "test", "UTC", false)

	sport := response.SportSettings[0]
	if sport.ThresholdPaceMetersPerSecond == nil || *sport.ThresholdPaceMetersPerSecond != 3.5714285 || sport.PaceUnitsSource != "FEET" || sport.Meta["unknown_unit"] != "FEET" || len(sport.PaceZonesPercentOfThreshold) != 2 || sport.PaceZonesPercentOfThreshold[0] != 77.5 || sport.PaceZonesPercentOfThreshold[1] != 100 {
		t.Fatalf("unknown pace fallback = %+v, want raw unit, m/s threshold, and unchanged percentages", sport)
	}
}

func TestNewResponseTreatsNonePaceUnitsAsKnownMPSFallback(t *testing.T) {
	response := NewResponse(intervals.AthleteWithSportSettings{
		ID: "i12345",
		SportSettings: []intervals.SportSettings{{
			ThresholdPace: 3.5714285,
			PaceUnits:     "NONE",
			PaceZones:     []float64{77.5, 100},
		}},
	}, "test", "UTC", false)

	sport := response.SportSettings[0]
	if sport.ThresholdPaceMetersPerSecond == nil || *sport.ThresholdPaceMetersPerSecond != 3.5714285 || sport.PaceUnitsSource != "NONE" || sport.Meta != nil {
		t.Fatalf("NONE pace fallback = %+v, want known m/s fallback without unknown-unit metadata", sport)
	}
	if len(sport.PaceZonesPercentOfThreshold) != 2 || sport.PaceZonesPercentOfThreshold[0] != 77.5 || sport.PaceZonesPercentOfThreshold[1] != 100 {
		t.Fatalf("NONE pace zones = %#v, want unchanged percentages", sport.PaceZonesPercentOfThreshold)
	}
}

func TestNewResponseFallsBackToMPSWhenPaceDisplayOverflows(t *testing.T) {
	response := NewResponse(intervals.AthleteWithSportSettings{
		ID: "i12345",
		SportSettings: []intervals.SportSettings{{
			ThresholdPace: math.SmallestNonzeroFloat64,
			PaceUnits:     "MINS_KM",
		}},
	}, "test", "UTC", false)

	sport := response.SportSettings[0]
	if sport.ThresholdPaceSecondsPerKM != nil || sport.ThresholdPaceMetersPerSecond == nil || *sport.ThresholdPaceMetersPerSecond != math.SmallestNonzeroFloat64 {
		t.Fatalf("overflowing pace display = %+v, want finite m/s fallback", sport)
	}
}

func TestNewResponsePaceMetadataUsesStorageAndPercentageSemantics(t *testing.T) {
	response := NewResponse(intervals.AthleteWithSportSettings{ID: "i12345"}, "test", "UTC", false)

	if !strings.Contains(response.Meta.PaceConvention, "meters per second") || !strings.Contains(response.Meta.PaceConvention, "presentation-only") {
		t.Fatalf("pace convention = %q, want m/s storage and presentation-only pace units", response.Meta.PaceConvention)
	}
	if !strings.Contains(response.Meta.ZoneBoundaryConvention, "pace_zones_percent_of_threshold") || !strings.Contains(response.Meta.ZoneBoundaryConvention, "percentage-of-threshold") {
		t.Fatalf("zone convention = %q, want percentage pace-zone semantics", response.Meta.ZoneBoundaryConvention)
	}
}

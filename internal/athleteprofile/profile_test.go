package athleteprofile

import (
	"math"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

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

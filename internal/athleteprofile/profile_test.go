package athleteprofile

import (
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

func TestNewResponsePaceMetadataUsesStorageAndPercentageSemantics(t *testing.T) {
	response := NewResponse(intervals.AthleteWithSportSettings{ID: "i12345"}, "test", "UTC", false)

	if !strings.Contains(response.Meta.PaceConvention, "meters per second") || !strings.Contains(response.Meta.PaceConvention, "presentation-only") {
		t.Fatalf("pace convention = %q, want m/s storage and presentation-only pace units", response.Meta.PaceConvention)
	}
	if !strings.Contains(response.Meta.ZoneBoundaryConvention, "pace_zones_percent_of_threshold") || !strings.Contains(response.Meta.ZoneBoundaryConvention, "percentage-of-threshold") {
		t.Fatalf("zone convention = %q, want percentage pace-zone semantics", response.Meta.ZoneBoundaryConvention)
	}
}

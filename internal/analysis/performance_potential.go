package analysis

import "strings"

const (
	// PerformancePotentialFormulaRef documents that the surface summarizes explicit upstream fields without estimating hidden thresholds.
	PerformancePotentialFormulaRef = "icuvisor.performance_potential.v1.explicit_profile_thresholds_plus_upstream_curves"
)

// PerformancePotentialUnavailable describes an unavailable source, threshold, or estimate without inventing a value.
type PerformancePotentialUnavailable struct {
	Field          string `json:"field"`
	Status         string `json:"status"`
	Reason         string `json:"reason"`
	RequiredSource string `json:"required_source,omitempty"`
}

// PerformancePotentialSourceStatus records whether a summarized source was queried and usable.
type PerformancePotentialSourceStatus struct {
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
	Unit   string `json:"unit,omitempty"`
}

// PerformancePotentialSportFamily classifies a sport for safe per-sport curve and threshold handling.
func PerformancePotentialSportFamily(sport string) string {
	switch normalizePerformancePotentialSport(sport) {
	case "ride", "virtualride", "bike", "bikeride", "cycle", "cycling", "indoorcycling", "mountainbike", "mtb", "gravelride", "ebikeride":
		return "cycling"
	case "run", "virtualrun", "trailrun", "treadmill", "walk", "hike":
		return "running"
	case "swim", "openwaterswim", "poolswim":
		return "swimming"
	case "row", "rowing", "kayak", "canoe":
		return "rowing"
	default:
		return "other"
	}
}

// PerformancePotentialSupportsPower reports whether a sport should request power-duration anchors by default.
func PerformancePotentialSupportsPower(sport string) bool {
	return PerformancePotentialSportFamily(sport) == "cycling"
}

// PerformancePotentialSupportsPace reports whether a sport should request pace-distance anchors by default.
func PerformancePotentialSupportsPace(sport string) bool {
	switch PerformancePotentialSportFamily(sport) {
	case "running", "swimming", "rowing":
		return true
	default:
		return false
	}
}

// PerformancePotentialSupportsHeartRate reports whether HR anchors are meaningful for the sport family.
func PerformancePotentialSupportsHeartRate(sport string) bool {
	switch PerformancePotentialSportFamily(sport) {
	case "cycling", "running", "swimming", "rowing":
		return true
	default:
		return false
	}
}

func normalizePerformancePotentialSport(sport string) string {
	return strings.ToLower(strings.NewReplacer("_", "", "-", "", " ", "").Replace(strings.TrimSpace(sport)))
}

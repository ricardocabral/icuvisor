// Package units provides typed intervals.icu unit handling.
package units

import (
	"log/slog"
	"strings"
)

// Unit is an intervals.icu unit enum value observed on activity reads.
type Unit string

const (
	// UnitUnknown represents an upstream unit icuvisor does not yet know.
	UnitUnknown Unit = "UNKNOWN"

	// Distance units.
	UnitM  Unit = "M"
	UnitKM Unit = "KM"
	UnitMI Unit = "MI"
	UnitYD Unit = "YD"

	// Pace units.
	UnitMinsKM   Unit = "MINS_KM"
	UnitMinsMile Unit = "MINS_MILE"
	UnitSecs100M Unit = "SECS_100M"
	UnitSecs500M Unit = "SECS_500M"

	// Speed units.
	UnitKMH Unit = "KMH"
	UnitMPH Unit = "MPH"
	UnitMS  Unit = "MS"

	// Time units.
	UnitSecs  Unit = "SECS"
	UnitMins  Unit = "MINS"
	UnitHours Unit = "HOURS"

	// Power units.
	UnitWatts      Unit = "WATTS"
	UnitWKG        Unit = "WKG"
	UnitPercentFTP Unit = "PERCENT_FTP"

	// Heart-rate units.
	UnitBPM          Unit = "BPM"
	UnitPercentHR    Unit = "PERCENT_HR"
	UnitPercentLTHR  Unit = "PERCENT_LTHR"
	UnitPercentMaxHR Unit = "PERCENT_MAX_HR"

	// Miscellaneous units.
	UnitRPE     Unit = "RPE"
	UnitZ1      Unit = "Z1"
	UnitZ2      Unit = "Z2"
	UnitZ3      Unit = "Z3"
	UnitZ4      Unit = "Z4"
	UnitZ5      Unit = "Z5"
	UnitZ6      Unit = "Z6"
	UnitZ7      Unit = "Z7"
	UnitPercent Unit = "PERCENT"
	UnitKCAL    Unit = "KCAL"
	UnitKJ      Unit = "KJ"
)

var knownUnits = map[string]Unit{
	string(UnitM):            UnitM,
	string(UnitKM):           UnitKM,
	string(UnitMI):           UnitMI,
	string(UnitYD):           UnitYD,
	string(UnitMinsKM):       UnitMinsKM,
	string(UnitMinsMile):     UnitMinsMile,
	string(UnitSecs100M):     UnitSecs100M,
	string(UnitSecs500M):     UnitSecs500M,
	string(UnitKMH):          UnitKMH,
	string(UnitMPH):          UnitMPH,
	string(UnitMS):           UnitMS,
	string(UnitSecs):         UnitSecs,
	string(UnitMins):         UnitMins,
	string(UnitHours):        UnitHours,
	string(UnitWatts):        UnitWatts,
	string(UnitWKG):          UnitWKG,
	string(UnitPercentFTP):   UnitPercentFTP,
	string(UnitBPM):          UnitBPM,
	string(UnitPercentHR):    UnitPercentHR,
	string(UnitPercentLTHR):  UnitPercentLTHR,
	string(UnitPercentMaxHR): UnitPercentMaxHR,
	string(UnitRPE):          UnitRPE,
	string(UnitZ1):           UnitZ1,
	string(UnitZ2):           UnitZ2,
	string(UnitZ3):           UnitZ3,
	string(UnitZ4):           UnitZ4,
	string(UnitZ5):           UnitZ5,
	string(UnitZ6):           UnitZ6,
	string(UnitZ7):           UnitZ7,
	string(UnitPercent):      UnitPercent,
	string(UnitKCAL):         UnitKCAL,
	string(UnitKJ):           UnitKJ,
}

// ParseUnit parses an intervals.icu unit token without failing on unknown values.
// It trims surrounding whitespace, matches known uppercase values case-sensitively,
// returns an empty raw string for known units, and returns UnitUnknown with the
// exact trimmed upstream token for unknown units. Callers must preserve the raw
// return value whenever the parsed Unit is UnitUnknown.
func ParseUnit(value string) (Unit, string) {
	raw := strings.TrimSpace(value)
	if unit, ok := knownUnits[raw]; ok {
		return unit, ""
	}
	slog.Default().Warn("unknown intervals.icu unit", "unit", raw)
	return UnitUnknown, raw
}

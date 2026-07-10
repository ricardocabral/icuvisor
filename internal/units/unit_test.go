package units

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestParseUnitKnownMembers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want Unit
	}{
		{name: "meters", in: "M", want: UnitM},
		{name: "kilometers", in: "KM", want: UnitKM},
		{name: "miles", in: "MI", want: UnitMI},
		{name: "yards", in: "YD", want: UnitYD},
		{name: "minutes per kilometer", in: "MINS_KM", want: UnitMinsKM},
		{name: "minutes per mile", in: "MINS_MILE", want: UnitMinsMile},
		{name: "seconds per 100 meters", in: "SECS_100M", want: UnitSecs100M},
		{name: "seconds per 100 yards", in: "SECS_100Y", want: UnitSecs100Y},
		{name: "seconds per 500 meters", in: "SECS_500M", want: UnitSecs500M},
		{name: "seconds per 400 meters", in: "SECS_400M", want: UnitSecs400M},
		{name: "seconds per 250 meters", in: "SECS_250M", want: UnitSecs250M},
		{name: "kilometers per hour", in: "KMH", want: UnitKMH},
		{name: "miles per hour", in: "MPH", want: UnitMPH},
		{name: "meters per second", in: "MS", want: UnitMS},
		{name: "seconds", in: "SECS", want: UnitSecs},
		{name: "minutes", in: "MINS", want: UnitMins},
		{name: "hours", in: "HOURS", want: UnitHours},
		{name: "watts", in: "WATTS", want: UnitWatts},
		{name: "watts per kilogram", in: "WKG", want: UnitWKG},
		{name: "percent ftp", in: "PERCENT_FTP", want: UnitPercentFTP},
		{name: "beats per minute", in: "BPM", want: UnitBPM},
		{name: "percent heart rate", in: "PERCENT_HR", want: UnitPercentHR},
		{name: "percent threshold heart rate", in: "PERCENT_LTHR", want: UnitPercentLTHR},
		{name: "percent max heart rate", in: "PERCENT_MAX_HR", want: UnitPercentMaxHR},
		{name: "rpe", in: "RPE", want: UnitRPE},
		{name: "zone 1", in: "Z1", want: UnitZ1},
		{name: "zone 2", in: "Z2", want: UnitZ2},
		{name: "zone 3", in: "Z3", want: UnitZ3},
		{name: "zone 4", in: "Z4", want: UnitZ4},
		{name: "zone 5", in: "Z5", want: UnitZ5},
		{name: "zone 6", in: "Z6", want: UnitZ6},
		{name: "zone 7", in: "Z7", want: UnitZ7},
		{name: "percent", in: "PERCENT", want: UnitPercent},
		{name: "kilocalories", in: "KCAL", want: UnitKCAL},
		{name: "kilojoules", in: "KJ", want: UnitKJ},
		{name: "trimmed known token", in: "  KM  ", want: UnitKM},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, raw := ParseUnit(tc.in)
			if got != tc.want {
				t.Fatalf("ParseUnit(%q) unit = %q, want %q", tc.in, got, tc.want)
			}
			if raw != "" {
				t.Fatalf("ParseUnit(%q) raw = %q, want empty", tc.in, raw)
			}
		})
	}
}

func TestParseUnitUnknownPreservesRawAndLogs(t *testing.T) {
	tests := []struct {
		name string
		in   string
		raw  string
	}{
		{name: "empty", in: "", raw: ""},
		{name: "mixed invalid casing", in: "mins_km", raw: "mins_km"},
		{name: "future token", in: "FEET", raw: "FEET"},
		{name: "trimmed unknown token", in: "  FTP_PER_KG  ", raw: "FTP_PER_KG"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var logs bytes.Buffer
			previous := slog.Default()
			t.Cleanup(func() { slog.SetDefault(previous) })
			slog.SetDefault(slog.New(slog.NewTextHandler(&logs, &slog.HandlerOptions{Level: slog.LevelWarn})))

			got, raw := ParseUnit(tc.in)
			if got != UnitUnknown {
				t.Fatalf("ParseUnit(%q) unit = %q, want %q", tc.in, got, UnitUnknown)
			}
			if raw != tc.raw {
				t.Fatalf("ParseUnit(%q) raw = %q, want %q", tc.in, raw, tc.raw)
			}

			logLine := logs.String()
			if !strings.Contains(logLine, "level=WARN") || !strings.Contains(logLine, "msg=\"unknown intervals.icu unit\"") {
				t.Fatalf("unknown unit log = %q, want WARN with unknown-unit message", logLine)
			}
			if tc.raw != "" && !strings.Contains(logLine, "unit="+tc.raw) {
				t.Fatalf("unknown unit log = %q, want raw unit %q", logLine, tc.raw)
			}
		})
	}
}

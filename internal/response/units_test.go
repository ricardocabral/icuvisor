package response

import (
	"math"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/units"
)

func TestUnitSystemFromPreferredUnits(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want UnitSystem
		ok   bool
	}{
		{name: "metric", in: "metric", want: UnitSystemMetric, ok: true},
		{name: "kilometers", in: "kilometers", want: UnitSystemMetric, ok: true},
		{name: "imperial", in: "imperial", want: UnitSystemImperial, ok: true},
		{name: "miles", in: "miles", want: UnitSystemImperial, ok: true},
		{name: "empty", in: "", ok: false},
		{name: "unknown", in: "furlongs", ok: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := UnitSystemFromPreferredUnits(tt.in)
			if got != tt.want || ok != tt.ok {
				t.Fatalf("UnitSystemFromPreferredUnits(%q) = %q, %t; want %q, %t", tt.in, got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestUnitSystemFromProfileFallbacks(t *testing.T) {
	tests := []struct {
		name                  string
		preferredUnits        string
		measurementPreference string
		weightPrefLB          bool
		want                  UnitSystem
		ok                    bool
	}{
		{name: "preferred wins", preferredUnits: "miles", measurementPreference: "metric", want: UnitSystemImperial, ok: true},
		{name: "measurement fallback", measurementPreference: "metric", weightPrefLB: true, want: UnitSystemMetric, ok: true},
		{name: "weight fallback", weightPrefLB: true, want: UnitSystemImperial, ok: true},
		{name: "unknown absent", preferredUnits: "furlongs", ok: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := UnitSystemFromProfile(tt.preferredUnits, tt.measurementPreference, tt.weightPrefLB)
			if got != tt.want || ok != tt.ok {
				t.Fatalf("UnitSystemFromProfile() = %q, %t; want %q, %t", got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestToPreferredConvertsDistanceSpeedAndRunPace(t *testing.T) {
	tests := []struct {
		name       string
		value      float64
		unit       units.Unit
		system     UnitSystem
		wantValue  float64
		wantUnit   units.Unit
		wantSuffix string
		converted  bool
	}{
		{name: "kilometers to miles", value: 1, unit: units.UnitKM, system: UnitSystemImperial, wantValue: 0.621371192237334, wantUnit: units.UnitMI, wantSuffix: "mi", converted: true},
		{name: "miles to kilometers", value: 1, unit: units.UnitMI, system: UnitSystemMetric, wantValue: 1.609344, wantUnit: units.UnitKM, wantSuffix: "km", converted: true},
		{name: "meters to kilometers", value: 500, unit: units.UnitM, system: UnitSystemMetric, wantValue: 0.5, wantUnit: units.UnitKM, wantSuffix: "km", converted: true},
		{name: "yards to miles", value: 1760, unit: units.UnitYD, system: UnitSystemImperial, wantValue: 1, wantUnit: units.UnitMI, wantSuffix: "mi", converted: true},
		{name: "kmh to mph", value: 36, unit: units.UnitKMH, system: UnitSystemImperial, wantValue: 22.369362920544024, wantUnit: units.UnitMPH, wantSuffix: "mph", converted: true},
		{name: "mph to kmh", value: 10, unit: units.UnitMPH, system: UnitSystemMetric, wantValue: 16.09344, wantUnit: units.UnitKMH, wantSuffix: "kmh", converted: true},
		{name: "meters per second to kmh", value: 10, unit: units.UnitMS, system: UnitSystemMetric, wantValue: 36, wantUnit: units.UnitKMH, wantSuffix: "kmh", converted: true},
		{name: "min per km to min per mile", value: 5, unit: units.UnitMinsKM, system: UnitSystemImperial, wantValue: 8.04672, wantUnit: units.UnitMinsMile, wantSuffix: "minutes_per_mile", converted: true},
		{name: "min per mile to min per km", value: 8, unit: units.UnitMinsMile, system: UnitSystemMetric, wantValue: 4.970969537898672, wantUnit: units.UnitMinsKM, wantSuffix: "minutes_per_km", converted: true},
		{name: "metric min per km unchanged", value: 5, unit: units.UnitMinsKM, system: UnitSystemMetric, wantValue: 5, wantUnit: units.UnitMinsKM, wantSuffix: "minutes_per_km", converted: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToPreferred(tt.value, tt.unit, tt.system)
			assertClose(t, got.Value, tt.wantValue)
			if got.Unit != tt.wantUnit || got.FieldSuffix != tt.wantSuffix || got.Converted != tt.converted {
				t.Fatalf("ToPreferred() = %+v, want unit %q suffix %q converted %t", got, tt.wantUnit, tt.wantSuffix, tt.converted)
			}
			if got.OriginalValue != tt.value || got.OriginalUnit != tt.unit || got.OriginalUnitLabel != string(tt.unit) {
				t.Fatalf("ToPreferred() original = %+v, want value %v unit %q", got, tt.value, tt.unit)
			}
		})
	}
}

func TestToPreferredPreservesSportSpecificAndUnknownUnits(t *testing.T) {
	swim := ToPreferred(82, units.UnitSecs100M, UnitSystemImperial)
	if swim.Value != 82 || swim.Unit != units.UnitSecs100M || swim.FieldSuffix != "seconds_per_100m" || swim.Converted {
		t.Fatalf("swim pace = %+v, want pass-through sec/100m", swim)
	}
	row := ToPreferred(115, units.UnitSecs500M, UnitSystemMetric)
	if row.Value != 115 || row.Unit != units.UnitSecs500M || row.FieldSuffix != "seconds_per_500m" || row.Converted {
		t.Fatalf("row pace = %+v, want pass-through sec/500m", row)
	}
	unknown := ToPreferredWithRaw(7, units.UnitUnknown, "FEET", UnitSystemImperial)
	if unknown.Value != 7 || unknown.Unit != units.UnitUnknown || unknown.UnitLabel != "FEET" || unknown.UnknownUnit != "FEET" || unknown.Converted {
		t.Fatalf("unknown unit = %+v, want raw FEET pass-through", unknown)
	}
}

func TestUnitSystemDistanceHelpers(t *testing.T) {
	if got := UnitSystemMetric.DistanceFieldName("distance"); got != "distance_km" {
		t.Fatalf("metric field = %q", got)
	}
	if got := UnitSystemImperial.DistanceFieldName("distance_km"); got != "distance_mi" {
		t.Fatalf("imperial field = %q", got)
	}
	if got := UnitSystemMetric.ConvertDistanceKM(10); got != 10 {
		t.Fatalf("metric distance = %v", got)
	}
	if got := UnitSystemImperial.ConvertDistanceKM(10); got < 6.2137 || got > 6.2138 {
		t.Fatalf("imperial distance = %v", got)
	}
}

func assertClose(t *testing.T, got float64, want float64) {
	t.Helper()
	if math.Abs(got-want) > 0.000001 {
		t.Fatalf("value = %.12f, want %.12f", got, want)
	}
}

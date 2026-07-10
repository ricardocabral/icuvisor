package analysis

import (
	"errors"
	"math"
	"reflect"
	"testing"
)

func TestZoneEnergyContract(t *testing.T) {
	t.Run("pins model visible constants", func(t *testing.T) {
		if ZoneEnergyMethod != "left_endpoint_power_timestamp_integration" {
			t.Fatalf("method = %q", ZoneEnergyMethod)
		}
		if ZoneEnergyFormulaRef != "icuvisor://analysis-formulas#power_zone_mechanical_work" {
			t.Fatalf("formula ref = %q", ZoneEnergyFormulaRef)
		}
		if ZoneEnergyMaxIntervalSeconds != 60 {
			t.Fatalf("max interval = %d", ZoneEnergyMaxIntervalSeconds)
		}
		wantBoundaries := []string{
			"Mechanical work from recorded power is not metabolic energy, calorie expenditure, or food calories.",
			"Left-endpoint integration; the final sample contributes no duration or work.",
			"Intervals longer than 60 seconds and invalid samples are skipped; missing power is not interpolated.",
			"Raw stream samples are never returned.",
		}
		if !reflect.DeepEqual(ZoneEnergyBoundaries, wantBoundaries) {
			t.Fatalf("boundaries = %#v", ZoneEnergyBoundaries)
		}
	})

	t.Run("validates boundaries without sorting or repair", func(t *testing.T) {
		tests := []struct {
			name       string
			boundaries []float64
			wantErr    bool
		}{
			{name: "initial zero", boundaries: []float64{0, 100, 200}},
			{name: "positive first boundary", boundaries: []float64{100, 200}},
			{name: "missing", wantErr: true},
			{name: "negative", boundaries: []float64{-1, 100}, wantErr: true},
			{name: "duplicate", boundaries: []float64{100, 100}, wantErr: true},
			{name: "descending", boundaries: []float64{200, 100}, wantErr: true},
			{name: "non finite", boundaries: []float64{100, math.Inf(1)}, wantErr: true},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				err := ValidatePowerZoneConfig(PowerZoneConfig{BoundariesWatts: tc.boundaries})
				if tc.wantErr && !errors.Is(err, ErrInvalidPowerZoneConfig) {
					t.Fatalf("error = %v, want ErrInvalidPowerZoneConfig", err)
				}
				if !tc.wantErr && err != nil {
					t.Fatalf("error = %v", err)
				}
			})
		}
	})

	t.Run("defines mismatch and short input diagnostics", func(t *testing.T) {
		tests := []struct {
			name  string
			input ZoneEnergyInput
			want  ZoneEnergyDiagnostics
		}{
			{
				name:  "power longer",
				input: ZoneEnergyInput{PowerWatts: []float64{100, 110, 120}, TimestampsSeconds: []float64{0, 1}},
				want:  ZoneEnergyDiagnostics{InputSamples: 3, AlignedSamples: 2, MisalignedSamples: 1, SkippedIntervals: 2},
			},
			{
				name:  "time longer",
				input: ZoneEnergyInput{PowerWatts: []float64{100}, TimestampsSeconds: []float64{0, 1, 2}},
				want:  ZoneEnergyDiagnostics{InputSamples: 3, AlignedSamples: 1, MisalignedSamples: 2, SkippedIntervals: 2},
			},
			{
				name:  "one aligned sample",
				input: ZoneEnergyInput{PowerWatts: []float64{100}, TimestampsSeconds: []float64{0}},
				want:  ZoneEnergyDiagnostics{InputSamples: 1, AlignedSamples: 1},
			},
			{
				name: "empty",
				want: ZoneEnergyDiagnostics{},
			},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				if got := ZoneEnergyInputDiagnostics(tc.input); !reflect.DeepEqual(got, tc.want) {
					t.Fatalf("diagnostics = %+v, want %+v", got, tc.want)
				}
			})
		}
	})
}

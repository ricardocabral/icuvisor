package analysis

import (
	"errors"
	"math"
	"reflect"
	"testing"
)

func TestComputeZoneEnergy(t *testing.T) {
	standardConfig := PowerZoneConfig{
		Sport:           "Ride",
		SportSettingID:  7,
		BoundariesWatts: []float64{0, 100, 200},
		Names:           []string{"Recovery", "Endurance", "Tempo"},
	}

	tests := []struct {
		name             string
		input            ZoneEnergyInput
		wantSeconds      float64
		wantKJ           float64
		wantZoneSeconds  []float64
		wantZoneKJ       []float64
		wantTimeShares   []float64
		wantEnergyShares []float64
	}{
		{
			name: "irregular timestamps coasting and exact boundaries",
			input: ZoneEnergyInput{
				TimestampsSeconds: []float64{0, 2, 5, 9},
				PowerWatts:        []float64{0, 100, 200, 999},
				ZoneConfig:        standardConfig,
			},
			wantSeconds:      9,
			wantKJ:           1.1,
			wantZoneSeconds:  []float64{2, 3, 4},
			wantZoneKJ:       []float64{0, 0.3, 0.8},
			wantTimeShares:   []float64{0.2222, 0.3333, 0.4445},
			wantEnergyShares: []float64{0, 0.2727, 0.7273},
		},
		{
			name: "sub-boundary coasting uses explicit below bucket",
			input: ZoneEnergyInput{
				TimestampsSeconds: []float64{0, 10, 20},
				PowerWatts:        []float64{0, 100, 999},
				ZoneConfig: PowerZoneConfig{
					BoundariesWatts: []float64{100, 200},
					Names:           []string{"Easy", "Hard"},
				},
			},
			wantSeconds:      20,
			wantKJ:           1,
			wantZoneSeconds:  []float64{10, 10, 0},
			wantZoneKJ:       []float64{0, 1, 0},
			wantTimeShares:   []float64{0.5, 0.5, 0},
			wantEnergyShares: []float64{0, 1, 0},
		},
		{
			name: "joules convert to kilojoules",
			input: ZoneEnergyInput{
				TimestampsSeconds: []float64{0, 4},
				PowerWatts:        []float64{250, 0},
				ZoneConfig:        standardConfig,
			},
			wantSeconds:      4,
			wantKJ:           1,
			wantZoneSeconds:  []float64{0, 0, 4},
			wantZoneKJ:       []float64{0, 0, 1},
			wantTimeShares:   []float64{0, 0, 1},
			wantEnergyShares: []float64{0, 0, 1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ComputeZoneEnergy(tc.input)
			if err != nil {
				t.Fatalf("ComputeZoneEnergy() error = %v", err)
			}
			if got.TotalSeconds != tc.wantSeconds || got.TotalKJ != tc.wantKJ {
				t.Fatalf("totals = (%v s, %v kJ), want (%v s, %v kJ)", got.TotalSeconds, got.TotalKJ, tc.wantSeconds, tc.wantKJ)
			}
			if got.Diagnostics.UsableIntervals != len(tc.input.PowerWatts)-1 || got.Diagnostics.SkippedIntervals != 0 {
				t.Fatalf("diagnostics = %+v", got.Diagnostics)
			}
			if len(got.Zones) != len(tc.wantZoneSeconds) {
				t.Fatalf("zone count = %d, want %d", len(got.Zones), len(tc.wantZoneSeconds))
			}
			for i, zone := range got.Zones {
				if zone.Seconds != tc.wantZoneSeconds[i] || zone.KJ != tc.wantZoneKJ[i] ||
					zone.TimeShare != tc.wantTimeShares[i] || zone.EnergyShare != tc.wantEnergyShares[i] {
					t.Errorf("zone %d = %+v", i, zone)
				}
			}
		})
	}
}

func TestComputeZoneEnergyInvalidSamples(t *testing.T) {
	config := PowerZoneConfig{BoundariesWatts: []float64{0, 100}}
	t.Run("classifies invalid intervals in deterministic precedence", func(t *testing.T) {
		got, err := ComputeZoneEnergy(ZoneEnergyInput{
			TimestampsSeconds: []float64{math.NaN(), 0, 0, -1, 100, 101, 102, 103},
			PowerWatts:        []float64{100, math.Inf(1), 100, 100, math.NaN(), -1, 50, 999},
			ZoneConfig:        config,
		})
		if err != nil {
			t.Fatalf("ComputeZoneEnergy() error = %v", err)
		}
		want := ZoneEnergyDiagnostics{
			InputSamples:              8,
			AlignedSamples:            8,
			UsableIntervals:           1,
			SkippedIntervals:          6,
			SkippedNonFiniteTimestamp: 1,
			SkippedDuplicateTimestamp: 1,
			SkippedReversedTimestamp:  1,
			SkippedLargeGap:           1,
			SkippedNonFinitePower:     1,
			SkippedNegativePower:      1,
		}
		if !reflect.DeepEqual(got.Diagnostics, want) {
			t.Fatalf("diagnostics = %+v, want %+v", got.Diagnostics, want)
		}
		if got.TotalSeconds != 1 || got.TotalKJ != 0.05 {
			t.Fatalf("totals = (%v s, %v kJ)", got.TotalSeconds, got.TotalKJ)
		}
	})

	t.Run("misaligned streams are not guessed", func(t *testing.T) {
		got, err := ComputeZoneEnergy(ZoneEnergyInput{
			TimestampsSeconds: []float64{0, 1},
			PowerWatts:        []float64{100, 100, 100},
			ZoneConfig:        config,
		})
		if err != nil {
			t.Fatalf("ComputeZoneEnergy() error = %v", err)
		}
		want := ZoneEnergyDiagnostics{InputSamples: 3, AlignedSamples: 2, SkippedIntervals: 2, MisalignedSamples: 1}
		if !reflect.DeepEqual(got.Diagnostics, want) || got.TotalSeconds != 0 || got.TotalKJ != 0 {
			t.Fatalf("result = %+v", got)
		}
	})

	t.Run("short aligned stream has no guessed duration", func(t *testing.T) {
		got, err := ComputeZoneEnergy(ZoneEnergyInput{
			TimestampsSeconds: []float64{0},
			PowerWatts:        []float64{100},
			ZoneConfig:        config,
		})
		if err != nil {
			t.Fatalf("ComputeZoneEnergy() error = %v", err)
		}
		if got.Diagnostics.UsableIntervals != 0 || got.Diagnostics.SkippedIntervals != 0 || got.TotalSeconds != 0 {
			t.Fatalf("result = %+v", got)
		}
	})
}

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

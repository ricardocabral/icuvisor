package workoutdoc

// SyntaxSpec describes the structured workout DSL supported by Serialize.
type SyntaxSpec struct {
	Features    []SyntaxFeature
	Limitations []SyntaxLimitation
}

// SyntaxFeature describes one supported workout DSL feature.
type SyntaxFeature struct {
	Key         string
	Title       string
	Description string
	Examples    []SyntaxExample
}

// SyntaxExample is a representative structured step rendered through Serialize.
type SyntaxExample struct {
	Key         string
	Description string
	Step        Step
}

// SyntaxLimitation describes a serializer limitation callers must honor.
type SyntaxLimitation struct {
	Key         string
	Description string
}

// WorkoutSyntaxSpec returns the workout DSL syntax supported by this package.
func WorkoutSyntaxSpec() SyntaxSpec {
	return SyntaxSpec{
		Features: []SyntaxFeature{
			{
				Key:         "duration_steps",
				Title:       "Duration steps",
				Description: "A simple step needs a positive duration or a distance. Durations serialize as h/m/s tokens.",
				Examples: []SyntaxExample{{
					Key:         "duration_percent_ftp",
					Description: "Duration step with a percent-FTP power target.",
					Step:        Step{Description: "Warmup", Duration: 600, Power: targetValue(55, "PERCENT_FTP")},
				}},
			},
			{
				Key:         "distance_steps",
				Title:       "Distance steps",
				Description: "Distance steps serialize with canonical mtr, km, or mi suffixes.",
				Examples: []SyntaxExample{
					{Key: "distance_mtr", Description: "Meter distance canonicalizes to mtr.", Step: Step{Description: "Stride", Distance: &Length{Value: 400, Unit: "meters"}, Power: targetValue(120, "PERCENT_FTP")}},
					{Key: "distance_km", Description: "Kilometer distance canonicalizes to km.", Step: Step{Description: "Tempo", Distance: &Length{Value: 5, Unit: "kilometers"}, Pace: targetRange(92, 96, "PERCENT_THRESHOLD")}},
					{Key: "distance_mi", Description: "Mile distance canonicalizes to mi.", Step: Step{Description: "Cooldown", Distance: &Length{Value: 1, Unit: "miles"}, Freeride: true}},
				},
			},
			{
				Key:         "repeats",
				Title:       "Repeat blocks",
				Description: "Repeat blocks use an Nx header and indented child steps.",
				Examples: []SyntaxExample{{
					Key:         "repeat_block",
					Description: "Two child steps repeated three times.",
					Step: Step{Description: "Main set", Reps: 3, Steps: []Step{
						{Description: "Hard", Duration: 120, Power: targetRange(105, 115, "PERCENT_FTP"), Cadence: targetRange(95, 105, "RPM")},
						{Description: "Easy", Duration: 60, Freeride: true},
					}},
				}},
			},
			{
				Key:         "freeride",
				Title:       "Free-ride steps",
				Description: "Freeride is a primary target and cannot be combined with another primary target.",
				Examples:    []SyntaxExample{{Key: "freeride", Description: "Open target free ride.", Step: Step{Description: "Open", Duration: 300, Freeride: true}}},
			},
			{
				Key:         "ramps",
				Title:       "Ramp targets",
				Description: "Ramp steps use start and end target bounds and serialize with a ramp prefix.",
				Examples:    []SyntaxExample{{Key: "power_ramp", Description: "Power ramp from one percent-FTP target to another.", Step: Step{Description: "Build", Duration: 480, Ramp: true, Power: targetRamp(70, 95, "PERCENT_FTP")}}},
			},
			{
				Key:         "cadence_targets",
				Title:       "Cadence targets",
				Description: "Cadence is an optional secondary target in rpm and may be scalar or range.",
				Examples:    []SyntaxExample{{Key: "cadence_range", Description: "Cadence range appended after the primary target.", Step: Step{Description: "Spin", Duration: 180, Power: targetValue(60, "PERCENT_FTP"), Cadence: targetRange(100, 110, "RPM")}}},
			},
			{
				Key:         "power_targets",
				Title:       "Power targets",
				Description: "Power targets support percent FTP, watts, power zones, scalar values, and ranges.",
				Examples: []SyntaxExample{
					{Key: "power_percent", Description: "Percent FTP scalar.", Step: Step{Description: "Endurance", Duration: 600, Power: targetValue(75, "PERCENT_FTP")}},
					{Key: "power_percent_range", Description: "Percent FTP range.", Step: Step{Description: "Sweet spot", Duration: 600, Power: targetRange(88, 94, "PERCENT_FTP")}},
					{Key: "power_watts", Description: "Watts scalar.", Step: Step{Description: "Erg", Duration: 300, Power: targetValue(250, "WATTS")}},
					{Key: "power_zone", Description: "Power zone range.", Step: Step{Description: "Zone work", Duration: 900, Power: targetRange(2, 3, "POWER_ZONE")}},
				},
			},
			{
				Key:         "heart_rate_targets",
				Title:       "Heart-rate targets",
				Description: "Heart-rate targets support percent max HR, percent LTHR, bpm, HR zones, scalar values, and ranges.",
				Examples: []SyntaxExample{
					{Key: "hr_percent", Description: "Percent max HR scalar.", Step: Step{Description: "Aerobic", Duration: 600, HR: targetValue(80, "PERCENT_HR")}},
					{Key: "hr_lthr", Description: "Percent LTHR range.", Step: Step{Description: "Threshold HR", Duration: 600, HR: targetRange(95, 99, "PERCENT_LTHR")}},
					{Key: "hr_bpm", Description: "BPM scalar.", Step: Step{Description: "Cap", Duration: 300, HR: targetValue(150, "BPM")}},
					{Key: "hr_zone", Description: "HR zone range.", Step: Step{Description: "Zone HR", Duration: 600, HR: targetRange(2, 3, "HR_ZONE")}},
				},
			},
			{
				Key:         "pace_targets",
				Title:       "Pace targets",
				Description: "Pace targets support percent threshold pace, pace zones, numeric PACE values, and non-ramp text pace labels.",
				Examples: []SyntaxExample{
					{Key: "pace_percent", Description: "Percent threshold pace scalar.", Step: Step{Description: "Cruise", Duration: 600, Pace: targetValue(95, "PERCENT_THRESHOLD")}},
					{Key: "pace_zone", Description: "Pace zone range.", Step: Step{Description: "Pace zone", Duration: 600, Pace: targetRange(2, 3, "PACE_ZONE")}},
					{Key: "pace_numeric", Description: "Numeric PACE unit as currently serialized.", Step: Step{Description: "Numeric pace", Duration: 300, Pace: targetValue(5, "PACE")}},
					{Key: "pace_text", Description: "Text pace label.", Step: Step{Description: "Marathon", Duration: 1200, Pace: &Target{Text: "Marathon Pace"}}},
				},
			},
			{
				Key:         "rpe_targets",
				Title:       "RPE targets",
				Description: "RPE targets support scalar values and ranges.",
				Examples: []SyntaxExample{
					{Key: "rpe_scalar", Description: "RPE scalar.", Step: Step{Description: "Steady", Duration: 600, RPE: targetValue(6, "RPE")}},
					{Key: "rpe_range", Description: "RPE range.", Step: Step{Description: "Progression", Duration: 600, RPE: targetRange(6, 8, "RPE")}},
				},
			},
		},
		Limitations: []SyntaxLimitation{
			{Key: "one_primary_target", Description: "Each simple step can contain only one primary target among power, heart rate, pace, RPE, or freeride."},
			{Key: "ramp_requires_numeric_target", Description: "Ramp requires a power, heart-rate, pace, or RPE target with start and end bounds; text targets cannot be used for ramps."},
			{Key: "freeride_not_ramp", Description: "Freeride cannot be combined with ramp or another primary target."},
			{Key: "repeat_fields", Description: "Repeat blocks require reps greater than zero and child steps, cannot be nested, and cannot also carry simple-step fields."},
			{Key: "simple_step_duration_or_distance", Description: "Simple steps require a positive duration or a supported distance."},
		},
	}
}

func targetValue(value float64, units string) *Target {
	return &Target{Value: &value, Units: units}
}

func targetRange(minValue, maxValue float64, units string) *Target {
	return &Target{Min: &minValue, Max: &maxValue, Units: units}
}

func targetRamp(startValue, endValue float64, units string) *Target {
	return &Target{Start: &startValue, End: &endValue, Units: units}
}

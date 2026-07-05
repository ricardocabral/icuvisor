package workoutdoc

// SyntaxSpec describes the structured workout DSL supported by Serialize.
type SyntaxSpec struct {
	CheatSheet     SyntaxCheatSheet
	Features       []SyntaxFeature
	Limitations    []SyntaxLimitation
	CommonMistakes []SyntaxMistake
}

// SyntaxCheatSheet is a compact at-a-glance summary rendered at the top of the resource.
type SyntaxCheatSheet struct {
	Form     string
	Examples []SyntaxCheatExample
}

// SyntaxCheatExample is a minimal labelled DSL snippet for the cheat sheet.
type SyntaxCheatExample struct {
	Label string
	DSL   string
}

// SyntaxMistake captures a recurring authoring mistake worth flagging.
type SyntaxMistake struct {
	Key         string
	Description string
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

// DistanceUnitSyntax describes distance units accepted by Serialize.
type DistanceUnitSyntax struct {
	Key         string
	Canonical   string
	Aliases     []string
	Description string
}

// TargetUnitSyntax describes a target unit form accepted by Serialize.
type TargetUnitSyntax struct {
	Key         string
	Family      string
	Units       []string
	Prefix      string
	Suffix      string
	Zone        bool
	Description string
}

var workoutDistanceUnits = []DistanceUnitSyntax{
	{Key: "distance_mtr", Canonical: "mtr", Aliases: []string{"m", "meter", "meters", "metre", "metres", "mtr"}, Description: "Meters serialize with the canonical `mtr` suffix."},
	{Key: "distance_km", Canonical: "km", Aliases: []string{"km", "kilometer", "kilometers", "kilometre", "kilometres"}, Description: "Kilometers serialize with the canonical `km` suffix."},
	{Key: "distance_mi", Canonical: "mi", Aliases: []string{"mi", "mile", "miles"}, Description: "Miles serialize with the canonical `mi` suffix."},
	{Key: "distance_yd", Canonical: "yd", Aliases: []string{"yd", "yard", "yards"}, Description: "Yards serialize with the canonical `yd` suffix for pool-swim distances."},
}

var workoutTargetUnits = []TargetUnitSyntax{
	{Key: "power_percent_ftp", Family: "power", Units: []string{"", "PERCENT_FTP", "%FTP"}, Suffix: "%", Description: "Power as percent FTP; blank units default to percent FTP."},
	{Key: "power_watts", Family: "power", Units: []string{"WATTS", "WATT", "W"}, Suffix: "w", Description: "Absolute power in watts."},
	{Key: "power_zone", Family: "power", Units: []string{"ZONE", "POWER_ZONE"}, Zone: true, Description: "Power zones as `Zn` or `Zlow-Zhigh`."},
	{Key: "hr_lthr", Family: "hr", Units: []string{"PERCENT_LTHR", "%LTHR", "LTHR"}, Suffix: "% LTHR", Description: "Heart rate as percent lactate-threshold HR."},
	{Key: "hr_percent", Family: "hr", Units: []string{"PERCENT_HR", "PERCENT_MAX_HR", "%HR", "HR"}, Suffix: "% HR", Description: "Heart rate as percent max HR."},
	{Key: "hr_bpm", Family: "hr", Units: []string{"BPM"}, Suffix: "bpm", Description: "Heart rate in beats per minute."},
	{Key: "hr_zone", Family: "hr", Units: []string{"ZONE", "HR_ZONE"}, Suffix: " HR", Zone: true, Description: "Heart-rate zones."},
	{Key: "pace_percent", Family: "pace", Units: []string{"", "PERCENT_THRESHOLD", "PERCENT_THRESHOLD_PACE", "PERCENT_PACE", "%PACE"}, Suffix: "% Pace", Description: "Pace as percent threshold pace; blank units default to percent threshold pace."},
	{Key: "pace_zone", Family: "pace", Units: []string{"ZONE", "PACE_ZONE"}, Suffix: " Pace", Zone: true, Description: "Pace zones."},
	{Key: "pace_mins_km", Family: "pace", Units: []string{"MINS_KM"}, Suffix: "/km Pace", Description: "Absolute running pace in seconds per kilometer, serialized as `mm:ss/km Pace`."},
	{Key: "pace_mins_mile", Family: "pace", Units: []string{"MINS_MILE"}, Suffix: "/mi Pace", Description: "Absolute running pace in seconds per mile, serialized as `mm:ss/mi Pace`."},
	{Key: "pace_numeric", Family: "pace", Units: []string{"PACE"}, Suffix: " Pace", Description: "Numeric PACE values as currently emitted by the serializer."},
	{Key: "rpe", Family: "rpe", Units: []string{"", "RPE"}, Prefix: "RPE ", Description: "Rating of perceived exertion scalar or range."},
}

// WorkoutDistanceUnitSyntax returns distance units accepted by Serialize.
func WorkoutDistanceUnitSyntax() []DistanceUnitSyntax {
	out := make([]DistanceUnitSyntax, len(workoutDistanceUnits))
	copy(out, workoutDistanceUnits)
	return out
}

// WorkoutTargetUnitSyntax returns primary target unit forms accepted by Serialize.
func WorkoutTargetUnitSyntax() []TargetUnitSyntax {
	out := make([]TargetUnitSyntax, len(workoutTargetUnits))
	copy(out, workoutTargetUnits)
	return out
}

// WorkoutSyntaxSpec returns the workout DSL syntax supported by this package.
func WorkoutSyntaxSpec() SyntaxSpec {
	return SyntaxSpec{
		CheatSheet: SyntaxCheatSheet{
			Form: "Simple step: `- [description] [duration|distance] [primary target] [optional cadence]`. In structured WorkoutDoc JSON, step `description` is only a label/comment; put duration or distance in its own field, not in the label. Repeat block: `Nx` header with two-space-indented child steps. Use one primary target per step (power OR HR OR pace OR RPE OR freeride).",
			Examples: []SyntaxCheatExample{
				{Label: "Duration step", DSL: "- Endurance 10m 75%"},
				{Label: "Distance step", DSL: "- Stride 400mtr 120%"},
				{Label: "Yard swim step", DSL: "- Swim 100yd 95% Pace"},
				{Label: "Repeat block", DSL: "Main set 3x\n  - Hard 2m 105-115% 95-105rpm\n  - Easy 1m freeride"},
				{Label: "Ramp", DSL: "- Build 8m ramp 70-95%"},
			},
		},
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
				Description: "Distance steps serialize with canonical mtr, km, mi, or yd suffixes.",
				Examples: []SyntaxExample{
					{Key: "distance_mtr", Description: "Meter distance canonicalizes to mtr.", Step: Step{Description: "Stride", Distance: &Length{Value: 400, Unit: "meters"}, Power: targetValue(120, "PERCENT_FTP")}},
					{Key: "distance_km", Description: "Kilometer distance canonicalizes to km.", Step: Step{Description: "Tempo", Distance: &Length{Value: 5, Unit: "kilometers"}, Pace: targetRange(92, 96, "PERCENT_THRESHOLD")}},
					{Key: "distance_mi", Description: "Mile distance canonicalizes to mi.", Step: Step{Description: "Cooldown", Distance: &Length{Value: 1, Unit: "miles"}, Freeride: true}},
					{Key: "distance_yd", Description: "Yard distance canonicalizes to yd.", Step: Step{Description: "Swim", Distance: &Length{Value: 100, Unit: "yards"}, Pace: targetValue(95, "PERCENT_THRESHOLD")}},
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
				Description: "Pace targets support percent threshold pace, pace zones, absolute seconds-per-km or seconds-per-mile values, numeric PACE values, and non-ramp text pace labels.",
				Examples: []SyntaxExample{
					{Key: "pace_percent", Description: "Percent threshold pace scalar.", Step: Step{Description: "Cruise", Duration: 600, Pace: targetValue(95, "PERCENT_THRESHOLD")}},
					{Key: "pace_zone", Description: "Pace zone range.", Step: Step{Description: "Pace zone", Duration: 600, Pace: targetRange(2, 3, "PACE_ZONE")}},
					{Key: "pace_mins_km", Description: "Absolute seconds-per-km pace.", Step: Step{Description: "Metric pace", Duration: 300, Pace: targetValue(300, "MINS_KM")}},
					{Key: "pace_mins_mile", Description: "Absolute seconds-per-mile pace.", Step: Step{Description: "Imperial pace", Duration: 480, Pace: targetValue(480, "MINS_MILE")}},
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
			{Key: "step_description_label_only", Description: "Structured WorkoutDoc step descriptions are labels/comments only. Do not include duration or distance tokens there; use the duration or distance fields so the serialized DSL has exactly one duration/distance source."},
		},
		CommonMistakes: []SyntaxMistake{
			{Key: "m_is_minutes", Description: "`m` is minutes, never meters. Use `mtr` for meters (e.g. `500mtr`, not `500m`)."},
			{Key: "no_duration_or_distance_in_step_description", Description: "In structured WorkoutDoc JSON, do not put tokens like `2h15m`, `45m`, `400mtr`, or `5km` in a step description. Use exactly one source: duration seconds or distance fields."},
			{Key: "one_primary_target_per_step", Description: "One primary target per step. Use power OR HR OR pace OR RPE (plus optional cadence). Mixing primary targets in one step is rejected."},
			{Key: "no_nested_repeats", Description: "No nested repeats. An `Nx` block cannot contain another `Nx` block."},
			{Key: "repeat_header_carries_only_reps", Description: "Repeat headers carry only `Nx` and an optional label. Duration and targets belong on the child steps, not the header."},
			{Key: "ramps_need_numeric_targets", Description: "Ramps cannot use text or zone-label pace targets. Use percentages or absolute pace bounds for the start/end."},
			{Key: "sport_specific_labels", Description: "Structured step descriptions are emitted verbatim. For run or swim workouts, choose run/swim labels and avoid cycling-only words like ride or spin unless that prose came from the user."},
			{Key: "prose_and_steps_coexist", Description: "`workout_doc` and `description` coexist in the same description field — prose, headers, and comments pass through untouched while structured-step lines are validated. You do not need a separate event or note to attach coach/athlete prose alongside structure."},
			{Key: "preflight_validate", Description: "Pre-flight: when uncertain about syntax, call `validate_workout` (when registered) with the proposed `workout_doc` and/or `description` before writing to get a diagnostic. If the tool is not available the write tools still apply the same parser server-side."},
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

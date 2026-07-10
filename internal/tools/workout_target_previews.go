package tools

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/response"
	"github.com/ricardocabral/icuvisor/internal/units"
)

type workoutTargetPreviewContext struct {
	Profile    *intervals.AthleteWithSportSettings
	UnitSystem response.UnitSystem
	Sport      string
	Indoor     *bool
}

type workoutTargetPreviewRow struct {
	Step        int    `json:"step"`
	Path        string `json:"path"`
	Description string `json:"description,omitempty"`
	Family      string `json:"family"`
	Target      string `json:"target"`
	Preview     string `json:"preview"`
	Basis       string `json:"basis"`
	RepeatReps  int    `json:"repeat_reps,omitempty"`
}

type workoutTargetBounds struct {
	Values []float64
}

func workoutPreviewContextForEvent(event intervals.Event, profile intervals.AthleteWithSportSettings, unitSystem response.UnitSystem) workoutTargetPreviewContext {
	return workoutTargetPreviewContext{Profile: &profile, UnitSystem: unitSystem, Sport: stringValue(event.Type), Indoor: event.Indoor}
}

func workoutPreviewContextForWorkout(workout intervals.Workout, profile intervals.AthleteWithSportSettings, unitSystem response.UnitSystem) workoutTargetPreviewContext {
	return workoutTargetPreviewContext{Profile: &profile, UnitSystem: unitSystem, Sport: stringValue(workout.Type), Indoor: workout.Indoor}
}

func workoutTargetPreviews(value any, ctx workoutTargetPreviewContext) []workoutTargetPreviewRow {
	if ctx.Profile == nil {
		return nil
	}
	setting, ok := selectWorkoutPreviewSportSetting(ctx.Profile.SportSettings, ctx.Sport)
	if !ok {
		return nil
	}
	steps, ok := workoutDocSteps(value)
	if !ok {
		return nil
	}
	previews := make([]workoutTargetPreviewRow, 0)
	walkWorkoutPreviewSteps(steps, setting, ctx, nil, 0, &previews)
	return previews
}

func workoutDocSteps(value any) ([]any, bool) {
	switch typed := value.(type) {
	case map[string]any:
		steps, ok := typed["steps"].([]any)
		return steps, ok
	case []any:
		return typed, true
	default:
		return nil, false
	}
}

func walkWorkoutPreviewSteps(steps []any, setting intervals.SportSettings, ctx workoutTargetPreviewContext, parent []int, repeatReps int, previews *[]workoutTargetPreviewRow) {
	for index, raw := range steps {
		stepMap, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		pathParts := append(append([]int(nil), parent...), index+1)
		path := workoutPreviewPath(pathParts)
		stepNumber := index + 1
		description := anyString(stepMap["description"])
		for _, family := range []string{"power", "hr", "pace"} {
			targetMap, ok := stepMap[family].(map[string]any)
			if !ok {
				continue
			}
			if preview, ok := workoutTargetPreviewForFamily(family, targetMap, setting, ctx); ok {
				preview.Step = stepNumber
				preview.Path = path
				preview.Description = description
				if repeatReps > 0 {
					preview.RepeatReps = repeatReps
				}
				*previews = append(*previews, preview)
			}
		}
		if children, ok := stepMap["steps"].([]any); ok {
			childRepeat := repeatReps
			if reps := intFromAny(stepMap["reps"]); reps > 0 {
				childRepeat = reps
			}
			walkWorkoutPreviewSteps(children, setting, ctx, pathParts, childRepeat, previews)
		}
	}
}

func workoutTargetPreviewForFamily(family string, target map[string]any, setting intervals.SportSettings, ctx workoutTargetPreviewContext) (workoutTargetPreviewRow, bool) {
	if strings.TrimSpace(anyString(target["text"])) != "" {
		return workoutTargetPreviewRow{}, false
	}
	bounds, ok := workoutTargetNumericBounds(target)
	if !ok {
		return workoutTargetPreviewRow{}, false
	}
	units := strings.ToUpper(strings.TrimSpace(anyString(target["units"])))
	switch family {
	case "power":
		if units != "" && units != "PERCENT_FTP" && units != "%FTP" {
			return workoutTargetPreviewRow{}, false
		}
		ftp := setting.FTP
		if ctx.Indoor != nil && *ctx.Indoor && setting.IndoorFTP > 0 {
			ftp = setting.IndoorFTP
		}
		if ftp <= 0 {
			return workoutTargetPreviewRow{}, false
		}
		return workoutTargetPreviewRow{Family: "power", Target: percentTargetLabel(bounds, "% FTP"), Preview: integerTargetPreview(bounds, float64(ftp), "W"), Basis: fmt.Sprintf("ftp %d W", ftp)}, true
	case "hr":
		switch units {
		case "PERCENT_LTHR", "%LTHR", "LTHR":
			lthr := firstNonZero(setting.LTHR, setting.FTHR)
			if lthr <= 0 {
				return workoutTargetPreviewRow{}, false
			}
			return workoutTargetPreviewRow{Family: "hr", Target: percentTargetLabel(bounds, "% LTHR"), Preview: integerTargetPreview(bounds, float64(lthr), "bpm"), Basis: fmt.Sprintf("lthr %d bpm", lthr)}, true
		case "PERCENT_HR", "PERCENT_MAX_HR", "%HR", "HR":
			if setting.MaxHR <= 0 {
				return workoutTargetPreviewRow{}, false
			}
			return workoutTargetPreviewRow{Family: "hr", Target: percentTargetLabel(bounds, "% HR"), Preview: integerTargetPreview(bounds, float64(setting.MaxHR), "bpm"), Basis: fmt.Sprintf("max_hr %d bpm", setting.MaxHR)}, true
		default:
			return workoutTargetPreviewRow{}, false
		}
	case "pace":
		if units != "" && units != "PERCENT_THRESHOLD" && units != "PERCENT_THRESHOLD_PACE" && units != "PERCENT_PACE" && units != "%PACE" {
			return workoutTargetPreviewRow{}, false
		}
		threshold := firstNonZeroFloat(setting.ThresholdPace, setting.PaceThreshold)
		if threshold <= 0 {
			return workoutTargetPreviewRow{}, false
		}
		preview, basis, ok := paceTargetPreview(bounds, threshold, setting.PaceUnits, ctx.UnitSystem)
		if !ok {
			return workoutTargetPreviewRow{}, false
		}
		return workoutTargetPreviewRow{Family: "pace", Target: percentTargetLabel(bounds, "% Pace"), Preview: preview, Basis: basis}, true
	default:
		return workoutTargetPreviewRow{}, false
	}
}

func workoutTargetNumericBounds(target map[string]any) (workoutTargetBounds, bool) {
	if value, ok := floatFromAny(target["value"]); ok {
		return workoutTargetBounds{Values: []float64{value}}, true
	}
	if min, okMin := floatFromAny(target["min"]); okMin {
		if max, okMax := floatFromAny(target["max"]); okMax {
			return workoutTargetBounds{Values: []float64{min, max}}, true
		}
	}
	if start, okStart := floatFromAny(target["start"]); okStart {
		if end, okEnd := floatFromAny(target["end"]); okEnd {
			return workoutTargetBounds{Values: []float64{start, end}}, true
		}
	}
	return workoutTargetBounds{}, false
}

func selectWorkoutPreviewSportSetting(settings []intervals.SportSettings, sport string) (intervals.SportSettings, bool) {
	if setting, ok := findSportSetting(settings, sport); ok {
		return setting, true
	}
	if len(settings) == 1 {
		return settings[0], true
	}
	return intervals.SportSettings{}, false
}

func percentTargetLabel(bounds workoutTargetBounds, suffix string) string {
	return formatTargetNumbers(bounds) + suffix
}

func integerTargetPreview(bounds workoutTargetBounds, basis float64, suffix string) string {
	values := make([]float64, 0, len(bounds.Values))
	for _, percent := range bounds.Values {
		values = append(values, math.Round(basis*percent/100))
	}
	return formatIntegerValues(values) + " " + suffix
}

func paceTargetPreview(bounds workoutTargetBounds, thresholdMetersPerSecond float64, sourceUnit string, unitSystem response.UnitSystem) (string, string, bool) {
	paceUnit, suffix := preferredPacePreviewUnit(sourceUnit, unitSystem)
	basisSeconds, ok := response.PaceSecondsFromMetersPerSecond(thresholdMetersPerSecond, paceUnit)
	if !ok || !paceSecondsAreFormattable(basisSeconds) {
		return "", "", false
	}
	values := make([]float64, 0, len(bounds.Values))
	for _, percent := range bounds.Values {
		if percent <= 0 || math.IsNaN(percent) || math.IsInf(percent, 0) {
			return "", "", false
		}
		seconds := basisSeconds * 100 / percent
		if !paceSecondsAreFormattable(seconds) {
			return "", "", false
		}
		values = append(values, math.Round(seconds))
	}
	return formatPaceValues(values, suffix), fmt.Sprintf("threshold pace %s", formatPaceSeconds(math.Round(basisSeconds), suffix)), true
}

func paceSecondsAreFormattable(seconds float64) bool {
	if seconds <= 0 || math.IsNaN(seconds) || math.IsInf(seconds, 0) {
		return false
	}
	return math.Round(seconds) < math.Exp2(float64(strconv.IntSize-1))
}

func preferredPacePreviewUnit(sourceUnit string, unitSystem response.UnitSystem) (units.Unit, string) {
	paceUnit, _ := units.ParseUnit(sourceUnit)
	switch paceUnit {
	case units.UnitMinsKM:
		return paceUnit, "/km"
	case units.UnitMinsMile:
		return paceUnit, "/mi"
	case units.UnitSecs100M:
		return paceUnit, "/100m"
	case units.UnitSecs100Y:
		return paceUnit, "/100y"
	case units.UnitSecs500M:
		return paceUnit, "/500m"
	case units.UnitSecs400M:
		return paceUnit, "/400m"
	case units.UnitSecs250M:
		return paceUnit, "/250m"
	}
	if unitSystem == response.UnitSystemImperial {
		return units.UnitMinsMile, "/mi"
	}
	return units.UnitMinsKM, "/km"
}

func formatTargetNumbers(bounds workoutTargetBounds) string {
	values := make([]string, 0, len(bounds.Values))
	for _, value := range bounds.Values {
		values = append(values, formatFloatCompact(value))
	}
	return strings.Join(values, "-")
}

func formatIntegerValues(values []float64) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, fmt.Sprintf("%.0f", value))
	}
	return strings.Join(parts, "-")
}

func formatPaceValues(values []float64, suffix string) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, formatPaceSeconds(value, suffix))
	}
	return strings.Join(parts, "-")
}

func formatPaceSeconds(seconds float64, suffix string) string {
	whole := int(math.Round(seconds))
	minutes := whole / 60
	remaining := whole % 60
	return fmt.Sprintf("%d:%02d%s", minutes, remaining, suffix)
}

func formatFloatCompact(value float64) string {
	if math.Abs(value-math.Round(value)) < 0.0000001 {
		return fmt.Sprintf("%.0f", math.Round(value))
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", value), "0"), ".")
}

func workoutPreviewPath(parts []int) string {
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		values = append(values, fmt.Sprintf("%d", part))
	}
	return strings.Join(values, ".")
}

func floatFromAny(value any) (float64, bool) {
	switch typed := value.(type) {
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case int:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case json.Number:
		parsed, err := typed.Float64()
		return parsed, err == nil
	default:
		return 0, false
	}
}

func intFromAny(value any) int {
	if parsed, ok := floatFromAny(value); ok {
		return int(parsed)
	}
	return 0
}

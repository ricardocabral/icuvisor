package tools

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/workoutdoc"
)

// workoutDocUnrenderedWarning explains that intervals.icu stored the write but did not
// parse the uploaded workout_doc into a structured workout, so it renders as plain text.
const workoutDocUnrenderedWarning = "intervals.icu saved this but did not parse the uploaded workout_doc into structured steps; it will display as plain text without graphical interval segments. The serialized workout DSL may not match the upstream workout grammar."

const workoutDocPartialFidelityWarning = "intervals.icu saved this and parsed structured steps, but the returned workout_doc differs from the uploaded workout_doc; it may have partially parsed the DSL and dropped, reordered, or changed some structured fields."

type upstreamWorkoutStep struct {
	value map[string]any
	used  bool
}

// workoutDocRenderWarning returns a warning when a structured workout_doc with steps was
// uploaded but the upstream response shows it was not parsed, or parsed only partially.
func workoutDocRenderWarning(uploaded *workoutdoc.WorkoutDoc, upstreamDoc any) string {
	if uploaded == nil || len(uploaded.Steps) == 0 {
		return ""
	}
	if !workoutDocHasSteps(upstreamDoc) {
		return workoutDocUnrenderedWarning
	}
	if !reflect.DeepEqual(uploadedWorkoutDocSignature(uploaded.Steps), upstreamWorkoutDocSignature(upstreamDoc)) {
		return workoutDocPartialFidelityWarning
	}
	return ""
}

func workoutDocLossyFields(uploaded *workoutdoc.WorkoutDoc, upstreamDoc any) []string {
	if uploaded == nil || len(uploaded.Steps) == 0 {
		return nil
	}
	upstream := flattenUpstreamWorkoutSteps(upstreamDoc)
	if len(upstream) == 0 {
		return []string{"workout_doc.steps"}
	}
	lossy := make([]string, 0)
	for index, step := range uploaded.Steps {
		collectWorkoutStepLossiness(step, fmt.Sprintf("steps[%d]", index), upstream, &lossy)
	}
	return uniqueWorkoutLossyFields(lossy)
}

func flattenUpstreamWorkoutSteps(value any) []*upstreamWorkoutStep {
	steps, ok := workoutDocSteps(value)
	if !ok {
		return nil
	}
	out := make([]*upstreamWorkoutStep, 0, len(steps))
	var walk func([]any)
	walk = func(items []any) {
		for _, item := range items {
			step, ok := item.(map[string]any)
			if !ok {
				continue
			}
			out = append(out, &upstreamWorkoutStep{value: step})
			if children, ok := step["steps"].([]any); ok {
				walk(children)
			}
		}
	}
	walk(steps)
	return out
}

func collectWorkoutStepLossiness(uploaded workoutdoc.Step, path string, upstream []*upstreamWorkoutStep, lossy *[]string) {
	match := matchUpstreamWorkoutStep(uploaded, upstream)
	if match == nil {
		if uploaded.Reps > 0 || len(uploaded.Steps) > 0 {
			*lossy = append(*lossy, path+".reps", path+".steps")
			for index, child := range uploaded.Steps {
				collectWorkoutStepLossiness(child, fmt.Sprintf("%s.steps[%d]", path, index), upstream, lossy)
			}
			return
		}
		*lossy = append(*lossy, path)
		return
	}
	if uploaded.Reps > 0 || len(uploaded.Steps) > 0 {
		if int(math.Round(anyFloat(match["reps"]))) != uploaded.Reps {
			*lossy = append(*lossy, path+".reps")
		}
		if _, ok := match["steps"].([]any); !ok {
			*lossy = append(*lossy, path+".steps")
		}
		for index, child := range uploaded.Steps {
			collectWorkoutStepLossiness(child, fmt.Sprintf("%s.steps[%d]", path, index), upstream, lossy)
		}
		return
	}
	if uploaded.Duration > 0 && int(math.Round(anyFloat(match["duration"]))) != uploaded.Duration {
		*lossy = append(*lossy, path+".duration")
	}
	if uploaded.Distance != nil {
		if distance, ok := optionalFloat(firstPresent(match, "distance")); !ok || math.Abs(distance-uploadedDistanceMeters(uploaded.Distance)) > 0.01 {
			*lossy = append(*lossy, path+".distance")
		}
	}
	for _, target := range []struct {
		family   string
		uploaded *workoutdoc.Target
	}{
		{family: "power", uploaded: uploaded.Power},
		{family: "hr", uploaded: uploaded.HR},
		{family: "pace", uploaded: uploaded.Pace},
		{family: "rpe", uploaded: uploaded.RPE},
		{family: "cadence", uploaded: uploaded.Cadence},
	} {
		if target.uploaded == nil {
			continue
		}
		upstreamTarget, ok := match[target.family].(map[string]any)
		if !ok || !targetsPreserveMeaning(target.family, *target.uploaded, upstreamTarget) {
			*lossy = append(*lossy, path+"."+target.family)
		}
	}
	if uploaded.Freeride && !anyBool(match["freeride"]) {
		*lossy = append(*lossy, path+".freeride")
	}
	if uploaded.PressLap && !anyBool(firstPresent(match, "press_lap", "until_lap_press")) {
		*lossy = append(*lossy, path+".press_lap")
	}
	if uploaded.Ramp && !anyBool(match["ramp"]) {
		*lossy = append(*lossy, path+".ramp")
	}
}

func matchUpstreamWorkoutStep(uploaded workoutdoc.Step, upstream []*upstreamWorkoutStep) map[string]any {
	wantText := signatureText(uploaded.Description)
	wantRepeat := uploaded.Reps > 0 || len(uploaded.Steps) > 0
	for _, candidate := range upstream {
		if candidate.used {
			continue
		}
		gotText := signatureText(anyString(firstPresent(candidate.value, "text", "description")))
		_, gotRepeat := candidate.value["steps"].([]any)
		if gotText == wantText && (wantText != "" || wantRepeat == gotRepeat) {
			candidate.used = true
			return candidate.value
		}
	}
	return nil
}

func targetsPreserveMeaning(family string, uploaded workoutdoc.Target, upstream map[string]any) bool {
	uploadedValues, uploadedOK := workoutTargetValues(uploaded.Value, uploaded.Min, uploaded.Max, uploaded.Start, uploaded.End)
	upstreamValues, upstreamOK := workoutTargetValuesFromMap(upstream)
	if !uploadedOK || !upstreamOK || len(uploadedValues) != len(upstreamValues) {
		return false
	}
	for index := range uploadedValues {
		if math.Abs(uploadedValues[index]-upstreamValues[index]) > 0.000001 {
			return false
		}
	}
	return canonicalWorkoutTargetUnit(family, uploaded.Units) == canonicalWorkoutTargetUnit(family, anyString(upstream["units"]))
}

func workoutTargetValues(value, minValue, maxValue, start, end *float64) ([]float64, bool) {
	if value != nil {
		return []float64{*value}, true
	}
	if minValue != nil && maxValue != nil {
		return []float64{*minValue, *maxValue}, true
	}
	if start != nil && end != nil {
		return []float64{*start, *end}, true
	}
	return nil, false
}

func workoutTargetValuesFromMap(target map[string]any) ([]float64, bool) {
	if value, ok := optionalFloat(target["value"]); ok {
		return []float64{value}, true
	}
	for _, keys := range [][2]string{{"min", "max"}, {"start", "end"}} {
		lo, okLo := optionalFloat(target[keys[0]])
		hi, okHi := optionalFloat(target[keys[1]])
		if okLo && okHi {
			return []float64{lo, hi}, true
		}
	}
	return nil, false
}

func canonicalWorkoutTargetUnit(family string, value string) string {
	unit := strings.ToUpper(strings.TrimSpace(value))
	switch unit {
	case "":
		switch family {
		case "power":
			return "PERCENT_FTP"
		case "pace":
			return "PERCENT_THRESHOLD"
		case "rpe":
			return "RPE"
		case "cadence":
			return "RPM"
		}
		return ""
	case "PERCENT_FTP", "%FTP":
		return "PERCENT_FTP"
	case "WATTS", "WATT", "W":
		return "WATTS"
	case "PERCENT_LTHR", "%LTHR", "LTHR":
		return "PERCENT_LTHR"
	case "PERCENT_HR", "PERCENT_MAX_HR", "%HR", "HR":
		return "PERCENT_HR"
	case "PERCENT_THRESHOLD", "PERCENT_THRESHOLD_PACE", "PERCENT_PACE", "%PACE":
		return "PERCENT_THRESHOLD"
	case "ZONE":
		return strings.ToUpper(family) + "_ZONE"
	case "POWER_ZONE":
		return "POWER_ZONE"
	case "HR_ZONE":
		return "HR_ZONE"
	case "PACE_ZONE":
		return "PACE_ZONE"
	case "MINS_KM", "SECS/KM":
		return "MINS_KM"
	case "MINS_MILE", "SECS/MI":
		return "MINS_MILE"
	case "SECS_100M", "SECS/100M":
		return "SECS_100M"
	case "SECS_100Y", "SECS/100Y":
		return "SECS_100Y"
	case "RPM", "CADENCE":
		return "RPM"
	case "RPE":
		return "RPE"
	default:
		return unit
	}
}

func uniqueWorkoutLossyFields(values []string) []string {
	seen := make(map[string]bool, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

// workoutDocHasSteps reports whether an upstream workout_doc payload parsed into at least one step.
func workoutDocHasSteps(value any) bool {
	switch typed := value.(type) {
	case map[string]any:
		steps, ok := typed["steps"].([]any)
		return ok && len(steps) > 0
	case []any:
		return len(typed) > 0
	default:
		return false
	}
}

func uploadedWorkoutDocSignature(steps []workoutdoc.Step) []string {
	signature := make([]string, 0, len(steps))
	for _, step := range steps {
		if step.Reps > 0 || len(step.Steps) > 0 {
			signature = append(signature, fmt.Sprintf("repeat|text=%s|reps=%d", signatureText(step.Description), step.Reps))
			signature = append(signature, uploadedWorkoutDocSignature(step.Steps)...)
			continue
		}
		signature = append(signature, simpleStepSignature(signatureStep{
			text:          step.Description,
			duration:      step.Duration,
			distance:      uploadedDistanceMeters(step.Distance),
			targetFamily:  uploadedTargetFamily(step),
			cadence:       step.Cadence != nil,
			freeride:      step.Freeride,
			pressLap:      step.PressLap,
			distanceBased: step.Distance != nil,
		}))
	}
	return signature
}

func upstreamWorkoutDocSignature(value any) []string {
	switch typed := value.(type) {
	case map[string]any:
		return upstreamWorkoutDocSignature(typed["steps"])
	case []any:
		signature := make([]string, 0, len(typed))
		for _, item := range typed {
			step, ok := item.(map[string]any)
			if !ok {
				continue
			}
			if childSteps, ok := step["steps"].([]any); ok {
				signature = append(signature, fmt.Sprintf("repeat|text=%s|reps=%d", signatureText(anyString(firstPresent(step, "text", "description"))), int(math.Round(anyFloat(firstPresent(step, "reps"))))))
				signature = append(signature, upstreamWorkoutDocSignature(childSteps)...)
				continue
			}
			distance, hasDistance := optionalFloat(firstPresent(step, "distance"))
			signature = append(signature, simpleStepSignature(signatureStep{
				text:          anyString(firstPresent(step, "text", "description")),
				duration:      int(math.Round(anyFloat(firstPresent(step, "duration")))),
				distance:      distance,
				targetFamily:  upstreamTargetFamily(step),
				cadence:       firstPresent(step, "cadence") != nil,
				freeride:      anyBool(firstPresent(step, "freeride")),
				pressLap:      anyBool(firstPresent(step, "press_lap")),
				distanceBased: hasDistance,
			}))
		}
		return signature
	default:
		return nil
	}
}

type signatureStep struct {
	text          string
	duration      int
	distance      float64
	targetFamily  string
	cadence       bool
	freeride      bool
	pressLap      bool
	distanceBased bool
}

func simpleStepSignature(step signatureStep) string {
	measure := fmt.Sprintf("duration=%d", step.duration)
	if step.distanceBased {
		measure = "distance_m=" + formatSignatureFloat(step.distance)
	}
	return strings.Join([]string{
		"step",
		"text=" + signatureText(step.text),
		measure,
		"target=" + step.targetFamily,
		"cadence=" + strconv.FormatBool(step.cadence),
		"freeride=" + strconv.FormatBool(step.freeride),
		"press_lap=" + strconv.FormatBool(step.pressLap),
	}, "|")
}

func uploadedTargetFamily(step workoutdoc.Step) string {
	switch {
	case step.Power != nil:
		return "power"
	case step.HR != nil:
		return "hr"
	case step.Pace != nil:
		return "pace"
	case step.RPE != nil:
		return "rpe"
	case step.Freeride:
		return "freeride"
	default:
		return ""
	}
}

func upstreamTargetFamily(step map[string]any) string {
	for _, key := range []string{"power", "hr", "pace", "rpe"} {
		if firstPresent(step, key) != nil {
			return key
		}
	}
	if anyBool(firstPresent(step, "freeride")) {
		return "freeride"
	}
	return ""
}

func uploadedDistanceMeters(distance *workoutdoc.Length) float64 {
	if distance == nil {
		return 0
	}
	switch strings.ToLower(strings.TrimSpace(distance.Unit)) {
	case "m", "meter", "meters", "metre", "metres", "mtr":
		return distance.Value
	case "km", "kilometer", "kilometers", "kilometre", "kilometres":
		return distance.Value * 1000
	case "mi", "mile", "miles":
		return distance.Value * 1609.34
	case "yrd", "yd", "yard", "yards":
		return distance.Value * 0.9144
	default:
		return distance.Value
	}
}

func signatureText(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func formatSignatureFloat(value float64) string {
	rounded := math.Round(value*100) / 100
	if math.Trunc(rounded) == rounded {
		return fmt.Sprintf("%.0f", rounded)
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", rounded), "0"), ".")
}

func firstPresent(values map[string]any, keys ...string) any {
	for _, key := range keys {
		if value, ok := values[key]; ok && value != nil {
			return value
		}
	}
	return nil
}

func optionalFloat(value any) (float64, bool) {
	if value == nil {
		return 0, false
	}
	return anyFloat(value), true
}

func anyFloat(value any) float64 {
	switch typed := value.(type) {
	case float64:
		return typed
	case float32:
		return float64(typed)
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	default:
		return 0
	}
}

func anyBool(value any) bool {
	typed, _ := value.(bool)
	return typed
}

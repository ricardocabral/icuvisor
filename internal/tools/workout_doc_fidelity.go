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

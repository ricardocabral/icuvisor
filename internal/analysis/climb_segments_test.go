package analysis

import (
	"math"
	"testing"
)

func climbTestStream(values []float64) ClimbStream {
	return ClimbStream{Values: append([]float64(nil), values...), Valid: validClimbValues(values), Present: true, RawLength: len(values)}
}

func climbTestStreamWithValidity(values []float64, valid []bool) ClimbStream {
	return ClimbStream{Values: append([]float64(nil), values...), Valid: append([]bool(nil), valid...), Present: true, RawLength: len(values)}
}

func validClimbValues(values []float64) []bool {
	valid := make([]bool, len(values))
	for i, value := range values {
		valid[i] = finite(value)
	}
	return valid
}

func climbTestInput(distance, altitude []float64) ClimbSegmentsInput {
	return ClimbSegmentsInput{
		Distance:                 distanceStream(distance),
		Altitude:                 climbTestStream(altitude),
		MinGradePercent:          3,
		MinElevationGainM:        0,
		MaxGapDistanceM:          100,
		MaxBridgedElevationLossM: 5,
	}
}

func distanceStream(values []float64) ClimbStream {
	return climbTestStream(values)
}

func TestAnalyzeClimbSegmentsPreservesNonGridEndpointsAndInterpolates(t *testing.T) {
	input := climbTestInput([]float64{0.3, 2.3}, []float64{100, 101})
	input.MinGradePercent = 3
	input.MinElevationGainM = 0.5
	input.Time = climbTestStream([]float64{10, 20})
	got, err := AnalyzeClimbSegments(input)
	if err != nil {
		t.Fatalf("AnalyzeClimbSegments() error = %v", err)
	}
	if got.DataQuality.Status != ClimbStatusOK || len(got.Segments) != 1 {
		t.Fatalf("result = %#v, want one ok segment", got)
	}
	segment := got.Segments[0]
	if segment.StartDistanceM != 0.3 || segment.EndDistanceM != 2.3 || segment.DistanceM != 2 || segment.ElevationGainM != 1 || segment.AverageGradePercent != 50 {
		t.Fatalf("segment = %#v, want preserved endpoints and six-decimal metrics", segment)
	}
	if segment.DurationSeconds == nil || *segment.DurationSeconds != 10 || segment.VAMMPerHour == nil || *segment.VAMMPerHour != 360 {
		t.Fatalf("time metrics = %#v, want duration 10 and VAM 360", segment)
	}
}

func TestAnalyzeClimbSegmentsCollapsesDuplicateDistanceWithoutPauseGain(t *testing.T) {
	input := climbTestInput([]float64{0, 0, 10}, []float64{100, 105, 110})
	input.MinElevationGainM = 5
	input.HeartRate = climbTestStream([]float64{140, 180, 160})
	got, err := AnalyzeClimbSegments(input)
	if err != nil {
		t.Fatalf("AnalyzeClimbSegments() error = %v", err)
	}
	if len(got.Segments) != 1 {
		t.Fatalf("segments = %#v, want one segment", got.Segments)
	}
	segment := got.Segments[0]
	if segment.DistanceM != 10 || segment.ElevationGainM != 10 || segment.AverageHeartRateBPM == nil || *segment.AverageHeartRateBPM != 150 {
		t.Fatalf("segment = %#v, want duplicate-selected first source HR and no pause gain", segment)
	}
}

func TestAnalyzeClimbSegmentsBridgesOnlyWhenMergedGuardsPass(t *testing.T) {
	input := climbTestInput([]float64{0, 10, 20, 30}, []float64{0, 10, 9, 19})
	input.MinElevationGainM = 15
	got, err := AnalyzeClimbSegments(input)
	if err != nil {
		t.Fatalf("AnalyzeClimbSegments() error = %v", err)
	}
	if len(got.Segments) != 1 || got.Segments[0].DistanceM != 30 || got.Segments[0].ElevationGainM != 19 {
		t.Fatalf("bridged result = %#v, want one post-bridge segment", got.Segments)
	}

	input.MaxBridgedElevationLossM = 0
	got, err = AnalyzeClimbSegments(input)
	if err != nil {
		t.Fatalf("loss-guard analysis error = %v", err)
	}
	if len(got.Segments) != 0 || got.DataQuality.Status != ClimbStatusNoClimb {
		t.Fatalf("loss-guard result = %#v, want filtered separate runs", got)
	}

	input = climbTestInput([]float64{0, 10, 110, 120}, []float64{0, 10, 9, 19})
	input.MinGradePercent = 20
	input.MinElevationGainM = 0
	input.MaxGapDistanceM = 100
	got, err = AnalyzeClimbSegments(input)
	if err != nil {
		t.Fatalf("grade-guard analysis error = %v", err)
	}
	if len(got.Segments) != 2 || got.Segments[0].DistanceM != 10 || got.Segments[1].StartDistanceM != 110 {
		t.Fatalf("grade-guard segments = %#v, want two unmerged candidates", got.Segments)
	}
}

func TestAnalyzeClimbSegmentsLocalizesNullWindowsAndReportsQuality(t *testing.T) {
	input := climbTestInput([]float64{0, 10, 20, 30, 40}, []float64{0, 10, 0, 20, 30})
	input.Altitude = climbTestStreamWithValidity(input.Altitude.Values, []bool{true, true, false, true, true})
	input.Altitude.NullCount = 1
	input.MinElevationGainM = 5
	got, err := AnalyzeClimbSegments(input)
	if err != nil {
		t.Fatalf("AnalyzeClimbSegments() error = %v", err)
	}
	if got.DataQuality.Status != ClimbStatusNull || len(got.Segments) != 2 {
		t.Fatalf("null result = %#v, want null quality and two local segments", got)
	}
	if got.Segments[0].StartDistanceM != 0 || got.Segments[1].StartDistanceM != 30 {
		t.Fatalf("segments = %#v, want ordered windows", got.Segments)
	}
	if got.DataQuality.BrokenWindows != 1 || got.DataQuality.NullAltitudeSamples != 1 {
		t.Fatalf("quality = %#v, want one broken/null sample", got.DataQuality)
	}
}

func TestAnalyzeClimbSegmentsQualityPrecedenceAndOptionalCoverage(t *testing.T) {
	missing := climbTestInput(nil, nil)
	missing.Distance = ClimbStream{Present: true, DataState: "empty"}
	missing.Altitude = ClimbStream{Present: true, DataState: "empty"}
	got, err := AnalyzeClimbSegments(missing)
	if err != nil || got.DataQuality.Status != ClimbStatusMissing || got.DataQuality.UsableSamples != 0 {
		t.Fatalf("missing = %#v, err %v", got, err)
	}

	invalid := climbTestInput([]float64{0, 10}, []float64{0, 10})
	invalid.Distance.Valid[1] = false
	invalid.Distance.NullCount = 1
	got, err = AnalyzeClimbSegments(invalid)
	if err != nil || got.DataQuality.Status != ClimbStatusInvalidDistance || len(got.Segments) != 0 {
		t.Fatalf("invalid distance = %#v, err %v", got, err)
	}
	invalid = climbTestInput([]float64{0, 10, 5, 20}, []float64{0, 1, 2, 3})
	got, err = AnalyzeClimbSegments(invalid)
	if err != nil || got.DataQuality.Status != ClimbStatusInvalidDistance {
		t.Fatalf("non-monotone distance = %#v, err %v", got, err)
	}
	invalid.Altitude = ClimbStream{Present: true, DataState: "data_null"}
	got, err = AnalyzeClimbSegments(invalid)
	if err != nil || got.DataQuality.Status != ClimbStatusInvalidDistance {
		t.Fatalf("non-monotone null altitude = %#v, err %v", got, err)
	}

	noisy := climbTestInput([]float64{0, 10, 20}, []float64{0, 20, 21})
	noisy.MinElevationGainM = 5
	got, err = AnalyzeClimbSegments(noisy)
	if err != nil || got.DataQuality.Status != ClimbStatusNoisy || got.DataQuality.NoisyTransitions != 10 {
		t.Fatalf("noisy = %#v, err %v", got, err)
	}
	noisyBridge := climbTestInput([]float64{0, 10, 20, 30}, []float64{0, 10, 21, 31})
	noisyBridge.MinElevationGainM = 0
	got, err = AnalyzeClimbSegments(noisyBridge)
	if err != nil || len(got.Segments) != 2 || got.Segments[0].EndDistanceM != 10 || got.Segments[1].StartDistanceM != 20 {
		t.Fatalf("noisy bridge result = %#v, err %v, want hard evidence boundary", got, err)
	}

	input := climbTestInput([]float64{0, 10, 20}, []float64{0, 1, 2})
	input.MinElevationGainM = 1
	input.HeartRate = climbTestStreamWithValidity([]float64{100, math.NaN(), 120}, []bool{true, false, true})
	input.HeartRate.InvalidCount = 1
	input.Watts = ClimbStream{Present: true, DataState: "null"}
	got, err = AnalyzeClimbSegments(input)
	if err != nil {
		t.Fatalf("optional analysis error = %v", err)
	}
	if got.DataQuality.OptionalStreams["heart_rate"].Status != ClimbOptionalNonFinite || got.DataQuality.OptionalStreams["heart_rate"].CoveragePercent != 66.666667 {
		t.Fatalf("heart rate quality = %#v", got.DataQuality.OptionalStreams["heart_rate"])
	}
	if got.DataQuality.OptionalStreams["watts"].Status != ClimbOptionalNull {
		t.Fatalf("watts quality = %#v", got.DataQuality.OptionalStreams["watts"])
	}
	if got.Segments[0].AverageHeartRateBPM == nil || *got.Segments[0].AverageHeartRateBPM != 110 {
		t.Fatalf("partial heart rate metric = %#v, want finite mean", got.Segments[0])
	}
	mismatch := climbTestInput([]float64{0, 10, 20}, []float64{0, 1, 2})
	mismatch.MinElevationGainM = 1
	mismatch.HeartRate = climbTestStream([]float64{100, 110})
	mismatch.Watts = climbTestStream([]float64{200, 210})
	got, err = AnalyzeClimbSegments(mismatch)
	if err != nil || got.DataQuality.OptionalStreams["heart_rate"].Status != ClimbOptionalLengthMismatch || got.DataQuality.OptionalStreams["watts"].Status != ClimbOptionalLengthMismatch {
		t.Fatalf("mismatched optional quality = %#v, err %v", got.DataQuality.OptionalStreams, err)
	}
	if got.Segments[0].AverageHeartRateBPM != nil || got.Segments[0].AveragePowerWatts != nil {
		t.Fatalf("mismatched optional metrics = %#v, want omitted", got.Segments[0])
	}

	vamInput := climbTestInput([]float64{0.3, 2.3}, []float64{0, 1})
	vamInput.MinElevationGainM = 0.5
	vamInput.Time = climbTestStream([]float64{0, 1.23456789})
	got, err = AnalyzeClimbSegments(vamInput)
	if err != nil || got.Segments[0].VAMMPerHour == nil {
		t.Fatalf("VAM result = %#v, err %v", got, err)
	}
	wantVAM := round6(3600 / 1.23456789)
	if *got.Segments[0].VAMMPerHour != wantVAM {
		t.Fatalf("VAM = %v, want unrounded-duration result %v", *got.Segments[0].VAMMPerHour, wantVAM)
	}
}

package analysis

import (
	"errors"
	"fmt"
	"math"
	"sort"
)

const (
	DefaultClimbMinGradePercent   = 3.0
	DefaultClimbMinElevationGainM = 30.0
	DefaultClimbMaxGapDistanceM   = 100.0
	DefaultClimbMaxBridgedLossM   = 5.0
	ClimbResampleM                = 1.0
	ClimbSegmentLimit             = 100
)

const (
	ClimbStatusOK               = "ok"
	ClimbStatusNoClimb          = "no_climb"
	ClimbStatusMissing          = "missing"
	ClimbStatusNull             = "null"
	ClimbStatusInsufficient     = "insufficient"
	ClimbStatusFlat             = "flat"
	ClimbStatusNoisy            = "noisy"
	ClimbStatusInvalidDistance  = "invalid_distance"
	ClimbOptionalAbsent         = "absent"
	ClimbOptionalNull           = "null"
	ClimbOptionalEmpty          = "empty"
	ClimbOptionalAllNull        = "all_null"
	ClimbOptionalLengthMismatch = "length_mismatch"
	ClimbOptionalNonFinite      = "non_finite"
	ClimbOptionalNonMonotone    = "non_monotone"
	ClimbOptionalPartial        = "partial"
	ClimbOptionalOK             = "ok"
)

var ErrInvalidClimbSegmentsInput = errors.New("invalid climb segments input")

// ClimbStream is the typed stream boundary used by the climb analyzer.
// Valid preserves null and non-finite evidence instead of treating it as zero.
type ClimbStream struct {
	Values       []float64
	Valid        []bool
	Present      bool
	NullCount    int
	InvalidCount int
	AllNull      bool
	RawLength    int
	DataState    string
}

// ClimbSegmentsInput contains normalized streams and effective analyzer parameters.
type ClimbSegmentsInput struct {
	Distance                 ClimbStream
	Altitude                 ClimbStream
	Time                     ClimbStream
	HeartRate                ClimbStream
	Watts                    ClimbStream
	MinGradePercent          float64
	MinElevationGainM        float64
	MaxGapDistanceM          float64
	MaxBridgedElevationLossM float64
}

// ClimbParameters records the bounded parameters used for one analysis.
type ClimbParameters struct {
	MinGradePercent          float64 `json:"min_grade_percent"`
	MinElevationGainM        float64 `json:"min_elevation_gain_m"`
	MaxGapDistanceM          float64 `json:"max_gap_distance_m"`
	MaxBridgedElevationLossM float64 `json:"max_bridged_elevation_loss_m"`
}

// ClimbOptionalQuality describes coverage and validity for an optional stream.
type ClimbOptionalQuality struct {
	Status          string  `json:"status"`
	InputSamples    int     `json:"input_samples"`
	FiniteSamples   int     `json:"finite_samples"`
	UsedSamples     int     `json:"used_samples"`
	CoveragePercent float64 `json:"coverage_percent"`
}

// ClimbDataQuality reports deterministic stream quality diagnostics.
type ClimbDataQuality struct {
	Status                 string                          `json:"status"`
	InputSamples           int                             `json:"input_samples"`
	UsableSamples          int                             `json:"usable_samples"`
	NullAltitudeSamples    int                             `json:"null_altitude_samples"`
	InvalidAltitudeSamples int                             `json:"invalid_altitude_samples"`
	BrokenWindows          int                             `json:"broken_windows"`
	NoisyTransitions       int                             `json:"noisy_transitions"`
	OptionalStreams        map[string]ClimbOptionalQuality `json:"optional_streams"`
}

// ClimbSegment is one bounded, source-honest climb summary.
type ClimbSegment struct {
	StartDistanceM      float64  `json:"start_distance_m"`
	EndDistanceM        float64  `json:"end_distance_m"`
	DistanceM           float64  `json:"distance_m"`
	ElevationGainM      float64  `json:"elevation_gain_m"`
	AverageGradePercent float64  `json:"average_grade_percent"`
	DurationSeconds     *float64 `json:"duration_seconds,omitempty"`
	VAMMPerHour         *float64 `json:"vam_m_per_hour,omitempty"`
	AverageHeartRateBPM *float64 `json:"average_heart_rate_bpm,omitempty"`
	AveragePowerWatts   *float64 `json:"average_power_watts,omitempty"`
}

// ClimbSegmentsResult is the concise climb analyzer result.
type ClimbSegmentsResult struct {
	Segments    []ClimbSegment   `json:"segments"`
	DataQuality ClimbDataQuality `json:"data_quality"`
	Parameters  ClimbParameters  `json:"parameters"`
}

type climbPoint struct {
	distance  float64
	altitude  float64
	source    int
	time      float64
	timeValid bool
}

type climbRun struct {
	start         climbPoint
	end           climbPoint
	bridgeBlocked bool
}

// AnalyzeClimbSegments computes deterministic climb segments from aligned streams.
func AnalyzeClimbSegments(input ClimbSegmentsInput) (ClimbSegmentsResult, error) {
	params, err := normalizeClimbParameters(input)
	if err != nil {
		return ClimbSegmentsResult{}, err
	}
	result := ClimbSegmentsResult{Segments: make([]ClimbSegment, 0), Parameters: params}
	quality := ClimbDataQuality{OptionalStreams: make(map[string]ClimbOptionalQuality)}
	optional := map[string]ClimbStream{"time": input.Time, "heart_rate": input.HeartRate, "watts": input.Watts}

	distancePresent := climbStreamPresent(input.Distance)
	altitudePresent := climbStreamPresent(input.Altitude)
	distanceLen := climbStreamLength(input.Distance)
	altitudeLen := climbStreamLength(input.Altitude)
	quality.InputSamples = altitudeLen
	for key, stream := range optional {
		quality.OptionalStreams[key] = climbOptionalQuality(stream, distanceLen, false)
	}
	result.DataQuality = quality

	if input.Distance.DataState == ClimbOptionalNull || input.Distance.DataState == "data_null" || input.Distance.DataState == ClimbOptionalAllNull || input.Distance.AllNull {
		result.DataQuality.Status = ClimbStatusInvalidDistance
		return result, nil
	}
	if !distancePresent || distanceLen == 0 {
		result.DataQuality.Status = ClimbStatusMissing
		return result, nil
	}
	if !climbStreamValuesUsable(input.Distance, distanceLen) {
		result.DataQuality.Status = ClimbStatusInvalidDistance
		return result, nil
	}
	if input.Altitude.DataState == "data_null" || input.Altitude.DataState == ClimbOptionalAllNull || input.Altitude.AllNull {
		result.DataQuality.Status = ClimbStatusNull
		result.DataQuality.NullAltitudeSamples = input.Altitude.NullCount
		if input.Altitude.AllNull && result.DataQuality.NullAltitudeSamples == 0 {
			result.DataQuality.NullAltitudeSamples = altitudeLen
		}
		return result, nil
	}
	if !altitudePresent || altitudeLen == 0 {
		result.DataQuality.Status = ClimbStatusMissing
		return result, nil
	}
	if distanceLen != altitudeLen {
		result.DataQuality.Status = ClimbStatusInvalidDistance
		return result, nil
	}
	if input.Altitude.DataState == ClimbOptionalNull {
		result.DataQuality.Status = ClimbStatusNull
		result.DataQuality.NullAltitudeSamples = input.Altitude.NullCount
		return result, nil
	}
	selected := collapseClimbSamples(input.Distance, input.Altitude, optional, distanceLen)
	altitudeInvalid, nullAltitude, invalidAltitude := climbAltitudeEvidence(input.Altitude, altitudeLen)
	result.DataQuality.NullAltitudeSamples = nullAltitude
	result.DataQuality.InvalidAltitudeSamples = invalidAltitude
	if altitudeInvalid {
		result.DataQuality.Status = ClimbStatusNull
	}

	finitePoints := 0
	for _, point := range selected {
		if point.altitudeValid {
			finitePoints++
		}
	}
	result.DataQuality.UsableSamples = finitePoints
	if finitePoints < 2 || !hasPositiveAltitudeDistance(selected) {
		if result.DataQuality.Status == "" {
			result.DataQuality.Status = ClimbStatusInsufficient
		}
		return result, nil
	}
	if allFiniteAltitudesEqual(selected) {
		if result.DataQuality.Status == "" {
			result.DataQuality.Status = ClimbStatusFlat
		}
		return result, nil
	}

	windows := finiteClimbWindows(selected)
	result.DataQuality.BrokenWindows = countBrokenClimbWindows(selected)
	allRuns := make([]climbRun, 0)
	noisyTransitions := 0
	finiteTransitions := 0
	for _, window := range windows {
		points := resampleClimbWindow(window, input.Time)
		runs, noisy, transitions := climbRuns(points, params.MinGradePercent)
		allRuns = append(allRuns, bridgeClimbRuns(runs, params)...)
		noisyTransitions += noisy
		finiteTransitions += transitions
	}
	result.DataQuality.NoisyTransitions = noisyTransitions
	segments := make([]ClimbSegment, 0, len(allRuns))
	for _, run := range allRuns {
		segment := buildClimbSegment(run, selected, optional, quality.OptionalStreams)
		if segment.ElevationGainM+0 < params.MinElevationGainM {
			continue
		}
		segments = append(segments, segment)
	}
	sort.SliceStable(segments, func(i, j int) bool {
		if segments[i].StartDistanceM != segments[j].StartDistanceM {
			return segments[i].StartDistanceM < segments[j].StartDistanceM
		}
		return segments[i].EndDistanceM < segments[j].EndDistanceM
	})
	if len(segments) > ClimbSegmentLimit {
		segments = segments[:ClimbSegmentLimit]
	}
	result.Segments = segments
	if result.DataQuality.Status == "" {
		if finiteTransitions > 0 && noisyTransitions*2 >= finiteTransitions {
			result.DataQuality.Status = ClimbStatusNoisy
		} else if len(segments) == 0 {
			result.DataQuality.Status = ClimbStatusNoClimb
		} else {
			result.DataQuality.Status = ClimbStatusOK
		}
	}
	for key, stream := range optional {
		q := result.DataQuality.OptionalStreams[key]
		q.UsedSamples = usedClimbSamples(stream, selected, segments)
		result.DataQuality.OptionalStreams[key] = q
	}
	return result, nil
}

// ComputeClimbSegments is an alias kept for analyzer callers using Compute naming.
func ComputeClimbSegments(input ClimbSegmentsInput) (ClimbSegmentsResult, error) {
	return AnalyzeClimbSegments(input)
}

func normalizeClimbParameters(input ClimbSegmentsInput) (ClimbParameters, error) {
	params := ClimbParameters{MinGradePercent: input.MinGradePercent, MinElevationGainM: input.MinElevationGainM, MaxGapDistanceM: input.MaxGapDistanceM, MaxBridgedElevationLossM: input.MaxBridgedElevationLossM}
	values := []struct {
		name  string
		value float64
		min   float64
		max   float64
	}{
		{"min_grade_percent", params.MinGradePercent, 0.1, 100},
		{"min_elevation_gain_m", params.MinElevationGainM, 0, 100000},
		{"max_gap_distance_m", params.MaxGapDistanceM, 0, 10000},
		{"max_bridged_elevation_loss_m", params.MaxBridgedElevationLossM, 0, 1000},
	}
	for _, item := range values {
		if !finite(item.value) || item.value < item.min || item.value > item.max {
			return ClimbParameters{}, fmt.Errorf("%w: %s must be finite and between %g and %g", ErrInvalidClimbSegmentsInput, item.name, item.min, item.max)
		}
	}
	return params, nil
}

func climbStreamPresent(stream ClimbStream) bool {
	return stream.Present || len(stream.Values) > 0 || stream.RawLength > 0 || stream.AllNull || stream.DataState != ""
}

func climbStreamLength(stream ClimbStream) int {
	if stream.RawLength > 0 {
		return stream.RawLength
	}
	return len(stream.Values)
}

func climbStreamValidAt(stream ClimbStream, index, length int) bool {
	if index < 0 || index >= len(stream.Values) || index >= length {
		return false
	}
	if len(stream.Valid) == length {
		return stream.Valid[index] && finite(stream.Values[index])
	}
	return finite(stream.Values[index])
}

func climbStreamValuesUsable(stream ClimbStream, length int) bool {
	if len(stream.Values) != length {
		return false
	}
	var previous float64
	for i := 0; i < length; i++ {
		if !climbStreamValidAt(stream, i, length) {
			return false
		}
		if i > 0 && stream.Values[i] < previous {
			return false
		}
		previous = stream.Values[i]
	}
	return true
}

type selectedClimbPoint struct {
	distance      float64
	altitude      float64
	altitudeValid bool
	source        int
}

func collapseClimbSamples(distance, altitude ClimbStream, optional map[string]ClimbStream, length int) []selectedClimbPoint {
	out := make([]selectedClimbPoint, 0, length)
	for i := 0; i < length; {
		j := i + 1
		for j < length && distance.Values[j] == distance.Values[i] {
			j++
		}
		selected := i
		for k := i; k < j; k++ {
			if climbStreamValidAt(altitude, k, length) {
				selected = k
				break
			}
		}
		point := selectedClimbPoint{distance: distance.Values[selected], source: selected, altitudeValid: climbStreamValidAt(altitude, selected, length)}
		if point.altitudeValid {
			point.altitude = altitude.Values[selected]
		}
		out = append(out, point)
		i = j
	}
	return out
}

func climbAltitudeEvidence(altitude ClimbStream, length int) (invalid bool, nullCount, invalidCount int) {
	for i := 0; i < length; i++ {
		valid := climbStreamValidAt(altitude, i, length)
		if valid {
			continue
		}
		if len(altitude.Valid) == length && altitude.Valid[i] && !finite(altitude.Values[i]) {
			invalidCount++
			continue
		}
		if altitude.InvalidCount > 0 && altitude.NullCount == 0 {
			invalidCount++
		} else if altitude.NullCount > 0 && nullCount < altitude.NullCount {
			nullCount++
		} else if !finiteValueAt(altitude, i) {
			invalidCount++
		} else {
			nullCount++
		}
	}
	if altitude.AllNull && nullCount == 0 {
		nullCount = length
	}
	return nullCount+invalidCount > 0, nullCount, invalidCount
}

func finiteValueAt(stream ClimbStream, index int) bool {
	return index >= 0 && index < len(stream.Values) && finite(stream.Values[index])
}

func hasPositiveAltitudeDistance(points []selectedClimbPoint) bool {
	var previous selectedClimbPoint
	have := false
	for _, point := range points {
		if !point.altitudeValid {
			have = false
			continue
		}
		if have && point.distance > previous.distance {
			return true
		}
		previous, have = point, true
	}
	return false
}

func allFiniteAltitudesEqual(points []selectedClimbPoint) bool {
	var first float64
	set := false
	for _, point := range points {
		if !point.altitudeValid {
			continue
		}
		if !set {
			first, set = point.altitude, true
			continue
		}
		if point.altitude != first {
			return false
		}
	}
	return set
}

func finiteClimbWindows(points []selectedClimbPoint) [][]selectedClimbPoint {
	windows := make([][]selectedClimbPoint, 0)
	current := make([]selectedClimbPoint, 0)
	for _, point := range points {
		if !point.altitudeValid {
			if len(current) > 0 {
				windows = append(windows, current)
				current = nil
			}
			continue
		}
		current = append(current, point)
	}
	if len(current) > 0 {
		windows = append(windows, current)
	}
	return windows
}

func countBrokenClimbWindows(points []selectedClimbPoint) int {
	broken := 0
	inBroken := false
	for _, point := range points {
		if !point.altitudeValid {
			if !inBroken {
				broken++
				inBroken = true
			}
			continue
		}
		inBroken = false
	}
	return broken
}

func resampleClimbWindow(window []selectedClimbPoint, timeStream ClimbStream) []climbPoint {
	if len(window) == 0 {
		return nil
	}
	coords := make([]float64, 0, len(window)+int(math.Max(0, math.Floor(window[len(window)-1].distance)-math.Ceil(window[0].distance)))+1)
	for _, source := range window {
		coords = append(coords, source.distance)
	}
	startGrid := int(math.Ceil(window[0].distance))
	endGrid := int(math.Floor(window[len(window)-1].distance))
	for distance := startGrid; distance <= endGrid; distance++ {
		coords = append(coords, float64(distance))
	}
	sort.Float64s(coords)
	unique := coords[:0]
	for _, coordinate := range coords {
		if len(unique) == 0 || coordinate != unique[len(unique)-1] {
			unique = append(unique, coordinate)
		}
	}
	points := make([]climbPoint, 0, len(unique))
	for _, coordinate := range unique {
		left := sort.Search(len(window), func(i int) bool { return window[i].distance >= coordinate })
		if left < len(window) && window[left].distance == coordinate {
			points = append(points, climbPoint{distance: coordinate, altitude: window[left].altitude, source: window[left].source, time: streamValue(timeStream, window[left].source), timeValid: streamValid(timeStream, window[left].source)})
			continue
		}
		if left == 0 || left >= len(window) {
			continue
		}
		a, b := window[left-1], window[left]
		ratio := (coordinate - a.distance) / (b.distance - a.distance)
		point := climbPoint{distance: coordinate, altitude: a.altitude + ratio*(b.altitude-a.altitude), source: -1}
		if streamValid(timeStream, a.source) && streamValid(timeStream, b.source) {
			point.time = streamValue(timeStream, a.source) + ratio*(streamValue(timeStream, b.source)-streamValue(timeStream, a.source))
			point.timeValid = true
		}
		points = append(points, point)
	}
	return points
}

func streamValid(stream ClimbStream, index int) bool {
	length := climbStreamLength(stream)
	return climbStreamValidAt(stream, index, length)
}

func streamValue(stream ClimbStream, index int) float64 {
	if index < 0 || index >= len(stream.Values) {
		return 0
	}
	return stream.Values[index]
}

func climbRuns(points []climbPoint, minGrade float64) ([]climbRun, int, int) {
	if len(points) < 2 {
		return nil, 0, 0
	}
	runs := make([]climbRun, 0)
	var current *climbRun
	noisy, transitions := 0, 0
	for i := 1; i < len(points); i++ {
		distance := points[i].distance - points[i-1].distance
		if distance <= 0 {
			continue
		}
		transitions++
		grade := 100 * (points[i].altitude - points[i-1].altitude) / distance
		if math.Abs(grade) > 100 {
			noisy++
			if current != nil {
				current.bridgeBlocked = true
				runs = append(runs, *current)
				current = nil
			}
			continue
		}
		if grade >= minGrade {
			if current == nil {
				current = &climbRun{start: points[i-1], end: points[i]}
			} else {
				current.end = points[i]
			}
			continue
		}
		if current != nil {
			runs = append(runs, *current)
			current = nil
		}
	}
	if current != nil {
		runs = append(runs, *current)
	}
	return runs, noisy, transitions
}

func bridgeClimbRuns(runs []climbRun, params ClimbParameters) []climbRun {
	if len(runs) < 2 {
		return runs
	}
	out := make([]climbRun, 0, len(runs))
	current := runs[0]
	for _, next := range runs[1:] {
		gap := next.start.distance - current.end.distance
		loss := math.Max(0, current.end.altitude-next.start.altitude)
		mergedGain := next.end.altitude - current.start.altitude
		mergedDistance := next.end.distance - current.start.distance
		mergedGrade := math.Inf(-1)
		if mergedDistance > 0 {
			mergedGrade = 100 * mergedGain / mergedDistance
		}
		if !current.bridgeBlocked && gap <= params.MaxGapDistanceM && loss <= params.MaxBridgedElevationLossM && mergedGrade >= params.MinGradePercent && mergedGain >= params.MinElevationGainM {
			current.end = next.end
			continue
		}
		out = append(out, current)
		current = next
	}
	return append(out, current)
}

func buildClimbSegment(run climbRun, selected []selectedClimbPoint, optional map[string]ClimbStream, qualities map[string]ClimbOptionalQuality) ClimbSegment {
	gain := run.end.altitude - run.start.altitude
	distance := run.end.distance - run.start.distance
	segment := ClimbSegment{StartDistanceM: round6(run.start.distance), EndDistanceM: round6(run.end.distance), DistanceM: round6(distance), ElevationGainM: round6(gain), AverageGradePercent: round6(100 * gain / distance)}
	if q := qualities["time"]; q.Status == ClimbOptionalOK {
		if run.start.timeValid && run.end.timeValid {
			rawDuration := run.end.time - run.start.time
			if finite(rawDuration) && rawDuration > 0 {
				duration := round6(rawDuration)
				segment.DurationSeconds = &duration
				vam := round6(gain / rawDuration * 3600)
				segment.VAMMPerHour = &vam
			}
		}
	}
	if metricStreamAvailable(qualities["heart_rate"]) {
		if value, ok := optionalMean(optional["heart_rate"], selected, run.start.distance, run.end.distance); ok {
			value = round6(value)
			segment.AverageHeartRateBPM = &value
		}
	}
	if metricStreamAvailable(qualities["watts"]) {
		if value, ok := optionalMean(optional["watts"], selected, run.start.distance, run.end.distance); ok {
			value = round6(value)
			segment.AveragePowerWatts = &value
		}
	}
	return segment
}

func metricStreamAvailable(quality ClimbOptionalQuality) bool {
	switch quality.Status {
	case ClimbOptionalAbsent, ClimbOptionalNull, ClimbOptionalEmpty, ClimbOptionalAllNull, ClimbOptionalLengthMismatch:
		return false
	default:
		return true
	}
}

func optionalMean(stream ClimbStream, selected []selectedClimbPoint, start, end float64) (float64, bool) {
	length := climbStreamLength(stream)
	if length == 0 || len(stream.Values) != length {
		return 0, false
	}
	values := make([]float64, 0)
	for _, point := range selected {
		if point.distance < start || point.distance > end || !streamValid(stream, point.source) {
			continue
		}
		values = append(values, stream.Values[point.source])
	}
	if len(values) == 0 {
		return 0, false
	}
	return mean(values), true
}

func usedClimbSamples(stream ClimbStream, selected []selectedClimbPoint, segments []ClimbSegment) int {
	if climbOptionalStatus(stream, len(selected)) == ClimbOptionalAbsent || len(stream.Values) == 0 {
		return 0
	}
	seen := map[int]struct{}{}
	for _, segment := range segments {
		for _, point := range selected {
			if point.distance >= segment.StartDistanceM && point.distance <= segment.EndDistanceM && streamValid(stream, point.source) {
				seen[point.source] = struct{}{}
			}
		}
	}
	return len(seen)
}

func climbOptionalQuality(stream ClimbStream, expectedLength int, _ bool) ClimbOptionalQuality {
	length := climbStreamLength(stream)
	q := ClimbOptionalQuality{Status: climbOptionalStatus(stream, expectedLength), InputSamples: length}
	for i := 0; i < len(stream.Values); i++ {
		if streamValid(stream, i) {
			q.FiniteSamples++
		}
	}
	if q.InputSamples > 0 {
		q.CoveragePercent = round6(float64(q.FiniteSamples) * 100 / float64(q.InputSamples))
	}
	return q
}

func climbOptionalStatus(stream ClimbStream, expectedLength int) string {
	if !climbStreamPresent(stream) {
		return ClimbOptionalAbsent
	}
	if stream.DataState == "null" || stream.DataState == "data_null" {
		return ClimbOptionalNull
	}
	if stream.AllNull || stream.DataState == "all_null" {
		return ClimbOptionalAllNull
	}
	length := climbStreamLength(stream)
	if length == 0 {
		return ClimbOptionalEmpty
	}
	if expectedLength > 0 && length != expectedLength {
		return ClimbOptionalLengthMismatch
	}
	if stream.InvalidCount > 0 {
		return ClimbOptionalNonFinite
	}
	for i := 0; i < len(stream.Values); i++ {
		if !streamValid(stream, i) {
			if !finiteValueAt(stream, i) {
				return ClimbOptionalNonFinite
			}
			return ClimbOptionalPartial
		}
	}
	if expectedLength > 0 {
		var previous float64
		have := false
		for i := 0; i < len(stream.Values); i++ {
			if !streamValid(stream, i) {
				continue
			}
			if have && stream.Values[i] < previous {
				return ClimbOptionalNonMonotone
			}
			previous, have = stream.Values[i], true
		}
	}
	return ClimbOptionalOK
}

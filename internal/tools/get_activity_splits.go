package tools

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/analysis"
	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/response"
	"github.com/ricardocabral/icuvisor/internal/streams"
)

const (
	splitBaseTypes       = "distance,time"
	splitMetricTypes     = "heart_rate,watts,cadence,altitude"
	split100mUserMessage = "split_unit 100m requires a Swim activity, a SECS_100M Swim sport setting, and a distance stream explicitly marked in meters"
)

var splitMetricKeys = []string{"heart_rate", "watts", "cadence", "altitude"}

type activitySplitsBuild struct {
	Splits         []activitySplitRow
	Source         string
	IntervalSource analysis.IntervalSource
	Diagnostics    []dataAvailabilityDiagnostic
	SplitUnit      string
	Units          map[string]string
	Unavailable    *activityUnavailable
	UnavailableErr error
}

type splitBaseStreams struct {
	Distance    []float64
	Time        []float64
	DistanceRow intervals.ActivityStream
}

type splitMetricStreams struct {
	Rows map[string]intervals.ActivityStream
}

type manualSplit struct {
	Row      activitySplitRow
	Interval intervals.ActivityInterval
}

type splitSegment struct {
	StartTime float64
	EndTime   float64
	StartDist float64
	EndDist   float64
}

func buildActivitySplits(ctx context.Context, args getActivitySplitsRequest, profile intervals.AthleteWithSportSettings, streamsClient ActivityStreamsClient, intervalsClient ActivityIntervalsClient, detailsClient ActivityDetailsClient, unitSystem response.UnitSystem) (activitySplitsBuild, error) {
	requestedUnit := strings.ToLower(strings.TrimSpace(args.SplitUnit))
	explicit100m := requestedUnit == "100m"
	activity, detailsErr := detailsClient.GetActivity(ctx, args.ActivityID)
	if detailsErr != nil && isContextError(detailsErr) {
		return activitySplitsBuild{}, detailsErr
	}

	candidateSwim := detailsErr == nil && swimActivity(profile, activity)
	unsupportedSwimPace := detailsErr == nil && unsupportedSwimPaceUnits(profile, activity)
	if explicit100m && !candidateSwim {
		return activitySplitsBuild{}, NewUserError(split100mUserMessage, errors.New("activity or Swim sport settings do not prove 100 m semantics"))
	}

	intervalsDTO, intervalErr := intervalsClient.GetActivityIntervals(ctx, args.ActivityID)
	if intervalErr != nil && isContextError(intervalErr) {
		return activitySplitsBuild{}, intervalErr
	}
	classification := analysis.IntervalSourceUnknown
	if intervalErr == nil {
		classification = classifyActivityIntervalsDTO(intervalsDTO).Source
	}
	diagnostics := make([]dataAvailabilityDiagnostic, 0, 8)
	if intervalErr != nil {
		diagnostics = append(diagnostics, splitDiagnostic("interval_source_unavailable", args.ActivityID, "The interval endpoint was unavailable; no upstream interval rows were retained and valid stream boundaries will be used for virtual splits.", nil, nil, nil))
	} else if len(intervalsDTO.ICUIntervals) > 0 && hasPositiveSplitIntervals(intervalsDTO) && classification == analysis.IntervalSourceUnknown {
		diagnostics = append(diagnostics, splitDiagnostic("interval_source_unknown", args.ActivityID, "Upstream interval rows have distance and duration, but their source could not be established; they remain upstream interval rows rather than fixed-distance rows.", nil, nil, nil))
	}

	baseRows, baseErr := streamsClient.GetActivityStreams(ctx, intervals.ActivityStreamsParams{ActivityID: args.ActivityID, Types: []string{"distance", "time"}, IncludeDefaults: true})
	if baseErr != nil && isContextError(baseErr) {
		return activitySplitsBuild{}, baseErr
	}
	if baseErr != nil {
		if explicit100m {
			return activitySplitsBuild{}, NewUserError(split100mUserMessage, baseErr)
		}
		diagnostics = append(diagnostics, splitDiagnostic("base_stream_unavailable", args.ActivityID, "Distance/time streams could not be fetched; upstream interval rows remain available but virtual splits and stream enrichment are unavailable.", []string{"distance", "time"}, nil, nil))
		if requestedUnit == "" {
			if diagnostic, ok := omittedSwimSemanticsDiagnostic(args.ActivityID, detailsErr, activity, candidateSwim, unsupportedSwimPace); ok {
				diagnostics = append(diagnostics, diagnostic)
			}
		}
		manualRows := manualSplitsFromIntervals(intervalsDTO, requestedOrDefaultSplitUnit(args.SplitUnit, unitSystem), classification)
		if len(manualRows) > 0 {
			rows := manualRowsOnly(manualRows)
			return activitySplitsBuild{Splits: rows, Source: "manual_intervals", IntervalSource: classification, Diagnostics: append(diagnostics, manualSourceDiagnostics(args.ActivityID, classification, rows)...), SplitUnit: requestedOrDefaultSplitUnit(args.SplitUnit, unitSystem), Units: splitUnits(requestedOrDefaultSplitUnit(args.SplitUnit, unitSystem), unitSystem)}, nil
		}
		if explicit100m {
			return activitySplitsBuild{}, NewUserError(split100mUserMessage, baseErr)
		}
		unavailable, unavailableErr := detectActivityUnavailable(ctx, detailsClient, args.ActivityID, baseErr)
		if unavailableErr != nil {
			return activitySplitsBuild{}, unavailableErr
		}
		return activitySplitsBuild{IntervalSource: classification, Diagnostics: diagnostics, SplitUnit: requestedOrDefaultSplitUnit(args.SplitUnit, unitSystem), Units: splitUnits(requestedOrDefaultSplitUnit(args.SplitUnit, unitSystem), unitSystem), Unavailable: &unavailable}, nil
	}

	base, baseDiagnostics, baseOK := parseSplitBase(baseRows, args.ActivityID)
	diagnostics = append(diagnostics, baseDiagnostics...)
	if explicit100m && (!candidateSwim || !distanceStreamHasMeterEvidence(base.DistanceRow)) {
		return activitySplitsBuild{}, NewUserError(split100mUserMessage, errors.New("distance stream does not carry explicit meter evidence"))
	}
	splitUnit := requestedOrDefaultSplitUnit(args.SplitUnit, unitSystem)
	if requestedUnit == "" && candidateSwim && distanceStreamHasMeterEvidence(base.DistanceRow) {
		splitUnit = "100m"
	} else if requestedUnit == "" {
		if diagnostic, ok := omittedSwimSemanticsDiagnostic(args.ActivityID, detailsErr, activity, candidateSwim, unsupportedSwimPace); ok {
			diagnostics = append(diagnostics, diagnostic)
		}
	}

	manualRows := manualSplitsFromIntervals(intervalsDTO, splitUnit, classification)
	if len(manualRows) > 0 {
		rows := manualRowsOnly(manualRows)
		diagnostics = append(diagnostics, manualSourceDiagnostics(args.ActivityID, classification, rows)...)
		metricRows, metricErr := fetchSplitMetricStreams(ctx, streamsClient, args.ActivityID)
		if metricErr != nil {
			if isContextError(metricErr) {
				return activitySplitsBuild{}, metricErr
			}
			diagnostics = append(diagnostics, splitDiagnostic("metric_stream_unavailable", args.ActivityID, "Optional metric streams were unavailable; source interval rows remain without stream-derived enrichment.", splitMetricKeys, nil, nil))
		} else if len(base.Time) >= 2 {
			diagnostics = append(diagnostics, validateMetricStreams(args.ActivityID, metricRows, len(base.Time))...)
			diagnostics = append(diagnostics, enrichManualSplitRows(manualRows, base, metricRows, args.ActivityID)...)
		}
		return activitySplitsBuild{Splits: manualRowsOnly(manualRows), Source: "manual_intervals", IntervalSource: classification, Diagnostics: diagnostics, SplitUnit: splitUnit, Units: splitUnits(splitUnit, unitSystem)}, nil
	}

	if !baseOK {
		if explicit100m {
			return activitySplitsBuild{}, NewUserError(split100mUserMessage, errors.New("distance/time streams are missing or not aligned"))
		}
		return activitySplitsBuild{Splits: []activitySplitRow{}, Source: "virtual_streams", IntervalSource: classification, Diagnostics: diagnostics, SplitUnit: splitUnit, Units: splitUnits(splitUnit, unitSystem)}, nil
	}

	metricRows, metricErr := fetchSplitMetricStreams(ctx, streamsClient, args.ActivityID)
	if metricErr != nil {
		if isContextError(metricErr) {
			return activitySplitsBuild{}, metricErr
		}
		diagnostics = append(diagnostics, splitDiagnostic("metric_stream_unavailable", args.ActivityID, "Optional metric streams were unavailable; virtual split duration and pace remain source-derived.", splitMetricKeys, nil, nil))
	} else {
		diagnostics = append(diagnostics, validateMetricStreams(args.ActivityID, metricRows, len(base.Distance))...)
	}

	rows, splitDiagnostics := virtualSplitRows(base, metricRows, splitUnit, args.ActivityID)
	diagnostics = append(diagnostics, splitDiagnostics...)
	return activitySplitsBuild{Splits: rows, Source: "virtual_streams", IntervalSource: classification, Diagnostics: diagnostics, SplitUnit: splitUnit, Units: splitUnits(splitUnit, unitSystem)}, nil
}

func requestedOrDefaultSplitUnit(requested string, unitSystem response.UnitSystem) string {
	if strings.EqualFold(strings.TrimSpace(requested), "100m") {
		return "100m"
	}
	return normalizeSplitUnit(requested, unitSystem)
}

func isSwimActivity(activity intervals.Activity) bool {
	return activity.Type != nil && strings.EqualFold(strings.TrimSpace(*activity.Type), "swim")
}

func swimActivity(profile intervals.AthleteWithSportSettings, activity intervals.Activity) bool {
	if !isSwimActivity(activity) {
		return false
	}
	for _, setting := range profile.SportSettings {
		matched := false
		if len(setting.Types) > 0 {
			for _, sport := range setting.Types {
				if strings.EqualFold(strings.TrimSpace(sport), "swim") {
					matched = true
					break
				}
			}
		} else {
			matched = strings.EqualFold(strings.TrimSpace(setting.Type), "swim")
		}
		if matched && strings.EqualFold(strings.TrimSpace(setting.PaceLoadType), "swim") && strings.EqualFold(strings.TrimSpace(setting.PaceUnits), "secs_100m") {
			return true
		}
	}
	return false
}

func unsupportedSwimPaceUnits(profile intervals.AthleteWithSportSettings, activity intervals.Activity) bool {
	if !isSwimActivity(activity) {
		return false
	}
	for _, setting := range profile.SportSettings {
		matched := false
		if len(setting.Types) > 0 {
			for _, sport := range setting.Types {
				if strings.EqualFold(strings.TrimSpace(sport), "swim") {
					matched = true
					break
				}
			}
		} else {
			matched = strings.EqualFold(strings.TrimSpace(setting.Type), "swim")
		}
		if matched && !strings.EqualFold(strings.TrimSpace(setting.PaceUnits), "secs_100m") {
			return true
		}
	}
	return false
}

func omittedSwimSemanticsDiagnostic(activityID string, detailsErr error, activity intervals.Activity, candidateSwim, unsupported bool) (dataAvailabilityDiagnostic, bool) {
	if detailsErr != nil {
		return splitDiagnostic("swim_semantics_unavailable", activityID, "Activity sport details were unavailable, so normal km/mi split semantics were retained.", []string{"activity.type"}, nil, []string{"activity.type"}), true
	}
	if !isSwimActivity(activity) {
		return dataAvailabilityDiagnostic{}, false
	}
	if unsupported {
		return splitDiagnostic("unsupported_swim_pace_units", activityID, "The Swim sport setting uses unsupported or unknown pace units; normal km/mi splits are retained instead of assuming 100 m semantics.", []string{"sport_settings.pace_units"}, []string{"pace_units"}, nil), true
	}
	missing := []string{"sport_settings"}
	if candidateSwim {
		missing = []string{"distance_unit"}
	}
	return splitDiagnostic("swim_semantics_unavailable", activityID, "Swim 100 m semantics were not proven by the activity settings and distance metadata; normal km/mi splits are retained.", []string{"activity.type", "sport_settings", "distance"}, nil, missing), true
}

func hasPositiveSplitIntervals(dto intervals.IntervalsDTO) bool {
	for _, interval := range dto.ICUIntervals {
		if interval.Distance != nil && interval.Duration != nil && finitePositive(*interval.Distance) && finitePositive(*interval.Duration) {
			return true
		}
	}
	return false
}

func manualSplitsFromIntervals(dto intervals.IntervalsDTO, splitUnit string, source analysis.IntervalSource) []manualSplit {
	rows := make([]manualSplit, 0, len(dto.ICUIntervals))
	for _, interval := range dto.ICUIntervals {
		if interval.Distance == nil || interval.Duration == nil || !finitePositive(*interval.Distance) || !finitePositive(*interval.Duration) {
			continue
		}
		row := newSplitRow(len(rows)+1, *interval.Distance, *interval.Duration, splitUnit)
		if splitUnit == "100m" {
			pacePer100M := roundSplitMetric(*interval.Duration * 100 / *interval.Distance)
			row.PaceSecondsPer100M = &pacePer100M
		}
		row.DistanceBasis = "upstream_interval_distance"
		row.Provenance = splitRowProvenance(source)
		if interval.AverageHR != nil && finiteNumber(*interval.AverageHR) {
			value := roundSplitMetric(*interval.AverageHR)
			row.AverageHeartRateBPM = &value
		}
		if interval.AveragePower != nil && finiteNumber(*interval.AveragePower) {
			value := roundSplitMetric(*interval.AveragePower)
			row.AveragePowerWatts = &value
		}
		rows = append(rows, manualSplit{Row: row, Interval: interval})
	}
	return rows
}

func manualRowsOnly(rows []manualSplit) []activitySplitRow {
	out := make([]activitySplitRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.Row)
	}
	return out
}

func splitRowProvenance(source analysis.IntervalSource) string {
	switch source {
	case analysis.IntervalSourceStructuredWorkout:
		return "structured_workout_interval"
	case analysis.IntervalSourceDeviceLaps:
		return "device_lap"
	case analysis.IntervalSourceManualAdded:
		return "manual_interval"
	case analysis.IntervalSourceMixed:
		return "mixed_interval"
	default:
		return "unknown_interval"
	}
}

func manualSourceDiagnostics(activityID string, source analysis.IntervalSource, rows []activitySplitRow) []dataAvailabilityDiagnostic {
	if len(rows) == 0 {
		return nil
	}
	switch source {
	case analysis.IntervalSourceDeviceLaps:
		return []dataAvailabilityDiagnostic{splitDiagnostic("device_lap_not_fixed_distance", activityID, "Device or auto-lap rows are preserved as upstream interval distances and are not exact fixed-distance splits.", nil, nil, nil)}
	case analysis.IntervalSourceMixed:
		return []dataAvailabilityDiagnostic{splitDiagnostic("mixed_interval_source", activityID, "Interval rows contain mixed source evidence; rows remain upstream interval distances rather than fixed-distance claims.", nil, nil, nil)}
	}
	return nil
}

func parseSplitBase(rows []intervals.ActivityStream, activityID string) (splitBaseStreams, []dataAvailabilityDiagnostic, bool) {
	base := splitBaseStreams{}
	distanceRow, distanceOK := findSplitStream(rows, "distance")
	timeRow, timeOK := findSplitStream(rows, "time")
	distanceShapeOK := distanceOK && validSplitBaseChannel(distanceRow.Data, distanceRow.Raw, distanceRow.AllNull)
	timeShapeOK := timeOK && validSplitBaseChannel(timeRow.Data, timeRow.Raw, timeRow.AllNull)
	if distanceShapeOK {
		base.Distance = append([]float64(nil), distanceRow.Data...)
		base.DistanceRow = distanceRow
	}
	if timeShapeOK {
		base.Time = append([]float64(nil), timeRow.Data...)
	}

	if !distanceOK || !timeOK || !distanceShapeOK || !timeShapeOK {
		missing := make([]string, 0, 2)
		if !distanceOK || !distanceShapeOK {
			missing = append(missing, "distance")
		}
		if !timeOK || !timeShapeOK {
			missing = append(missing, "time")
		}
		return base, []dataAvailabilityDiagnostic{splitDiagnostic("base_stream_unavailable", activityID, "Distance/time streams are missing, invalid, or unavailable for fixed-distance boundaries.", []string{"distance", "time"}, nil, missing)}, false
	}

	diagnostics := splitBaseMonotonicDiagnostics(base, activityID)
	if len(diagnostics) > 0 {
		return base, diagnostics, false
	}
	if len(base.Distance) < 2 || len(base.Time) < 2 || len(base.Distance) != len(base.Time) {
		return base, []dataAvailabilityDiagnostic{splitDiagnostic("base_stream_unavailable", activityID, "Distance/time streams are not aligned or do not contain enough samples for fixed-distance boundaries.", []string{"distance", "time"}, nil, []string{"distance", "time"})}, false
	}

	for i := 1; i < len(base.Distance); i++ {
		if base.Distance[i] == base.Distance[i-1] && base.Time[i] > base.Time[i-1] {
			return base, []dataAvailabilityDiagnostic{splitDiagnostic("paused_samples_present", activityID, "Duplicate distance samples include elapsed pause time; split durations remain elapsed rather than moving time.", []string{"distance", "time"}, nil, nil)}, true
		}
	}
	return base, nil, true
}

func validSplitBaseChannel(values []float64, raw map[string]any, allNull bool) bool {
	return len(values) >= 2 && !allNull && !rawArrayHasNull(raw, "data")
}

func splitBaseMonotonicDiagnostics(base splitBaseStreams, activityID string) []dataAvailabilityDiagnostic {
	diagnostics := make([]dataAvailabilityDiagnostic, 0, 2)
	distanceInvalid := false
	for i, value := range base.Distance {
		if !finiteNumber(value) || value < 0 || (i > 0 && value < base.Distance[i-1]) {
			distanceInvalid = true
			break
		}
	}
	timeInvalid := false
	for i, value := range base.Time {
		if !finiteNumber(value) || (i > 0 && value < base.Time[i-1]) {
			timeInvalid = true
			break
		}
	}
	if distanceInvalid {
		diagnostics = append(diagnostics, splitDiagnostic("non_monotonic_distance", activityID, "Distance samples are non-finite, negative, or decreasing; no fixed-distance split is fabricated.", []string{"distance"}, nil, []string{"distance"}))
	}
	if timeInvalid {
		diagnostics = append(diagnostics, splitDiagnostic("non_monotonic_time", activityID, "Elapsed-time samples are non-finite or decreasing; no fixed-distance split is fabricated.", []string{"time"}, nil, []string{"time"}))
	}
	return diagnostics
}

func findSplitStream(rows []intervals.ActivityStream, wanted string) (intervals.ActivityStream, bool) {
	for _, row := range rows {
		key, _ := streams.CanonicalKey(firstNonEmpty(row.Type, row.Name))
		if key == wanted {
			return row, true
		}
	}
	return intervals.ActivityStream{}, false
}

func fetchSplitMetricStreams(ctx context.Context, client ActivityStreamsClient, activityID string) (splitMetricStreams, error) {
	rows, err := client.GetActivityStreams(ctx, intervals.ActivityStreamsParams{ActivityID: activityID, Types: []string{"heart_rate", "watts", "cadence", "altitude"}, IncludeDefaults: true})
	if err != nil {
		return splitMetricStreams{}, err
	}
	out := splitMetricStreams{Rows: make(map[string]intervals.ActivityStream, len(rows))}
	for _, row := range rows {
		key, _ := streams.CanonicalKey(firstNonEmpty(row.Type, row.Name))
		if _, wanted := splitMetricSet()[key]; wanted {
			if _, exists := out.Rows[key]; !exists {
				out.Rows[key] = row
			}
		}
	}
	return out, nil
}

func splitMetricSet() map[string]bool {
	return map[string]bool{"heart_rate": true, "watts": true, "cadence": true, "altitude": true}
}

func validateMetricStreams(activityID string, metrics splitMetricStreams, baseLength int) []dataAvailabilityDiagnostic {
	available := make([]string, 0, len(splitMetricKeys))
	for _, key := range splitMetricKeys {
		if _, ok := metrics.Rows[key]; ok {
			available = append(available, key)
		}
	}
	diagnostics := make([]dataAvailabilityDiagnostic, 0, len(splitMetricKeys))
	for _, key := range splitMetricKeys {
		row, ok := metrics.Rows[key]
		if !ok {
			diagnostics = append(diagnostics, splitDiagnostic("missing_metric_stream", activityID, "Optional split metric stream is absent; the metric is omitted rather than zero-filled.", []string{key}, available, []string{key}))
			continue
		}
		if row.AllNull || rawArrayHasNull(row.Raw, "data") {
			diagnostics = append(diagnostics, splitDiagnostic("metric_channel_invalid", activityID, "Optional split metric stream contains null/all-null samples and is omitted rather than interpolated.", []string{key}, available, []string{key}))
			continue
		}
		if len(row.Data) != baseLength || len(row.Data) < 2 {
			diagnostics = append(diagnostics, splitDiagnostic("metric_channel_length_mismatch", activityID, "Optional split metric stream is not index-aligned with distance/time and is omitted.", []string{key}, available, []string{key}))
			continue
		}
		invalid := false
		for _, value := range row.Data {
			if !finiteNumber(value) {
				invalid = true
				break
			}
		}
		if invalid {
			diagnostics = append(diagnostics, splitDiagnostic("metric_channel_invalid", activityID, "Optional split metric stream contains non-finite samples and is omitted rather than interpolated.", []string{key}, available, []string{key}))
		}
	}
	return diagnostics
}

func virtualSplitRows(base splitBaseStreams, metrics splitMetricStreams, splitUnit, activityID string) ([]activitySplitRow, []dataAvailabilityDiagnostic) {
	step := splitDistanceMeters(splitUnit)
	if len(base.Distance) < 2 || base.Distance[0] > 0.001 || base.Distance[len(base.Distance)-1] < step {
		return []activitySplitRow{}, []dataAvailabilityDiagnostic{splitDiagnostic("insufficient_split_coverage", activityID, "Distance/time streams do not cover a complete zero-origin fixed-distance split; no partial row was fabricated.", []string{"distance", "time"}, nil, nil)}
	}
	rows := make([]activitySplitRow, 0)
	diagnostics := []dataAvailabilityDiagnostic{}
	for index := 0; ; index++ {
		start := float64(index) * step
		end := start + step
		if end > base.Distance[len(base.Distance)-1]+0.001 {
			break
		}
		startTime, okStart := interpolateMonotonic(base.Distance, base.Time, start)
		endTime, okEnd := interpolateMonotonic(base.Distance, base.Time, end)
		if !okStart || !okEnd || endTime <= startTime {
			continue
		}
		row := newSplitRow(len(rows)+1, step, endTime-startTime, splitUnit)
		row.Provenance = "virtual_fixed_distance"
		row.DistanceBasis = "fixed_distance_boundary"
		row.PaceSecondsPer100M = nil
		if splitUnit == "100m" {
			pace := roundSplitMetric(endTime - startTime)
			row.PaceSecondsPer100M = &pace
		}
		segment := splitSegment{StartTime: startTime, EndTime: endTime, StartDist: start, EndDist: end}
		diagnostics = append(diagnostics, enrichVirtualSplitRow(&row, segment, base, metrics, activityID)...)
		rows = append(rows, row)
	}
	if len(rows) == 0 {
		diagnostics = append(diagnostics, splitDiagnostic("insufficient_split_coverage", activityID, "No positive-duration fixed-distance segment could be established from aligned distance/time samples.", []string{"distance", "time"}, nil, nil))
	}
	return rows, diagnostics
}

func enrichVirtualSplitRow(row *activitySplitRow, segment splitSegment, base splitBaseStreams, metrics splitMetricStreams, activityID string) []dataAvailabilityDiagnostic {
	diagnostics := []dataAvailabilityDiagnostic{}
	for _, key := range splitMetricKeys {
		stream, ok := metrics.Rows[key]
		if !ok || !metricStreamUsable(stream, len(base.Time)) {
			continue
		}
		value, valid := aggregateMetric(stream.Data, base.Time, segment.StartTime, segment.EndTime, key == "altitude")
		if !valid {
			field := splitMetricOutputField(key)
			diagnostics = append(diagnostics, splitDiagnostic("metric_insufficient_coverage", activityID, "Optional metric does not fully cover this split boundary and was omitted.", []string{key}, nil, []string{field}))
			continue
		}
		setSplitMetric(row, key, value)
	}
	return diagnostics
}

func enrichManualSplitRows(rows []manualSplit, base splitBaseStreams, metrics splitMetricStreams, activityID string) []dataAvailabilityDiagnostic {
	diagnostics := []dataAvailabilityDiagnostic{}
	for index := range rows {
		segment, ok := manualSegment(rows[index].Interval, base)
		if !ok {
			diagnostic := splitDiagnostic("manual_boundary_unavailable", activityID, "Manual interval has no usable distance, elapsed-time, or aligned sample-index boundary for stream enrichment.", nil, nil, []string{"splits[" + strconv.Itoa(index) + "]"})
			diagnostic.IntervalID = anyString(rows[index].Interval.ID)
			diagnostics = append(diagnostics, diagnostic)
			continue
		}
		for _, key := range []string{"cadence", "altitude"} {
			stream, exists := metrics.Rows[key]
			if !exists || !metricStreamUsable(stream, len(base.Time)) {
				continue
			}
			value, valid := aggregateMetric(stream.Data, base.Time, segment.StartTime, segment.EndTime, key == "altitude")
			if !valid {
				diagnostics = append(diagnostics, splitDiagnostic("metric_insufficient_coverage", activityID, "Optional metric does not fully cover this manual interval boundary and was omitted.", []string{key}, nil, []string{"splits[" + strconv.Itoa(index) + "]." + splitMetricOutputField(key)}))
				continue
			}
			setSplitMetric(&rows[index].Row, key, value)
		}
	}
	return diagnostics
}

func manualSegment(interval intervals.ActivityInterval, base splitBaseStreams) (splitSegment, bool) {
	if interval.StartDistance != nil && interval.EndDistance != nil && finiteNumber(*interval.StartDistance) && finiteNumber(*interval.EndDistance) && *interval.EndDistance > *interval.StartDistance && validSplitDistanceTimeline(base) {
		startTime, okStart := interpolateMonotonic(base.Distance, base.Time, *interval.StartDistance)
		endTime, okEnd := interpolateMonotonic(base.Distance, base.Time, *interval.EndDistance)
		if okStart && okEnd && endTime > startTime {
			return splitSegment{StartTime: startTime, EndTime: endTime, StartDist: *interval.StartDistance, EndDist: *interval.EndDistance}, true
		}
	}
	if interval.StartTime != nil && interval.EndTime != nil && validSplitTimeTimeline(base.Time) {
		start, errStart := strconv.ParseFloat(strings.TrimSpace(*interval.StartTime), 64)
		end, errEnd := strconv.ParseFloat(strings.TrimSpace(*interval.EndTime), 64)
		if errStart == nil && errEnd == nil && finiteNumber(start) && finiteNumber(end) && end > start {
			return splitSegment{StartTime: start, EndTime: end}, true
		}
	}
	if interval.StartIndex != nil && interval.EndIndex != nil && *interval.StartIndex >= 0 && *interval.EndIndex >= *interval.StartIndex && *interval.EndIndex < len(base.Time) && validSplitTimeTimeline(base.Time) {
		startIndex, endIndex := *interval.StartIndex, *interval.EndIndex
		if base.Time[endIndex] > base.Time[startIndex] {
			segment := splitSegment{StartTime: base.Time[startIndex], EndTime: base.Time[endIndex]}
			if len(base.Distance) == len(base.Time) {
				segment.StartDist = base.Distance[startIndex]
				segment.EndDist = base.Distance[endIndex]
			}
			return segment, true
		}
	}
	return splitSegment{}, false
}

func validSplitTimeTimeline(values []float64) bool {
	if len(values) < 2 {
		return false
	}
	for i, value := range values {
		if !finiteNumber(value) || (i > 0 && value < values[i-1]) {
			return false
		}
	}
	return true
}

func validSplitDistanceTimeline(base splitBaseStreams) bool {
	if len(base.Distance) < 2 || len(base.Distance) != len(base.Time) || !validSplitTimeTimeline(base.Time) {
		return false
	}
	for i, value := range base.Distance {
		if !finiteNumber(value) || value < 0 || (i > 0 && value < base.Distance[i-1]) {
			return false
		}
	}
	return true
}

func metricStreamUsable(stream intervals.ActivityStream, length int) bool {
	if len(stream.Data) != length || stream.AllNull || rawArrayHasNull(stream.Raw, "data") {
		return false
	}
	for _, value := range stream.Data {
		if !finiteNumber(value) {
			return false
		}
	}
	return true
}

func aggregateMetric(values, times []float64, start, end float64, elevation bool) (float64, bool) {
	if len(values) != len(times) || len(values) < 2 || end <= start || start < times[0]-0.001 || end > times[len(times)-1]+0.001 {
		return 0, false
	}
	pointsX := []float64{start}
	pointsY := []float64{interpolateMonotonicOrNaN(times, values, start)}
	if !finiteNumber(pointsY[0]) {
		return 0, false
	}
	for i := 0; i < len(times); i++ {
		if times[i] > start && times[i] < end && (i+1 == len(times) || times[i+1] != times[i]) {
			pointsX = append(pointsX, times[i])
			pointsY = append(pointsY, values[i])
		}
	}
	endValue := interpolateMonotonicOrNaN(times, values, end)
	if !finiteNumber(endValue) {
		return 0, false
	}
	pointsX = append(pointsX, end)
	pointsY = append(pointsY, endValue)
	if elevation {
		gain := 0.0
		for i := 1; i < len(pointsY); i++ {
			if delta := pointsY[i] - pointsY[i-1]; delta > 0 {
				gain += delta
			}
		}
		return gain, true
	}
	area := 0.0
	for i := 1; i < len(pointsX); i++ {
		area += (pointsX[i] - pointsX[i-1]) * (pointsY[i] + pointsY[i-1]) / 2
	}
	return area / (end - start), true
}

func interpolateMonotonic(x, y []float64, target float64) (float64, bool) {
	value := interpolateMonotonicOrNaN(x, y, target)
	return value, finiteNumber(value)
}

func interpolateMonotonicOrNaN(x, y []float64, target float64) float64 {
	if len(x) == 0 || len(x) != len(y) || target < x[0]-0.001 || target > x[len(x)-1]+0.001 {
		return math.NaN()
	}
	if target <= x[0] {
		last := 0
		for last+1 < len(x) && x[last+1] == x[0] {
			last++
		}
		return y[last]
	}
	for i := 1; i < len(x); i++ {
		if x[i] >= target {
			if x[i] == target {
				last := i
				for last+1 < len(x) && x[last+1] == target {
					last++
				}
				return y[last]
			}
			span := x[i] - x[i-1]
			if span <= 0 {
				continue
			}
			ratio := (target - x[i-1]) / span
			return y[i-1] + ratio*(y[i]-y[i-1])
		}
	}
	return y[len(y)-1]
}

func splitMetricOutputField(key string) string {
	switch key {
	case "heart_rate":
		return "average_heart_rate_bpm"
	case "watts":
		return "average_power_watts"
	case "cadence":
		return "average_cadence_rpm"
	case "altitude":
		return "elevation_gain_m"
	default:
		return key
	}
}

func setSplitMetric(row *activitySplitRow, key string, value float64) {
	value = roundSplitMetric(value)
	switch key {
	case "heart_rate":
		row.AverageHeartRateBPM = &value
	case "watts":
		row.AveragePowerWatts = &value
	case "cadence":
		row.AverageCadenceRPM = &value
	case "altitude":
		row.ElevationGainM = &value
	}
}

func splitUnits(splitUnit string, unitSystem response.UnitSystem) map[string]string {
	if splitUnit == "100m" {
		return map[string]string{
			"system":     "metric",
			"distance":   "100m",
			"pace":       "sec/100m",
			"elevation":  "m",
			"heart_rate": "bpm",
			"power":      "W",
			"cadence":    "rpm",
		}
	}
	units := unitSystem.Metadata()
	units["elevation"] = "m"
	units["heart_rate"] = "bpm"
	units["power"] = "W"
	units["cadence"] = "rpm"
	return units
}

func distanceStreamHasMeterEvidence(row intervals.ActivityStream) bool {
	if row.Raw == nil {
		return false
	}
	found := false
	for rawKey, value := range row.Raw {
		key := strings.ToLower(strings.TrimSpace(rawKey))
		if key != "unit" && key != "units" && key != "distance_unit" {
			continue
		}
		text := strings.ToLower(strings.TrimSpace(fmt.Sprint(value)))
		switch text {
		case "m", "meter", "meters", "metre", "metres":
			found = true
		default:
			return false
		}
	}
	return found
}

func splitAlgorithm(source string) string {
	if source == "manual_intervals" {
		return "upstream interval distance/duration with source classification; optional aligned stream enrichment"
	}
	return "fixed-distance boundaries from cumulative distance/time with elapsed-time interpolation; optional trapezoidal stream metrics and positive altitude deltas"
}

func splitDiagnostic(reason, activityID, message string, requested, available, missing []string) dataAvailabilityDiagnostic {
	return dataAvailabilityDiagnostic{Reason: reason, ActivityID: activityID, Message: message, Requested: requested, Available: available, MissingFields: missing}
}

func finiteNumber(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func finitePositive(value float64) bool {
	return finiteNumber(value) && value > 0
}

func roundSplitMetric(value float64) float64 {
	return math.Round(value*10) / 10
}

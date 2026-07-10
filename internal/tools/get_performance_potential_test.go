package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/athleteprofile"
	"github.com/ricardocabral/icuvisor/internal/intervals"
)

func TestGetPerformancePotentialCyclingAndRunningContracts(t *testing.T) {
	t.Parallel()

	client := newFakeFitnessMetricsClient(t)
	client.profile = intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC", SportSettings: []intervals.SportSettings{
		{Types: []string{"Ride"}, FTP: 260, IndoorFTP: 250, WPrime: 21000, PMax: 1050, LTHR: 172, MaxHR: 190},
		{Types: []string{"Run"}, FTP: 999, LTHR: 181, MaxHR: 195, ThresholdPace: 3.5714285, PaceUnits: "MINS_KM"},
	}}
	tool := newGetPerformancePotentialTool(client, client, "test", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"start_date":"2026-05-01","end_date":"2026-05-07","sports":["Ride","Run"],"power_duration_seconds":[60],"hr_duration_seconds":[60],"pace_distance_meters":[1000]}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	payload := resultMap(t, result)
	sports := payload["sports"].([]any)
	if len(sports) != 2 {
		t.Fatalf("sports = %#v, want two", sports)
	}
	ride := sports[0].(map[string]any)
	rideThresholds := ride["thresholds"].(map[string]any)
	if ride["sport_family"] != "cycling" || rideThresholds["ftp_watts"] != float64(260) || rideThresholds["indoor_ftp_watts"] != float64(250) {
		t.Fatalf("ride thresholds = %#v", rideThresholds)
	}
	if _, ok := rideThresholds["threshold_pace_seconds_per_km"]; ok {
		t.Fatalf("ride copied pace threshold into cycling row: %#v", rideThresholds)
	}
	if cp := rideThresholds["critical_power"].(map[string]any); cp["status"] != "unsupported" {
		t.Fatalf("critical_power = %#v, want unsupported", cp)
	}
	rideCurves := ride["curve_anchors"].(map[string]any)
	power := rideCurves["power"].(map[string]any)
	if power["status"] != "available" || power["unit"] != "W" || power["points"].([]any)[0].(map[string]any)["watts"] != float64(420) {
		t.Fatalf("ride power anchors = %#v", power)
	}
	if pace := rideCurves["pace"].(map[string]any); pace["status"] != "unsupported" {
		t.Fatalf("ride pace anchors = %#v, want unsupported", pace)
	}

	run := sports[1].(map[string]any)
	runThresholds := run["thresholds"].(map[string]any)
	if run["sport_family"] != "running" || runThresholds["threshold_pace_seconds_per_km"] == nil || runThresholds["lthr_bpm"] != float64(181) {
		t.Fatalf("run thresholds = %#v", runThresholds)
	}
	if pace := runThresholds["threshold_pace_seconds_per_km"].(float64); pace < 279.9999 || pace > 280.0001 {
		t.Fatalf("run thresholds = %#v", runThresholds)
	}
	if _, ok := runThresholds["ftp_watts"]; ok {
		t.Fatalf("run copied FTP into pace row: %#v", runThresholds)
	}
	runCurves := run["curve_anchors"].(map[string]any)
	if power := runCurves["power"].(map[string]any); power["status"] != "unsupported" {
		t.Fatalf("run power anchors = %#v, want unsupported", power)
	}
	pace := runCurves["pace"].(map[string]any)
	pacePoint := pace["points"].([]any)[0].(map[string]any)
	if pace["preferred_unit"] != "seconds_per_km" || pacePoint["pace_seconds_per_km"] != float64(230) {
		t.Fatalf("run pace anchors = %#v", pace)
	}
}

func TestGetPerformancePotentialUnavailableNoZeroFillAndFullGating(t *testing.T) {
	t.Parallel()

	client := newFakeFitnessMetricsClient(t)
	client.profile = intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", SportSettings: []intervals.SportSettings{{Types: []string{"Run"}}}}
	client.curves["Run:pace"] = distanceCurveSet(t, []float64{1000}, []float64{0})
	client.curves["Run:hr"] = curveSet(t, []float64{60}, []float64{0})
	tool := newGetPerformancePotentialTool(client, client, "test", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"start_date":"2026-05-01","end_date":"2026-05-07","sports":["Run"],"hr_duration_seconds":[60],"pace_distance_meters":[1000]}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	text := resultText(t, result)
	if strings.Contains(text, `"threshold_pace_seconds_per_km":0`) || strings.Contains(text, `"lthr_bpm":0`) || strings.Contains(text, `"heart_rate_bpm":0`) {
		t.Fatalf("response zero-filled unavailable fields: %s", text)
	}
	payload := resultMap(t, result)
	row := payload["sports"].([]any)[0].(map[string]any)
	if _, ok := row["full"]; ok {
		t.Fatalf("full payload present in terse response: %#v", row["full"])
	}
	unavailable := row["unavailable"].([]any)
	if !unavailableHasField(unavailable, "threshold_pace") || !unavailableHasField(unavailable, "lthr_bpm") || !unavailableHasField(unavailable, "curve_anchors.pace") || !unavailableHasField(unavailable, "curve_anchors.heart_rate") {
		t.Fatalf("unavailable = %#v, want threshold and curve caveats", unavailable)
	}
	curves := row["curve_anchors"].(map[string]any)
	if curves["pace"].(map[string]any)["status"] != "unavailable" || curves["heart_rate"].(map[string]any)["status"] != "unavailable" {
		t.Fatalf("curve statuses = %#v", curves)
	}

	fullResult, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"start_date":"2026-05-01","end_date":"2026-05-07","sports":["Run"],"include_full":true}`)})
	if err != nil {
		t.Fatalf("full Handler() error = %v", err)
	}
	fullRow := resultMap(t, fullResult)["sports"].([]any)[0].(map[string]any)
	if _, ok := fullRow["full"].(map[string]any); !ok {
		t.Fatalf("full payload missing when requested: %#v", fullRow)
	}
}

func TestGetPerformancePotentialSwimUsesSportSpecificPaceUnit(t *testing.T) {
	t.Parallel()

	client := newFakeFitnessMetricsClient(t)
	client.profile = intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "imperial", SportSettings: []intervals.SportSettings{{Types: []string{"Swim"}, LTHR: 150, ThresholdPace: 2, PaceUnits: "SECS_100M"}}}
	client.curves["Swim:pace"] = distanceCurveSet(t, []float64{100}, []float64{90})
	client.curves["Swim:hr"] = curveSet(t, []float64{60}, []float64{150})
	tool := newGetPerformancePotentialTool(client, client, "test", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"start_date":"2026-05-01","end_date":"2026-05-07","sports":["Swim"],"pace_distance_meters":[100],"hr_duration_seconds":[60]}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	row := resultMap(t, result)["sports"].([]any)[0].(map[string]any)
	thresholds := row["thresholds"].(map[string]any)
	if thresholds["pace_distance_unit"] != "100m" || thresholds["threshold_pace_seconds_per_100m"] != float64(50) {
		t.Fatalf("swim thresholds = %#v", thresholds)
	}
	pace := row["curve_anchors"].(map[string]any)["pace"].(map[string]any)
	point := pace["points"].([]any)[0].(map[string]any)
	if pace["preferred_unit"] != "seconds_per_100m" || point["pace_seconds_per_100m"] != float64(90) {
		t.Fatalf("swim pace anchors = %#v", pace)
	}
}

func TestAssignPerformancePotentialPaceThresholdsKeepsEveryRecognizedDisplay(t *testing.T) {
	km := 280.0
	mile := 450.0
	meters := 50.0
	yards := 45.72
	fiveHundred := 125.0
	fourHundred := 100.0
	twoHundredFifty := 62.5
	metersPerSecond := 3.5714285

	profileSport := &athleteprofile.Sport{
		ThresholdPaceSecondsPerKM:    &km,
		ThresholdPaceSecondsPerMile:  &mile,
		ThresholdPaceSecondsPer100M:  &meters,
		ThresholdPaceSecondsPer100Y:  &yards,
		ThresholdPaceSecondsPer500M:  &fiveHundred,
		ThresholdPaceSecondsPer400M:  &fourHundred,
		ThresholdPaceSecondsPer250M:  &twoHundredFifty,
		ThresholdPaceMetersPerSecond: &metersPerSecond,
	}
	var thresholds performancePotentialThresholds
	var thresholdContext performancePotentialThresholdContext
	assignPerformancePotentialPaceThresholds(&thresholds, &thresholdContext, profileSport)

	if thresholds.ThresholdPaceSecondsPer100Y != &yards || thresholds.ThresholdPaceSecondsPer400M != &fourHundred || thresholds.ThresholdPaceSecondsPer250M != &twoHundredFifty || thresholds.ThresholdPaceMetersPerSecond != &metersPerSecond {
		t.Fatalf("pace thresholds = %+v, want every recognized pace display and m/s fallback", thresholds)
	}
	if !hasPerformancePotentialPaceThreshold(thresholds) {
		t.Fatal("hasPerformancePotentialPaceThreshold() = false, want true")
	}
	wantFields := map[string]bool{
		"threshold_pace_seconds_per_km":    false,
		"threshold_pace_seconds_per_mile":  false,
		"threshold_pace_seconds_per_100m":  false,
		"threshold_pace_seconds_per_100y":  false,
		"threshold_pace_seconds_per_500m":  false,
		"threshold_pace_seconds_per_400m":  false,
		"threshold_pace_seconds_per_250m":  false,
		"threshold_pace_meters_per_second": false,
	}
	for _, threshold := range thresholdContext.AnaerobicThreshold {
		if _, ok := wantFields[threshold.Field]; ok {
			wantFields[threshold.Field] = true
		}
	}
	for field, found := range wantFields {
		if !found {
			t.Fatalf("anaerobic threshold fields = %#v, missing %q", thresholdContext.AnaerobicThreshold, field)
		}
	}
}

func TestGetPerformancePotentialSourceFailuresAndContextCancellation(t *testing.T) {
	t.Parallel()

	client := newFakeFitnessMetricsClient(t)
	client.profile = intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", SportSettings: []intervals.SportSettings{{Types: []string{"Ride"}, FTP: 250, LTHR: 170}}}
	client.curveErrs["Ride:power"] = errors.New("upstream unavailable")
	tool := newGetPerformancePotentialTool(client, client, "test", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"start_date":"2026-05-01","end_date":"2026-05-07","sports":["Ride"],"power_duration_seconds":[60],"hr_duration_seconds":[60]}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	row := resultMap(t, result)["sports"].([]any)[0].(map[string]any)
	sources := row["sources"].(map[string]any)
	if sources["power_curves"].(map[string]any)["status"] != "unavailable" {
		t.Fatalf("sources = %#v, want non-cancellation source failure as unavailable", sources)
	}

	client.curveErrs["Ride:power"] = context.Canceled
	_, err = tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"start_date":"2026-05-01","end_date":"2026-05-07","sports":["Ride"],"power_duration_seconds":[60]}`)})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Handler() error = %v, want context.Canceled", err)
	}
}

func TestGetPerformancePotentialMatchesRawSportTypeFallback(t *testing.T) {
	t.Parallel()

	client := newFakeFitnessMetricsClient(t)
	client.profile = intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", SportSettings: []intervals.SportSettings{{Type: "Ride", FTP: 255, LTHR: 171}}}
	tool := newGetPerformancePotentialTool(client, client, "test", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"start_date":"2026-05-01","end_date":"2026-05-07","sports":["Ride"],"power_duration_seconds":[60],"hr_duration_seconds":[60]}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	thresholds := resultMap(t, result)["sports"].([]any)[0].(map[string]any)["thresholds"].(map[string]any)
	if thresholds["ftp_watts"] != float64(255) || thresholds["lthr_bpm"] != float64(171) {
		t.Fatalf("thresholds = %#v, want raw type fallback match", thresholds)
	}
}

func unavailableHasField(values []any, field string) bool {
	for _, value := range values {
		row, ok := value.(map[string]any)
		if ok && row["field"] == field {
			return true
		}
	}
	return false
}

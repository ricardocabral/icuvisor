package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

type fakeFitnessMetricsClient struct {
	fakeProfileClient
	summaries    []intervals.SummaryWithCats
	summaryCalls []intervals.AthleteSummaryParams
	powerCalls   []intervals.CurveParams
	curves       map[string]intervals.DataCurveSet
}

func (f *fakeFitnessMetricsClient) ListAthleteSummary(_ context.Context, params intervals.AthleteSummaryParams) ([]intervals.SummaryWithCats, error) {
	f.summaryCalls = append(f.summaryCalls, params)
	return append([]intervals.SummaryWithCats(nil), f.summaries...), nil
}

func (f *fakeFitnessMetricsClient) ListAthletePowerCurves(_ context.Context, params intervals.CurveParams) (intervals.DataCurveSet, error) {
	f.powerCalls = append(f.powerCalls, params)
	return f.curves[params.Sport+":power"], nil
}

func (f *fakeFitnessMetricsClient) ListAthleteHRCurves(_ context.Context, params intervals.CurveParams) (intervals.DataCurveSet, error) {
	return f.curves[params.Sport+":hr"], nil
}

func (f *fakeFitnessMetricsClient) ListAthletePaceCurves(_ context.Context, params intervals.CurveParams) (intervals.DataCurveSet, error) {
	return f.curves[params.Sport+":pace"], nil
}

func newFakeFitnessMetricsClient(t *testing.T) *fakeFitnessMetricsClient {
	t.Helper()

	profile := intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "America/Sao_Paulo"}
	client := &fakeFitnessMetricsClient{fakeProfileClient: fakeProfileClient{profile: profile}}
	client.summaries = decodeSummaries(t, `[
		{"date":"2026-05-02","fitness":72,"fatigue":80,"form":-8,"time":3600,"moving_time":3500,"distance":40000,"training_load":90,"timeInZones":[10,20],"timeInZonesTot":30,"byCategory":[{"category":"Ride","time":3600,"distance":40000,"training_load":90}]},
		{"date":"2026-05-01","fitness":70,"fatigue":78,"form":-8,"time":1800,"moving_time":1700,"distance":5000,"training_load":40,"timeInZones":[5,15],"timeInZonesTot":20,"byCategory":[{"category":"Run","time":1800,"distance":5000,"training_load":40}]}
	]`)
	client.curves = map[string]intervals.DataCurveSet{
		"Ride:power": curveSet(t, []float64{60, 300}, []float64{420, 310}),
		"Ride:hr":    curveSet(t, []float64{60, 300}, []float64{178, 165}),
		"Ride:pace":  distanceCurveSet(t, []float64{1000}, []float64{240}),
		"Run:power":  curveSet(t, []float64{60, 300}, []float64{390, 300}),
		"Run:hr":     curveSet(t, []float64{60, 300}, []float64{182, 170}),
		"Run:pace":   distanceCurveSet(t, []float64{1000}, []float64{230}),
	}

	return client
}

func decodeSummaries(t *testing.T, text string) []intervals.SummaryWithCats {
	t.Helper()
	var out []intervals.SummaryWithCats
	if err := json.Unmarshal([]byte(text), &out); err != nil {
		t.Fatalf("decode summaries: %v", err)
	}
	return out
}

func curveSet(t *testing.T, secs []float64, values []float64) intervals.DataCurveSet {
	t.Helper()
	data, _ := json.Marshal(map[string]any{"list": []map[string]any{{"id": "r", "secs": secs, "values": values, "activity_id": []string{"a1", "a2", "a3"}}}, "activities": map[string]any{}})
	return decodeCurveSet(t, data)
}

func decodeCurveSet(t *testing.T, data []byte) intervals.DataCurveSet {
	t.Helper()
	var out intervals.DataCurveSet
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("decode curve set: %v", err)
	}
	return out
}

func distanceCurveSet(t *testing.T, distances []float64, values []float64) intervals.DataCurveSet {
	t.Helper()
	data, _ := json.Marshal(map[string]any{"list": []map[string]any{{"id": "r", "distance": distances, "values": values, "activity_id": []string{"a1", "a2", "a3"}}}, "activities": map[string]any{}})
	var out intervals.DataCurveSet
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("decode distance curve set: %v", err)
	}
	return out
}

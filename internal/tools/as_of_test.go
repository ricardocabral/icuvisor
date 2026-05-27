package tools

import (
	"testing"
	"time"
)

func TestCurrentDayAsOfMetadataRangePredicate(t *testing.T) {
	t.Parallel()

	clock := func() time.Time { return time.Date(2026, 5, 25, 2, 30, 0, 0, time.UTC) }
	tests := []struct {
		name        string
		oldest      string
		newest      string
		wantAsOf    bool
		wantDate    string
		wantWeekday string
	}{
		{name: "closed range includes local today", oldest: "2026-05-24", newest: "2026-05-24", wantAsOf: true, wantDate: "2026-05-24", wantWeekday: "Sunday"},
		{name: "open ended activity range includes local today", oldest: "2026-05-01T06:00:00", newest: "", wantAsOf: true, wantDate: "2026-05-24", wantWeekday: "Sunday"},
		{name: "past range excludes local today", oldest: "2026-05-01", newest: "2026-05-23", wantAsOf: false},
		{name: "future range excludes local today", oldest: "2026-05-25", newest: "2026-05-30", wantAsOf: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := currentDayAsOfMetadata(clock, "America/Sao_Paulo", tt.oldest, tt.newest)
			if err != nil {
				t.Fatalf("currentDayAsOfMetadata() error = %v", err)
			}
			if (got != nil) != tt.wantAsOf {
				t.Fatalf("currentDayAsOfMetadata() present = %v, want %v", got != nil, tt.wantAsOf)
			}
			if got != nil && (got.AsOfDate != tt.wantDate || got.AsOfWeekday != tt.wantWeekday) {
				t.Fatalf("as-of metadata = %#v, want date %s weekday %s", got, tt.wantDate, tt.wantWeekday)
			}
		})
	}
}

func assertSaoPauloAsOfMeta(t *testing.T, meta map[string]any) {
	t.Helper()
	if meta["as_of"] != "2026-05-24T23:30:00-03:00" || meta["as_of_date"] != "2026-05-24" || meta["as_of_weekday"] != "Sunday" || meta["timezone"] != "America/Sao_Paulo" {
		t.Fatalf("as-of meta = %#v, want Sao Paulo 2026-05-24 anchor", meta)
	}
}

func assertNoAsOfMeta(t *testing.T, meta map[string]any) {
	t.Helper()
	for _, key := range []string{"as_of", "as_of_date", "as_of_weekday"} {
		if _, ok := meta[key]; ok {
			t.Fatalf("%s present in meta for non-current-day range: %#v", key, meta)
		}
	}
}

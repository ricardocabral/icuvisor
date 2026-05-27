package response

import (
	"strings"
	"testing"
	"time"
)

func TestAsOfMetadataInTimezone(t *testing.T) {
	tests := []struct {
		name     string
		instant  time.Time
		timezone string
		want     AsOfMetadata
		wantErr  string
	}{
		{
			name:     "positive offset date differs from UTC",
			instant:  time.Date(2026, 5, 11, 15, 30, 0, 0, time.UTC),
			timezone: "Pacific/Kiritimati",
			want:     AsOfMetadata{AsOf: "2026-05-12T05:30:00+14:00", AsOfDate: "2026-05-12", AsOfWeekday: "Tuesday", Timezone: "Pacific/Kiritimati"},
		},
		{
			name:     "negative offset date differs from UTC",
			instant:  time.Date(2026, 5, 12, 2, 30, 0, 0, time.UTC),
			timezone: "America/Sao_Paulo",
			want:     AsOfMetadata{AsOf: "2026-05-11T23:30:00-03:00", AsOfDate: "2026-05-11", AsOfWeekday: "Monday", Timezone: "America/Sao_Paulo"},
		},
		{
			name:     "trimmed timezone",
			instant:  time.Date(2026, 5, 11, 15, 0, 0, 0, time.UTC),
			timezone: " Europe/Lisbon ",
			want:     AsOfMetadata{AsOf: "2026-05-11T16:00:00+01:00", AsOfDate: "2026-05-11", AsOfWeekday: "Monday", Timezone: "Europe/Lisbon"},
		},
		{
			name:     "empty defaults UTC",
			instant:  time.Date(2026, 5, 11, 15, 0, 0, 0, time.UTC),
			timezone: "",
			want:     AsOfMetadata{AsOf: "2026-05-11T15:00:00Z", AsOfDate: "2026-05-11", AsOfWeekday: "Monday", Timezone: "UTC"},
		},
		{
			name:     "invalid timezone errors",
			instant:  time.Date(2026, 5, 11, 15, 0, 0, 0, time.UTC),
			timezone: "Not/AZone",
			wantErr:  "loading athlete timezone",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AsOfMetadataInTimezone(tt.instant, tt.timezone)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("AsOfMetadataInTimezone() error = %v, want containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("AsOfMetadataInTimezone() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("AsOfMetadataInTimezone() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestRenderDateInTimezone(t *testing.T) {
	instant := time.Date(2026, 5, 12, 2, 30, 0, 0, time.UTC)
	got, err := RenderDateInTimezone(instant, "America/Sao_Paulo")
	if err != nil {
		t.Fatalf("RenderDateInTimezone() error = %v", err)
	}
	if got != "2026-05-11" {
		t.Fatalf("RenderDateInTimezone() = %q, want 2026-05-11", got)
	}
}

func TestRenderTimeInTimezone(t *testing.T) {
	instant := time.Date(2026, 5, 11, 15, 0, 0, 0, time.UTC)
	tests := []struct {
		name     string
		timezone string
		want     string
		wantErr  string
	}{
		{name: "configured zone", timezone: "America/Sao_Paulo", want: "2026-05-11T12:00:00-03:00"},
		{name: "empty defaults UTC", timezone: "", want: "2026-05-11T15:00:00Z"},
		{name: "invalid", timezone: "Not/AZone", wantErr: "loading athlete timezone"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RenderTimeInTimezone(instant, tt.timezone)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("RenderTimeInTimezone() error = %v, want containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("RenderTimeInTimezone() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("RenderTimeInTimezone() = %q, want %q", got, tt.want)
			}
		})
	}
}

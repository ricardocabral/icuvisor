package response

import (
	"strings"
	"testing"
	"time"
)

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

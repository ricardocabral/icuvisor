package analysis

import (
	"fmt"
	"time"
)

const DateLayout = time.DateOnly

// Window is an inclusive athlete-local date range.
type Window struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// ParsedWindow contains parsed inclusive date bounds.
type ParsedWindow struct {
	Window
	Start time.Time
	End   time.Time
	Days  int
}

// ParseWindow validates an inclusive athlete-local YYYY-MM-DD window.
func ParseWindow(window Window, maxDays int) (ParsedWindow, error) {
	start, err := time.Parse(DateLayout, window.StartDate)
	if err != nil {
		return ParsedWindow{}, fmt.Errorf("start_date must be YYYY-MM-DD: %w", err)
	}
	end, err := time.Parse(DateLayout, window.EndDate)
	if err != nil {
		return ParsedWindow{}, fmt.Errorf("end_date must be YYYY-MM-DD: %w", err)
	}
	if end.Before(start) {
		return ParsedWindow{}, fmt.Errorf("end_date must be on or after start_date")
	}
	days := int(end.Sub(start).Hours()/24) + 1
	if maxDays > 0 && days > maxDays {
		return ParsedWindow{}, fmt.Errorf("window must be at most %d days", maxDays)
	}
	return ParsedWindow{Window: Window{StartDate: start.Format(DateLayout), EndDate: end.Format(DateLayout)}, Start: start, End: end, Days: days}, nil
}

// DefaultBaselineWindow returns the same-length inclusive window immediately before current.
func DefaultBaselineWindow(current ParsedWindow) Window {
	end := current.Start.AddDate(0, 0, -1)
	start := end.AddDate(0, 0, -(current.Days - 1))
	return Window{StartDate: start.Format(DateLayout), EndDate: end.Format(DateLayout)}
}

// ShiftWindow shifts a window by whole days.
func ShiftWindow(window ParsedWindow, days int) Window {
	return Window{StartDate: window.Start.AddDate(0, 0, days).Format(DateLayout), EndDate: window.End.AddDate(0, 0, days).Format(DateLayout)}
}

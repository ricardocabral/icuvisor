package tools

import (
	"strings"
	"time"

	"github.com/ricardocabral/icuvisor/internal/response"
)

func currentDayAsOfMetadata(now func() time.Time, timezoneName string, oldest string, newest string) (*response.AsOfMetadata, error) {
	if now == nil {
		now = time.Now
	}
	asOf, err := response.AsOfMetadataInTimezone(now(), timezoneName)
	if err != nil {
		return nil, err
	}
	if !dateRangeIncludesLocalDate(oldest, newest, asOf.AsOfDate) {
		return nil, nil
	}
	return &asOf, nil
}

func dateRangeIncludesLocalDate(oldest string, newest string, localDate string) bool {
	oldestDate, ok := asOfLocalDatePrefix(oldest)
	if !ok || localDate < oldestDate {
		return false
	}
	newestDate, ok := asOfLocalDatePrefix(newest)
	if !ok {
		return strings.TrimSpace(newest) == ""
	}
	return localDate <= newestDate
}

func asOfLocalDatePrefix(value string) (string, bool) {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) < len(time.DateOnly) {
		return "", false
	}
	date := trimmed[:len(time.DateOnly)]
	return date, validDate(date)
}

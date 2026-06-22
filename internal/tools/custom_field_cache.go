package tools

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/intervals"
)

// activityCustomFieldItemType is the intervals.icu custom-item type for
// athlete-defined activity custom fields.
const (
	activityCustomFieldItemType     = "ACTIVITY_FIELD"
	maxSelectedActivityCustomFields = 20
	maxCustomFieldHintLength        = 64
)

// ActivityCustomFieldClient lists custom-item definitions so activity reads can
// discover athlete-defined activity custom field codes.
type ActivityCustomFieldClient interface {
	ListCustomItems(context.Context) ([]intervals.CustomItem, error)
}

// customFieldCache memoizes activity custom field codes per target athlete so
// activity reads do not re-fetch custom-item definitions on every call.
type customFieldCache struct {
	mu      sync.RWMutex
	entries map[string][]string
}

func newCustomFieldCache() *customFieldCache {
	return &customFieldCache{entries: map[string][]string{}}
}

// activityFieldCodes returns the athlete's activity custom field codes, fetching
// and caching them on first use. A nil client yields no codes; a fetch failure
// degrades to no codes so activity reads still succeed without custom fields.
func (c *customFieldCache) activityFieldCodes(ctx context.Context, client ActivityCustomFieldClient) []string {
	codes, err := c.lookupActivityFieldCodes(ctx, client)
	if err != nil {
		return nil
	}
	return codes
}

func (c *customFieldCache) lookupActivityFieldCodes(ctx context.Context, client ActivityCustomFieldClient) ([]string, error) {
	if client == nil || ctx.Err() != nil {
		return nil, ctx.Err()
	}
	key, err := customFieldCacheKey(ctx)
	if err != nil {
		return nil, err
	}
	if c != nil {
		c.mu.RLock()
		cached, ok := c.entries[key]
		c.mu.RUnlock()
		if ok {
			return append([]string(nil), cached...), nil
		}
	}
	items, err := client.ListCustomItems(ctx)
	if err != nil {
		return nil, err
	}
	codes := activityCustomFieldCodes(items)
	if c != nil {
		c.mu.Lock()
		c.entries[key] = append([]string(nil), codes...)
		c.mu.Unlock()
	}
	return codes, nil
}

func selectedActivityCustomFieldCodes(ctx context.Context, client ActivityCustomFieldClient, cache *customFieldCache, requested []string) ([]string, error) {
	requested = compactCustomFieldCodes(requested)
	if len(requested) == 0 {
		return nil, nil
	}
	if len(requested) > maxSelectedActivityCustomFields {
		return nil, tooManyActivityCustomFieldsError{count: len(requested)}
	}
	seenRequested := make(map[string]bool, len(requested))
	selected := make([]string, 0, len(requested))
	for _, code := range requested {
		if seenRequested[code] {
			continue
		}
		seenRequested[code] = true
		selected = append(selected, code)
	}
	if client == nil {
		return selected, nil
	}
	available, err := cache.lookupActivityFieldCodes(ctx, client)
	if err != nil {
		return nil, err
	}
	availableSet := make(map[string]bool, len(available))
	for _, code := range available {
		availableSet[code] = true
	}
	unknown := make([]string, 0)
	for _, code := range selected {
		if !availableSet[code] {
			unknown = append(unknown, code)
		}
	}
	if len(unknown) > 0 {
		return nil, unknownActivityCustomFieldsError{unknown: unknown, available: available}
	}
	return selected, nil
}

type tooManyActivityCustomFieldsError struct {
	count int
}

func (e tooManyActivityCustomFieldsError) Error() string {
	return fmt.Sprintf("custom_fields accepts at most %d entries; got %d", maxSelectedActivityCustomFields, e.count)
}

type unknownActivityCustomFieldsError struct {
	unknown   []string
	available []string
}

func (e unknownActivityCustomFieldsError) Error() string {
	unknown := limitedFieldList(e.unknown, 3)
	available := limitedFieldList(e.available, 8)
	if available == "" {
		available = "none"
	}
	if len(e.unknown) == 1 {
		return fmt.Sprintf("unknown activity custom field %q; available: %s", fieldHint(e.unknown[0]), available)
	}
	return fmt.Sprintf("unknown activity custom fields: %s; available: %s", unknown, available)
}

func activityCustomFieldSelectionMessage(err error, fallback string) string {
	var tooMany tooManyActivityCustomFieldsError
	if errors.As(err, &tooMany) {
		return tooMany.Error()
	}
	var unknown unknownActivityCustomFieldsError
	if errors.As(err, &unknown) {
		return unknown.Error()
	}
	return fallback
}

func limitedFieldList(values []string, limit int) string {
	if len(values) == 0 || limit <= 0 {
		return ""
	}
	trimmed := values
	suffix := ""
	if len(values) > limit {
		trimmed = values[:limit]
		suffix = fmt.Sprintf(", +%d more", len(values)-limit)
	}
	hints := make([]string, 0, len(trimmed))
	for _, value := range trimmed {
		hints = append(hints, fieldHint(value))
	}
	return strings.Join(hints, ", ") + suffix
}

func fieldHint(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	if len(value) <= maxCustomFieldHintLength {
		return value
	}
	return value[:maxCustomFieldHintLength] + "..."
}

func compactCustomFieldCodes(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func customFieldCacheKey(ctx context.Context) (string, error) {
	athleteID, ok := intervals.TargetAthleteIDFromContext(ctx)
	if !ok {
		return defaultGearCacheKey, nil
	}
	normalized, err := config.NormalizeAthleteID(athleteID)
	if err != nil {
		return "", fmt.Errorf("normalizing target athlete ID for custom field cache: %w", err)
	}
	return normalized, nil
}

// activityCustomFieldCodes extracts the field codes declared by ACTIVITY_FIELD
// custom-item definitions. Each code is the top-level key the field occupies in
// an activity payload.
func activityCustomFieldCodes(items []intervals.CustomItem) []string {
	seen := map[string]bool{}
	codes := make([]string, 0, len(items))
	for _, item := range items {
		if !strings.EqualFold(strings.TrimSpace(stringValue(item.Type)), activityCustomFieldItemType) {
			continue
		}
		content, ok := item.Content.(map[string]any)
		if !ok {
			continue
		}
		code := anyString(content["field"])
		if code == "" || seen[code] {
			continue
		}
		seen[code] = true
		codes = append(codes, code)
	}
	if len(codes) == 0 {
		return nil
	}
	sort.Strings(codes)
	return codes
}

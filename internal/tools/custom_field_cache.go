package tools

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/intervals"
)

// activityCustomFieldItemType is the intervals.icu custom-item type for
// athlete-defined activity custom fields.
const activityCustomFieldItemType = "ACTIVITY_FIELD"

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
	if client == nil || ctx.Err() != nil {
		return nil
	}
	key, err := customFieldCacheKey(ctx)
	if err != nil {
		return nil
	}
	if c != nil {
		c.mu.RLock()
		cached, ok := c.entries[key]
		c.mu.RUnlock()
		if ok {
			return append([]string(nil), cached...)
		}
	}
	items, err := client.ListCustomItems(ctx)
	if err != nil {
		return nil
	}
	codes := activityCustomFieldCodes(items)
	if c != nil {
		c.mu.Lock()
		c.entries[key] = append([]string(nil), codes...)
		c.mu.Unlock()
	}
	return codes
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

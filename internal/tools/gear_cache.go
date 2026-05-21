package tools

import (
	"context"
	"fmt"
	"sync"

	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/intervals"
)

const defaultGearCacheKey = "default"

type gearListCache struct {
	mu      sync.RWMutex
	entries map[string][]intervals.Gear
}

type gearListCacheResult struct {
	gear      []intervals.Gear
	cached    bool
	refreshed bool
}

func newGearListCache() *gearListCache {
	return &gearListCache{entries: map[string][]intervals.Gear{}}
}

func (c *gearListCache) get(ctx context.Context, key string, refresh bool, fetch func(context.Context) ([]intervals.Gear, error)) (gearListCacheResult, error) {
	if err := ctx.Err(); err != nil {
		return gearListCacheResult{}, err
	}
	if c == nil {
		gear, err := fetch(ctx)
		if err != nil {
			return gearListCacheResult{}, err
		}
		return gearListCacheResult{gear: cloneGearList(gear), refreshed: true}, nil
	}
	if !refresh {
		c.mu.RLock()
		cached, ok := c.entries[key]
		c.mu.RUnlock()
		if ok {
			return gearListCacheResult{gear: cloneGearList(cached), cached: true}, nil
		}
	}
	gear, err := fetch(ctx)
	if err != nil {
		return gearListCacheResult{}, err
	}
	if err := ctx.Err(); err != nil {
		return gearListCacheResult{}, err
	}
	cloned := cloneGearList(gear)
	c.mu.Lock()
	c.entries[key] = cloned
	c.mu.Unlock()
	return gearListCacheResult{gear: cloneGearList(cloned), refreshed: true}, nil
}

func gearCacheKey(ctx context.Context) (string, error) {
	athleteID, ok := intervals.TargetAthleteIDFromContext(ctx)
	if !ok {
		return defaultGearCacheKey, nil
	}
	normalized, err := config.NormalizeAthleteID(athleteID)
	if err != nil {
		return "", fmt.Errorf("normalizing target athlete ID for gear cache: %w", err)
	}
	return normalized, nil
}

func cloneGearList(gear []intervals.Gear) []intervals.Gear {
	return append([]intervals.Gear(nil), gear...)
}

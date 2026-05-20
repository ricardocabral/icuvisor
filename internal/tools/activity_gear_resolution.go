package tools

import (
	"context"
	"errors"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

const (
	gearResolutionResolved          = "resolved"
	gearResolutionNameMissing       = "name_missing"
	gearResolutionUnresolved        = "unresolved"
	gearResolutionLookupUnavailable = "lookup_unavailable"
)

type activityGearResolution struct {
	GearID string
	Name   string
	Status string
}

func resolveActivityGear(ctx context.Context, client GearListClient, cache *gearListCache, activities []intervals.Activity) (map[string]activityGearResolution, error) {
	gearIDs := make(map[string]struct{})
	for _, activity := range activities {
		gearID := strings.TrimSpace(activity.GearID)
		if gearID != "" {
			gearIDs[gearID] = struct{}{}
		}
	}
	if len(gearIDs) == 0 {
		return nil, nil
	}
	if client == nil {
		return unavailableGearResolutions(activities), nil
	}
	cacheKey, err := gearCacheKey(ctx)
	if err != nil {
		return unavailableGearResolutions(activities), nil
	}
	result, err := cache.get(ctx, cacheKey, false, client.ListGear)
	if err != nil {
		if isContextError(err) || errors.Is(ctx.Err(), context.Canceled) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, firstNonNilError(ctx.Err(), err)
		}
		return unavailableGearResolutions(activities), nil
	}
	byID := make(map[string]intervals.Gear, len(result.gear))
	for _, gear := range result.gear {
		if gear.ID != "" {
			byID[gear.ID] = gear
		}
	}
	out := make(map[string]activityGearResolution)
	for _, activity := range activities {
		gearID := strings.TrimSpace(activity.GearID)
		if gearID == "" {
			continue
		}
		resolution := activityGearResolution{GearID: gearID, Status: gearResolutionUnresolved}
		if gear, ok := byID[gearID]; ok {
			name := strings.TrimSpace(stringValue(gear.Name))
			if name == "" {
				resolution.Status = gearResolutionNameMissing
			} else {
				resolution.Status = gearResolutionResolved
				resolution.Name = name
			}
		}
		out[activity.ID] = resolution
	}
	return out, nil
}

func unavailableGearResolutions(activities []intervals.Activity) map[string]activityGearResolution {
	out := make(map[string]activityGearResolution)
	for _, activity := range activities {
		gearID := strings.TrimSpace(activity.GearID)
		if gearID == "" {
			continue
		}
		out[activity.ID] = activityGearResolution{GearID: gearID, Status: gearResolutionLookupUnavailable}
	}
	return out
}

func firstNonNilError(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

package tools

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

const (
	activityDetailWithGearFixture = "../intervals/testdata/activity_detail_with_gear.json"
	activityListWithGearFixture   = "../intervals/testdata/activity_list_with_gear.json"
	gearListFixture               = "../intervals/testdata/gear_list.json"
)

func TestActivityGearFixtureHasIDOnlyGear(t *testing.T) {
	t.Parallel()

	activity := loadSingleActivityFixtureFile(t, activityDetailWithGearFixture)
	if activity.GearID != "123" {
		t.Fatalf("activity.GearID = %q, want numeric fixture gear ID 123", activity.GearID)
	}
	for _, key := range []string{"gear", "gear_name", "gearName"} {
		if _, ok := activity.Raw[key]; ok {
			t.Fatalf("activity fixture raw = %#v, want no embedded gear name via %q", activity.Raw, key)
		}
	}
}

func TestResolveActivityGearUsesFullGearListForNumericID(t *testing.T) {
	t.Parallel()

	activity := loadSingleActivityFixtureFile(t, activityDetailWithGearFixture)
	client := &fixtureGearListClient{gear: loadGearFixtureFile(t, gearListFixture)}
	cache := newGearListCache()
	resolutions, err := resolveActivityGear(context.Background(), client, cache, []intervals.Activity{activity})
	if err != nil {
		t.Fatalf("resolveActivityGear() error = %v", err)
	}
	resolution := resolutions["a-bike"]
	if resolution.GearID != activity.GearID || resolution.GearID != "123" || resolution.Name != "Synthetic Active Bike" || resolution.Status != gearResolutionResolved {
		t.Fatalf("resolution = %#v, want retained numeric gear ID resolved from full gear list", resolution)
	}
	cached, err := resolveActivityGear(context.Background(), client, cache, []intervals.Activity{activity})
	if err != nil {
		t.Fatalf("resolveActivityGear() cached error = %v", err)
	}
	if cached[activity.ID].Name != "Synthetic Active Bike" || cached[activity.ID].GearID != activity.GearID || cached[activity.ID].Status != gearResolutionResolved {
		t.Fatalf("cached resolution = %#v, want retained ID and resolved name", cached[activity.ID])
	}
	if client.calls != 1 {
		t.Fatalf("gear list calls = %d, want one fetch followed by cache reuse", client.calls)
	}
}

func loadSingleActivityFixtureFile(t *testing.T, path string) intervals.Activity {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read activity fixture %s: %v", path, err)
	}
	var activity intervals.Activity
	if err := json.Unmarshal(data, &activity); err != nil {
		t.Fatalf("decode activity fixture %s: %v", path, err)
	}
	return activity
}

func loadGearFixtureFile(t *testing.T, path string) []intervals.Gear {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read gear fixture %s: %v", path, err)
	}
	var gear []intervals.Gear
	if err := json.Unmarshal(data, &gear); err != nil {
		t.Fatalf("decode gear fixture %s: %v", path, err)
	}
	return gear
}

type fixtureGearListClient struct {
	gear  []intervals.Gear
	calls int
}

func (c *fixtureGearListClient) ListGear(context.Context) ([]intervals.Gear, error) {
	c.calls++
	return c.gear, nil
}

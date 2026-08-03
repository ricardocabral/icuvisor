package intervals

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestGearFixtureContract(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("testdata/gear_list.json")
	if err != nil {
		t.Fatalf("read gear fixture: %v", err)
	}
	var original []map[string]any
	if err := json.Unmarshal(data, &original); err != nil {
		t.Fatalf("decode original fixture: %v", err)
	}
	var gear []Gear
	if err := json.Unmarshal(data, &gear); err != nil {
		t.Fatalf("decode Gear fixture: %v", err)
	}
	if len(gear) != len(original) {
		t.Fatalf("decoded gear len = %d, want %d", len(gear), len(original))
	}

	tests := []struct {
		name        string
		index       int
		id          string
		gearName    string
		gearType    string
		retired     *string
		retiredNull bool
	}{
		{name: "active", index: 0, id: "123", gearName: "Synthetic Active Bike", gearType: "Bike", retiredNull: true},
		{name: "retired", index: 1, id: "shoe-7", gearName: "Synthetic Retired Shoes", gearType: "Shoes", retired: stringPointer("2025-11-30")},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			item := gear[tc.index]
			if item.ID != tc.id || item.Name == nil || *item.Name != tc.gearName || item.Type == nil || *item.Type != tc.gearType {
				t.Fatalf("Gear = %#v, want id/name/type %q/%q/%q", item, tc.id, tc.gearName, tc.gearType)
			}
			if tc.retired == nil {
				if item.Retired != nil {
					t.Fatalf("Gear.Retired = %q, want nil", *item.Retired)
				}
			} else if item.Retired == nil || *item.Retired != *tc.retired {
				t.Fatalf("Gear.Retired = %v, want %q", item.Retired, *tc.retired)
			}
			if tc.retiredNull {
				value, ok := item.Raw["retired"]
				if !ok || value != nil {
					t.Fatalf("Gear.Raw[retired] = %#v, want present nil", value)
				}
			}

			roundTrip, err := json.Marshal(item.Raw)
			if err != nil {
				t.Fatalf("marshal raw gear: %v", err)
			}
			var got map[string]any
			if err := json.Unmarshal(roundTrip, &got); err != nil {
				t.Fatalf("unmarshal raw gear round trip: %v", err)
			}
			if !reflect.DeepEqual(got, original[tc.index]) {
				t.Fatalf("raw round trip = %#v, want %#v", got, original[tc.index])
			}
		})
	}
}

func stringPointer(value string) *string {
	return &value
}

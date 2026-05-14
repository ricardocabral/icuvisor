package customitemschemas

import (
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

func TestValidateContentAgainstReadSchemaRejectsUnknownKeysAndWrongKinds(t *testing.T) {
	t.Parallel()

	itemType := "FITNESS_CHART"
	items := []intervals.CustomItem{{Type: &itemType, Content: map[string]any{"series": []any{map[string]any{"field": "ctl"}}, "layout": map[string]any{"height": float64(240)}}}}
	if _, err := ValidateContentAgainstReadSchema(items, itemType, map[string]any{"series": []any{map[string]any{"field": "atl"}}, "layout": map[string]any{"height": float64(260)}}, true); err != nil {
		t.Fatalf("ValidateContentAgainstReadSchema() valid content error = %v", err)
	}
	_, err := ValidateContentAgainstReadSchema(items, itemType, map[string]any{"series": []any{map[string]any{"field": "atl"}}, "layout": map[string]any{"height": "tall"}}, true)
	if err == nil || !strings.Contains(err.Error(), "content.layout.height must be number") {
		t.Fatalf("wrong kind error = %v, want layout height kind error", err)
	}
	_, err = ValidateContentAgainstReadSchema(items, itemType, map[string]any{"series": []any{map[string]any{"field": "atl"}}, "layout": map[string]any{"height": float64(260)}, "extra": true}, true)
	if err == nil || !strings.Contains(err.Error(), "content.extra is not in the readable schema") {
		t.Fatalf("unknown key error = %v, want readable schema error", err)
	}
}

func TestFamiliesHaveSamplesAndInferredPaths(t *testing.T) {
	t.Parallel()

	for _, family := range Families() {
		if family.Key == "" || family.Title == "" || family.Description == "" || len(family.Items) == 0 {
			t.Fatalf("family is incomplete: %#v", family)
		}
		for _, item := range family.Items {
			if item.ItemType == "" || item.Description == "" {
				t.Fatalf("item descriptor is incomplete: %#v", item)
			}
			if item.Sample == nil && item.SharesSchemaWith == "" {
				t.Fatalf("item descriptor %s has no sample or alias", item.ItemType)
			}
			if item.Sample == nil {
				continue
			}
			schema, err := InferContentSchema([]map[string]any{item.Sample})
			if err != nil {
				t.Fatalf("InferContentSchema(%s) error = %v", item.ItemType, err)
			}
			if len(SchemaPaths(schema)) < 2 {
				t.Fatalf("SchemaPaths(%s) too short: %#v", item.ItemType, SchemaPaths(schema))
			}
		}
	}
}

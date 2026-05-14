package resources

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/customitemschemas"
)

func TestCustomItemSchemasMarkdownGolden(t *testing.T) {
	t.Parallel()

	got, err := CustomItemSchemasMarkdown()
	if err != nil {
		t.Fatalf("CustomItemSchemasMarkdown() error = %v", err)
	}
	want, err := os.ReadFile("testdata/custom_item_schemas.md")
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	if got != string(want) {
		t.Fatalf("CustomItemSchemasMarkdown() mismatch with testdata/custom_item_schemas.md")
	}
}

func TestCustomItemSchemasMarkdownCoversDescriptorsAndInferredPaths(t *testing.T) {
	t.Parallel()

	markdown, err := CustomItemSchemasMarkdown()
	if err != nil {
		t.Fatalf("CustomItemSchemasMarkdown() error = %v", err)
	}
	families := customitemschemas.Families()
	for _, family := range families {
		for _, want := range []string{"`" + family.Key + "`", family.Title, family.Description} {
			if !strings.Contains(markdown, want) {
				t.Fatalf("markdown missing %q for family %s", want, family.Key)
			}
		}
		for _, item := range family.Items {
			if !strings.Contains(markdown, "### `"+item.ItemType+"`") {
				t.Fatalf("markdown missing item_type subsection %s for family %s", item.ItemType, family.Key)
			}
			sample := item.Sample
			if item.SharesSchemaWith != "" {
				if !strings.Contains(markdown, "Shares schema with `"+item.SharesSchemaWith+"`") {
					t.Fatalf("markdown missing alias %s -> %s", item.ItemType, item.SharesSchemaWith)
				}
				var found bool
				sample, found = testSampleForItem(families, item.SharesSchemaWith)
				if !found {
					t.Fatalf("alias %s points to unknown %s", item.ItemType, item.SharesSchemaWith)
				}
			}
			if sample == nil {
				t.Fatalf("item_type %s has no sample or alias", item.ItemType)
			}
			schema, err := customitemschemas.InferContentSchema([]map[string]any{sample})
			if err != nil {
				t.Fatalf("InferContentSchema(%s) error = %v", item.ItemType, err)
			}
			for _, path := range customitemschemas.SchemaPaths(schema) {
				if !strings.Contains(markdown, "`"+path.Path+"`: "+path.Kind) {
					t.Fatalf("markdown missing path %s:%s for item %s", path.Path, path.Kind, item.ItemType)
				}
			}
		}
	}
}

func testSampleForItem(families []customitemschemas.FamilyDescriptor, itemType string) (map[string]any, bool) {
	for _, family := range families {
		for _, item := range family.Items {
			if item.ItemType == itemType && item.Sample != nil {
				return item.Sample, true
			}
		}
	}
	return nil, false
}

func TestNewRegistryRegistersCustomItemSchemasResource(t *testing.T) {
	t.Parallel()

	registrar := &captureRegistrar{}
	if err := NewRegistry().Register(context.Background(), registrar); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	var resource Resource
	for _, candidate := range registrar.resources {
		if candidate.URI == CustomItemSchemasURI {
			resource = candidate
			break
		}
	}
	if resource.URI == "" {
		t.Fatalf("registered resources = %#v, missing %s", registrar.resources, CustomItemSchemasURI)
	}
	if resource.Name != "custom_item_schemas" || resource.Title != "Custom item schemas" || resource.MIMEType != CustomItemSchemasMIMEType {
		t.Fatalf("resource metadata = %#v, want custom item schemas metadata", resource)
	}
	result, err := resource.Handler(context.Background(), Request{URI: CustomItemSchemasURI})
	if err != nil {
		t.Fatalf("resource handler error = %v", err)
	}
	if result.URI != CustomItemSchemasURI || result.MIMEType != CustomItemSchemasMIMEType || !strings.Contains(result.Text, "# Custom item content schemas") {
		t.Fatalf("resource handler result = %#v, want URI/MIME/markdown", result)
	}
}

func TestCustomItemSchemasResourceHonorsCanceledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := CustomItemSchemasResource().Handler(ctx, Request{URI: CustomItemSchemasURI})
	if err == nil {
		t.Fatal("handler error = nil, want context cancellation")
	}
}

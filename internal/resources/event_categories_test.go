package resources

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

func TestEventCategoriesMarkdownGolden(t *testing.T) {
	t.Parallel()

	got, err := EventCategoriesMarkdown()
	if err != nil {
		t.Fatalf("EventCategoriesMarkdown() error = %v", err)
	}
	want, err := os.ReadFile("testdata/event_categories.md")
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	if got != string(want) {
		t.Fatalf("EventCategoriesMarkdown() mismatch with testdata/event_categories.md")
	}
}

func TestEventCategoriesDescriptorRenderedExactlyOnce(t *testing.T) {
	t.Parallel()

	markdown, err := EventCategoriesMarkdown()
	if err != nil {
		t.Fatalf("EventCategoriesMarkdown() error = %v", err)
	}
	seen := make(map[string]struct{})
	for _, category := range intervals.EventCategories() {
		if strings.TrimSpace(category.Value) == "" || strings.TrimSpace(category.Description) == "" {
			t.Fatalf("category descriptor has empty field: %#v", category)
		}
		if _, exists := seen[category.Value]; exists {
			t.Fatalf("duplicate category in descriptor: %s", category.Value)
		}
		seen[category.Value] = struct{}{}
		marker := "| `" + category.Value + "` |"
		if strings.Count(markdown, marker) != 1 {
			t.Fatalf("markdown contains %q %d times, want exactly once", marker, strings.Count(markdown, marker))
		}
		if !strings.Contains(markdown, category.Description) {
			t.Fatalf("markdown missing description for %s", category.Value)
		}
	}
}

func TestNewRegistryRegistersEventCategoriesResource(t *testing.T) {
	t.Parallel()

	registrar := &captureRegistrar{}
	if err := NewRegistry().Register(context.Background(), registrar); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	var resource Resource
	for _, candidate := range registrar.resources {
		if candidate.URI == EventCategoriesURI {
			resource = candidate
			break
		}
	}
	if resource.URI == "" {
		t.Fatalf("registered resources = %#v, missing %s", registrar.resources, EventCategoriesURI)
	}
	if resource.Name != "event_categories" || resource.Title != "Event categories" || resource.MIMEType != EventCategoriesMIMEType {
		t.Fatalf("resource metadata = %#v, want event categories metadata", resource)
	}
	result, err := resource.Handler(context.Background(), Request{URI: EventCategoriesURI})
	if err != nil {
		t.Fatalf("resource handler error = %v", err)
	}
	if result.URI != EventCategoriesURI || result.MIMEType != EventCategoriesMIMEType || !strings.Contains(result.Text, "# Event categories") {
		t.Fatalf("resource handler result = %#v, want URI/MIME/markdown", result)
	}
}

func TestEventCategoriesResourceHonorsCanceledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := EventCategoriesResource().Handler(ctx, Request{URI: EventCategoriesURI})
	if err == nil {
		t.Fatal("handler error = nil, want context cancellation")
	}
}

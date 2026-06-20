package main

import (
	"strings"
	"testing"
)

func TestDiffSpecsDetectsAddedPaths(t *testing.T) {
	diff, err := diffSpecs(fixtureSpec([]string{"/api/v1/athlete/{id}"}, []string{"Athlete"}), fixtureSpec([]string{"/api/v1/athlete/{id}", "/api/v1/athlete/{id}/activities"}, []string{"Athlete"}))
	if err != nil {
		t.Fatalf("diffSpecs returned error: %v", err)
	}
	if got, want := strings.Join(diff.Added, ","), "/api/v1/athlete/{id}/activities"; got != want {
		t.Fatalf("added paths = %q, want %q", got, want)
	}
	if len(diff.Removed) != 0 {
		t.Fatalf("removed paths = %v, want none", diff.Removed)
	}
	if len(diff.SchemasAdded) != 0 || len(diff.SchemasRemoved) != 0 {
		t.Fatalf("schema drift = added %v removed %v, want none", diff.SchemasAdded, diff.SchemasRemoved)
	}
}

func TestDiffSpecsDetectsRemovedPaths(t *testing.T) {
	diff, err := diffSpecs(fixtureSpec([]string{"/api/v1/athlete/{id}", "/api/v1/athlete/{id}/events"}, []string{"Athlete"}), fixtureSpec([]string{"/api/v1/athlete/{id}"}, []string{"Athlete"}))
	if err != nil {
		t.Fatalf("diffSpecs returned error: %v", err)
	}
	if got, want := strings.Join(diff.Removed, ","), "/api/v1/athlete/{id}/events"; got != want {
		t.Fatalf("removed paths = %q, want %q", got, want)
	}
	if len(diff.Added) != 0 {
		t.Fatalf("added paths = %v, want none", diff.Added)
	}
	if len(diff.SchemasAdded) != 0 || len(diff.SchemasRemoved) != 0 {
		t.Fatalf("schema drift = added %v removed %v, want none", diff.SchemasAdded, diff.SchemasRemoved)
	}
}

func TestDiffSpecsDetectsSchemaOnlyDrift(t *testing.T) {
	diff, err := diffSpecs(fixtureSpec([]string{"/api/v1/athlete/{id}"}, []string{"Athlete", "OldModel"}), fixtureSpec([]string{"/api/v1/athlete/{id}"}, []string{"Athlete", "AthleteWithTags"}))
	if err != nil {
		t.Fatalf("diffSpecs returned error: %v", err)
	}
	if len(diff.Added) != 0 || len(diff.Removed) != 0 {
		t.Fatalf("path drift = added %v removed %v, want none", diff.Added, diff.Removed)
	}
	if got, want := strings.Join(diff.SchemasAdded, ","), "AthleteWithTags"; got != want {
		t.Fatalf("added schemas = %q, want %q", got, want)
	}
	if got, want := strings.Join(diff.SchemasRemoved, ","), "OldModel"; got != want {
		t.Fatalf("removed schemas = %q, want %q", got, want)
	}
}

func TestDiffSpecsDetectsCombinedPathAndSchemaDrift(t *testing.T) {
	diff, err := diffSpecs(fixtureSpec([]string{"/api/v1/athlete/{id}", "/api/v1/legacy"}, []string{"Athlete", "LegacyModel"}), fixtureSpec([]string{"/api/v1/athlete/{id}", "/api/v1/athlete/{id}/activities"}, []string{"Athlete", "Activity"}))
	if err != nil {
		t.Fatalf("diffSpecs returned error: %v", err)
	}
	if got, want := strings.Join(diff.Added, ","), "/api/v1/athlete/{id}/activities"; got != want {
		t.Fatalf("added paths = %q, want %q", got, want)
	}
	if got, want := strings.Join(diff.Removed, ","), "/api/v1/legacy"; got != want {
		t.Fatalf("removed paths = %q, want %q", got, want)
	}
	if got, want := strings.Join(diff.SchemasAdded, ","), "Activity"; got != want {
		t.Fatalf("added schemas = %q, want %q", got, want)
	}
	if got, want := strings.Join(diff.SchemasRemoved, ","), "LegacyModel"; got != want {
		t.Fatalf("removed schemas = %q, want %q", got, want)
	}
}

func TestRenderMarkdownNoChangeOutput(t *testing.T) {
	diff, err := diffSpecs(fixtureSpec([]string{"/api/v1/athlete/{id}"}, []string{"Athlete"}), fixtureSpec([]string{"/api/v1/athlete/{id}"}, []string{"Athlete"}))
	if err != nil {
		t.Fatalf("diffSpecs returned error: %v", err)
	}
	out := renderMarkdown(diff, "baseline.json", "latest.json")
	for _, want := range []string{
		"Added paths: 0",
		"Removed paths: 0",
		"Added schemas: 0",
		"Removed schemas: 0",
		"No added endpoint paths detected.",
		"No removed endpoint paths detected.",
		"No added schema names detected.",
		"No removed schema names detected.",
		"human triage aid",
		"Schema-name drift is only a signal",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("markdown output missing %q:\n%s", want, out)
		}
	}
}

func fixtureSpec(paths, schemas []string) []byte {
	var b strings.Builder
	b.WriteString(`{"openapi":"3.0.0","info":{"title":"fixture","version":"test"},"paths":{`)
	for i, path := range paths {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(`"`)
		b.WriteString(path)
		b.WriteString(`":{}`)
	}
	b.WriteString(`},"components":{"schemas":{`)
	for i, schema := range schemas {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(`"`)
		b.WriteString(schema)
		b.WriteString(`":{}`)
	}
	b.WriteString(`}}}`)
	return []byte(b.String())
}

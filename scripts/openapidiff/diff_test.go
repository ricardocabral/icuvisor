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
	if !diff.hasStructuralDrift() {
		t.Fatal("hasStructuralDrift = false, want true for added path")
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
	if !diff.hasStructuralDrift() {
		t.Fatal("hasStructuralDrift = false, want true for removed path")
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
	if !diff.hasStructuralDrift() {
		t.Fatal("hasStructuralDrift = false, want true for schema-name drift")
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

func TestDiffSpecsIgnoresMetadataOnlyChanges(t *testing.T) {
	baseline := []byte(`{
		"openapi": "3.0.0",
		"info": {"title": "intervals", "version": "baseline", "description": "baseline docs"},
		"paths": {
			"/api/v1/athlete/{id}": {
				"get": {
					"summary": "Get athlete",
					"description": "Baseline endpoint documentation",
					"responses": {"200": {"description": "ok"}}
				}
			}
		},
		"components": {
			"schemas": {
				"Athlete": {
					"type": "object",
					"description": "Baseline schema documentation",
					"example": {"name": "baseline"}
				}
			}
		}
	}`)
	latest := []byte(`{
		"openapi": "3.0.0",
		"info": {"title": "intervals.icu", "version": "latest", "description": "latest docs"},
		"paths": {
			"/api/v1/athlete/{id}": {
				"get": {
					"summary": "Fetch athlete profile",
					"description": "Updated endpoint documentation",
					"responses": {"200": {"description": "success"}}
				}
			}
		},
		"components": {
			"schemas": {
				"Athlete": {
					"type": "object",
					"description": "Updated schema documentation",
					"example": {"name": "latest"}
				}
			}
		}
	}`)

	diff, err := diffSpecs(baseline, latest)
	if err != nil {
		t.Fatalf("diffSpecs returned error: %v", err)
	}
	assertNoStructuralDrift(t, diff)

	out := renderMarkdown(diff, "baseline.json", "latest.json")
	for _, want := range []string{
		"No structural OpenAPI key drift detected",
		"Metadata, descriptions, examples, or formatting may have changed",
		"method/request/response/field-level semantic edits are outside this key-level check",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("markdown output missing %q:\n%s", want, out)
		}
	}
}

func TestDiffSpecsIgnoresFormattingAndOrderOnlyChanges(t *testing.T) {
	baseline := []byte(`{"openapi":"3.0.0","info":{"title":"fixture","version":"test"},"paths":{"/api/v1/b":{},"/api/v1/a":{}},"components":{"schemas":{"Beta":{},"Alpha":{}}}}`)
	latest := []byte(`{
		"components": {"schemas": {"Alpha": {}, "Beta": {}}},
		"paths": {
			"/api/v1/a": {},
			"/api/v1/b": {}
		},
		"info": {"version": "test", "title": "fixture"},
		"openapi": "3.0.0"
	}`)

	diff, err := diffSpecs(baseline, latest)
	if err != nil {
		t.Fatalf("diffSpecs returned error: %v", err)
	}
	assertNoStructuralDrift(t, diff)
}

func TestRenderMarkdownNoChangeOutput(t *testing.T) {
	diff, err := diffSpecs(fixtureSpec([]string{"/api/v1/athlete/{id}"}, []string{"Athlete"}), fixtureSpec([]string{"/api/v1/athlete/{id}"}, []string{"Athlete"}))
	if err != nil {
		t.Fatalf("diffSpecs returned error: %v", err)
	}
	out := renderMarkdown(diff, "baseline.json", "latest.json")
	for _, want := range []string{
		"No structural OpenAPI key drift detected",
		"the `paths` key inventory and `components.schemas` names are unchanged",
		"Metadata, descriptions, examples, or formatting may have changed",
		"method/request/response/field-level semantic edits are outside this key-level check",
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

func TestRenderMarkdownStructuralDriftOutput(t *testing.T) {
	diff := endpointDiff{Added: []string{"/api/v1/new"}}
	out := renderMarkdown(diff, "baseline.json", "latest.json")
	if !strings.Contains(out, "Structural OpenAPI key drift detected") {
		t.Fatalf("markdown output missing structural-drift classification:\n%s", out)
	}
}

func assertNoStructuralDrift(t *testing.T, diff endpointDiff) {
	t.Helper()
	if len(diff.Added) != 0 || len(diff.Removed) != 0 || len(diff.SchemasAdded) != 0 || len(diff.SchemasRemoved) != 0 {
		t.Fatalf("structural drift = added paths %v removed paths %v added schemas %v removed schemas %v, want none", diff.Added, diff.Removed, diff.SchemasAdded, diff.SchemasRemoved)
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

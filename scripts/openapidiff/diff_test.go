package main

import (
	"strings"
	"testing"
)

func TestDiffSpecsDetectsAddedPaths(t *testing.T) {
	diff, err := diffSpecs(fixtureSpec("/api/v1/athlete/{id}"), fixtureSpec("/api/v1/athlete/{id}", "/api/v1/athlete/{id}/activities"))
	if err != nil {
		t.Fatalf("diffSpecs returned error: %v", err)
	}
	if got, want := strings.Join(diff.Added, ","), "/api/v1/athlete/{id}/activities"; got != want {
		t.Fatalf("added paths = %q, want %q", got, want)
	}
	if len(diff.Removed) != 0 {
		t.Fatalf("removed paths = %v, want none", diff.Removed)
	}
}

func TestDiffSpecsDetectsRemovedPaths(t *testing.T) {
	diff, err := diffSpecs(fixtureSpec("/api/v1/athlete/{id}", "/api/v1/athlete/{id}/events"), fixtureSpec("/api/v1/athlete/{id}"))
	if err != nil {
		t.Fatalf("diffSpecs returned error: %v", err)
	}
	if got, want := strings.Join(diff.Removed, ","), "/api/v1/athlete/{id}/events"; got != want {
		t.Fatalf("removed paths = %q, want %q", got, want)
	}
	if len(diff.Added) != 0 {
		t.Fatalf("added paths = %v, want none", diff.Added)
	}
}

func TestRenderMarkdownNoChangeOutput(t *testing.T) {
	diff, err := diffSpecs(fixtureSpec("/api/v1/athlete/{id}"), fixtureSpec("/api/v1/athlete/{id}"))
	if err != nil {
		t.Fatalf("diffSpecs returned error: %v", err)
	}
	out := renderMarkdown(diff, "baseline.json", "latest.json")
	for _, want := range []string{
		"Added paths: 0",
		"Removed paths: 0",
		"No added endpoint paths detected.",
		"No removed endpoint paths detected.",
		"human triage aid",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("markdown output missing %q:\n%s", want, out)
		}
	}
}

func fixtureSpec(paths ...string) []byte {
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
	b.WriteString(`}}`)
	return []byte(b.String())
}

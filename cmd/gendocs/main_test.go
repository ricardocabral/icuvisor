package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunWritesGeneratedDocsGolden(t *testing.T) {
	t.Parallel()

	outDir := filepath.Join(t.TempDir(), "data")
	toolsOut := filepath.Join(outDir, "tools.json")
	schemasOut := filepath.Join(outDir, "tool_schemas.json")
	if err := run([]string{"--out", toolsOut, "--schemas-out", schemasOut}); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	assertGoldenFile(t, toolsOut, filepath.Join("testdata", "tools.golden.json"))
	assertGoldenFile(t, schemasOut, filepath.Join("testdata", "tool_schemas.golden.json"))
}

func TestRunWritesDeterministicSchemaCatalog(t *testing.T) {
	t.Parallel()

	outDir := t.TempDir()
	firstTools := filepath.Join(outDir, "first-tools.json")
	firstSchemas := filepath.Join(outDir, "first-schemas.json")
	secondTools := filepath.Join(outDir, "second-tools.json")
	secondSchemas := filepath.Join(outDir, "second-schemas.json")
	if err := run([]string{"--out", firstTools, "--schemas-out", firstSchemas}); err != nil {
		t.Fatalf("first run() error = %v", err)
	}
	if err := run([]string{"--out", secondTools, "--schemas-out", secondSchemas}); err != nil {
		t.Fatalf("second run() error = %v", err)
	}

	assertSameFile(t, firstTools, secondTools)
	assertSameFile(t, firstSchemas, secondSchemas)
}

func TestGeneratedSchemaCatalogIncludesKeyFieldsAndExamples(t *testing.T) {
	t.Parallel()

	outDir := t.TempDir()
	schemasOut := filepath.Join(outDir, "tool_schemas.json")
	if err := run([]string{"--out", filepath.Join(outDir, "tools.json"), "--schemas-out", schemasOut}); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	raw, err := os.ReadFile(schemasOut)
	if err != nil {
		t.Fatalf("reading schema catalog: %v", err)
	}
	var catalog map[string]generatedToolSchema
	if err := json.Unmarshal(raw, &catalog); err != nil {
		t.Fatalf("unmarshal schema catalog: %v", err)
	}

	assertArgument(t, catalog, "get_activity_details", "include_full")
	assertArgument(t, catalog, "add_or_update_event", "date")
	assertArgument(t, catalog, "add_or_update_event", "category")
	workoutDoc := assertArgument(t, catalog, "create_workout", "workout_doc")
	if workoutDoc.Type != "object" || workoutDoc.AdditionalProperties == "" {
		t.Fatalf("create_workout.workout_doc = %#v, want object with additional_properties summary", workoutDoc)
	}
	content := assertArgument(t, catalog, "create_custom_item", "content")
	if content.Type != "object" || content.AdditionalProperties == "" {
		t.Fatalf("create_custom_item.content = %#v, want object with additional_properties summary", content)
	}

	for _, toolName := range []string{"add_or_update_event", "create_workout", "update_workout", "create_custom_item"} {
		examples := catalog[toolName].Examples
		if len(examples) == 0 || len(examples) > 2 {
			t.Fatalf("%s examples length = %d, want 1..2", toolName, len(examples))
		}
	}

	text := string(raw)
	for _, forbidden := range []string{"ICUVISOR_API_KEY", "api_key", "apikey", os.Getenv("HOME"), outDir} {
		if forbidden == "" {
			continue
		}
		if strings.Contains(strings.ToLower(text), strings.ToLower(forbidden)) {
			t.Fatalf("schema catalog contains forbidden text %q", forbidden)
		}
	}
}

type generatedToolSchema struct {
	Name      string                    `json:"name"`
	Arguments []generatedSchemaArgument `json:"arguments"`
	Examples  []any                     `json:"examples"`
}

type generatedSchemaArgument struct {
	Name                 string `json:"name"`
	Required             bool   `json:"required"`
	Type                 string `json:"type"`
	AdditionalProperties string `json:"additional_properties"`
}

func assertArgument(t *testing.T, catalog map[string]generatedToolSchema, toolName, argName string) generatedSchemaArgument {
	t.Helper()
	tool, ok := catalog[toolName]
	if !ok {
		t.Fatalf("schema catalog missing %s", toolName)
	}
	for _, arg := range tool.Arguments {
		if arg.Name == argName {
			return arg
		}
	}
	t.Fatalf("%s missing argument %s", toolName, argName)
	return generatedSchemaArgument{}
}

func assertGoldenFile(t *testing.T, gotPath, wantPath string) {
	t.Helper()
	got, err := os.ReadFile(gotPath)
	if err != nil {
		t.Fatalf("reading generated file: %v", err)
	}
	want, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("reading golden file: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("generated file differs from golden\n got: %s\nwant: %s", got, want)
	}
}

func assertSameFile(t *testing.T, firstPath, secondPath string) {
	t.Helper()
	first, err := os.ReadFile(firstPath)
	if err != nil {
		t.Fatalf("reading first file: %v", err)
	}
	second, err := os.ReadFile(secondPath)
	if err != nil {
		t.Fatalf("reading second file: %v", err)
	}
	if !bytes.Equal(first, second) {
		t.Fatalf("generated files differ\nfirst: %s\nsecond: %s", first, second)
	}
}

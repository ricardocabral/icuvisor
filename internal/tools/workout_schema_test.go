package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/safety"
)

func TestWorkoutRelatedInputSchemasRemainNonRecursive(t *testing.T) {
	t.Parallel()

	registrar := &collectingRegistrar{}
	registry := NewRegistryWithOptions(newNoNetworkIntervalsClient(t), RegistryOptions{
		Version:          "test",
		TimezoneFallback: "UTC",
		Capability:       safety.NewCapability(safety.ModeFull),
		Toolset:          safety.ToolsetFull,
	})
	if err := registry.Register(context.Background(), registrar); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	wantWorkoutDocTools := map[string]bool{
		addOrUpdateEventName:     false,
		createWorkoutName:        false,
		setActivityIntervalsName: false,
		updateWorkoutName:        false,
		validateWorkoutName:      false,
	}

	for _, tool := range registrar.tools {
		schema, ok := tool.InputSchema.(map[string]any)
		if !ok {
			t.Fatalf("%s InputSchema type = %T, want map[string]any", tool.Name, tool.InputSchema)
		}
		properties, _ := schema["properties"].(map[string]any)
		if _, ok := properties["workout_doc"]; !ok {
			continue
		}
		if _, ok := wantWorkoutDocTools[tool.Name]; !ok {
			t.Fatalf("unexpected registered tool %s accepts workout_doc", tool.Name)
		}
		wantWorkoutDocTools[tool.Name] = true

		if _, err := json.Marshal(schema); err != nil {
			t.Fatalf("%s input schema did not JSON-marshal cleanly: %v", tool.Name, err)
		}
		assertNonRecursiveWorkoutSchema(t, tool.Name, schema)
		assertWorkoutSyntaxResourceGuidance(t, tool, schema)
	}

	var missing []string
	for name, seen := range wantWorkoutDocTools {
		if !seen {
			missing = append(missing, name)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Fatalf("registered workout_doc tools missing from schema walk: %v", missing)
	}
}

func assertNonRecursiveWorkoutSchema(t *testing.T, toolName string, schema map[string]any) {
	t.Helper()

	var walk func(path string, node any)
	walk = func(path string, node any) {
		switch typed := node.(type) {
		case map[string]any:
			if ref, ok := typed["$ref"].(string); ok {
				lowerRef := strings.ToLower(ref)
				if strings.Contains(lowerRef, "workoutdoc") || strings.Contains(lowerRef, "workout_doc") || strings.Contains(lowerRef, "step") {
					t.Fatalf("%s input schema contains recursive workout-related $ref at %s: %s", toolName, path, ref)
				}
			}
			if properties, ok := typed["properties"].(map[string]any); ok {
				if stepsSchema, ok := properties["steps"]; ok && schemaHasNestedStepExpansion(stepsSchema) {
					t.Fatalf("%s input schema inlines nested steps.items.properties.steps at %s.properties.steps", toolName, path)
				}
				for name, propertySchema := range properties {
					walk(path+".properties."+name, propertySchema)
				}
			}
			for _, key := range []string{"items", "additionalProperties", "not"} {
				if child, ok := typed[key]; ok {
					walk(path+"."+key, child)
				}
			}
			for _, key := range []string{"allOf", "anyOf", "oneOf"} {
				if children, ok := typed[key].([]any); ok {
					for i, child := range children {
						walk(fmt.Sprintf("%s.%s[%d]", path, key, i), child)
					}
				}
			}
		case []any:
			for i, child := range typed {
				walk(fmt.Sprintf("%s[%d]", path, i), child)
			}
		}
	}

	walk("$", schema)
}

func schemaHasNestedStepExpansion(schema any) bool {
	typed, ok := schema.(map[string]any)
	if !ok {
		return false
	}
	items, ok := typed["items"].(map[string]any)
	if !ok {
		return false
	}
	properties, ok := items["properties"].(map[string]any)
	if !ok {
		return false
	}
	_, ok = properties["steps"]
	return ok
}

func assertWorkoutSyntaxResourceGuidance(t *testing.T, tool Tool, schema map[string]any) {
	t.Helper()

	var descriptions []string
	appendSchemaDescriptions(&descriptions, schema)
	contract := strings.ToLower(tool.Description + "\n" + strings.Join(descriptions, "\n"))
	if !strings.Contains(contract, "icuvisor://workout-syntax") {
		t.Fatalf("%s workout_doc schema contract does not point to icuvisor://workout-syntax", tool.Name)
	}
}

func appendSchemaDescriptions(out *[]string, node any) {
	typed, ok := node.(map[string]any)
	if !ok {
		return
	}
	if description, ok := typed["description"].(string); ok {
		*out = append(*out, description)
	}
	if properties, ok := typed["properties"].(map[string]any); ok {
		for _, child := range properties {
			appendSchemaDescriptions(out, child)
		}
	}
	if items, ok := typed["items"]; ok {
		appendSchemaDescriptions(out, items)
	}
	if additionalProperties, ok := typed["additionalProperties"]; ok {
		appendSchemaDescriptions(out, additionalProperties)
	}
	for _, key := range []string{"allOf", "anyOf", "oneOf"} {
		children, ok := typed[key].([]any)
		if !ok {
			continue
		}
		for _, child := range children {
			appendSchemaDescriptions(out, child)
		}
	}
}

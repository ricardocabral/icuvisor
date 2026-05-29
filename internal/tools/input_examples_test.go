package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/safety"
)

func TestRegisteredV03WriteToolsExposeInputExamples(t *testing.T) {
	client, err := intervals.NewClient(intervals.Options{Config: config.Config{APIKey: "example", AthleteID: "i12345"}, Version: "test"})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	registrar := &collectingRegistrar{}
	if err := NewRegistryWithOptions(client, RegistryOptions{Version: "test", TimezoneFallback: "UTC", Capability: safety.NewCapability(safety.ModeFull)}).Register(context.Background(), registrar); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	covered := map[string]bool{
		addOrUpdateEventName:    false,
		createWorkoutName:       false,
		updateWorkoutName:       false,
		createCustomItemName:    false,
		updateCustomItemName:    false,
		applyTrainingPlanName:   false,
		updateWellnessName:      false,
		updateSportSettingsName: false,
	}
	simpleWritesWithoutExamples := map[string]string{
		addActivityMessageName:  "simple free-text activity comment writer",
		linkActivityToEventName: "simple activity/event ID linker",
		updateActivityName:      "simple sparse activity name/description updater",
	}

	for _, tool := range registrar.tools {
		if tool.Requirement != RequirementWrite {
			continue
		}
		if reason, ok := simpleWritesWithoutExamples[tool.Name]; ok {
			t.Logf("%s does not require input_examples: %s", tool.Name, reason)
			continue
		}
		schema, ok := tool.InputSchema.(map[string]any)
		if !ok {
			t.Fatalf("%s InputSchema type = %T, want map[string]any", tool.Name, tool.InputSchema)
		}
		examples := schemaInputExamples(t, schema)
		if len(examples) == 0 {
			t.Fatalf("%s input_examples is empty", tool.Name)
		}
		if _, ok := covered[tool.Name]; ok {
			covered[tool.Name] = true
		}
	}

	for name, found := range covered {
		if !found {
			t.Fatalf("registered write tool %s was not checked for input_examples", name)
		}
	}
}

func TestWorkoutWriteDescriptionSchemaAdvertisesMergedWorkoutDoc(t *testing.T) {
	registrar := &collectingRegistrar{}
	if err := NewRegistryWithOptions(newNoNetworkIntervalsClient(t), RegistryOptions{Version: "test", TimezoneFallback: "UTC", Capability: safety.NewCapability(safety.ModeFull)}).Register(context.Background(), registrar); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	for _, name := range []string{addOrUpdateEventName, createWorkoutName, updateWorkoutName} {
		t.Run(name, func(t *testing.T) {
			tool := findTool(t, registrar.tools, name)
			schema, ok := tool.InputSchema.(map[string]any)
			if !ok {
				t.Fatalf("%s InputSchema type = %T, want map[string]any", tool.Name, tool.InputSchema)
			}
			props := schemaProperties(t, schema)
			descriptionField := schemaPropertyDescription(t, props, "description")
			workoutDocField := schemaPropertyDescription(t, props, "workout_doc")
			contractText := normalizeContractText(tool.Description + "\n" + descriptionField + "\n" + workoutDocField)

			for _, forbidden := range []string{"mutually exclusive", "mutually-exclusive", "exclusive with", "not both", "do not supply both", "cannot be supplied together", "either description or workout_doc", "choose description or workout_doc"} {
				if strings.Contains(contractText, forbidden) {
					t.Fatalf("%s schema contract contains contradictory wording %q:\n%s", tool.Name, forbidden, contractText)
				}
			}
			if !strings.Contains(contractText, "description") || !strings.Contains(contractText, "workout_doc") {
				t.Fatalf("%s schema contract should mention both description and workout_doc:\n%s", tool.Name, contractText)
			}
			if !strings.Contains(contractText, "may be supplied with workout_doc") && !strings.Contains(contractText, "merged with description when both are supplied") && !strings.Contains(contractText, "sentinel") {
				t.Fatalf("%s schema contract lost merge/coexistence or sentinel guidance:\n%s", tool.Name, contractText)
			}
			for _, want := range []string{"preview", "total duration", "key steps", "target intensities", "load/distance/time", "preserved"} {
				if !strings.Contains(contractText, want) {
					t.Fatalf("%s schema contract missing preview guidance %q:\n%s", tool.Name, want, contractText)
				}
			}
			if strings.Contains(contractText, "confirm") {
				t.Fatalf("%s schema contract introduced model-controlled confirm wording:\n%s", tool.Name, contractText)
			}
		})
	}
}

func TestComplexWriteToolInputExamplesValidateAgainstSchema(t *testing.T) {
	targets := []struct {
		name   string
		schema map[string]any
	}{
		{name: addOrUpdateEventName, schema: addOrUpdateEventInputSchema()},
		{name: createWorkoutName, schema: createWorkoutInputSchema()},
		{name: updateWorkoutName, schema: updateWorkoutInputSchema()},
		{name: createCustomItemName, schema: createCustomItemInputSchema()},
		{name: updateCustomItemName, schema: updateCustomItemInputSchema()},
		{name: applyTrainingPlanName, schema: applyTrainingPlanInputSchema(safety.NewCapability(safety.ModeSafe))},
		{name: updateWellnessName, schema: updateWellnessInputSchema()},
		{name: updateSportSettingsName, schema: updateSportSettingsInputSchema()},
	}

	for _, target := range targets {
		t.Run(target.name, func(t *testing.T) {
			examples := schemaInputExamples(t, target.schema)
			if len(examples) < 2 {
				t.Fatalf("input_examples length = %d, want at least 2", len(examples))
			}
			for i, example := range examples {
				roundTripped := jsonRoundTripExample(t, example)
				if err := validateExampleAgainstSchema(target.schema, roundTripped, target.name); err != nil {
					t.Fatalf("input_examples[%d] does not validate: %v", i, err)
				}
				assertNoRealAthleteData(t, roundTripped, fmt.Sprintf("%s.input_examples[%d]", target.name, i))
			}
		})
	}
}

func schemaProperties(t *testing.T, schema map[string]any) map[string]any {
	t.Helper()
	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("schema properties type = %T, want map[string]any", schema["properties"])
	}
	return props
}

func schemaPropertyDescription(t *testing.T, props map[string]any, name string) string {
	t.Helper()
	prop, ok := props[name].(map[string]any)
	if !ok {
		t.Fatalf("schema property %s type = %T, want map[string]any", name, props[name])
	}
	description, ok := prop["description"].(string)
	if !ok || strings.TrimSpace(description) == "" {
		t.Fatalf("schema property %s description = %#v, want non-empty string", name, prop["description"])
	}
	return description
}

func normalizeContractText(text string) string {
	return strings.Join(strings.Fields(strings.ToLower(text)), " ")
}

func schemaInputExamples(t *testing.T, schema map[string]any) []any {
	t.Helper()
	examples, ok := schema["input_examples"]
	if !ok {
		t.Fatal("input_examples missing")
	}
	mirrored, ok := schema["examples"]
	if !ok {
		t.Fatal("examples missing")
	}
	if !reflect.DeepEqual(examples, mirrored) {
		t.Fatalf("examples and input_examples differ\nexamples=%#v\ninput_examples=%#v", mirrored, examples)
	}
	switch typed := examples.(type) {
	case []any:
		return typed
	case []map[string]any:
		out := make([]any, 0, len(typed))
		for _, example := range typed {
			out = append(out, example)
		}
		return out
	default:
		t.Fatalf("input_examples type = %T, want []any or []map[string]any", examples)
		return nil
	}
}

func jsonRoundTripExample(t *testing.T, example any) any {
	t.Helper()
	encoded, err := json.Marshal(example)
	if err != nil {
		t.Fatalf("marshal example: %v", err)
	}
	var decoded any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal example: %v", err)
	}
	return decoded
}

func validateExampleAgainstSchema(schema map[string]any, value any, path string) error {
	if schemaType, ok := schema["type"].(string); ok {
		if err := validateJSONType(schemaType, value, path); err != nil {
			return err
		}
	}
	if enumValues, ok := anySlice(schema["enum"]); ok {
		matched := false
		for _, enumValue := range enumValues {
			if reflect.DeepEqual(value, enumValue) {
				matched = true
				break
			}
		}
		if !matched {
			return fmt.Errorf("%s = %v not in enum %v", path, value, enumValues)
		}
	}
	if min, ok := numberKeyword(schema["minimum"]); ok {
		valueNumber, ok := asNumber(value)
		if !ok || valueNumber < min {
			return fmt.Errorf("%s = %v below minimum %v", path, value, min)
		}
	}
	if min, ok := numberKeyword(schema["exclusiveMinimum"]); ok {
		valueNumber, ok := asNumber(value)
		if !ok || valueNumber <= min {
			return fmt.Errorf("%s = %v not greater than exclusiveMinimum %v", path, value, min)
		}
	}
	if max, ok := numberKeyword(schema["maximum"]); ok {
		valueNumber, ok := asNumber(value)
		if !ok || valueNumber > max {
			return fmt.Errorf("%s = %v above maximum %v", path, value, max)
		}
	}
	if schemaType, _ := schema["type"].(string); schemaType == "object" {
		object, _ := value.(map[string]any)
		for _, required := range stringSlice(schema["required"]) {
			if _, ok := object[required]; !ok {
				return fmt.Errorf("%s missing required property %q", path, required)
			}
		}
		properties, _ := schema["properties"].(map[string]any)
		if additional, ok := schema["additionalProperties"].(bool); ok && !additional {
			for key := range object {
				if _, ok := properties[key]; !ok {
					return fmt.Errorf("%s has additional property %q", path, key)
				}
			}
		}
		for key, child := range properties {
			childValue, ok := object[key]
			if !ok {
				continue
			}
			childSchema, ok := child.(map[string]any)
			if !ok {
				return fmt.Errorf("%s.%s schema type = %T, want map[string]any", path, key, child)
			}
			if err := validateExampleAgainstSchema(childSchema, childValue, path+"."+key); err != nil {
				return err
			}
		}
	}
	if schemaType, _ := schema["type"].(string); schemaType == "array" {
		array, _ := value.([]any)
		if minItems, ok := numberKeyword(schema["minItems"]); ok && float64(len(array)) < minItems {
			return fmt.Errorf("%s length = %d below minItems %v", path, len(array), minItems)
		}
		itemSchema, ok := schema["items"].(map[string]any)
		if ok {
			for i, item := range array {
				if err := validateExampleAgainstSchema(itemSchema, item, fmt.Sprintf("%s[%d]", path, i)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func validateJSONType(schemaType string, value any, path string) error {
	switch schemaType {
	case "object":
		if _, ok := value.(map[string]any); !ok {
			return fmt.Errorf("%s type = %T, want object", path, value)
		}
	case "array":
		if _, ok := value.([]any); !ok {
			return fmt.Errorf("%s type = %T, want array", path, value)
		}
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("%s type = %T, want string", path, value)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("%s type = %T, want boolean", path, value)
		}
	case "integer":
		if number, ok := asNumber(value); !ok || math.Trunc(number) != number {
			return fmt.Errorf("%s type = %T(%v), want integer", path, value, value)
		}
	case "number":
		if _, ok := asNumber(value); !ok {
			return fmt.Errorf("%s type = %T, want number", path, value)
		}
	default:
		return fmt.Errorf("%s schema has unsupported type %q", path, schemaType)
	}
	return nil
}

func anySlice(value any) ([]any, bool) {
	switch typed := value.(type) {
	case []any:
		return typed, true
	case []string:
		out := make([]any, 0, len(typed))
		for _, item := range typed {
			out = append(out, item)
		}
		return out, true
	default:
		return nil, false
	}
}

func stringSlice(value any) []string {
	switch typed := value.(type) {
	case []string:
		return typed
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

func numberKeyword(value any) (float64, bool) {
	return asNumber(value)
}

func asNumber(value any) (float64, bool) {
	switch typed := value.(type) {
	case int:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	default:
		return 0, false
	}
}

var athleteLikeID = regexp.MustCompile(`(?i)^i\d{4,}$`)

func assertNoRealAthleteData(t *testing.T, value any, path string) {
	t.Helper()
	sensitiveKeys := map[string]struct{}{
		"api_key":    {},
		"apikey":     {},
		"athlete_id": {},
		"password":   {},
		"secret":     {},
		"token":      {},
	}
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			if _, sensitive := sensitiveKeys[strings.ToLower(key)]; sensitive {
				t.Fatalf("%s.%s contains sensitive or real-athlete field", path, key)
			}
			assertNoRealAthleteData(t, child, path+"."+key)
		}
	case []any:
		for i, child := range typed {
			assertNoRealAthleteData(t, child, fmt.Sprintf("%s[%d]", path, i))
		}
	case string:
		lower := strings.ToLower(typed)
		if strings.Contains(lower, "api_key") || strings.Contains(lower, "bearer ") || strings.Contains(lower, "password") || strings.Contains(lower, "secret") || athleteLikeID.MatchString(typed) {
			t.Fatalf("%s contains sensitive-looking fixture value %q", path, typed)
		}
	}
}

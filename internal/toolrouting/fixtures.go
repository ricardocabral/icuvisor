package toolrouting

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
)

// Fixture contains prompt cases for first-tool routing smoke evaluations.
type Fixture struct {
	Version int    `json:"version"`
	Cases   []Case `json:"cases"`
}

// Case describes one model prompt and the first tool icuvisor expects a model to select.
type Case struct {
	ID                string  `json:"id"`
	Prompt            string  `json:"prompt"`
	ExpectedFirstTool *string `json:"expected_first_tool"`
	CatalogMode       string  `json:"catalog_mode"`
	Toolset           string  `json:"toolset"`
	DeleteMode        string  `json:"delete_mode"`
	Notes             string  `json:"notes,omitempty"`
}

// Result captures the first tool selected for a case, or NoTool when the model should not call a tool.
type Result struct {
	CaseID     string `json:"case_id"`
	Expected   string `json:"expected"`
	Actual     string `json:"actual"`
	NoTool     bool   `json:"no_tool"`
	Pass       bool   `json:"pass"`
	Detail     string `json:"detail,omitempty"`
	RawMessage string `json:"raw_message,omitempty"`
}

const NoToolName = "<no_tool>"

var errUnknownCatalogMode = errors.New("unknown catalog mode")

// LoadFixture decodes and validates routing smoke-eval cases.
func LoadFixture(r io.Reader, knownTools map[string]struct{}) (Fixture, error) {
	var fixture Fixture
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&fixture); err != nil {
		return Fixture{}, fmt.Errorf("decoding routing fixture: %w", err)
	}
	if err := validateFixture(fixture, knownTools); err != nil {
		return Fixture{}, err
	}
	return fixture, nil
}

// CompareResult compares an observed first tool or no-tool result to a case expectation.
func CompareResult(c Case, actualFirstTool string, rawMessage string) Result {
	expected := expectedName(c)
	actual := strings.TrimSpace(actualFirstTool)
	noTool := actual == ""
	if noTool {
		actual = NoToolName
	}
	pass := actual == expected
	detail := ""
	if !pass {
		detail = fmt.Sprintf("expected %s, got %s", expected, actual)
	}
	return Result{
		CaseID:     c.ID,
		Expected:   expected,
		Actual:     actual,
		NoTool:     noTool,
		Pass:       pass,
		Detail:     detail,
		RawMessage: rawMessage,
	}
}

func validateFixture(fixture Fixture, knownTools map[string]struct{}) error {
	if fixture.Version != 1 {
		return fmt.Errorf("routing fixture version = %d, want 1", fixture.Version)
	}
	if len(fixture.Cases) == 0 {
		return errors.New("routing fixture has no cases")
	}
	seen := make(map[string]struct{}, len(fixture.Cases))
	for i, c := range fixture.Cases {
		if strings.TrimSpace(c.ID) == "" {
			return fmt.Errorf("case #%d has empty id", i)
		}
		if _, ok := seen[c.ID]; ok {
			return fmt.Errorf("case %q is duplicated", c.ID)
		}
		seen[c.ID] = struct{}{}
		if strings.TrimSpace(c.Prompt) == "" {
			return fmt.Errorf("case %q has empty prompt", c.ID)
		}
		if err := validateCatalogFields(c); err != nil {
			return fmt.Errorf("case %q: %w", c.ID, err)
		}
		if c.ExpectedFirstTool != nil {
			expected := strings.TrimSpace(*c.ExpectedFirstTool)
			if expected == "" {
				return fmt.Errorf("case %q has blank expected_first_tool", c.ID)
			}
			if _, ok := knownTools[expected]; !ok {
				return fmt.Errorf("case %q expects unknown tool %q", c.ID, expected)
			}
		}
	}
	return nil
}

func validateCatalogFields(c Case) error {
	if !slices.Contains([]string{"core_safe", "core_full_delete", "full_safe", "full_full_delete"}, c.CatalogMode) {
		return fmt.Errorf("%w %q", errUnknownCatalogMode, c.CatalogMode)
	}
	switch c.Toolset {
	case "core", "full":
	default:
		return fmt.Errorf("toolset = %q, want core or full", c.Toolset)
	}
	switch c.DeleteMode {
	case "safe", "full":
	default:
		return fmt.Errorf("delete_mode = %q, want safe or full", c.DeleteMode)
	}
	wantPrefix := c.Toolset + "_"
	wantSuffix := "safe"
	if c.DeleteMode == "full" {
		wantSuffix = "full_delete"
	}
	if c.CatalogMode != wantPrefix+wantSuffix {
		return fmt.Errorf("catalog_mode %q does not match toolset/delete_mode %s/%s", c.CatalogMode, c.Toolset, c.DeleteMode)
	}
	return nil
}

func expectedName(c Case) string {
	if c.ExpectedFirstTool == nil {
		return NoToolName
	}
	return strings.TrimSpace(*c.ExpectedFirstTool)
}

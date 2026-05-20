package analysis

import (
	"sort"
	"strings"
)

const (
	// MissingActionSkip is the default missing-data policy for analyzer windows.
	MissingActionSkip = "skip"
	// MinBaselineSamples is the default minimum sample count for baseline analyzers.
	MinBaselineSamples = 7
	// MinCorrelationSamples is the default minimum sample count for correlation analyzers.
	MinCorrelationSamples = 14
)

// AnalyzerMeta is the mandatory analyzer _meta contract emitted by analyzer tools.
type AnalyzerMeta struct {
	Method             string   `json:"method"`
	SourceTools        []string `json:"source_tools"`
	N                  int      `json:"n"`
	MissingDays        int      `json:"missing_days"`
	MissingAction      string   `json:"missing_action"`
	InsufficientSample bool     `json:"insufficient_sample"`
	FormulaRef         string   `json:"formula_ref,omitempty"`
}

// AnalyzerMetaInput describes analyzer metadata before normalization.
type AnalyzerMetaInput struct {
	Method        string
	SourceTools   []string
	N             int
	MissingDays   int
	MissingAction string
	MinSamples    int
	FormulaRef    string
}

// NewAnalyzerMeta normalizes analyzer metadata while preserving mandatory fields.
func NewAnalyzerMeta(input AnalyzerMetaInput) AnalyzerMeta {
	n := input.N
	if n < 0 {
		n = 0
	}
	missingDays := input.MissingDays
	if missingDays < 0 {
		missingDays = 0
	}
	missingAction := strings.TrimSpace(input.MissingAction)
	if missingAction == "" {
		missingAction = MissingActionSkip
	}
	return AnalyzerMeta{
		Method:             strings.TrimSpace(input.Method),
		SourceTools:        NormalizeSourceTools(input.SourceTools),
		N:                  n,
		MissingDays:        missingDays,
		MissingAction:      missingAction,
		InsufficientSample: InsufficientSample(n, input.MinSamples),
		FormulaRef:         strings.TrimSpace(input.FormulaRef),
	}
}

// NormalizeSourceTools trims, deduplicates, and sorts source tool names.
func NormalizeSourceTools(sourceTools []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(sourceTools))
	for _, sourceTool := range sourceTools {
		trimmed := strings.TrimSpace(sourceTool)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	sort.Strings(out)
	return out
}

// InsufficientSample reports whether a usable sample count misses a minimum.
func InsufficientSample(n int, minN int) bool {
	return minN > 0 && n < minN
}

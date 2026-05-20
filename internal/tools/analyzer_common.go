package tools

import (
	"fmt"

	"github.com/ricardocabral/icuvisor/internal/analysis"
	"github.com/ricardocabral/icuvisor/internal/response"
)

type analyzerResponseInput struct {
	Result any
	Series any
	Meta   analysis.AnalyzerMetaInput
}

type analyzerResponsePayload struct {
	Result any                   `json:"result,omitempty"`
	Series any                   `json:"series,omitempty"`
	Meta   analysis.AnalyzerMeta `json:"_meta"`
}

func newAnalyzerResponsePayload(input analyzerResponseInput, includeFull bool) analyzerResponsePayload {
	payload := analyzerResponsePayload{Result: input.Result, Meta: analysis.NewAnalyzerMeta(input.Meta)}
	if includeFull {
		payload.Series = input.Series
	}
	return payload
}

func shapeAnalyzerResponse(input analyzerResponseInput, includeFull bool, version string, debugMetadata bool, queryType string, unitSystem response.UnitSystem, shaping ...responseShaping) (any, error) {
	shapeCfg := responseShapingOrDefault(shaping)
	return response.Shape(newAnalyzerResponsePayload(input, includeFull), shapeCfg.options(includeFull, nil, version, debugMetadata, queryType, unitSystem))
}

func encodeAnalyzerResponse(input analyzerResponseInput, includeFull bool, version string, debugMetadata bool, queryType string, unitSystem response.UnitSystem, shaping ...responseShaping) (Result, error) {
	shaped, err := shapeAnalyzerResponse(input, includeFull, version, debugMetadata, queryType, unitSystem, shaping...)
	if err != nil {
		return Result{}, fmt.Errorf("shaping %s response: %w", queryType, err)
	}
	return TextResult(shaped), nil
}

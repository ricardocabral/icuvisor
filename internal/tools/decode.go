package tools

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// DecodeStrict decodes raw JSON arguments into T without accepting unknown fields.
func DecodeStrict[T any](raw json.RawMessage) (T, error) {
	var out T
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return out, nil
	}
	if trimmed[0] != '{' {
		return out, errors.New("arguments must be a JSON object")
	}
	decoder := json.NewDecoder(bytes.NewReader(trimmed))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&out); err != nil {
		if errors.Is(err, io.EOF) {
			return out, nil
		}
		return out, fmt.Errorf("decoding arguments: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return out, errors.New("unexpected trailing JSON")
	}
	return out, nil
}

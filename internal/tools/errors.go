package tools

import (
	"errors"
	"fmt"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

// ErrInvalidInput marks tool input-validation failures.
var ErrInvalidInput = errors.New("invalid input")

// validationErrorMessage returns a short, actionable LLM-facing message when err wraps an
// intervals.ValidationError (HTTP 422). The second return value is false when err is not a
// validation error.
func validationErrorMessage(err error) (string, bool) {
	var ve *intervals.ValidationError
	if !errors.As(err, &ve) {
		return "", false
	}
	if ve.Field != "" {
		return fmt.Sprintf(
			"intervals.icu rejected field %q: create it under Settings > Custom Fields in intervals.icu, or omit it from this call",
			ve.Field,
		), true
	}
	return "intervals.icu rejected the request (HTTP 422): check that all fields are configured in your intervals.icu account settings", true
}

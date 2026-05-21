package intervals

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var (
	// ErrUnauthorized indicates intervals.icu rejected the configured credentials.
	ErrUnauthorized = errors.New("intervals.icu unauthorized")
	// ErrNotFound indicates intervals.icu could not find the requested resource.
	ErrNotFound = errors.New("intervals.icu not found")
	// ErrRateLimited indicates intervals.icu rate limited the request.
	ErrRateLimited = errors.New("intervals.icu rate limited")
	// ErrUpstream indicates intervals.icu returned a transient upstream error.
	ErrUpstream = errors.New("intervals.icu upstream error")
	// ErrValidation indicates intervals.icu rejected the request body (HTTP 422).
	ErrValidation = errors.New("intervals.icu validation error")
	// ErrResponseTooLarge indicates intervals.icu returned a response body above the client safety cap.
	ErrResponseTooLarge = errors.New("intervals.icu response too large")
	// ErrTargetAthleteMismatch indicates an object response is not owned by the request target athlete.
	ErrTargetAthleteMismatch = errors.New("intervals.icu target athlete mismatch")
	// ErrUnsupportedWellnessFeel indicates intervals.icu does not accept feel on wellness writes.
	ErrUnsupportedWellnessFeel = errors.New("field_not_writable: feel (not accepted by intervals.icu wellness write)")
)

// Error describes a structured intervals.icu API error.
type Error struct {
	StatusCode int
	Kind       error
	RetryAfter time.Duration
}

func (e *Error) Error() string {
	if e == nil {
		return "intervals.icu error"
	}
	if e.StatusCode == 0 {
		return e.Kind.Error()
	}
	return fmt.Sprintf("%s (HTTP %d)", e.Kind, e.StatusCode)
}

// Unwrap exposes the stable sentinel error for errors.Is.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Kind
}

// ValidationError wraps a 422 rejection from intervals.icu and names the offending field when the
// upstream body contains enough context to parse one.
type ValidationError struct {
	// Field is the intervals.icu field name extracted from the upstream rejection body, or empty
	// when the body does not name a specific field.
	Field string
	// UpstreamMessage is the raw upstream rejection text, retained for logging only.
	UpstreamMessage string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: field %q rejected by intervals.icu", ErrValidation, e.Field)
	}
	return ErrValidation.Error()
}

// Unwrap allows errors.Is(err, ErrValidation) to match.
func (e *ValidationError) Unwrap() error {
	return ErrValidation
}

// validationFieldRe matches patterns like:
//   - "Unrecognized wellness field [SomeName]"
//   - "Unknown field: SomeName"
//   - "Field 'SomeName' is not valid"
var validationFieldRe = regexp.MustCompile(`(?i)(?:field[s]?[:\s]+['\[]?|unrecognized\s+\w+\s+field\s+\[?)([A-Za-z_]\w*)`)

// parseValidationError reads the 422 response body, attempts to extract a field name, and
// returns a ValidationError. The body reader is always drained and must be closed by the caller.
func parseValidationError(body io.Reader) *ValidationError {
	raw, err := io.ReadAll(io.LimitReader(body, 4096))
	if err != nil || len(raw) == 0 {
		return &ValidationError{}
	}
	text := strings.TrimSpace(string(raw))

	// Try JSON first: {"error":"..."} or {"message":"..."} shapes.
	var jsonBody struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	if jsonErr := json.Unmarshal(raw, &jsonBody); jsonErr == nil {
		candidate := jsonBody.Error
		if candidate == "" {
			candidate = jsonBody.Message
		}
		if candidate != "" {
			text = candidate
		}
	}

	ve := &ValidationError{UpstreamMessage: text}
	if m := validationFieldRe.FindStringSubmatch(text); len(m) >= 2 {
		ve.Field = m[1]
	}
	return ve
}

func errorForStatus(statusCode int, retryAfter time.Duration) *Error {
	kind := ErrUpstream
	switch statusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		kind = ErrUnauthorized
	case http.StatusNotFound:
		kind = ErrNotFound
	case http.StatusTooManyRequests:
		kind = ErrRateLimited
	case http.StatusUnprocessableEntity:
		kind = ErrValidation
	}
	return &Error{StatusCode: statusCode, Kind: kind, RetryAfter: retryAfter}
}

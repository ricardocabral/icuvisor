package intervals

import (
	"errors"
	"fmt"
	"net/http"
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
	// ErrResponseTooLarge indicates intervals.icu returned a response body above the client safety cap.
	ErrResponseTooLarge = errors.New("intervals.icu response too large")
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

func errorForStatus(statusCode int, retryAfter time.Duration) *Error {
	kind := ErrUpstream
	switch statusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		kind = ErrUnauthorized
	case http.StatusNotFound:
		kind = ErrNotFound
	case http.StatusTooManyRequests:
		kind = ErrRateLimited
	}
	return &Error{StatusCode: statusCode, Kind: kind, RetryAfter: retryAfter}
}

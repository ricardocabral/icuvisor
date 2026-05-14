package prompts

import (
	"context"
	"errors"
)

// Registry registers MCP prompts with a registrar.
type Registry interface {
	Register(context.Context, Registrar) error
}

// Registrar accepts prompt definitions during server construction.
type Registrar interface {
	AddPrompt(Prompt) error
}

// Handler renders a prompt from string arguments supplied by the MCP client.
type Handler func(context.Context, Request) (Result, error)

// Prompt describes one MCP prompt without exposing SDK-specific types.
type Prompt struct {
	Name        string
	Title       string
	Description string
	Arguments   []Argument
	Handler     Handler
}

// Argument describes one string argument accepted by an MCP prompt.
type Argument struct {
	Name        string
	Title       string
	Description string
	Required    bool
}

// Request contains a prompt render request.
type Request struct {
	Name      string
	Arguments map[string]string
}

// Result is returned from a prompt handler.
type Result struct {
	Description string
	Messages    []Message
}

// Message contains one rendered prompt message.
type Message struct {
	Role Role
	Text string
}

// Role identifies the role for a rendered prompt message.
type Role string

const (
	// RoleUser is a user-authored prompt message.
	RoleUser Role = "user"
)

// UserError carries a short public prompt error and optional internal cause.
type UserError struct {
	Message string
	Err     error
}

// Error returns the short public message safe to show to an LLM.
func (e *UserError) Error() string {
	return e.Message
}

// Unwrap returns the internal cause, if any.
func (e *UserError) Unwrap() error {
	return e.Err
}

// NewUserError creates a user-facing prompt error with an optional internal cause.
func NewUserError(message string, err error) *UserError {
	return &UserError{Message: message, Err: err}
}

// PublicErrorMessage reports the short public message for err, if it has one.
func PublicErrorMessage(err error) (string, bool) {
	var userErr *UserError
	if errors.As(err, &userErr) && userErr.Message != "" {
		return userErr.Message, true
	}
	return "", false
}

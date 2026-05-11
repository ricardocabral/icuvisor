package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// Registry registers the MCP tools exposed by icuvisor.
type Registry interface {
	Register(context.Context, Registrar) error
}

// NewRegistry creates the default v0.1 tool registry.
func NewRegistry(profileClient ProfileClient, version string) Registry {
	return &defaultRegistry{profileClient: profileClient, version: normalizeVersion(version)}
}

type defaultRegistry struct {
	profileClient ProfileClient
	version       string
}

func (r *defaultRegistry) Register(ctx context.Context, registrar Registrar) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if r.profileClient == nil {
		return fmt.Errorf("registering %s: missing profile client", getAthleteProfileName)
	}
	return registrar.AddTool(newGetAthleteProfileTool(r.profileClient, r.version))
}

// Registrar accepts tool definitions from a Registry.
type Registrar interface {
	AddTool(Tool) error
}

// Handler handles a tool call using raw JSON arguments.
type Handler func(context.Context, Request) (Result, error)

// Tool describes one MCP tool without exposing SDK-specific types.
type Tool struct {
	Name         string
	Description  string
	InputSchema  any
	OutputSchema any
	Handler      Handler
}

// Request carries an MCP tool call to a Handler.
type Request struct {
	Name      string
	Arguments json.RawMessage
}

// Result is returned from a Handler.
type Result struct {
	Content           []Content
	StructuredContent any
	IsError           bool
}

// Content is a user-visible MCP response content item.
type Content struct {
	Type ContentType
	Text string
}

// ContentType identifies supported response content kinds.
type ContentType string

const (
	// ContentTypeText is plain text response content.
	ContentTypeText ContentType = "text"
)

// UserError carries a short public message and an optional internal cause.
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

// NewUserError creates a user-facing tool error with an optional internal cause.
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

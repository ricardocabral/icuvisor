package tools

import (
	"context"
	"encoding/json"
)

// Registry registers the MCP tools exposed by icuvisor.
type Registry interface {
	Register(context.Context, Registrar) error
}

// Registrar accepts tool definitions from a Registry.
type Registrar interface {
	AddTool(Tool) error
}

// Handler handles a tool call using raw JSON arguments.
type Handler func(context.Context, Request) (Result, error)

// Tool describes one MCP tool without exposing SDK-specific types.
type Tool struct {
	Name        string
	Description string
	InputSchema any
	Handler     Handler
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

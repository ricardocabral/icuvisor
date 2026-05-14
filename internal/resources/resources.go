package resources

import "context"

// Registry registers MCP resources with a registrar.
type Registry interface {
	Register(context.Context, Registrar) error
}

// Registrar accepts resource definitions during server construction.
type Registrar interface {
	AddResource(Resource) error
}

// Resource describes one readable MCP resource.
type Resource struct {
	URI         string
	Name        string
	Title       string
	Description string
	MIMEType    string
	Handler     Handler
}

// Handler reads a resource.
type Handler func(context.Context, Request) (Result, error)

// Request contains a resource read request.
type Request struct {
	URI string
}

// Result contains text resource contents.
type Result struct {
	URI      string
	MIMEType string
	Text     string
}

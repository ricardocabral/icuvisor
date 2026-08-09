package mcp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ricardocabral/icuvisor/internal/resources"
)

const genericResourceErrorMessage = "resource read failed; try again or check icuvisor logs"

type safeResourceRegistrar struct {
	server          *sdkmcp.Server
	logger          *slog.Logger
	uris            map[string]struct{}
	registeredCount int
}

func (r *safeResourceRegistrar) AddResource(resource resources.Resource) error {
	if err := r.validateResource(resource); err != nil {
		return err
	}
	r.uris[resource.URI] = struct{}{}

	return withPanicRecovery(fmt.Sprintf("registering resource %q", resource.URI), func() error {
		r.server.AddResource(&sdkmcp.Resource{
			URI:         resource.URI,
			Name:        resource.Name,
			Title:       resource.Title,
			Description: resource.Description,
			MIMEType:    resource.MIMEType,
		}, func(ctx context.Context, req *sdkmcp.ReadResourceRequest) (*sdkmcp.ReadResourceResult, error) {
			result, err := resource.Handler(ctx, resources.Request{URI: req.Params.URI})
			if err != nil {
				if isResourceNotFound(err) {
					return nil, err
				}
				r.logger.Error("resource handler failed", "resource_uri", resource.URI, "error", err)
				return nil, errors.New(genericResourceErrorMessage)
			}
			return &sdkmcp.ReadResourceResult{Contents: []*sdkmcp.ResourceContents{{
				URI:      stringOrDefault(result.URI, req.Params.URI),
				MIMEType: stringOrDefault(result.MIMEType, resource.MIMEType),
				Text:     result.Text,
			}}}, nil
		})
		r.registeredCount++
		return nil
	})
}

func (r *safeResourceRegistrar) validateResource(resource resources.Resource) error {
	if resource.URI == "" {
		return errors.New("resource is missing a URI")
	}
	parsed, err := url.Parse(resource.URI)
	if err != nil || !parsed.IsAbs() || parsed.Scheme != "icuvisor" {
		return fmt.Errorf("invalid resource URI %q; use absolute icuvisor:// URI", resource.URI)
	}
	if _, exists := r.uris[resource.URI]; exists {
		return fmt.Errorf("duplicate resource URI %q", resource.URI)
	}
	if resource.Name == "" {
		return fmt.Errorf("resource %q is missing a name", resource.URI)
	}
	if resource.Title == "" {
		return fmt.Errorf("resource %q is missing a title", resource.URI)
	}
	if resource.Description == "" {
		return fmt.Errorf("resource %q is missing a description", resource.URI)
	}
	if resource.MIMEType == "" {
		return fmt.Errorf("resource %q is missing a MIME type", resource.URI)
	}
	if resource.Handler == nil {
		return fmt.Errorf("resource %q is missing a handler", resource.URI)
	}
	return nil
}

func isResourceNotFound(err error) bool {
	var rpcErr *jsonrpc.Error
	return errors.As(err, &rpcErr) && rpcErr.Code == jsonrpc.CodeInvalidParams
}

func stringOrDefault(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

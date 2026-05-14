package resources

import (
	"context"
	"time"
)

// ResourceOptions configures the default MCP resource registry.
type ResourceOptions struct {
	Version           string
	TimezoneFallback  string
	DebugMetadata     bool
	AthleteProfileTTL time.Duration
	Now               func() time.Time
}

type staticRegistry struct {
	entries []Resource
}

// NewRegistry returns the default MCP resource registry with static resources.
func NewRegistry() Registry {
	return NewRegistryWithOptions(nil, ResourceOptions{})
}

// NewRegistryWithOptions returns the default MCP resource registry.
func NewRegistryWithOptions(profileClient ProfileClient, opts ResourceOptions) Registry {
	entries := []Resource{WorkoutSyntaxResource(), EventCategoriesResource(), CustomItemSchemasResource()}
	if profileClient != nil {
		entries = append(entries, AthleteProfileResource(profileClient, opts))
	}
	return staticRegistry{entries: entries}
}

func (r staticRegistry) Register(ctx context.Context, registrar Registrar) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	for _, resource := range r.entries {
		if err := registrar.AddResource(resource); err != nil {
			return err
		}
	}
	return nil
}

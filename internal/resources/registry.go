package resources

import "context"

type staticRegistry struct {
	entries []Resource
}

// NewRegistry returns the default MCP resource registry.
func NewRegistry() Registry {
	return staticRegistry{entries: []Resource{WorkoutSyntaxResource(), EventCategoriesResource(), CustomItemSchemasResource()}}
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

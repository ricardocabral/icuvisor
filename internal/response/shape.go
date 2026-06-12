package response

import (
	"errors"
	"time"

	"github.com/ricardocabral/icuvisor/internal/response/jsonenc"
	"github.com/ricardocabral/icuvisor/internal/safety"
)

// Options controls response shaping at the MCP response boundary.
type Options struct {
	IncludeFull    bool
	RowCollections []string
	ServerVersion  string
	DebugMetadata  bool
	QueryType      string
	FetchedAt      time.Time
	UnitSystem     UnitSystem
	DeleteMode     safety.Mode
	Toolset        safety.Toolset
	CatalogHash    string
}

// Shape converts value through JSON tags and applies response-boundary shaping.
func Shape(value any, opts Options) (any, error) {
	jsonValue, err := jsonenc.Encode(value)
	if err != nil {
		return nil, err
	}
	root, ok := jsonValue.(map[string]any)
	if !ok {
		return nil, errors.New("response shape must be a JSON object wrapper")
	}
	return shapeRoot(root, opts), nil
}

func shapeRoot(root map[string]any, opts Options) map[string]any {
	if len(opts.RowCollections) > 0 {
		return shapeWrapperRow(root, opts)
	}
	return shapeRow(root, opts, true)
}

func shapeRows(values []any, opts Options, includeCommonMeta bool) []any {
	rows := make([]any, 0, len(values))
	for _, item := range values {
		if row, ok := item.(map[string]any); ok {
			rows = append(rows, shapeRow(row, opts, includeCommonMeta))
			continue
		}
		if opts.IncludeFull {
			rows = append(rows, item)
			continue
		}
		stripped, _ := stripNulls(item, "")
		rows = append(rows, stripped)
	}
	return rows
}

func shapeWrapperRow(row map[string]any, opts Options) map[string]any {
	collections := rowCollectionSet(opts.RowCollections)
	out := make(map[string]any, len(row))
	var missing []string
	for key, item := range row {
		itemPath := key
		if item == nil {
			if opts.IncludeFull {
				out[key] = nil
			} else {
				missing = append(missing, itemPath)
			}
			continue
		}
		if collections[key] {
			if values, ok := item.([]any); ok {
				out[key] = shapeRows(values, opts, false)
			} else {
				out[key] = item
			}
			continue
		}
		if opts.IncludeFull {
			out[key] = item
			continue
		}
		stripped, nestedMissing := stripNulls(item, itemPath)
		out[key] = stripped
		missing = append(missing, nestedMissing...)
	}
	return finalizeShapedRow(out, missing, opts, true, true)
}

func shapeRow(row map[string]any, opts Options, includeCommonMeta bool) map[string]any {
	var shaped map[string]any
	var missing []string
	if opts.IncludeFull {
		shaped = cloneMap(row)
	} else {
		stripped, strippedMissing := stripNulls(row, "")
		var ok bool
		shaped, ok = stripped.(map[string]any)
		if !ok {
			shaped = cloneMap(row)
		}
		missing = strippedMissing
	}
	return finalizeShapedRow(shaped, missing, opts, includeCommonMeta, includeCommonMeta)
}

func finalizeShapedRow(row map[string]any, missing []string, opts Options, includeDebugMetadata bool, includeCommonMeta bool) map[string]any {
	if opts.DebugMetadata {
		if includeDebugMetadata {
			addDebugMetadata(row, opts)
		}
	} else {
		if dropped, ok := dropDebugMetadata(row, "").(map[string]any); ok {
			row = dropped
		}
		missing = filterDebugMissing(missing)
	}
	if !opts.IncludeFull && len(missing) > 0 {
		addStripMeta(row, missing)
	}
	addScaleMeta(row)
	if includeCommonMeta {
		addCommonMeta(row, opts)
	}
	return row
}

func stripNulls(value any, path string) (any, []string) {
	stripped, missing, _ := walkJSON(value, path, walkRoot, stripNullVisitor)
	return stripped, missing
}

func stripNullVisitor(path string, value any, container jsonWalkContainer) jsonWalkDecision {
	if value != nil {
		return jsonWalkDecision{}
	}
	if container == walkMapValue {
		return jsonWalkDecision{Drop: true, Missing: []string{path}}
	}
	return jsonWalkDecision{Stop: true}
}

func cloneMap(row map[string]any) map[string]any {
	out := make(map[string]any, len(row))
	for key, value := range row {
		out[key] = value
	}
	return out
}

func rowCollectionSet(rowCollections []string) map[string]bool {
	collections := make(map[string]bool, len(rowCollections))
	for _, key := range rowCollections {
		if key != "" {
			collections[key] = true
		}
	}
	return collections
}

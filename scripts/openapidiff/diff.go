package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

type specKeys struct {
	Paths      map[string]json.RawMessage `json:"paths"`
	Components struct {
		Schemas map[string]json.RawMessage `json:"schemas"`
	} `json:"components"`
}

type endpointDiff struct {
	Added          []string
	Removed        []string
	SchemasAdded   []string
	SchemasRemoved []string
}

func (diff endpointDiff) hasStructuralDrift() bool {
	return len(diff.Added) > 0 || len(diff.Removed) > 0 || len(diff.SchemasAdded) > 0 || len(diff.SchemasRemoved) > 0
}

func readSpec(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading spec %s: %w", path, err)
	}
	return data, nil
}

func fetchSpec(url string) ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching latest spec: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fetching latest spec: HTTP %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 20*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("reading latest spec response: %w", err)
	}
	return data, nil
}

func diffSpecs(baseline, latest []byte) (endpointDiff, error) {
	baselineKeys, err := extractKeys(baseline)
	if err != nil {
		return endpointDiff{}, fmt.Errorf("baseline: %w", err)
	}
	latestKeys, err := extractKeys(latest)
	if err != nil {
		return endpointDiff{}, fmt.Errorf("latest: %w", err)
	}

	return endpointDiff{
		Added:          addedKeys(baselineKeys.Paths, latestKeys.Paths),
		Removed:        addedKeys(latestKeys.Paths, baselineKeys.Paths),
		SchemasAdded:   addedKeys(baselineKeys.Components.Schemas, latestKeys.Components.Schemas),
		SchemasRemoved: addedKeys(latestKeys.Components.Schemas, baselineKeys.Components.Schemas),
	}, nil
}

func addedKeys[T any](baseline, latest map[string]T) []string {
	var added []string
	for key := range latest {
		if _, ok := baseline[key]; !ok {
			added = append(added, key)
		}
	}
	sort.Strings(added)
	return added
}

func extractKeys(data []byte) (specKeys, error) {
	var spec specKeys
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&spec); err != nil {
		return specKeys{}, fmt.Errorf("decoding JSON: %w", err)
	}
	if len(spec.Paths) == 0 {
		return specKeys{}, fmt.Errorf("missing or empty paths object")
	}
	paths, err := compactKeys(spec.Paths, "paths", true)
	if err != nil {
		return specKeys{}, err
	}
	schemas, err := compactKeys(spec.Components.Schemas, "components.schemas", false)
	if err != nil {
		return specKeys{}, err
	}
	spec.Paths = paths
	spec.Components.Schemas = schemas
	return spec, nil
}

func compactKeys[T any](items map[string]T, label string, required bool) (map[string]T, error) {
	if len(items) == 0 {
		if required {
			return nil, fmt.Errorf("missing or empty %s object", label)
		}
		return map[string]T{}, nil
	}
	compacted := make(map[string]T, len(items))
	for key, value := range items {
		if strings.TrimSpace(key) == "" {
			continue
		}
		compacted[key] = value
	}
	if len(compacted) == 0 && required {
		return nil, fmt.Errorf("%s object contains no non-empty keys", label)
	}
	return compacted, nil
}

func renderMarkdown(diff endpointDiff, baselineSource, latestSource string) string {
	var b strings.Builder
	b.WriteString("# intervals.icu OpenAPI endpoint diff\n\n")
	b.WriteString("This report compares OpenAPI `paths` keys and `components.schemas` names. It is a human triage aid; it does not approve product scope or auto-generate icuvisor tools. Schema-name drift is only a signal to inspect upstream docs and models.\n\n")
	b.WriteString("## Classification\n\n")
	if diff.hasStructuralDrift() {
		b.WriteString("Structural OpenAPI key drift detected: one or more `paths` keys or `components.schemas` names were added or removed. Review the sections below before updating the baseline.\n\n")
	} else {
		b.WriteString("No structural OpenAPI key drift detected: the `paths` key inventory and `components.schemas` names are unchanged. Metadata, descriptions, examples, or formatting may have changed, and method/request/response/field-level semantic edits are outside this key-level check.\n\n")
	}
	fmt.Fprintf(&b, "- Baseline: `%s`\n", baselineSource)
	fmt.Fprintf(&b, "- Latest: `%s`\n", latestSource)
	fmt.Fprintf(&b, "- Added paths: %d\n", len(diff.Added))
	fmt.Fprintf(&b, "- Removed paths: %d\n", len(diff.Removed))
	fmt.Fprintf(&b, "- Added schemas: %d\n", len(diff.SchemasAdded))
	fmt.Fprintf(&b, "- Removed schemas: %d\n\n", len(diff.SchemasRemoved))

	writeListSection(&b, "Added paths", diff.Added, "No added endpoint paths detected.")
	writeListSection(&b, "Removed paths", diff.Removed, "No removed endpoint paths detected.")
	writeListSection(&b, "Added schemas", diff.SchemasAdded, "No added schema names detected.")
	writeListSection(&b, "Removed schemas", diff.SchemasRemoved, "No removed schema names detected.")

	b.WriteString("## Triage checklist\n\n")
	b.WriteString("1. Read the upstream OpenAPI docs for each changed path or schema name and confirm whether it is relevant to icuvisor's PRD/roadmap. Treat schema-only drift as a human triage signal, not implementation approval.\n")
	b.WriteString("2. For relevant additions, create a Taskplane/backlog task with endpoint, method, response-shaping, safety, schema/model, and fixture requirements.\n")
	b.WriteString("3. For removals or incompatible changes, check whether existing icuvisor tools depend on the path or schema and open a regression task if needed.\n")
	b.WriteString("4. After triage, intentionally update the pinned baseline spec in `scripts/openapidiff/baseline/` so future reports only show new drift.\n")
	return b.String()
}

func writeListSection(b *strings.Builder, title string, items []string, empty string) {
	b.WriteString("## " + title + "\n\n")
	if len(items) == 0 {
		b.WriteString(empty + "\n\n")
		return
	}
	for _, item := range items {
		b.WriteString("- `" + item + "`\n")
	}
	b.WriteString("\n")
}

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

type specPaths struct {
	Paths map[string]json.RawMessage `json:"paths"`
}

type endpointDiff struct {
	Added   []string
	Removed []string
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
	baselinePaths, err := extractPaths(baseline)
	if err != nil {
		return endpointDiff{}, fmt.Errorf("baseline: %w", err)
	}
	latestPaths, err := extractPaths(latest)
	if err != nil {
		return endpointDiff{}, fmt.Errorf("latest: %w", err)
	}

	var diff endpointDiff
	for path := range latestPaths {
		if _, ok := baselinePaths[path]; !ok {
			diff.Added = append(diff.Added, path)
		}
	}
	for path := range baselinePaths {
		if _, ok := latestPaths[path]; !ok {
			diff.Removed = append(diff.Removed, path)
		}
	}
	sort.Strings(diff.Added)
	sort.Strings(diff.Removed)
	return diff, nil
}

func extractPaths(data []byte) (map[string]struct{}, error) {
	var spec specPaths
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&spec); err != nil {
		var loose map[string]json.RawMessage
		if looseErr := json.Unmarshal(data, &loose); looseErr != nil {
			return nil, fmt.Errorf("decoding JSON: %w", err)
		}
		pathsRaw, ok := loose["paths"]
		if !ok {
			return nil, fmt.Errorf("missing paths object")
		}
		if err := json.Unmarshal(pathsRaw, &spec.Paths); err != nil {
			return nil, fmt.Errorf("decoding paths object: %w", err)
		}
	}
	if len(spec.Paths) == 0 {
		return nil, fmt.Errorf("missing or empty paths object")
	}
	paths := make(map[string]struct{}, len(spec.Paths))
	for path := range spec.Paths {
		if strings.TrimSpace(path) == "" {
			continue
		}
		paths[path] = struct{}{}
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("paths object contains no non-empty paths")
	}
	return paths, nil
}

func renderMarkdown(diff endpointDiff, baselineSource, latestSource string) string {
	var b strings.Builder
	b.WriteString("# intervals.icu OpenAPI endpoint diff\n\n")
	b.WriteString("This report compares only OpenAPI `paths` keys. It is a human triage aid; it does not approve product scope or auto-generate icuvisor tools.\n\n")
	fmt.Fprintf(&b, "- Baseline: `%s`\n", baselineSource)
	fmt.Fprintf(&b, "- Latest: `%s`\n", latestSource)
	fmt.Fprintf(&b, "- Added paths: %d\n", len(diff.Added))
	fmt.Fprintf(&b, "- Removed paths: %d\n\n", len(diff.Removed))

	writePathSection(&b, "Added paths", diff.Added, "No added endpoint paths detected.")
	writePathSection(&b, "Removed paths", diff.Removed, "No removed endpoint paths detected.")

	b.WriteString("## Triage checklist\n\n")
	b.WriteString("1. Read the upstream OpenAPI docs for each changed path and confirm whether the endpoint is relevant to icuvisor's PRD/roadmap.\n")
	b.WriteString("2. For relevant additions, create a Taskplane/backlog task with endpoint, method, response-shaping, safety, and fixture requirements.\n")
	b.WriteString("3. For removals or incompatible changes, check whether existing icuvisor tools depend on the path and open a regression task if needed.\n")
	b.WriteString("4. After triage, intentionally update the pinned baseline spec in `scripts/openapidiff/baseline/` so future reports only show new drift.\n")
	return b.String()
}

func writePathSection(b *strings.Builder, title string, paths []string, empty string) {
	b.WriteString("## " + title + "\n\n")
	if len(paths) == 0 {
		b.WriteString(empty + "\n\n")
		return
	}
	for _, path := range paths {
		b.WriteString("- `" + path + "`\n")
	}
	b.WriteString("\n")
}

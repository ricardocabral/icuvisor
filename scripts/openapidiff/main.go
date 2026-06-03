package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	baselinePath := flag.String("baseline", "scripts/openapidiff/baseline/intervals-openapi.json", "pinned baseline OpenAPI JSON spec")
	latestPath := flag.String("latest", "", "latest OpenAPI JSON spec file; omit when using -latest-url")
	latestURL := flag.String("latest-url", "", "opt-in URL to fetch the latest OpenAPI JSON spec")
	outputPath := flag.String("output", "-", "Markdown output path, or - for stdout")
	flag.Parse()

	if err := run(*baselinePath, *latestPath, *latestURL, *outputPath); err != nil {
		fmt.Fprintf(os.Stderr, "openapi endpoint diff: %v\n", err)
		os.Exit(1)
	}
}

func run(baselinePath, latestPath, latestURL, outputPath string) error {
	baselinePath = strings.TrimSpace(baselinePath)
	latestPath = strings.TrimSpace(latestPath)
	latestURL = strings.TrimSpace(latestURL)
	outputPath = strings.TrimSpace(outputPath)

	if baselinePath == "" {
		return fmt.Errorf("-baseline is required")
	}
	if latestPath == "" && latestURL == "" {
		return fmt.Errorf("provide -latest for offline diffing or -latest-url for an opt-in live fetch")
	}
	if latestPath != "" && latestURL != "" {
		return fmt.Errorf("provide only one of -latest or -latest-url")
	}

	baseline, err := readSpec(baselinePath)
	if err != nil {
		return err
	}
	latestSource := latestPath
	var latest []byte
	if latestPath != "" {
		latest, err = readSpec(latestPath)
	} else {
		latestSource = latestURL
		latest, err = fetchSpec(latestURL)
	}
	if err != nil {
		return err
	}

	diff, err := diffSpecs(baseline, latest)
	if err != nil {
		return err
	}
	markdown := renderMarkdown(diff, baselinePath, latestSource)
	if outputPath == "" || outputPath == "-" {
		fmt.Print(markdown)
		return nil
	}
	if err := os.WriteFile(outputPath, []byte(markdown), 0o600); err != nil {
		return fmt.Errorf("writing output %s: %w", outputPath, err)
	}
	return nil
}

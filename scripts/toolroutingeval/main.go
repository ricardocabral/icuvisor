package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/ricardocabral/icuvisor/internal/toolrouting"
)

func main() {
	fixturePath := flag.String("fixture", "internal/toolrouting/testdata/cases.json", "routing eval fixture path")
	jsonOutput := flag.Bool("json", false, "write JSON summary to stdout")
	outputPath := flag.String("output", "", "write JSON summary/report to this file")
	diffPath := flag.String("diff", "", "compare the current run against a saved JSON summary")
	includeRaw := flag.Bool("include-raw", false, "include raw provider messages in JSON output; off by default for safe baselines")
	flag.Parse()

	fixture, err := toolrouting.LoadFixtureFile(*fixturePath)
	if err != nil {
		fatalf("loading fixture: %v", err)
	}
	provider, providerName, configured, err := toolrouting.EnvProvider(os.Getenv, &http.Client{Timeout: 30 * time.Second})
	if err != nil {
		fatalf("configuring provider: %v", err)
	}
	if !configured {
		fmt.Fprintf(os.Stderr, "provider not configured; validating fixture and catalog only. Set %s=anthropic and %s to run live.\n", toolrouting.EnvRoutingEvalProvider, toolrouting.EnvAnthropicAPIKey)
	} else {
		fmt.Fprintf(os.Stderr, "running routing eval with provider %s\n", providerName)
	}

	out := os.Stdout
	if *jsonOutput || *outputPath != "" {
		out = os.Stderr
	}
	summary, err := toolrouting.Run(context.Background(), fixture, provider, out)
	if err != nil {
		fatalf("running eval: %v", err)
	}
	safeSummary := summary
	if !*includeRaw {
		safeSummary = toolrouting.StripRawMessages(summary)
	}

	var report any = safeSummary
	if *diffPath != "" {
		baseline, err := loadSummary(*diffPath)
		if err != nil {
			fatalf("loading diff baseline: %v", err)
		}
		diff := toolrouting.DiffResults(baseline, safeSummary)
		writeDiffReport(out, diff)
		report = struct {
			Summary toolrouting.EvalSummary `json:"summary"`
			Diff    toolrouting.DiffSummary `json:"diff"`
		}{Summary: safeSummary, Diff: diff}
	}
	if *jsonOutput || *outputPath != "" {
		if err := writeJSONReport(report, *outputPath); err != nil {
			fatalf("writing JSON summary: %v", err)
		}
	}
	if summary.Failed > 0 {
		os.Exit(1)
	}
}

func loadSummary(path string) (toolrouting.EvalSummary, error) {
	file, err := os.Open(path)
	if err != nil {
		return toolrouting.EvalSummary{}, err
	}
	defer file.Close()
	var summary toolrouting.EvalSummary
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&summary); err == nil && summary.Total > 0 {
		return summary, nil
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return toolrouting.EvalSummary{}, err
	}
	var wrapped struct {
		Summary toolrouting.EvalSummary `json:"summary"`
	}
	if err := json.NewDecoder(file).Decode(&wrapped); err != nil {
		return toolrouting.EvalSummary{}, err
	}
	if wrapped.Summary.Total == 0 && len(wrapped.Summary.Results) == 0 {
		return toolrouting.EvalSummary{}, fmt.Errorf("%s does not contain a routing eval summary", path)
	}
	return wrapped.Summary, nil
}

func writeDiffReport(out *os.File, diff toolrouting.DiffSummary) {
	classes := make([]string, 0, len(diff.Counts))
	for class := range diff.Counts {
		classes = append(classes, class)
	}
	sort.Strings(classes)
	fmt.Fprintln(out, "diff:")
	for _, class := range classes {
		fmt.Fprintf(out, "  %s=%d\n", class, diff.Counts[class])
	}
	for _, entry := range diff.Entries {
		if entry.Classification == toolrouting.DiffStillPassing {
			continue
		}
		fmt.Fprintf(out, "  %s %s\n", entry.Classification, entry.CaseID)
	}
}

func writeJSONReport(report any, outputPath string) error {
	var out io.Writer = os.Stdout
	var file *os.File
	if outputPath != "" {
		created, err := os.Create(outputPath)
		if err != nil {
			return err
		}
		file = created
		out = created
	}
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	err := encoder.Encode(report)
	if closeErr := closeFile(file); err == nil {
		err = closeErr
	}
	return err
}

func closeFile(file *os.File) error {
	if file == nil {
		return nil
	}
	return file.Close()
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}

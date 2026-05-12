//go:build ignore

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/toolchecks"
)

func main() {
	threshold := flag.Float64("threshold", toolchecks.DefaultConfusableThreshold, "maximum allowed token-Jaccard similarity between first description sentences in a cluster")
	summaryFile := flag.String("summary-file", os.Getenv("GITHUB_STEP_SUMMARY"), "optional Markdown summary file; defaults to GITHUB_STEP_SUMMARY")
	flag.Parse()

	catalog, err := toolchecks.GenerateToolCatalog()
	if err != nil {
		fmt.Fprintf(os.Stderr, "check confusable names: %v\n", err)
		os.Exit(1)
	}
	report := toolchecks.CheckConfusableCatalog(catalog, *threshold)
	writeConfusableOutput(report, *threshold, *summaryFile)
	if !report.OK() {
		os.Exit(1)
	}
}

func writeConfusableOutput(report toolchecks.ConfusableReport, threshold float64, summaryFile string) {
	var summary strings.Builder
	summary.WriteString("## Tool-name confusability\n\n")
	if report.OK() {
		fmt.Printf("confusable-name check: PASS (threshold %.2f)\n", threshold)
		summary.WriteString(fmt.Sprintf("- PASS: no clustered first-sentence pairs reached token-Jaccard threshold %.2f.\n", threshold))
	} else {
		fmt.Fprintf(os.Stderr, "confusable-name check: FAIL (%d pair(s) at or above %.2f)\n", len(report.Pairs), threshold)
		summary.WriteString(fmt.Sprintf("- FAIL: %d pair(s) at or above token-Jaccard threshold %.2f.\n", len(report.Pairs), threshold))
	}
	for _, pair := range report.Pairs {
		message := fmt.Sprintf("cluster=%s tools=%s,%s score=%.2f; rewrite one first sentence to emphasize access pattern and payload shape", pair.Cluster, pair.ToolA, pair.ToolB, pair.Score)
		fmt.Fprintln(os.Stderr, message)
		fmt.Println("::error::" + escapeAnnotation(message))
		summary.WriteString(fmt.Sprintf("  - `%s` / `%s` (%.2f): rewrite one first sentence to emphasize access pattern and payload shape.\n", pair.ToolA, pair.ToolB, pair.Score))
		summary.WriteString(fmt.Sprintf("    - `%s`: %s\n", pair.ToolA, pair.SentenceA))
		summary.WriteString(fmt.Sprintf("    - `%s`: %s\n", pair.ToolB, pair.SentenceB))
	}
	if strings.TrimSpace(summaryFile) != "" {
		if err := os.WriteFile(summaryFile, []byte(summary.String()), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "writing confusable-name summary: %v\n", err)
		}
	}
}

func escapeAnnotation(value string) string {
	value = strings.ReplaceAll(value, "%", "%25")
	value = strings.ReplaceAll(value, "\r", "%0D")
	value = strings.ReplaceAll(value, "\n", "%0A")
	return value
}

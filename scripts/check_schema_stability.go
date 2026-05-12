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
	snapshotDir := flag.String("snapshot-dir", toolchecks.DefaultSchemaSnapshotDir, "current tree snapshot directory to compare with generated live registry schemas")
	baselineDir := flag.String("baseline-dir", "", "pre-PR baseline snapshot directory for additive-only stability checks")
	requireBaseline := flag.Bool("require-baseline", false, "fail when -baseline-dir is empty")
	summaryFile := flag.String("summary-file", os.Getenv("GITHUB_STEP_SUMMARY"), "optional Markdown summary file; defaults to GITHUB_STEP_SUMMARY")
	flag.Parse()

	generated, err := toolchecks.GenerateSchemaSnapshots()
	if err != nil {
		fatal(err)
	}
	var reports []namedReport
	freshness, err := toolchecks.CheckSnapshotFreshness(*snapshotDir, generated)
	if err != nil {
		fatal(err)
	}
	reports = append(reports, namedReport{Name: "snapshot freshness", Report: freshness})
	if strings.TrimSpace(*baselineDir) == "" {
		if *requireBaseline {
			fatal(fmt.Errorf("-baseline-dir is required for CI stability checks"))
		}
	} else {
		stability, err := toolchecks.CheckSchemaStability(*baselineDir, *snapshotDir, generated)
		if err != nil {
			fatal(err)
		}
		reports = append(reports, namedReport{Name: "additive-only stability", Report: stability})
	}

	failed := false
	for _, report := range reports {
		if !report.Report.OK() {
			failed = true
		}
	}
	writeOutput(reports, *summaryFile)
	if failed {
		os.Exit(1)
	}
}

type namedReport struct {
	Name   string
	Report toolchecks.SchemaReport
}

func writeOutput(reports []namedReport, summaryFile string) {
	var summary strings.Builder
	summary.WriteString("## Tool schema stability\n\n")
	for _, named := range reports {
		if named.Report.OK() {
			fmt.Printf("%s: PASS\n", named.Name)
			summary.WriteString(fmt.Sprintf("- %s: PASS\n", named.Name))
		} else {
			fmt.Fprintf(os.Stderr, "%s: FAIL (%d issue(s))\n", named.Name, len(named.Report.Failures))
			summary.WriteString(fmt.Sprintf("- %s: FAIL (%d issue(s))\n", named.Name, len(named.Report.Failures)))
		}
		for _, added := range named.Report.Added {
			fmt.Printf("%s: new tool accepted: %s\n", named.Name, added)
		}
		for _, failure := range named.Report.Failures {
			line := formatFailure(failure)
			fmt.Fprintln(os.Stderr, line)
			fmt.Println(githubAnnotation(failure, line))
			summary.WriteString(fmt.Sprintf("  - `%s`: %s\n", failure.ToolName, failure.Message))
		}
	}
	if strings.TrimSpace(summaryFile) != "" {
		if err := os.WriteFile(summaryFile, []byte(summary.String()), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "writing schema stability summary: %v\n", err)
		}
	}
}

func formatFailure(f toolchecks.SchemaFailure) string {
	parts := []string{fmt.Sprintf("tool=%s", f.ToolName), fmt.Sprintf("kind=%s", f.Kind)}
	if f.Property != "" {
		parts = append(parts, fmt.Sprintf("property=%s", f.Property))
	}
	if f.Baseline != "" {
		parts = append(parts, fmt.Sprintf("baseline=%s", f.Baseline))
	}
	if f.Current != "" {
		parts = append(parts, fmt.Sprintf("current=%s", f.Current))
	}
	parts = append(parts, f.Message)
	return strings.Join(parts, " ")
}

func githubAnnotation(f toolchecks.SchemaFailure, message string) string {
	file := f.Current
	if file == "" {
		file = f.Baseline
	}
	if file == "" {
		return "::error::" + escapeAnnotation(message)
	}
	return fmt.Sprintf("::error file=%s::%s", escapeAnnotation(file), escapeAnnotation(message))
}

func escapeAnnotation(value string) string {
	value = strings.ReplaceAll(value, "%", "%25")
	value = strings.ReplaceAll(value, "\r", "%0D")
	value = strings.ReplaceAll(value, "\n", "%0A")
	return value
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "check schema stability: %v\n", err)
	os.Exit(1)
}

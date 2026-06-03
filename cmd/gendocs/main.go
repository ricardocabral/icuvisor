package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ricardocabral/icuvisor/internal/tools"
)

const (
	defaultToolsOut   = "web/data/tools.json"
	defaultSchemasOut = "web/data/tool_schemas.json"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "gendocs: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	flags := flag.NewFlagSet("gendocs", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	out := flags.String("out", defaultToolsOut, "path to write generated tool catalog JSON")
	schemasOut := flags.String("schemas-out", defaultSchemasOut, "path to write generated per-tool schema JSON")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %v", flags.Args())
	}
	if err := writeJSONFile(*out, tools.Catalog(), "tool catalog"); err != nil {
		return err
	}
	return writeJSONFile(*schemasOut, tools.SchemaCatalog(), "tool schema catalog")
}

func writeJSONFile(out string, value any, label string) error {
	if out == "" {
		return fmt.Errorf("missing output path for %s", label)
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal %s: %w", label, err)
	}
	data = append(data, '\n')
	return writeFileAtomic(out, data)
}

func writeFileAtomic(out string, data []byte) error {
	dir := filepath.Dir(out)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(out)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Rename(tmpName, out); err != nil {
		return fmt.Errorf("replace output file: %w", err)
	}
	return nil
}

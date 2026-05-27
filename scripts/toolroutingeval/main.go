package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ricardocabral/icuvisor/internal/toolrouting"
)

func main() {
	fixturePath := flag.String("fixture", "internal/toolrouting/testdata/cases.json", "routing eval fixture path")
	jsonOutput := flag.Bool("json", false, "write JSON summary to stdout")
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
	if *jsonOutput {
		out = os.Stderr
	}
	summary, err := toolrouting.Run(context.Background(), fixture, provider, out)
	if err != nil {
		fatalf("running eval: %v", err)
	}
	if *jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(summary); err != nil {
			fatalf("writing JSON summary: %v", err)
		}
	}
	if summary.Failed > 0 {
		os.Exit(1)
	}
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}

//go:build ignore

package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/ricardocabral/icuvisor/internal/toolchecks"
)

func main() {
	dir := flag.String("dir", toolchecks.DefaultSchemaSnapshotDir, "directory where per-tool JSON schema snapshots are written")
	flag.Parse()
	if err := toolchecks.WriteGeneratedSchemaSnapshots(context.Background(), *dir); err != nil {
		fmt.Fprintf(os.Stderr, "snapshot tool schemas: %v\n", err)
		os.Exit(1)
	}
}

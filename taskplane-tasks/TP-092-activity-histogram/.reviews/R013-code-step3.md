# R013 code review — Step 3: Tool registration and activation hint

Verdict: REVISE

## Findings

### P1 — Generated tool catalog artifacts were not updated

`internal/tools/catalog.go:105` now registers `get_activity_histogram`, so `tools.Catalog()` emits the new descriptor, but the generated catalog artifacts are still stale. Both `cmd/gendocs/testdata/tools.golden.json` and `web/data/tools.json` contain no `get_activity_histogram` entry (they jump from `get_activity_details` to `get_activity_intervals`). This breaks the gendocs golden test and leaves the website/catalog data unable to discover the newly registered tool.

Evidence: `go test ./internal/toolcatalog ./internal/tools ./cmd/gendocs` fails in `cmd/gendocs` with `generated catalog differs from golden`; the generated output includes:

```json
{
  "name": "get_activity_histogram",
  "group": "activities",
  "tier": "full",
  "safety": "read",
  "summary": "Summarize a single activity's power, heart-rate, or pace distribution into terse time-in-bucket histogram rows.",
  "anchor": "get_activity_histogram"
}
```

Please run/update the generated docs catalog (`make docs-tools` or equivalent) and commit at least `cmd/gendocs/testdata/tools.golden.json` and `web/data/tools.json` with the new descriptor.

## Verification run

- `go test ./internal/toolcatalog ./internal/tools ./cmd/gendocs` — FAIL (`cmd/gendocs` golden mismatch due missing `get_activity_histogram` in committed generated catalog)
- `go test ./internal/mcp` — PASS

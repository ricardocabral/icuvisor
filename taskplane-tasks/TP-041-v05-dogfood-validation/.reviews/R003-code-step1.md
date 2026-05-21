# Code Review — Step 1 Diagnostics subcommand

Result: **REVISE**

## Findings

### 1. Diagnostics still emits unsanitized config paths through the default logger

- **Severity:** High
- **Files:** `internal/app/diagnostics.go:36-43`, via `internal/config/config.go:167` and `internal/config/config.go:190-193`

`runDiagnosticsCommand` calls the normal `config.Load` directly when no test loader is injected. `config.Load` writes informational log records with raw config and env-file paths using `slog.Default()` (`config file loaded path=...`, `env file loaded/not found path=...`). In the real CLI, the default slog handler writes to stderr, so `icuvisor diagnostics` can print paths outside `opts.Stdout` and outside the redaction/sanitized output model.

This violates the Step 1 requirements/plan that diagnostics output is routed through `opts.Stdout` and that diagnostics prints only source labels, not config paths. It also undermines the no-secret/no-athlete-ID guarantee: if the config or env-file path contains an athlete ID or token-shaped string, diagnostics leaks it on stderr even though the structured stdout output is redacted.

Reproduction:

```bash
tmpdir=$(mktemp -d /tmp/i7777777diag.XXXX)
cat > "$tmpdir/config.json" <<'JSON'
{"athlete_id":"7777777","timezone":"UTC"}
JSON
INTERVALS_ICU_API_KEY=sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx \
INTERVALS_ICU_ATHLETE_ID=7777777 \
  go run ./cmd/icuvisor diagnostics --config "$tmpdir/config.json" >/tmp/diag.out 2>/tmp/diag.err
cat /tmp/diag.err
```

Observed stderr includes:

```text
INFO config file loaded path=/tmp/i7777777diag.../config.json
INFO env file not found path=.env
```

Please suppress or redirect config-loader logging for diagnostics, or make the loader logging redacted/diagnostics-aware. Add a test using the real loader (not an injected loader) that captures stderr/log output and asserts the fixture API key, raw/normalized athlete ID, token-shaped strings, and config/env-file paths are absent.

## Verification performed

- `git diff ee86173..HEAD --name-only`
- `git diff ee86173..HEAD`
- `go test ./internal/app ./internal/diagnostics ./internal/mcp`
- `go test ./...`

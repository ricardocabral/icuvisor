# Code Review R003 — Step 1: Resource registration plumbing

**Verdict: REVISE**

I reviewed the diff from `384b8eb7b958fb4fe4f6694e90917f2dc84a09b0..HEAD`, read the changed MCP/resource files, and ran:

- `go test ./internal/mcp ./internal/resources` — passes
- `make fmt-check` — passes
- `golangci-lint run ./internal/mcp ./internal/resources` — fails

## Findings

### 1. Lint failure in resource error construction blocks `make lint`

- **File:** `internal/mcp/server.go:303`
- **Severity:** Blocking

`golangci-lint` reports:

```text
internal/mcp/server.go:303:16: SA1006: printf-style function with dynamic format string and no further arguments should use print-style function instead (staticcheck)
            return nil, fmt.Errorf(genericResourceErrorMessage)
                        ^
```

This means the step does not satisfy the project verification requirements, and CI/`make lint` will fail. Since the error message is a constant with no formatting or wrapping, use `errors.New(genericResourceErrorMessage)` instead.

```go
return nil, errors.New(genericResourceErrorMessage)
```

`errors` is already imported in this file.

## Notes

The overall plumbing shape looks consistent with the approved plan: resources are registered through the SDK `AddResource` API, exposed through a small `internal/resources` registry boundary, validated at server construction, and protocol tests cover initialize/list/read, duplicate/invalid registration, unknown reads, and sanitized handler failures.

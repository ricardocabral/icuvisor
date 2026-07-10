# Code Review — TP-228 Step 4

## Verdict: APPROVE

The regression changes match the approved plan. The update contract now checks the complete raw query for both recalculation modes, preventing unsupported extra parameters from slipping through. The no-implicit-apply test rejects every non-update path and verifies exactly one update request. Existing direct apply coverage still requires the exact PUT/path, empty query, zero-byte body, and absent content type, while the tool test retains exact terse-error and zero-client-call assertions for legacy, unknown, and null arguments.

`go test ./internal/intervals ./internal/tools` passes.

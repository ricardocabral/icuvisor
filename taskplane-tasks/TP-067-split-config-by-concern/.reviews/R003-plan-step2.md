# Plan Review — TP-067 Step 2 (Mechanical split)

## Verdict: Approved

The Step 2 plan is appropriate for a low-risk mechanical split. It follows the Step 1 declaration inventory, keeps exported names/signatures stable, preserves `Load` as the public entry point in `config.go`, and sequences the extraction by concern with targeted tests after each move.

## What looks good

- The extraction order is sensible: small standalone helpers first (`path.go`, `athlete.go`, `redaction.go`, `httpbind.go`, `dotenv.go`, `write.go`), then the larger load/validation moves.
- The plan explicitly keeps `Load` as the public `config.go` wrapper while moving composition into `load.go`, matching the acceptance criterion.
- The validation step calls out preserving call order and error strings. This is the main behavioral risk, and the plan has the right guardrail.
- The plan preserves exact exported APIs for athlete normalization, HTTP bind validation, redaction/logging, write helpers, and default path resolution.
- Running targeted config tests after every extraction is sufficient for Step 2, with full `make` verification deferred to Step 3 as requested.

## Implementation guardrails

- Keep the commits truly mechanical: one extraction file per commit, with import cleanup only. Avoid mixing test reshuffles or opportunistic behavior fixes into the source-move commits.
- If tests are split during Step 2, make those moves mechanical too; do not rewrite assertions unless required by file movement.
- Preserve the current `.env` allow-list exactly when moving `recognizedEnvKey`; do not add `ICUVISOR_DEBUG_METADATA`, `ICUVISOR_ENV_FILE`, or other keys as part of this refactor.
- After moving `load.go` and `validate.go`, confirm `config.go` contains only constants/types/defaults, `Transport.String`, `Config` semantics helpers, and the thin public `Load` wrapper.
- Use targeted runs such as `go test ./internal/config` after each extraction; Step 3 can run the full `make build`, `make test`, `make test-race`, and `make lint` suite.

## Recommendation

Proceed with Step 2. The plan is sufficiently detailed for implementation; the main review focus should be verifying moved-code diffs and unchanged config/load/write behavior.

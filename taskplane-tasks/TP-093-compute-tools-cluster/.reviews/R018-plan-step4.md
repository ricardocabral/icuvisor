# R018 plan review — Step 4: Tests and verification

Verdict: APPROVE

The revised Step 4 plan addresses the blocking gaps from R017. It now names the contract-critical test areas by tool, includes truncation/partial-status precedence across the activity/event-backed paths, calls out source-priority and no-stream guarantees, broadens baseline status/calculation coverage, and covers deterministic compliance pairing/caution semantics. It also names a targeted `go test` command for the affected packages and explicitly defers the full `make test`/build/lint gate to Step 5, which is consistent with the task's separate verification step.

## Non-blocking suggestions

- When implementing the zone/load tests, make the fake clients fail on any raw-stream call or unexpected activity enumeration in the summary-preferred path. That will lock in the "do not fetch rows/streams and reduce manually" activation contract, not just the returned payload.
- Ensure at least one load-balance golden asserts `_meta.formula_ref`, polarization classification/state, and the moderate/high-zero undefined cases if those are not already covered by existing tests.
- For the docs freshness check, record the exact command or comparison used in `STATUS.md` if generated tool docs are touched again.

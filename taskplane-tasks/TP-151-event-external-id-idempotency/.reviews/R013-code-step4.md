# Code Review R013 — TP-151 Step 4

**Verdict: APPROVE**

No blocking findings.

The Step 4 changes refresh the `add_or_update_event` schema snapshot with the new optional `external_id` field, add an appropriate tool-routing fixture for explicit retry/idempotency language, and document the public contract/caveats in the cookbook pages and `[Unreleased]` changelog. The docs stay conservative about best-effort idempotency, no blank/clear semantics, upstream audit visibility, stable namespaces, and `apply_training_plan` same-day/upstream caveats.

Validation performed:

- `go run ./scripts/snapshot_tool_schemas.go` — no additional diff
- `go test ./internal/toolrouting`
- `cd web && hugo --environment development --quiet`
- `make docs-tools` — no additional diff

# Review: Step 4 plan

Verdict: **approved**.

The Step 4 plan is appropriate for the current tree. Step 1 already found that the `update_wellness` error contract is no longer only in the PRD: the implementation, PRD, generated catalog data, and website tool reference path all expose the same read-only-field literals. Since the README has already been slimmed to point users to the website catalog, the plan is also right not to reintroduce a one-off per-tool README bullet just to satisfy stale pre-TP-052 wording.

Verification context I checked:

- `internal/tools/update_wellness.go:18` includes the MCP description with both error literals.
- `internal/tools/update_wellness.go:183` treats those literals as validation errors, and `:203-206` returns `field_not_writable: sleepScore (device-managed)` / `field_not_writable: _native (bridge-managed)`.
- `docs/prd/PRD-icuvisor.md:252` matches those literals.
- `web/data/tools.json:259-264` includes the same summary for `update_wellness`.
- `web/content/reference/tools.md:8` renders the generated catalog, and `web/layouts/partials/tool-catalog.html` prints each tool `summary`, so `/reference/tools/#update_wellness` will surface the contract.
- `README.md:21` delegates users to the website tool catalog rather than maintaining a per-tool list.

What is strong:

- The plan checks code as the source of truth before claiming documentation correctness.
- It verifies the actual user-facing website surface, not just the implementation file.
- It explicitly recognizes that the old README bullet requirement was an interim path and is superseded by the current README/website split.

Execution notes, not blockers:

- In the Step 4 `Resolution:` section, record this as a no-op content step and include the key grep evidence for code, PRD, `web/data/tools.json`, and README delegation.
- Mention that no TP-051 amendment is needed because the existing generated catalog `summary` already renders the error contract; an `errors` field would only be needed if the summary stopped carrying it or the website did not display it.
- When Step 5 updates/validates changelog wording, avoid saying the contract was surfaced in a README bullet unless a README edit is actually made. Prefer wording that reflects the current surface, e.g. website/generated tool reference and PRD/code alignment.

No plan changes are required before executing Step 4.

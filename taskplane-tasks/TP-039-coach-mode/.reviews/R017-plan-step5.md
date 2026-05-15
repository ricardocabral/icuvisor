# R017 plan review — Step 5: Catalog-cache caveat + Tests

Verdict: **REVISE**

The Step 5 outline names the right deliverables, but it is not yet specific enough to protect the Step 3/4 security invariants. In particular, the plan must pin the catalog-cache behavior to the *actual visible catalog* after delete-mode, toolset-tier, and coach ACL composition. A test-only step that merely adds broad coverage can miss mismatches between `tools/list`, `select_athlete.allowed_tools`, `icuvisor_list_advanced_capabilities`, and call-time enforcement.

## Required revisions

1. **Make the cache-caveat implementation use the authoritative visibility calculation.**
   `select_athlete._meta.requires_new_conversation` and `allowed_tools` must be derived from the same post-gate visible catalog used for protocol `tools/list`: delete-mode → toolset-tier → coach ACL, plus the always-visible coach/control tools. Do not compute it from the coach evaluator alone, because that can list tools hidden by delete-mode/toolset or compute the wrong `requires_new_conversation` value. If a helper does not already exist at the right layer, Step 5 should introduce one or move the computation to MCP-side code rather than duplicating visibility logic in `internal/tools`.

2. **Document the caveat in `docs/coach-mode.md` now, even if Step 6 later expands the doc.**
   The Step 5 documentation should explicitly say:
   - MCP clients may cache the tool catalog for the current conversation;
   - `select_athlete` changes server-side routing immediately, but the model/client may not see a refreshed catalog until a new conversation/reconnect;
   - when `_meta.requires_new_conversation: true`, the user should start a new conversation/reconnect to avoid stale visible tools;
   - TP-040 notifications are the intended future improvement.
   Update `STATUS.md` with the caveat wording/location.

3. **Define the composition truth table as protocol-visible behavior, not only evaluator unit tests.**
   The table should cover representative tools across all three gates:
   - `delete_event`: requires delete-mode full, full toolset, and coach ACL allow;
   - `add_or_update_event`: requires write-enabled delete-mode (`safe` or `full`), appropriate toolset visibility, and coach ACL allow;
   - `get_athlete_profile`: read/core tool where only the coach ACL should veto.

   For each case, assert both catalog exposure (`tools/list`) and call-time behavior where the tool is registered but denied for the selected/per-call target. The property must be explicit: **any single gate's deny vetoes the tool**.

4. **Add regression coverage for `select_athlete` metadata under delete/toolset filtering.**
   Include cases where two athletes differ only on a tool hidden by the current toolset or delete-mode. In those cases `requires_new_conversation` should remain `false` if the actual visible catalog is unchanged, and `allowed_tools` must not include hidden full/delete tools. This is the main risk area for Step 5.

5. **End-to-end test must route real athlete-scoped requests through a fake intervals client.**
   Use two configured athletes, one full-access and one read-only, and assert the upstream request path/target for default selection, after `select_athlete`, and with a per-call `athlete_id` override. The fake client/server should also prove a read-only athlete cannot invoke write/delete tools and receives the existing enumeration-safe public target error, without leaking roster membership or credentials.

6. **Keep test assertions structured and race-safe.**
   Parse tool JSON responses rather than relying only on substring checks for `_meta`. Include a concurrent/two-session assertion if not already covered by Step 4 tests, and make the relevant selection-store/catalog tests safe under `go test -race`.

Once the plan includes these specifics, Step 5 should be safe to implement.

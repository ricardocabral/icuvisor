# R018 plan review — Step 5: Catalog-cache caveat + Tests

Verdict: **REVISE**

The updated Step 5 plan incorporates the main R017 direction, but it is still too high-level for the highest-risk parts of this step. In particular, it says to use the authoritative visible catalog, but does not pin down the regression cases that would catch the current class of bug: `select_athlete` metadata can be computed from coach ACL/toolcatalog names while `tools/list` is constrained by delete-mode and toolset registration gates.

## Required revisions

1. **Specify the authoritative visible-catalog helper and layer.**
   The plan should state that `select_athlete.allowed_tools` and `_meta.requires_new_conversation` are computed in `internal/mcp` from the same post-registration/post-filter set used by protocol `tools/list` for the active session athlete. Do not leave this as a generic “authoritative catalog” phrase, and do not keep using `internal/tools.visibleToolsForAthlete` over `toolcatalog.AthleteScopedToolNames()` because that cannot know delete-mode/toolset filtering. A good plan would name a helper like `visibleToolNamesForAthlete(athleteID)` on `safeRegistrar` (or equivalent) and inject/wrap the `select_athlete` handler there.

2. **Add explicit `select_athlete` metadata regressions under hidden-gate differences.**
   The plan must include cases where two athletes differ only on tools hidden by non-coach gates:
   - differ only on `delete_event` while delete-mode is not `full` → `requires_new_conversation` must be `false`, and `allowed_tools` must not include `delete_event`;
   - differ only on a full-toolset tool while `ICUVISOR_TOOLSET=core` → `requires_new_conversation` must be `false`, and `allowed_tools` must not include that full-only tool;
   - differ on a visible core read tool → `requires_new_conversation` must be `true`.

3. **Define exact assertions for protocol truth-table tests.**
   “Catalog exposure and call-time vetoes” is the right goal, but the plan should require exact assertions for `tools/list`, `select_athlete.allowed_tools`, and `icuvisor_list_advanced_capabilities` all agreeing with the same visible catalog. For call-time behavior, include both: hidden/unregistered by delete-mode/toolset yields the normal unknown-tool protocol path, and registered-for-some-athlete but denied for the selected/per-call athlete yields the enumeration-safe public target error.

4. **Require structured response parsing and exact tool-set comparisons.**
   Current tests already have substring checks around `requires_new_conversation`; Step 5 should explicitly replace/add structured JSON parsing for `select_athlete`, `list_athletes`, and any `_meta` assertions. Compare sorted tool-name sets exactly enough to catch hidden full/delete tools leaking into `allowed_tools`.

5. **Make the end-to-end fake-client scenario concrete.**
   The plan should name the routing checks: default athlete request routes to `i111`, after `select_athlete` routes to `i222`, and a per-call `athlete_id` override routes to the override without changing the session default. For read-only athlete write/delete denial, assert the fake intervals client is **not** called and the returned error is the existing enumeration-safe target error with no roster/credential details.

6. **Cover session isolation/race safety or explicitly cite existing coverage.**
   R017 asked for a concurrent/two-session assertion if not already covered. The revised plan does not address this. Either add a two-session test where selecting athlete B in one session does not change another session’s default, or explicitly point to an existing Step 4 test that already proves it. Keep these tests safe under `go test -race`.

Once these points are added, the Step 5 plan should be implementable without weakening the Step 3/4 security invariants.

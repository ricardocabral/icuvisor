# R010 plan review — Step 3: Tool registry plumbing

Verdict: **APPROVE**

The revised Step 3 plan in `STATUS.md` addresses the blocking issues from R009 and is now specific enough to implement safely. In particular, it now pins the coach ACL gate to `internal/mcp.safeRegistrar`, covers `registeredTools`/catalog-hash/skip-count behavior, calls out the `icuvisor_list_advanced_capabilities` leak path, requires context-scoped routing rather than shared-client mutation, binds activity pagination tokens to the resolved athlete, and adds drift tests for the central `athlete_id` wrapper.

Implementation notes to preserve during coding:

- Keep the gate order exactly as planned: delete-mode/capability, then toolset, then coach ACL; any deny must veto registration and be reflected in the exposed catalog hash.
- Use one enumeration-safe public target error for malformed IDs, unknown coach-roster IDs, and non-coach mismatches. Log details without raw athlete IDs.
- Make routing per-call/per-context or via immutable shallow client copies only. Do not mutate `intervals.Client.athleteID` on the shared client.
- Ensure `icuvisor_list_advanced_capabilities` is based on a coach-filtered view, not the pre-filter collector catalog, so coach-denied full-only tools are not discoverable.
- Treat resources deliberately: either route/gate `icuvisor://athlete-profile` consistently in coach mode or record the explicit deferral/rationale in `STATUS.md` and docs before closing the step.
- Add the planned drift tests over `toolcatalog.AthleteScopedToolNames()` so future tools cannot miss the `athlete_id` schema/wrapper path silently.

No further plan revisions are required before implementation.

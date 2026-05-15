# R009 plan review — Step 3: Tool registry plumbing

Verdict: **REVISE**

The Step 3 outline has the right high-level shape (coach ACL as a third gate, uniform `athlete_id`, strict-decoder stripping, and request routing), but it needs a few concrete design constraints before implementation. Without them, this step can accidentally leak coach-denied tools or route calls to the wrong athlete while still appearing to satisfy the checklist.

## Required revisions

1. **Place the coach ACL gate where it affects the actual exposed catalog, not only `internal/tools`.**
   Today delete-mode and toolset filtering happen in `internal/mcp.safeRegistrar`, while `internal/tools.defaultRegistry` still constructs/collects tools. The Step 3 plan should explicitly state where the three gates compose and how the resulting `registeredTools`, catalog hash, and skip counts are computed. A coach-denied tool must not be present in the SDK catalog or the hashed exposed catalog.

2. **Do not leak coach-denied tools through `icuvisor_list_advanced_capabilities`.**
   `list_advanced_capabilities` is built from the collector catalog before the MCP registrar skips tools for toolset/capability. That behavior is intentional for TP-030, but it is not safe for coach ACLs: the task says the LLM never sees disallowed tools for the active athlete. The plan must require `list_advanced_capabilities` to exclude coach-denied tools as well, or otherwise register a coach-filtered catalog source for that tool.

3. **Specify request routing without mutating the shared intervals client.**
   All intervals methods currently use `Client.athleteID`. The plan should require a concurrency-safe mechanism such as a shallow per-call/per-athlete client copy or an explicit request-scoped athlete resolver used by the client. Do not update `client.athleteID` in place; Streamable HTTP can have concurrent calls and that would race/cross-route athletes.

4. **Bind pagination tokens to the resolved athlete.**
   `get_activities` has an opaque `next_page_token`. Once `athlete_id` can be selected or overridden, the token must carry the resolved canonical athlete ID, and token validation must reject use with a different selected/supplied athlete. Otherwise page 2 can silently fetch against another athlete after `select_athlete` or a per-call override.

5. **Make the public `athlete_id` rejection message enumeration-safe.**
   The acceptance criteria require wrong format and not-in-roster/mismatch cases to be indistinguishable to the LLM. Add an explicit public error string and test it for: malformed ID, unknown coach-roster ID, and non-coach mismatch. Internal logs should not include raw athlete identifiers.

6. **Add schema/wrapper drift tests for every athlete-scoped tool.**
   Because every existing tool has strict decoding and hand-written schemas, the plan should prefer a central wrapper that:
   - adds the optional `athlete_id` schema property with the required description,
   - removes `athlete_id` before calling the tool's existing strict decoder,
   - resolves and stores the target for intervals routing.

   Add a test over `toolcatalog.AthleteScopedToolNames()` proving every registered athlete-scoped tool accepts/advertises `athlete_id`, and non-athlete tools do not accidentally get it.

7. **Address resource bypass explicitly.**
   `resources.AthleteProfileResource` can make an intervals.icu call independently of the tool ACL. The Step 3 plan should either gate/route athlete-scoped resources consistently in coach mode or explicitly defer with a documented security rationale. As written, a coach-denied `get_athlete_profile` tool could still be bypassed through the profile resource.

## Suggested acceptance tests for Step 3

- Coach mode on, default athlete allows `get_*` but denies a full-only tool: denied tool is absent from tools/list, absent from catalog hash inputs, and absent from `icuvisor_list_advanced_capabilities` output.
- Delete-mode deny, toolset deny, and coach ACL deny each independently veto a representative tool; skip counters/log metadata distinguish the gates.
- A call with `athlete_id: "67890"` reaches `/athlete/i67890/...`; the same call without `athlete_id` reaches the selected/default athlete; invalid/missing roster cases return the same public message.
- Concurrent calls for two athletes use the correct upstream paths under `go test -race`.
- `get_activities` continuation token created for athlete A is rejected when used with athlete B.

Once these details are added, the plan should be safe to implement.

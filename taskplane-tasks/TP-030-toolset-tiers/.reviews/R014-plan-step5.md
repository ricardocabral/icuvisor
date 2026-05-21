# R014 plan review — Step 5: `_meta` surfacing + docs

Verdict: **REVISE**

I reviewed `PROMPT.md`, `STATUS.md`, the completed Step 1-4 notes, and the existing response/meta plumbing. The Step 5 section currently only restates the acceptance checklist; it does not yet pin the implementation approach or tests. Because `_meta.toolset` is an API-surface change that must apply consistently across tool responses, the plan needs more detail before coding.

## What looks good

- The scope is correct for Step 5: response metadata, README documentation, and the `[Unreleased]` changelog entry.
- The intended chokepoint is also correct. TP-018 surfaced `_meta.delete_mode` through `internal/response` (`SetDeleteMode`, `DeleteMode`, and `addCommonMeta`), and Step 5 should extend that path rather than adding per-tool ad hoc metadata.
- Startup already resolves `toolset` in `internal/app/defaultStartServer` and passes it into MCP/tool registry options, so there is a clear source of truth to reuse.

## Required plan adjustments

1. **Pin the response metadata mechanism.**
   - Add an explicit plan to extend `internal/response` with process-level toolset state parallel to delete mode, e.g. `SetToolset(string|safety.Toolset)` and `Toolset() string`, defaulting to `core` via `safety.ParseToolset`.
   - `addCommonMeta` should add `meta["toolset"] = Toolset()` next to `server_version` and `delete_mode`, preserving existing caller `_meta` keys and stripping any caller-supplied/stale owned key the same way response-owned metadata is handled.
   - `internal/app/defaultStartServer` should call the new setter using the already-resolved startup `toolset`; do not re-read `ICUVISOR_TOOLSET` from handlers or from `response`.

2. **Define the boundary for tools that do not currently use `response.Shape`.**
   - Most tool responses flow through `response.Shape`/`encodeShaped`, but `icuvisor_list_advanced_capabilities` currently returns a typed response directly with its own `_meta` and JSON text.
   - The plan must say how `_meta.toolset` is added there so the “every response” acceptance criterion is true and text content stays consistent with `StructuredContent`.
   - Prefer either routing this tool through the shared response shaper or adding a tiny shared helper that injects response-owned common meta before both structured and text serialization. Do not hand-maintain a different toolset source in the tool handler.

3. **Specify overwrite/merge semantics for `_meta.toolset`.**
   - The response layer owns `server_version`, `delete_mode`, `toolset`, and units metadata. The plan should state that caller-supplied `_meta.toolset` is overwritten/normalized, not preserved.
   - Keep existing metadata such as counts, pagination tokens, scales, and `delete_mode_note` intact.

4. **Add tests that prove global and per-response behavior.**
   - Extend `internal/response/shaper_test.go` with table-driven coverage for default `core`, explicit `full`, invalid/empty fallback to `core`, and preservation of existing `_meta` keys while overwriting stale `toolset`.
   - Update existing expected JSON in response shaper tests that currently assert exact `_meta` maps, because they will now include `toolset` everywhere common meta is added.
   - Add or update an app/startup propagation test showing the resolved `ServerInfo.Toolset` calls the response setter, analogous to the Step 1 delete-mode propagation coverage.
   - Add `icuvisor_list_advanced_capabilities` handler/protocol coverage asserting `_meta.toolset` is present in both core and full modes and that the serialized text JSON matches structured content.

5. **Plan docs precisely.**
   - README should document `ICUVISOR_TOOLSET` near the existing delete/write safety mode section: `core` is the default, `full` enables the full catalog, invalid/empty values fall back to `core`, changes require restarting icuvisor, and `icuvisor_list_advanced_capabilities` remains available in core to explain hidden tools.
   - The README should also note that `ICUVISOR_TOOLSET=full` is orthogonal to `ICUVISOR_DELETE_MODE`; destructive tools still require the delete-mode gate.
   - Add a concise `[Unreleased]` changelog bullet mentioning the toolset tier env var/default core catalog, discoverability tool, and `_meta.toolset` response metadata.

6. **Record the Step 5 plan in `STATUS.md`.**
   - As with Steps 2-4, add a dedicated “Step 5 plan” note before implementation with the concrete files, metadata semantics, and tests above.
   - Keep Step 6 verification (`make test`, `make build`, `make lint`, race/manual checks) out of this step except for normal targeted tests while implementing.

Once these details are recorded, Step 5 should be low-risk: a small `internal/response` extension, one startup setter call, targeted coverage for direct-return tools, and the docs/changelog updates.

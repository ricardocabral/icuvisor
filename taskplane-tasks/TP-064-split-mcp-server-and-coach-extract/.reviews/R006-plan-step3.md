# R006 Plan Review — Step 3: Lift coach concerns to `internal/coach`

## Decision: REVISE

The proposed seam is in the right direction: keep `internal/coach` independent of `internal/tools`, expose coach-owned visibility decisions over tool names/metadata, and move the advanced-capabilities rendering out of `internal/mcp`. However, the plan is not yet precise enough to protect the coach ACL contract during the lift.

## Required revisions

1. **Cover request-time target authorization, not only catalog visibility.**
   The current Step 3 sketch moves `visibleForAthlete`/`visibleToolNamesForAthlete`/`AllowedForAny`, but leaves the main request-time coach enforcement in `mcp.resolveToolTarget` / `resolveAthleteID` (`HasAthlete`, selected/default athlete fallback, and `MustEvaluate`). That is still a coach concern and is explicitly part of the threat model: normalize/roster-check/ACL-check before any tool handler or upstream request. Extend the seam plan with either:
   - a coach-owned target resolver/authorizer used by `mcp` after raw JSON `athlete_id` stripping, or
   - an explicit justification for the minimal part that must remain in `mcp` as SDK/schema adaptation.

   Be careful with imports: `internal/config` already imports `internal/coach`, so `coach` cannot call `config.NormalizeAthleteID` directly without a cycle. If the resolver moves to `coach`, inject a normalizer callback or keep only normalization in `mcp` while moving roster/default/ACL decisions behind a coach API.

2. **Specify exact visibility semantics for non-athlete-scoped tools.**
   The sketch says `VisibleForAthlete` always allows `list_athletes`, `select_athlete`, and `icuvisor_list_advanced_capabilities`, then delegates athlete-scoped decisions to `Evaluator`. It must also preserve the current behavior that every non-athlete-scoped tool is allowed by coach ACL (`Evaluator.Evaluate` returns allowed when `!toolcatalog.IsAthleteScopedTool`). Otherwise future non-athlete-scoped tools could be hidden accidentally.

3. **Pin the advanced-capabilities catalog source and gate order.**
   The moved handler must use the same effective catalog as today: capability-allowed and coach-allowed-for-any tools collected before the toolset gate, excluding `icuvisor_list_advanced_capabilities`, then filtered per selected athlete and full-only at call time. Do not feed it `registeredTools` in core mode, because that would drop full-only tools and make `icuvisor_list_advanced_capabilities` useless.

   The plan should state the preserved composition explicitly: capability gate, coach registration/any-athlete gate, toolset gate, then per-selected-athlete visibility for `tools/list` and advanced-capabilities rows. Any deny remains final.

4. **Make the `tools` helper contract byte-compatible.**
   Moving `coachFilteredAdvancedCapabilitiesHandler` into `internal/tools` is a good way to avoid an import cycle, but the helper must reuse the existing `list_advanced_capabilities` response shaping/status text/argument rejection/row sorting/requirement formatting. The current duplicate in `mcp` is close but not identical to the `tools` implementation (`firstSentence` differs). The plan should require tests that compare the filtered helper behavior to existing `TestListAdvancedCapabilities*` expectations.

5. **Add concrete Step 3 tests to the plan.**
   In addition to rerunning the existing protocol gate, specify new/updated tests for:
   - `coach.ToolFilter` visibility: always-visible coach/catalog tools, non-athlete-scoped allowed, per-athlete allow/deny, unknown athlete deny for athlete-scoped tools, disabled evaluator behavior.
   - registration-time `AllowedForAny` behavior for a tool allowed by one athlete and denied for all athletes.
   - request-time target authorization remains enumeration-safe and still rejects malformed/out-of-roster/ACL-denied overrides with the same public message.
   - filtered advanced capabilities preserve invalid-argument behavior, full/core status text, sorting, `_meta.count`, and selected-athlete filtering.

## Notes

- Keep `internal/coach` free of `internal/tools` imports; the proposed `tools.NewFilteredAdvancedCapabilitiesHandler(... include func(context.Context, tools.Tool) bool)` shape is acceptable if the catalog source and output compatibility are nailed down.
- If `ToolFilter` or a target resolver is exported, add proper Go doc notes and update the Step 3 documentation note in `docs/coach-mode.md` as the prompt requests.
- `mcp` may still need to own SDK session adaptation, raw JSON mutation, `schemaWithAthleteID`, and `intervals.WithTargetAthleteID`; those are transport/schema boundaries rather than coach ACL policy.

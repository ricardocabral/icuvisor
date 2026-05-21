# Plan Review — TP-007 Step 6

Verdict: **Not approved yet**. The Step 6 section in `STATUS.md` only repeats the prompt checklist, so there is not enough implementation detail to review. This step is the integration point between the existing `get_athlete_profile` contract and the new `internal/response` chokepoint; it needs a concrete plan before coding.

## Blocking findings

1. **No concrete Step 6 implementation plan is recorded.**
   - `STATUS.md` marks Step 6 in progress, but the section only lists the four prompt bullets.
   - A reviewable plan should name the exact handler/DTO/test changes, not just state “replace ad-hoc shaping”.

2. **Metadata ownership needs to be explicitly resolved.**
   - Earlier steps made `internal/response` own `_meta.server_version` and `_meta.units`; Step 6 must state that `get_athlete_profile` will pass these through `response.Options` and not maintain a competing source of truth in tool-local metadata.
   - If tool-specific profile metadata remains under `_meta` (`athlete_id_format`, `timezone_convention`, `pace_convention`, `include_full`), the plan should say these are preserved while response-owned keys are overwritten by the shaper.

3. **`include_full` semantics for this tool are underspecified.**
   - The plan needs to define the exact default terse shape and the full-mode delta: which non-secret fields are added, whether null-valued profile fields should be visible in full mode, and how this interacts with existing `omitempty` tags.
   - If `get_athlete_profile` intentionally does not expose raw upstream nulls, record that decision and why it is compatible with the Step 2 include-full convention; otherwise plan the DTO/map changes needed so `response.Shape(... IncludeFull: true)` can preserve nulls.

4. **The response chokepoint and structured/text output contract are not specified.**
   - Step 6 should state that the handler builds the profile DTO, calls `response.Shape` once with `IncludeFull`, `ServerVersion`, `DebugMetadata`, `QueryType`, and the profile-derived `UnitSystem`, then uses the shaped value for both `StructuredContent` and the JSON text content.
   - This avoids the old pattern where the typed response was separately marshaled and could diverge from shaped structured content.

5. **Unit-source consistency needs to be part of the plan.**
   - The plan should say how `preferred_units` / measurement fallbacks determine both visible `units.measurement_preference` and response-level `_meta.units`, and that both default consistently when the profile is empty or unknown.
   - Tests should inspect `_meta.units` on the shaped map/JSON, not only the typed `GetAthleteProfileResponse`, since the shaper injects that metadata.

6. **Focused integration tests are not planned.**
   - Step 7 has broad convention tests, but Step 6 needs tool-level contract tests before merge.
   - Required Step 6 test cases should include: default `{}` arguments remain terse; `include_full: true` adds only the intended full-only, non-secret fields; unknown arguments such as `athlete_id`/`api_key` are rejected; response text and structured content match; `_meta.server_version` and `_meta.units` are present in both terse and full modes; debug metadata is gated by the registry option; and preferred-unit fallback behavior keeps visible units consistent with `_meta.units`.

## Required additions before approval

Please update `STATUS.md` with a Step 6 plan covering at least:

- The exact changes in `internal/tools/get_athlete_profile.go`, including which existing ad-hoc metadata/version/unit fields are removed, retained, or moved to `response.Options`.
- The final JSON shape for terse and `include_full: true` responses, including `_meta` ownership/collision rules.
- The exact `response.Shape` options the handler will pass and confirmation that shaped output feeds both MCP text and structured content.
- The `include_full` null/`omitempty` decision for this profile DTO.
- The tool-level tests to add or update in this step, with assertions for `_meta.server_version`, `_meta.units`, terse/full deltas, argument validation, and structured/text equivalence.

Once those decisions are recorded, the refactor should be straightforward to implement and review.

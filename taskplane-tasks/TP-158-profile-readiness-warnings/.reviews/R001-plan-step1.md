# R001 Plan Review — Step 1

**Verdict:** Approved with required design clarifications.

The Step 1 plan is directionally sound: the shared `internal/athleteprofile` shaping layer is the right place to add `_meta.warnings`, because both `get_athlete_profile` and `icuvisor://athlete-profile` flow through it. Keep the implementation typed and compact, not `map[string]any`.

Before coding, pin down these details to avoid false warnings or unstable output:

1. **Define the warning contract explicitly.** Use a typed `ReadinessWarning`/`Warnings []ReadinessWarning` field on `athleteprofile.Meta`, with stable fields such as `code`, `sport_types`, `message`, and optionally `action`/`missing_fields`. Do not include athlete IDs, raw upstream payloads, or response-owned `_meta` keys.
2. **Use a sport/metric applicability matrix.** Only warn for metrics relevant to that sport setting. For example, do not emit `missing_power_*` for Run/Swim or `missing_pace_*` for Ride unless the sport setting explicitly uses those metrics. Normalize from `Types`, with a fallback to `Type` when `Types` is empty.
3. **Avoid alias-related false positives.** Presence checks should account for the upstream aliases already represented in `intervals.SportSettings` (notably `ThresholdPace` vs `PaceThreshold`, and potentially HR threshold fields) or deliberately match the current output fields and document that choice in tests.
4. **Verify response shaping preserves warnings.** `_meta.warnings` should survive `response.Shape`/`addCommonMeta` in terse mode and should not be stripped or overwritten by common metadata.

Step 1 tests should include at least one missing-settings case and one complete-settings case through the shared/profile shape (the existing tool test is acceptable if no `internal/athleteprofile` test file is added). Broader resource/schema coverage can remain in Step 2 as planned.

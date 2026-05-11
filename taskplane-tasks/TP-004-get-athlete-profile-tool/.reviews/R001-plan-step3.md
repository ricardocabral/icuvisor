# Plan Review: TP-004 Step 3 — shape the response for v0.1

## Verdict

**Request changes before coding Step 3.** The Step 1 contract is detailed, but the current Step 3 plan in `STATUS.md` is only a checklist and leaves one response-shaping promise unresolved: timezone fallback. Tighten the Step 3 plan so the implementation cannot ship a response whose `_meta` claims behavior the tool cannot provide.

## Findings

### Blocking: timezone fallback is promised but not planned

The Step 1 response contract says `_meta.timezone_convention` is `IANA timezone from athlete profile when available; config timezone fallback otherwise`. However the Step 2 tool boundary is `ProfileClient.GetAthleteProfile(ctx)` plus `tools.NewRegistry(profileClient, version)`, so Step 3 currently has no access to the configured timezone if `profile.Timezone` is empty.

Before coding Step 3, amend the plan to choose one of these paths:

1. Extend the tool/registry construction with a non-secret configured timezone fallback and use it when the upstream profile has no timezone; or
2. Change the contract/meta wording to only claim `timezone` is emitted when present in the upstream profile.

Do not leave `_meta.timezone_convention` claiming a config fallback unless the implementation can actually perform it.

## Non-blocking notes

- Add a short Step 3 design paragraph under `STATUS.md` rather than relying only on the checklist. It should explicitly point to the Step 1 contract and call out the exact field-mapping rules for: normalized top-level `athlete_id`, `units.measurement_preference`, metric vs imperial pace keys, `_meta.server_version`, and the `include_full: true` delta.
- Keep the default response strictly to the terse fields already defined in Step 1. `sport_setting_id`, `sport_setting_athlete_id`, raw measurement preference source, fetched timestamps, raw upstream JSON, request URLs, headers, or credential-derived values must not appear unless the Step 1 contract explicitly allows them.
- For unit normalization, make the plan clear that the public response should use stable LLM-friendly values such as `metric`/`imperial`, `kg`/`lb`, and `celsius`/`fahrenheit`, not arbitrary lowercased upstream enum strings unless returned in an explicit `*_source` field.

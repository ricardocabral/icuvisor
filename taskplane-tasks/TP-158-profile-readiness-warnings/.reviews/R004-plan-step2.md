# R004 Plan Review — Step 2

**Verdict:** Approved with required test-contract clarifications.

The Step 2 plan targets the right surfaces: the public tool output, the shared athlete-profile resource, schema snapshot drift, and the write-tool guidance that warnings point to. Before implementing, tighten the plan in these ways:

1. **Test the shaped/serialized boundary, not only `NewResponse`.** Add handler-level `get_athlete_profile` coverage with default `{}` arguments and decode the `TextResult`, proving `_meta.warnings` survives terse response shaping. For complete settings, assert `warnings` is absent from serialized `_meta` (or at least empty after decoding), not just empty on the pre-shaped typed struct.
2. **Make the warning matrix explicit.** The missing ride/run/swim test should assert deterministic codes and sport scope: Ride gets power + HR warnings, Run gets HR + pace warnings, Swim gets HR + pace warnings; no pace warning for Ride unless applicability changes. Include sport_types in assertions so mixed `types` do not produce ambiguous warnings.
3. **Cover false-positive aliases in the “complete” case.** Include at least one complete setting that uses upstream alias fields (`FTHR` for HR threshold and `PaceThreshold` for pace threshold) so Step 1’s alias handling cannot regress.
4. **Resource test should use a warning-producing profile.** The existing resource/shared-shape test with a complete profile does not prove warnings are exposed through `icuvisor://athlete-profile`. Decode resource JSON and assert the same warning codes/action fields as the tool, or compare against `athleteprofile.Shape` using a missing-settings fixture plus explicit warning assertions.
5. **Align warning actions with actual write fields.** If action strings are touched, make them point to `update_sport_settings` fields the model can actually send: `ftp`, `threshold_hr`, `threshold_pace`, and `zones` (not raw LTHR/FTHR terminology as the only instruction). Do not imply API keys, credentials, or a model-controlled athlete credential.
6. **Refresh schema snapshots only if generated input schemas drift.** `internal/tools/schema_snapshot/get_athlete_profile.json` currently snapshots the input schema (including coach `athlete_id`), not the output schema/description. Run the generator/check if unsure, but avoid manual snapshot churn when only output description text changed.

Targeted verification from the prompt (`go test ./internal/tools ./internal/resources`) is sufficient for this step after the above tests are in place.

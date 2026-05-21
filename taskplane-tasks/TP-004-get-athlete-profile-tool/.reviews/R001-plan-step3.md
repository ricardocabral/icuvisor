# Plan Review: TP-004 Step 3 — shape the response for v0.1

## Verdict

**Approve.** The Step 3 plan now covers the response-shaping contract from Step 1 and addresses the key boundary issue for timezone fallback by explicitly adding a non-secret configured timezone to registry/tool construction.

## Notes for implementation

- When adding the configured timezone fallback, make the constructor/API change explicit and thread only `config.Config.Timezone` (not the full config or any credential-bearing value) into the tool layer. Use `profile.Timezone` first, then the fallback; only keep the current `_meta.timezone_convention` wording if that behavior is implemented.
- Keep the `include_full: true` delta exactly as planned: `measurement_preference_source`, `sport_setting_id`, and normalized `sport_setting_athlete_id` only. Do not let raw upstream payloads, request URLs, headers, fetched timestamps, or debug fields slip into either mode.
- Unit normalization should produce stable public values (`metric`/`imperial`, `kg`/`lb`, `celsius`/`fahrenheit`). Preserve raw upstream unit/preference strings only in the explicit source fields allowed by the Step 1 contract.
- For pace fields, continue using disambiguating unit-specific keys based on `pace_units` (`*_seconds_per_km` vs `*_seconds_per_mile`) and include `pace_units_source` only when present.

The remaining validation items belong in Step 4 tests, especially `_meta.server_version`, normalized athlete IDs, default vs `include_full`, and absence of secret/debug fields.

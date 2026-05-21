# Plan Review — Step 2: Test

Decision: **approved**

The Step 2 plan targets the right behavior: exercise `Config` through `slog` JSON output, verify the allowlisted structured attrs, and add a negative assertion that the API key value is never emitted. That is sufficient for this XS task, provided the tests are written narrowly around the approved allowlist from Step 1.

## Notes for implementation

- Assert the structured `cfg` value is a JSON object/group containing exactly the approved keys: `api_base_url`, `default_athlete_id`, `http_bind`, `coach_athletes_count`, `delete_mode`, and `toolset`.
- Include a direct `LogValue` check if practical: `cfg.LogValue().Kind() == slog.KindGroup`, then inspect `Group()` attrs. This makes the `slog.LogValuer` contract explicit instead of relying only on JSON handler behavior.
- Keep the redaction assertions stronger than just “no secret value”: the JSON should contain neither the `api_key` key nor the configured API key value. Given the Step 1 redaction decision, also assert raw athlete IDs/coach roster details are not emitted when present in the fixture.
- Prefer a local logger (`slog.New(slog.NewJSONHandler(&buf, nil))`) for the JSON test. If the test uses `slog.Default()` to match the prompt literally, restore the previous default and do not run that test in parallel to avoid cross-test log capture interference.
- Parse the JSON into a typed or `map[string]any` structure for key/value assertions; reserve `strings.Contains` checks for negative leak checks.

No plan changes are required before writing Step 2 tests.

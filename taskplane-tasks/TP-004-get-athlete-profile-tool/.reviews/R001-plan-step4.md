# Plan Review: TP-004 Step 4 — add tests

## Verdict

**Request changes before coding Step 4.** The checklist covers the high-level acceptance criteria, but it is too thin for the behavior already implemented in Steps 2–3. Step 4 is the main regression net for the tool contract, so the plan should explicitly include the response-shaping and argument-validation cases below.

## Blocking gaps

1. **Response-shaping tests are under-specified.** The plan mentions success and `_meta.server_version`/normalized athlete ID, but it does not explicitly cover Step 3 behavior that must not regress:
   - timezone precedence: upstream profile timezone first, configured fallback second, `config.DefaultTimezone` when no fallback is supplied;
   - unit normalization: `measurement_preference` to `metric`/`imperial`, weight `kg`/`lb`, temperature `celsius`/`fahrenheit`;
   - pace key selection: km fields for `MINS_KM`, mile fields for `MINS_MILE`, with `pace_distance_unit` and `pace_units_source`;
   - omission of fetched timestamps, raw upstream payloads, URLs, headers, credentials, and debug fields.

2. **Default vs `include_full` needs exact-delta assertions.** Since `include_full` is implemented, the plan should remove the conditional wording and assert the exact contract:
   - default response omits `measurement_preference_source`, `sport_setting_id`, and `sport_setting_athlete_id`;
   - `include_full: true` includes only those additional typed non-secret fields when present;
   - `sport_setting_athlete_id` is normalized to the `i12345` form.

3. **Strict runtime argument validation is not planned for tests.** Step 2 intentionally rejects unknown fields, JSON `null`, non-object arguments, and trailing JSON. Add table-driven handler tests for these cases and assert they return the short public invalid-arguments message without calling the fake intervals client.

4. **Error-path tests should distinguish upstream failures from cancellation.** The plan covers upstream error mapping, but Step 2 also preserved `context.Canceled`/`context.DeadlineExceeded` instead of converting them to credential errors. Add coverage for at least one cancellation path so future changes do not turn canceled requests into misleading auth/athlete-ID errors.

## Suggested test structure

- Add `internal/tools/get_athlete_profile_test.go` in package `tools` so tests can inspect the registered `Tool` and call the handler directly.
- Use a fake `ProfileClient` with call-count/context capture and configurable profile/error; do not hit the network.
- Use table-driven tests with `t.Run`, matching repository conventions.
- For registration metadata, collect tools through a fake `Registrar` and assert:
  - exactly one tool named `get_athlete_profile`;
  - the first sentence/description clearly targets profile/thresholds/zones;
  - input schema is an object with `additionalProperties: false`;
  - the only argument is `include_full` with boolean type/default false;
  - no schema properties named or containing `api_key`, `password`, `token`, `credential`, or `athlete_id`.
- For handler success, assert both `StructuredContent` and text JSON decode to the expected typed shape, including `_meta.server_version`, normalized `athlete_id`, units, timezone, sport thresholds/zones, and pace fields.

Once these cases are added to the Step 4 plan, it should be adequate to implement.

# Code Review — TP-009 Step 4

Decision: **Changes requested**

## Summary

`go test ./...` passes, but I cannot approve Step 4 yet. The new tools are registered and the happy-path unit tests cover basic metadata gating and metric/imperial virtual splits, but there are several contract and correctness gaps around stream decoding, key-to-upstream mapping, error handling, Strava-unavailable fallback, and split pace/validation.

## Blocking issues

1. **Stream decoding is not tolerant enough for real stream payloads**
   - `internal/intervals/activity_streams.go:22-25` models `data`/`data2` as `[]float64`, and `UnmarshalJSON` then decodes the full stream through that typed alias (`activity_streams.go:39-41`).
   - This will fail the whole `get_activity_streams` call when any returned stream contains `null` samples, array-valued samples (`valueTypeIsArray`, e.g. lat/lng-style streams), or other non-float values. This is especially problematic because even the default metadata-only tool path still fetches and decodes the samples before stripping them.
   - Please make the upstream model tolerant enough to preserve raw payloads and not fail metadata-only calls on heavy/irregular streams. For split derivation, extract numeric `distance`/`time` samples from a typed helper rather than forcing all stream channels into `[]float64`.

2. **Explicit stream keys are not mapped to the upstream `types` tokens**
   - `getActivityStreamsHandler` canonicalizes requested keys for filtering, but sends the original caller strings directly upstream (`internal/tools/get_activity_streams.go:99-101`). The splits fallback similarly sends canonical strings directly (`get_activity_streams.go:131`).
   - Per the Step 1 notes, callers may use canonical snake_case names or aliases, and the tool should map them to the best upstream `types` token. As written, a request like `keys:["heart_rate"]` or `keys:["watts"]` may issue `types=heart_rate` / `types=watts` instead of the documented/observed upstream tokens such as `HR`/`HeartRate` or `Power`, causing missing data despite valid public keys.
   - Please add a canonical/alias-to-upstream-token mapping and tests that assert the actual `ActivityStreamsParams.Types` for known aliases and for unknown pass-through values.

3. **`get_activity_splits` ignores interval errors and masks cancellation/credential failures**
   - `dto, _ := intervalsClient.GetActivityIntervals(ctx, args.ActivityID)` (`internal/tools/get_activity_streams.go:128`) discards every interval error, including `context.Canceled`, `context.DeadlineExceeded`, unauthorized/not-found, and other upstream failures, then proceeds to stream fallback.
   - This violates the project convention used by the other activity tools and can convert a real intervals failure into either a misleading streams error or an empty virtual split result.
   - Please handle the error explicitly: preserve context cancellation/deadline errors, return a user-facing fetch error for genuine upstream failures, and only fall back to streams when intervals were fetched successfully but contained no qualifying manual laps.

4. **Streams and splits do not implement the Strava-blocked unavailable fallback**
   - `get_activity_streams` returns `NewUserError(fetchActivityDetailsMessage, err)` for any streams error (`internal/tools/get_activity_streams.go:101-104`), and `get_activity_splits` does the same for streams fallback errors (`get_activity_streams.go:131-134`). There is no details-read confirmation path like `get_activity_intervals` uses for 404/403 Strava-blocked activities.
   - The TP-009 mission and acceptance criteria require Strava-blocked detection across the activity read cluster, returning `unavailable: { reason: "strava_tos", ... }` instead of propagating sparse/error rows.
   - Please add the same not-found/forbidden fallback-confirmation behavior for these tools, which likely requires passing an optional `ActivityDetailsClient` into the streams/splits constructors or otherwise reusing the existing details client from the registry.

5. **Manual split pace is wrong for non-unit-distance intervals**
   - `newSplitRow` sets `pace := duration` regardless of the interval distance (`internal/tools/get_activity_streams.go:270-272`). That is only correct for virtual splits where `meters` is exactly one split unit.
   - If manual intervals/laps are not exactly 1000 m or 1609.344 m, `pace_seconds` is emitted as raw duration rather than seconds per km/mile. For example, a 500 m interval in 120 s would report 120 s/km instead of 240 s/km.
   - Please compute pace as `duration / (meters / splitDistanceMeters(splitUnit))` and add regression coverage for a manual interval distance that is not one split unit.

6. **Virtual splits do not enforce the documented missing/non-monotonic stream behavior**
   - The new package doc says missing, non-monotonic, or mismatched distance/time streams return an empty split list (`internal/tools/doc.go`), but `virtualSplits` only checks empty/mismatched lengths (`internal/tools/get_activity_streams.go:239-241`). `interpolateTime` also accepts duplicate/decreasing distance spans by returning `times[i]` (`get_activity_streams.go:256-264`).
   - This can emit partial or incorrect splits for GPS glitches or reordered samples instead of the documented empty result.
   - Please validate distance and time streams are usable before interpolation: matching lengths, at least two samples, finite values, non-decreasing or strictly increasing as appropriate, and monotonic time. Add tests for non-monotonic distance/time and duplicate distance boundaries.

## Additional issues to address

- **Context cancellation is wrapped as user errors in the new handlers.** `get_activity_streams` and the profile/streams calls inside `get_activity_splits` should preserve `context.Canceled` / `context.DeadlineExceeded` the same way `get_activities`, details, and intervals do (`internal/tools/get_activity_streams.go:122-134`).
- **Input schemas lack argument descriptions.** `activityStreamsInputSchema` and `activitySplitsInputSchema` define properties without descriptions (`internal/tools/get_activity_streams.go:283-287`), which violates the MCP-server convention that every argument has an LLM-readable description including units/ranges/default semantics.
- **The `split_unit` enum is silently ignored for invalid values.** The schema restricts to `km`/`mi`, but `normalizeSplitUnit` defaults unrecognized values to preferred units (`internal/tools/get_activity_streams.go:194-205`) because the decoder does not validate. Either validate and return an invalid-arguments user error or explicitly document/encode the permissive behavior.
- **Response field `pace_seconds` is unit-ambiguous.** The activity tools have been careful to use unit-disambiguated names (`pace_seconds_per_km` / `pace_seconds_per_mile`). Splits include `split_unit`, but the field itself is still ambiguous; consider emitting unit-specific pace fields for consistency.
- **JSON encoding errors are ignored.** Both new handlers use `text, _ := json.Marshal(shaped)` (`internal/tools/get_activity_streams.go:111`, `143`) while the existing handlers wrap and return marshal errors.
- **Tests do not cover the actual streams client query construction or raw/null/array tolerance.** Please add httptest coverage for `/activity/{id}/streams`, `types`, `includeDefaults`, null/raw preservation, explicit key mapping, context errors, Strava fallback, manual split precedence, and stream edge cases.

## Commands run

- `git diff 6265cd298dff547d56b5d7bf357998b921813e39..HEAD --name-only`
- `git diff 6265cd298dff547d56b5d7bf357998b921813e39..HEAD`
- `go test ./internal/intervals ./internal/tools`
- `go test ./...`

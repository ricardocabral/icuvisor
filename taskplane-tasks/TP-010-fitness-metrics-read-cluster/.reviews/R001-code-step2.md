# Review R001 — Code review for Step 2

Decision: **Request changes**

I reviewed the diff from `5224829dbc105a1748a6f9465d15bc253b7846e4..HEAD`, read the changed tool/client code in context, and ran the validation commands.

Validation run:

- `go test ./...` — passes
- `make lint` — fails

## Blocking findings

1. **`make lint` fails on the new helper.**  
   `golangci-lint` reports `unparam` for `toolProfile` because its first return value is never used, and the only caller that binds it immediately discards it (`profile, unitSystem, timezone, err := ...; _ = profile`). This blocks the repository's required `make lint` gate.
   - `internal/tools/get_fitness.go:232`
   - `internal/tools/get_fitness.go:638`

   Suggested fix: make `toolProfile` return only `(response.UnitSystem, string, error)` or inline the profile fetch in the handlers.

2. **Default best-effort distance buckets are used but not reported in `_meta`.**  
   The Step 2 plan explicitly says requested/default buckets must be echoed in metadata. `decodeBestEffortsRequest` leaves `args.DistanceMeters` nil when omitted, `bestEffortsForSport` then applies per-sport defaults locally, but `_meta.distance_meters` is populated from the still-nil `args.DistanceMeters`. In terse default calls, consumers cannot tell which run/swim distance buckets were requested/checked.
   - Metadata built from nil value: `internal/tools/get_fitness.go:280`
   - Defaults applied later, per sport: `internal/tools/get_fitness.go:483-486`

   Suggested fix: record the effective distance buckets in metadata, likely per sport because swim defaults differ from run-style defaults, or otherwise make the contract explicit and test it.

3. **Runtime bucket validation silently drops invalid requested values.**  
   The schemas declare `minimum: 1`, but MCP callers can still send arbitrary JSON. `normalizePositiveInts` filters out non-positive values instead of rejecting them, so requests like `{"duration_seconds":[0,-5]}` are accepted and can produce empty curve output rather than the advertised “invalid curve arguments” error. This violates the plan's positive bucket validation requirement and makes bad prompts look like missing upstream data.
   - `internal/tools/get_fitness.go:350`
   - `internal/tools/get_fitness.go:368-369`
   - `internal/tools/get_fitness.go:410-423`

   Suggested fix: distinguish omitted/empty arrays from supplied invalid values and return a user error if any bucket is <= 0.

## Non-blocking note

`bestEffortRow.PaceValue` is unit-ambiguous (`pace_value`). The approved plan notes called out unit-disambiguated response keys for pace/distance values. If the upstream pace curve values have a stable unit, expose that in the field name (and/or metadata) before tests lock in the response shape.

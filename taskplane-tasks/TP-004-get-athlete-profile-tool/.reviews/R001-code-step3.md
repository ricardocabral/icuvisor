# R001 Code Review — Step 3: Shape the response for v0.1

## Verdict

**Request changes.** The response shaping is mostly aligned with the Step 1 contract, including stable unit labels, include-full boundaries, and profile-timezone precedence. One constructor/API issue still lets the tool emit metadata claiming a config timezone fallback when no fallback was actually wired.

## Findings

### Blocking: timezone fallback is optional while `_meta` always claims it exists

- **File:** `internal/tools/registry.go:16`
- **File:** `internal/tools/get_athlete_profile.go:263` and `internal/tools/get_athlete_profile.go:166`

`NewRegistry` accepts `timezoneFallback ...string`, so existing/future callers can still build the registry as `tools.NewRegistry(profileClient, version)` with no configured timezone. In that case, if intervals.icu returns an empty profile timezone, `profileTimezone` returns an empty string and the `timezone` field is omitted, but `_meta.timezone_convention` still says: `IANA timezone from athlete profile when available; config timezone fallback otherwise`.

That leaves the Step 1/Step 3 response contract unenforced: config loading always produces a timezone fallback (`config.Config.Timezone`, defaulting to UTC), but the registry API does not require it and does not default it.

Suggested fix: make the fallback explicit/non-optional in the registry constructor, or default the registry fallback to `config.DefaultTimezone` when no non-empty fallback is provided. When Step 5 wires the real server, pass only `config.Config.Timezone` into `tools.NewRegistry` rather than passing the full config.

## Notes

- `profileTimezone` correctly prefers the athlete profile timezone over the supplied fallback.
- The normalized measurement preference now preserves the explicit metric/imperial upstream preference independently from `weight_pref_lb`; pounds only affect `units.weight` unless no measurement preference is present.
- The default vs `include_full` response boundary stays within the agreed fields; no raw upstream payloads, request/debug fields, timestamps, or credentials were added.

## Checks run

- `git diff d0e5154..HEAD --name-only`
- `git diff d0e5154..HEAD`
- `go test ./...` — passed
- `gofmt -l internal/tools/get_athlete_profile.go internal/tools/registry.go` — no output
- `go vet ./...` — passed

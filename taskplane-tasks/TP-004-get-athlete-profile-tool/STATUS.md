# TP-004 — Status

**Issue:** v0.1 — get_athlete_profile tool
**Iteration:** 1
**Current Step:** Step 5: End-to-end local verification
**Last Updated:** 2026-05-11
**State:** ✅ Complete

### Step 1: Define the tool contract in STATUS.md

**Status:** ✅ Complete

- [x] Write intended description, arguments, and response shape
- [x] Do not accept API key as a tool parameter
- [x] Decide whether v0.1 needs `include_full`
- [x] Include units/timezone/athlete-ID conventions where available
- [x] Clarify pace field units/normalization for default and imperial athletes
- [x] Pin exact `include_full: true` response delta fields

#### Contract

**Tool name:** `get_athlete_profile`

**Description:** `Get the configured intervals.icu athlete profile, FTP/thresholds, zones, and sport settings. Use this for athlete identity, units, timezone, FTP, heart-rate thresholds, pace thresholds, and zone configuration; do not use it for activities, wellness, fitness trends, events, or workouts.`

**Arguments:** one optional argument only.

```json
{
  "include_full": false
}
```

- `include_full` (`boolean`, optional, default `false`): When true, include additional non-secret profile and sport-setting fields that icuvisor has typed from intervals.icu. Defaults to false; raw upstream payloads, credentials, request debug data, and fetched timestamps are never returned.
- Explicit non-arguments for v0.1: no `api_key`, `password`, token, or credential fields; no `athlete_id` argument because v0.1 uses the configured athlete from server config. Coach-mode athlete selection is out of scope.

**Terse default response shape:**

```json
{
  "athlete_id": "i12345",
  "name": "Example Athlete",
  "first_name": "Example",
  "last_name": "Athlete",
  "timezone": "America/Sao_Paulo",
  "locale": "en_US",
  "units": {
    "measurement_preference": "metric",
    "weight": "kg",
    "temperature": "celsius"
  },
  "sport_settings": [
    {
      "types": ["Ride"],
      "ftp_watts": 250,
      "indoor_ftp_watts": 240,
      "w_prime_joules": 20000,
      "p_max_watts": 900,
      "lthr_bpm": 170,
      "max_hr_bpm": 190,
      "power_zones_watts": [100, 150, 200],
      "power_zone_names": ["Z1", "Z2", "Z3"],
      "hr_zones_bpm": [120, 140, 160],
      "hr_zone_names": ["Z1", "Z2", "Z3"],
      "threshold_pace_seconds_per_km": 255.5,
      "pace_zones_seconds_per_km": [330, 300, 270],
      "pace_units_source": "MINS_KM",
      "pace_distance_unit": "km",
      "pace_zone_names": ["Z1", "Z2", "Z3"]
    }
  ],
  "_meta": {
    "server_version": "dev",
    "athlete_id_format": "i-prefixed intervals.icu athlete ID",
    "timezone_convention": "IANA timezone from athlete profile when available; config timezone fallback otherwise",
    "pace_convention": "paces are seconds per athlete pace distance unit; metric athletes receive threshold_pace_seconds_per_km/pace_zones_seconds_per_km, imperial athletes receive threshold_pace_seconds_per_mile/pace_zones_seconds_per_mile, and pace_units_source preserves the upstream enum such as MINS_KM or MINS_MILE",
    "include_full": false
  }
}
```

**Pace fields:** intervals.icu typed client currently exposes `threshold_pace`, `pace_units`, and `pace_zones`. v0.1 will not return decimal minutes. It will return the upstream numeric pace values as seconds per configured pace distance unit with unit-specific keys. If `pace_units` indicates miles (`MINS_MILE`), emit `threshold_pace_seconds_per_mile` and `pace_zones_seconds_per_mile` with `pace_distance_unit: "mile"`; otherwise emit `threshold_pace_seconds_per_km` and `pace_zones_seconds_per_km` with `pace_distance_unit: "km"`. Always include `pace_units_source` with the upstream enum/string when available.

**`include_full: true` exact response delta:** keep every terse-default field and additionally include only these typed fields when they are present:

- Top-level `measurement_preference_source` with the raw `measurement_preference` string from the typed intervals profile when it differs from the normalized `units.measurement_preference` value.
- Per sport setting `sport_setting_id` from typed `SportSettings.ID`.
- Per sport setting `sport_setting_athlete_id` normalized to `i12345` from typed `SportSettings.AthleteID`.

Do not return API keys, raw upstream JSON, HTTP headers, request URLs, Basic Auth usernames, fetched timestamps, or any untyped passthrough fields in either default or full mode.

**Error behavior:** return short LLM-facing messages such as `could not fetch athlete profile; check intervals.icu credentials and athlete ID` while wrapping/logging the detailed client error through existing MCP error handling.

### Step 2: Implement the typed tool

**Status:** ✅ Complete

- [x] Add typed request/response structs
- [x] Register exactly `get_athlete_profile`
- [x] Use a distinguishing first sentence
- [x] Include useful JSON Schema descriptions
- [x] Call intervals client with request context
- [x] Return short actionable LLM-facing errors
- [x] Add concrete registry constructor and fakeable profile-client interface
- [x] Pass normalized server version into tool responses
- [x] Enforce strict runtime argument validation with unknown-field rejection
- [x] Preserve Step 2/Step 3 boundary while leaving response-mapping hooks
- [x] Reject JSON `null` and other non-object arguments at runtime
- [x] Preserve mid-call context cancellation instead of mapping it to credential errors
- [x] Return an error instead of panicking for nil registrars

#### Step 2 design

- Add `internal/tools/get_athlete_profile.go` for the typed tool request, response structs, schemas, constructor, and handler.
- Extend `internal/tools/registry.go` with a concrete registry and constructor, tentatively `NewRegistry(profileClient ProfileClient, version string) Registry`. The constructor normalizes empty version to `dev`.
- Define a fakeable profile-client interface in `internal/tools`: `GetAthleteProfile(ctx context.Context) (intervals.AthleteWithSportSettings, error)`. The real `*intervals.Client` already satisfies it; tests can use stubs.
- Register exactly one tool, `get_athlete_profile`, from the concrete registry. App-level wiring to instantiate the intervals client and pass this registry may be completed in Step 5, but Step 2 must leave a usable constructor hook.
- Input schema is `type: object` with optional `include_full` boolean, default `false`, a clear description, and `additionalProperties: false`. Runtime decoding uses `json.Decoder.DisallowUnknownFields()` and returns `tools.NewUserError("invalid get_athlete_profile arguments; only include_full is supported", err)` for bad JSON or forbidden/unknown fields.
- Upstream/profile-client failures return `tools.NewUserError("could not fetch athlete profile; check intervals.icu credentials and athlete ID", err)`. The handler does not log directly and never includes upstream bodies, request URLs, secrets, config values, or raw athlete identifiers in public errors.
- Step 2 creates the typed response envelope and handler flow, including the version metadata hook. Step 3 owns final response shaping details: units, pace key selection, normalized IDs, terse/full field omission, and `_meta.server_version` assertions.

### Step 3: Shape the response for v0.1

**Status:** ✅ Complete

- [x] Return terse useful profile fields
- [x] Use disambiguating field names/metadata where applicable
- [x] Include `_meta.server_version`
- [x] Exclude fetched timestamps/debug cruft by default
- [x] Exclude secrets/raw upstream payloads by default
- [x] Add non-secret configured timezone fallback to registry/tool construction
- [x] Normalize public unit values to stable LLM-friendly strings
- [x] Keep explicit measurement preference independent from weight unit preference
- [x] Default missing registry timezone fallback to `config.DefaultTimezone`

#### Step 3 design

- Use the Step 1 contract as the response source of truth. Keep default output to normalized athlete identity, display/name fields, timezone, locale, unit preferences, sport thresholds/zones, pace fields, and `_meta`.
- Extend registry/tool construction with a non-secret configured timezone fallback from `config.Config.Timezone`. The response `timezone` uses `profile.Timezone` when present, otherwise the configured fallback; `_meta.timezone_convention` may claim config fallback only after this is implemented.
- Normalize top-level `athlete_id` and full-mode `sport_setting_athlete_id` through `config.NormalizeAthleteID` when possible.
- Normalize public unit strings to stable LLM-friendly values: `metric`/`imperial`, `kg`/`lb`, and `celsius`/`fahrenheit`. Preserve raw measurement preference only in `measurement_preference_source` when `include_full: true` and it differs from the normalized value.
- Use metric pace keys (`threshold_pace_seconds_per_km`, `pace_zones_seconds_per_km`) unless upstream `pace_units` indicates miles, then use mile keys (`threshold_pace_seconds_per_mile`, `pace_zones_seconds_per_mile`) and `pace_distance_unit: "mile"`. Always include `pace_units_source` when present.
- Always include `_meta.server_version`; normalize empty version to `dev`.
- Never return fetched timestamps, raw upstream JSON, request URLs, headers, credentials, or debug cruft. Only add `sport_setting_id`, `sport_setting_athlete_id`, and `measurement_preference_source` in `include_full: true` mode.

### Step 4: Add tests

**Status:** ✅ Complete

- [x] Test registration metadata and no secret args
- [x] Test successful handler with fake intervals client
- [x] Test include-full/default behavior if implemented
- [x] Test upstream error mapping
- [x] Test `_meta.server_version` and normalized athlete ID
- [x] Test timezone fallback precedence and default timezone fallback
- [x] Test unit normalization and pace key selection for km/mile settings
- [x] Test strict runtime argument validation rejects unknown fields, null, non-objects, and trailing JSON
- [x] Test cancellation is not mapped to credential/athlete-ID user errors
- [x] Test default/full response omits debug, raw, URL, header, credential, and timestamp fields
- [x] Assert full-mode forbidden/debug field omission with full-only fields present
- [x] Assert handler passes the original request context to the profile client

#### Step 4 design

- Add `internal/tools/get_athlete_profile_test.go` in package `tools` and call the registered handler directly through a fake registrar.
- Use a fake `ProfileClient` with call count, context capture, configurable profile, and configurable error. Tests must not hit the network.
- Registration metadata assertions: exactly one `get_athlete_profile` tool; first sentence targets athlete profile/thresholds/zones; input schema is an object with `additionalProperties: false`; only `include_full` is present with boolean type/default false; no property names contain `api_key`, `password`, `token`, `credential`, or `athlete_id`.
- Success assertions decode both `StructuredContent` and text JSON into `GetAthleteProfileResponse` and verify normalized `athlete_id`, `_meta.server_version`, timezone, units, sport thresholds/zones, and pace fields.
- Default/full assertions: default omits `measurement_preference_source`, `sport_setting_id`, and `sport_setting_athlete_id`; `include_full: true` includes only those additional typed fields when present and normalizes `sport_setting_athlete_id`.
- Argument validation assertions are table-driven and verify invalid args return `invalid get_athlete_profile arguments; only include_full is supported` without calling the fake client.
- Error assertions distinguish upstream failures, which map to `could not fetch athlete profile; check intervals.icu credentials and athlete ID`, from context cancellation/deadline errors, which propagate as cancellation errors.
- Negative output assertions marshal the response and check forbidden/debug substrings are absent: fetched timestamps, raw upstream payloads, URLs, headers, credentials, API keys, tokens, and Basic Auth details.

### Step 5: End-to-end local verification

**Status:** ✅ Complete

- [x] Exercise MCP stdio tool dispatch to `get_athlete_profile` with fake client/server
- [x] If local `.env` has `INTERVALS_ICU_ATHLETE_ID` and `INTERVALS_ICU_API_KEY`, run optional end-to-end MCP validation and record only non-sensitive result shape
- [x] Run `go fmt ./...`
- [x] Run `make test`
- [x] Run `make build`
- [x] Run `make lint` if available
- [x] Update `CHANGELOG.md`

## Discoveries

| Date | Finding | Impact |
| ---- | ------- | ------ |
| 2026-05-11 00:38 | Task started | Runtime V2 lane-runner execution |
| 2026-05-11 00:38 | Step 1 started | Define the tool contract in STATUS.md |
| 2026-05-11 00:45 | MIT Python reference not consulted for Step 1 | Contract derived from PRD, roadmap, existing config/client/MCP scaffolding, and task prompt |
| 2026-05-11 01:20 | Optional live MCP validation skipped | Local `.env` does not contain both `INTERVALS_ICU_ATHLETE_ID` and `INTERVALS_ICU_API_KEY`; no secrets read or printed |
| 2026-05-11 00:41 | Review R001 | plan Step 1: UNKNOWN |
| 2026-05-11 00:45 | Review R001 | code Step 1: UNKNOWN |
| 2026-05-11 00:48 | Review R001 | code Step 1: APPROVE |
| 2026-05-11 00:50 | Review R001 | plan Step 2: UNKNOWN |
| 2026-05-11 00:52 | Review R001 | plan Step 2: APPROVE |
| 2026-05-11 00:59 | Review R001 | code Step 2: UNKNOWN |
| 2026-05-11 01:03 | Review R001 | code Step 2: APPROVE |
| 2026-05-11 01:06 | Review R001 | plan Step 3: UNKNOWN |
| 2026-05-11 01:08 | Review R001 | plan Step 3: APPROVE |
| 2026-05-11 01:12 | Review R001 | code Step 3: UNKNOWN |
| 2026-05-11 01:16 | Review R001 | code Step 3: UNKNOWN |
| 2026-05-11 01:19 | Review R001 | code Step 3: APPROVE |
| 2026-05-11 01:21 | Review R001 | plan Step 4: UNKNOWN |
| 2026-05-11 01:22 | Review R001 | plan Step 4: APPROVE |
| 2026-05-11 01:29 | Review R001 | code Step 4: UNKNOWN |
| 2026-05-11 01:33 | Review R001 | code Step 4: APPROVE |

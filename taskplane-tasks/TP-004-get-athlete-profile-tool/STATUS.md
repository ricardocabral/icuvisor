# TP-004 — Status

**Issue:** v0.1 — get_athlete_profile tool
**Iteration:** 1
**Current Step:** Step 1: Define the tool contract in STATUS.md
**Last Updated:** 2026-05-11
**State:** Ready

## Step 1: Define the tool contract in STATUS.md

**Status:** 🟡 In Progress

- [x] Write intended description, arguments, and response shape
- [x] Do not accept API key as a tool parameter
- [x] Decide whether v0.1 needs `include_full`
- [x] Include units/timezone/athlete-ID conventions where available
- [ ] Clarify pace field units/normalization for default and imperial athletes
- [ ] Pin exact `include_full: true` response delta fields

### Contract

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
      "threshold_pace": 4.2,
      "pace_units": "min/km",
      "pace_zones": [5.5, 5.0, 4.5],
      "pace_zone_names": ["Z1", "Z2", "Z3"]
    }
  ],
  "_meta": {
    "server_version": "dev",
    "athlete_id_format": "i-prefixed intervals.icu athlete ID",
    "timezone_convention": "IANA timezone from athlete profile when available; config timezone fallback otherwise",
    "pace_convention": "threshold_pace and pace_zones use the intervals.icu pace value plus pace_units",
    "include_full": false
  }
}
```

**`include_full: true` additions:** include typed, non-secret identifiers that are useful for debugging sport-setting mapping, such as sport-setting `id` and normalized `athlete_id` per sport setting. Do not return API keys, raw upstream JSON, HTTP headers, or fetched timestamps.

**Error behavior:** return short LLM-facing messages such as `could not fetch athlete profile; check intervals.icu credentials and athlete ID` while wrapping/logging the detailed client error through existing MCP error handling.

## Step 2: Implement the typed tool

**Status:** ⬜ Not started

- [ ] Add typed request/response structs
- [ ] Register exactly `get_athlete_profile`
- [ ] Use a distinguishing first sentence
- [ ] Include useful JSON Schema descriptions
- [ ] Call intervals client with request context
- [ ] Return short actionable LLM-facing errors

## Step 3: Shape the response for v0.1

**Status:** ⬜ Not started

- [ ] Return terse useful profile fields
- [ ] Use disambiguating field names/metadata where applicable
- [ ] Include `_meta.server_version`
- [ ] Exclude fetched timestamps/debug cruft by default
- [ ] Exclude secrets/raw upstream payloads by default

## Step 4: Add tests

**Status:** ⬜ Not started

- [ ] Test registration metadata and no secret args
- [ ] Test successful handler with fake intervals client
- [ ] Test include-full/default behavior if implemented
- [ ] Test upstream error mapping
- [ ] Test `_meta.server_version` and normalized athlete ID

## Step 5: End-to-end local verification

**Status:** ⬜ Not started

- [ ] Exercise MCP stdio tool dispatch to `get_athlete_profile` with fake client/server
- [ ] If local `.env` has `INTERVALS_ICU_ATHLETE_ID` and `INTERVALS_ICU_API_KEY`, run optional end-to-end MCP validation and record only non-sensitive result shape
- [ ] Run `go fmt ./...`
- [ ] Run `make test`
- [ ] Run `make build`
- [ ] Run `make lint` if available
- [ ] Update `CHANGELOG.md`

## Discoveries

| Date | Finding | Impact |
| ---- | ------- | ------ |
| 2026-05-11 00:38 | Task started | Runtime V2 lane-runner execution |
| 2026-05-11 00:38 | Step 1 started | Define the tool contract in STATUS.md |
| 2026-05-11 00:45 | MIT Python reference not consulted for Step 1 | Contract derived from PRD, roadmap, existing config/client/MCP scaffolding, and task prompt |
| 2026-05-11 00:41 | Review R001 | plan Step 1: UNKNOWN |
| 2026-05-11 00:45 | Review R001 | code Step 1: UNKNOWN |

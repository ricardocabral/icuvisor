# Plan Review: TP-004 Step 1 — Define `get_athlete_profile` contract

## Verdict

**Needs changes before proceeding to Step 2.** The current STATUS.md only repeats the Step 1 checklist; it does not yet define the concrete tool contract that Step 2/3 will implement.

## Findings

### 1. Missing concrete contract in STATUS.md

- `STATUS.md:13-16` lists the required Step 1 tasks, but there is no intended tool description, input schema, output shape, or example response.
- The prompt explicitly requires writing those details in `STATUS.md` before implementation (`PROMPT.md:43-48`).

**Required adjustment:** Add a dedicated contract section under Step 1, e.g.:

- Tool name: exactly `get_athlete_profile`
- First-sentence description distinguishing profile/thresholds/zones from activities or wellness
- Arguments: exact fields, defaults, and JSON Schema descriptions
- Response shape: top-level fields, nested sport settings, `_meta`, units, timezone, and omitted fields
- Error shape/public messages, if known

### 2. `include_full` decision is unresolved

- `STATUS.md:15` says to decide whether v0.1 needs `include_full`, but the plan leaves it undecided.
- The PRD and contributing guidance lean toward `include_full: bool` for read tools (`PRD:212`, `CONTRIBUTING.md` MCP conventions), while the task prompt allows an explicit v0.1 decision as long as terse default is preserved.

**Required adjustment:** Make an explicit decision and record it. Recommended: include a single optional argument:

```json
{
  "include_full": false
}
```

with a schema description such as: `When true, include additional non-secret profile/sport-setting fields returned by intervals.icu. Defaults to false; raw upstream payloads and credentials are never returned.`

If the team chooses not to implement `include_full` in v0.1, STATUS.md should document the rationale and the conflict/exception against the broader PRD convention before coding.

### 3. Input contract should explicitly avoid credential and coach-mode arguments

The Step 1 checklist says not to accept an API key, but the contract should also make clear that v0.1 does **not** accept `athlete_id` as a tool argument. The configured athlete comes from server config; coach-mode athlete selection is out of scope for v0.1.

**Required adjustment:** Define arguments as either empty `{}` or `{ include_full?: boolean }`; explicitly list forbidden/non-arguments:

- no `api_key`
- no `password`/token fields
- no `athlete_id` in v0.1

The response should still emit the normalized `athlete_id` (`i12345`) per PRD response-shaping rules.

### 4. Response shape needs unit/timezone specifics before implementation

The plan currently says to include conventions “where available” but does not pin names or units. That leaves Step 2/3 open to inconsistent field naming.

**Required adjustment:** Record concrete response fields using disambiguating names. A terse default could be shaped along these lines:

```json
{
  "athlete_id": "i12345",
  "name": "Example Athlete",
  "timezone": "America/Sao_Paulo",
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
    "timezone_convention": "IANA timezone from athlete profile/config when available"
  }
}
```

The exact fields can differ, but STATUS.md should settle the naming, units, and omission rules before implementation.

### 5. `_meta.server_version` should be part of the Step 1 contract

The acceptance criteria require `_meta.server_version`, and Step 3 lists it, but Step 1 should include it in the response contract so tests and structs are designed correctly from the start.

**Required adjustment:** Put `_meta.server_version` in the planned response shape and note that no `fetched_at`, `query_type`, secrets, or raw upstream payloads are returned by default.

## Suggested Step 1 completion criteria

Before marking Step 1 complete, STATUS.md should contain:

1. A concrete tool description string, including the distinguishing first sentence.
2. The finalized arguments/schema decision (`include_full` yes/no, default, descriptions; no secrets; no v0.1 `athlete_id`).
3. A concrete terse response shape with normalized `athlete_id`, timezone, unit conventions, profile/name fields, sport settings, thresholds, zones, and `_meta.server_version`.
4. A default/full behavior statement and explicit non-secret/no-raw-payload guarantee.
5. A note in Discoveries if the MIT Python reference was consulted, or a statement that it was not consulted for this contract.

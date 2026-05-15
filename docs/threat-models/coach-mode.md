# Coach mode threat model

## Scope

This review covers the v0.5 coach-mode design where one locally configured coach-scoped intervals.icu API key can target multiple athletes by accepting an LLM-supplied `athlete_id` argument. The key is loaded from existing server configuration sources and is never accepted as a tool argument or returned in any tool response.

Out of scope for v0.5: per-athlete delegated credentials, hosted multi-tenant operation, wildcard or multi-athlete calls, and direct client-side calls from the LLM to intervals.icu.

## Assets

- Coach-scoped intervals.icu API key.
- Roster membership and athlete identifiers.
- Per-athlete tool permissions configured by the coach.
- Athlete data reachable through read tools.
- Athlete data mutable through write/delete tools.

## Trust boundaries

- The MCP client and LLM control tool names and tool arguments, including `athlete_id`.
- icuvisor owns configuration loading, credential retrieval, athlete-ID normalization, tool registration, ACL evaluation, and all upstream intervals.icu HTTP calls.
- intervals.icu owns final authorization for what the coach-scoped key can access upstream.

## Required security invariants

1. `athlete_id` is a target selector only. It is not a credential, cannot replace the configured API key, and is never interpolated into an auth header.
2. The intervals.icu API key is read only by server startup/configuration code and remains inside the local process. No tool schema contains an API-key argument, and no tool response includes the key or credential source details.
3. A request may target exactly one normalized athlete ID. Wildcards, comma-separated lists, empty strings, and malformed IDs are rejected before any upstream call.
4. In coach mode, the normalized athlete ID must match the configured coach roster before any tool-specific handler runs. In single-athlete mode, any explicit `athlete_id` must match the configured single athlete.
5. Per-athlete ACL evaluation uses the normalized target athlete ID and requested tool name at request entry. The caller cannot choose another athlete's ACL by spelling the ID differently because all accepted IDs normalize to canonical `i12345` form before lookup.
6. Registration-time catalog filtering and request-time authorization both compose with the process-global delete-mode gate and the toolset-tier gate. Any deny is final.
7. Error messages for malformed IDs and unknown roster members are intentionally the same class of error so callers cannot enumerate roster membership from differential failures.

## Threat analysis

### Credential exfiltration through `athlete_id`

`athlete_id` is only used to select an upstream path segment after normalization and roster authorization. It is never used as a username, password, bearer token, header name, query that includes credentials, log field containing the API key, or response field. The Basic Auth username remains the fixed intervals.icu API username and the password remains the server-held API key.

Mitigations:

- Keep API-key loading in config/keychain paths, not tool arguments.
- Redact or avoid credential-bearing fields in logs.
- Build intervals.icu requests through the typed client rather than exposing raw HTTP tools.
- Reject malformed athlete IDs before path construction.

Residual risk: if a future debug endpoint or log dump tool exposes process configuration, it could leak the key independently of coach mode. Such tooling must remain out of the MCP catalog or must explicitly redact secrets.

### ACL escalation by swapping `athlete_id`

An LLM might try to call a write-capable tool with an athlete ID whose ACL allows the tool, then include data meant for a different athlete, or switch from a read-only athlete to a full-access athlete. The request-time resolver prevents this by binding the normalized target athlete ID and ACL decision before the tool handler receives the request. The tool handler only receives the resolved target, not an arbitrary unvalidated string.

Mitigations:

- Evaluate `coach.Evaluator(athleteID, toolName)` after normalization and roster lookup.
- Use deny-by-default behavior when an athlete has no ACL or when a tool name is unknown.
- Apply the same authorization at request time even if the tool catalog shown to the MCP client is stale.
- Preserve delete-mode and toolset-tier denies; coach ACL cannot re-enable a tool denied by another gate.

Residual risk: MCP clients can cache tool catalogs per conversation. A stale visible catalog is treated as a UX/cache issue, not an authorization source; request-time ACL remains authoritative.

### Roster escape by guessing IDs

A coach-scoped intervals.icu key may be able to access more upstream athletes than the local config intends to expose. Therefore upstream authorization is not sufficient. In coach mode, icuvisor must treat the configured roster as the local authorization boundary and reject any normalized target ID absent from that roster before making an intervals.icu request.

Mitigations:

- Require non-empty coach roster when coach mode is on or auto-enables.
- Normalize configured roster IDs during config load and reject duplicates.
- Reject unconfigured IDs with a generic invalid-target error.
- Do not silently fall back to the default athlete when an explicit target is invalid.

Residual risk: if the coach intentionally configures an athlete, the LLM can operate within that athlete's allowed tool surface. This is the intended delegation model.

## Implementation requirements derived from the review

- Feature flag defaults to off; `list_athletes` and `select_athlete` are not registered outside coach mode.
- All athlete-scoped tools receive a consistent optional `athlete_id` argument whose value is normalized once at request entry.
- Request routing rejects a target absent from the relevant roster/single-athlete configuration before the intervals client builds an HTTP request.
- Per-athlete ACLs are deny-by-default for unknown tools and compose as: delete-mode gate, then toolset-tier gate, then coach ACL.
- Tool responses may include canonical athlete IDs and labels but must not include credential material.

## Endpoint probe

### Probe environment

- Date: 2026-05-15.
- Method: clean-room probe against public intervals.icu API documentation and unauthenticated HTTP responses. No GPL implementation was consulted.
- Local credential availability: `INTERVALS_ICU_API_KEY`, `INTERVALS_ICU_ATHLETE_ID`, `ICUVISOR_CONFIG`, and `ICUVISOR_ENV_FILE` were unset in the task environment, so an authenticated real-coach-key probe could not be completed.

### Public API documentation result

The public OpenAPI document at `GET https://intervals.icu/api/v1/docs` advertises:

- Path: `GET /api/v1/athlete/{id}/athlete-summary{ext}`
- Tag: `Athletes`
- Summary: `Summary information for followed athletes`
- Description note: when called with a bearer token, only the token's athlete is returned.
- Query parameters: `start`, `end`, and `tags`; no page-size, offset, cursor, or next-page parameter is advertised.
- Response shape: JSON array of `SummaryWithCats` objects. Roster-relevant fields include `athlete_id`, `athlete_name`, `email`, plus training summary fields such as `fitness`, `fatigue`, `form`, `rampRate`, `weight`, `timeInZones`, `byCategory`, and `mostRecentWellnessId`.

The current intervals client uses Basic Auth with username `API_KEY` and the configured API key as password. That auth style should be used for any future authenticated probe unless upstream docs require otherwise.

### Unauthenticated endpoint observations

Unauthenticated probes are not sufficient to confirm coach-account behavior, but they help distinguish documented paths from guesses:

| Candidate path                           | Observed status | Interpretation                                                                          |
| ---------------------------------------- | --------------: | --------------------------------------------------------------------------------------- |
| `/api/v1/athlete/0/athlete-summary`      |             401 | Documented followed-athlete summary endpoint exists and requires authentication.        |
| `/api/v1/athlete/0/athlete-summary.json` |             401 | Documented endpoint also accepts an `.json` extension form.                             |
| `/api/v1/athlete/i0/athlete-summary`     |             403 | `0` has special authenticated-user behavior; canonical `i0` is not equivalent.          |
| `/api/v1/athlete/0/followers`            |             401 | Path exists but is not documented as the roster summary endpoint in the OpenAPI result. |
| `/api/v1/athletes`                       |             401 | Path exists but is not documented as the roster summary endpoint in the OpenAPI result. |
| `/api/v1/coaches/athletes`               |             401 | Path exists but is not documented as the roster summary endpoint in the OpenAPI result. |
| `/api/v1/athlete/0/athletes`             |             404 | Not an exposed endpoint.                                                                |
| `/api/v1/athlete/0/coached-athletes`     |             404 | Not an exposed endpoint.                                                                |
| `/api/v1/athlete/0/clients`              |             404 | Not an exposed endpoint.                                                                |

### Temporary implementation fallback pending authenticated validation

Because this environment has no real coach key, the upstream roster endpoint remains unvalidated for v0.5 behavior. The authenticated coach-key probe is incomplete, not passed. Until a real coach account confirms auth, response shape, pagination behavior, and whether the endpoint returns only athletes intentionally exposed to the coach, implement `list_athletes` against the configured `coach.athletes[]` roster with `_meta.source: "config"` as a temporary fallback.

The intervals client may remain extensible for a later authenticated `GET /api/v1/athlete/0/athlete-summary` probe, but coach mode must not claim `_meta.source: "upstream"` or depend on that upstream endpoint until the blocked probe is completed.

Future maintainer probe requirements:

1. Obtain a real coach-scoped intervals.icu API key through the normal local credential flow. Do not paste it into a prompt, commit it, or pass it through an MCP tool argument.
2. Export it only in the local shell used for probing, for example `export INTERVALS_ICU_API_KEY='<redacted coach key>'`.
3. Run authenticated clean-room probes such as:

   ```sh
   curl -sS -u "API_KEY:${INTERVALS_ICU_API_KEY}" \
     'https://intervals.icu/api/v1/athlete/0/athlete-summary' | jq '.[0] | keys'

   curl -i -sS -u "API_KEY:${INTERVALS_ICU_API_KEY}" \
     'https://intervals.icu/api/v1/athlete/0/athlete-summary.json'
   ```

4. Record the exact path, auth style, response shape, roster filtering semantics, and any pagination or rate-limit headers before enabling `_meta.source: "upstream"`.

Pagination gap: the public spec for `athlete-summary` does not advertise pagination. If future authenticated testing returns large rosters or pagination metadata, update this document and the client contract before switching `_meta.source` to `"upstream"`.

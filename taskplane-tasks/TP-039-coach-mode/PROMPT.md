# TP-039 — Coach mode behind a feature flag, with per-athlete granular tool permissions

## Mission

Add coach-mode support so a coach holding a coach-scoped intervals.icu API key can talk to Claude (or any MCP client) about *multiple* athletes from one server, while keeping the credential never-on-the-wire (CLAUDE.md rule #6) and exposing only the tool surface the coach allowed *per athlete*. The server selects which athlete each call targets via an `athlete_id` argument the LLM passes; the credential stays in the keychain.

Concretely, this task ships:

1. **Feature flag** `ICUVISOR_COACH_MODE=on|off` (default `off`). When off, the server behaves exactly as today — single-athlete, no `list_athletes` / `select_athlete` tools registered, `athlete_id` arguments ignored or rejected per existing behavior.
2. **`list_athletes`** read tool — returns the coach's roster from the upstream coach endpoint (PRD §7.2.C; the current reference surfaces lack this — verify the intervals.icu endpoint via black-box probe as part of Step 1).
3. **`select_athlete`** session-state tool — sets the default `athlete_id` for subsequent tool calls in the same MCP session; any tool may still override per-call.
4. **Per-athlete tool ACL** loaded from the config file: `coach.athletes[].allowed_tools: [...]` and `coach.athletes[].denied_tools: [...]`. The server filters the tool catalog returned to the client *based on the active athlete* (PRD §7.1 Flow D: "granular per-tool permissions are enforced in the server before any intervals.icu call; the LLM never sees disallowed tools in its catalog"). Compose this filter with TP-018 delete-mode and TP-030 toolset-tier filters: a tool is exposed iff **all three** gates allow it.
5. **`athlete_id` argument** added consistently to every athlete-scoped tool in the catalog (most reads, all writes), defaulting to the selected athlete; emitted in canonical `i12345` form (TP-007 normalization).

PRD anchors: §5 Secondary segment "Coach with a roster"; §7.1 Flow D in full; §7.2.C `list_athletes`, `select_athlete`; §7.4 #5 (coach-mode credential-delegation threat model — "the coach-scoped API key is held only by the local binary and never passed as a tool parameter"); §6 Pains avoided ("Per-athlete SaaS signup for coaches").

ROADMAP positioning: v0.5 — Internal beta, fourth item ("Coach mode behind a feature flag, with per-athlete granular tool permissions"). The "internal beta" cohort includes "at least one coach" — this is the tool surface that coach uses.

Complexity: Blast radius 5 (every athlete-scoped tool grows an arg; the registry filter is third-gate), Pattern novelty 4 (multi-athlete session state inside MCP is new), Security 4 (credential-delegation threat model), Reversibility 2 = 15 → Review Level 4. Size: L.

## Dependencies

- **TP-018** — delete-mode registration-time filter (reuse the gate plumbing, compose into a three-gate filter).
- **TP-030** — toolset-tier filter (reuse and compose).
- **TP-002** — intervals.icu client (need to confirm/add the coach-roster endpoint).
- **TP-007** — athlete-ID normalization helper.
- **TP-038** — onboarding flow (coach key is entered the same way; document the variation, no code change needed unless `icuvisor setup --coach` becomes a future task).

Threat-model review (PRD §7.4 #5) is a Step 1 deliverable, not a follow-up.

## Context to Read First

- `CLAUDE.md` "MCP-server conventions" — `athlete_id` is a *target selector*, not a credential. Lock the policy.
- `docs/prd/PRD-icuvisor.md` §5, §7.1 Flow D, §7.2.C, §7.4 #5.
- `internal/safety/` — the registration-time gate established by TP-018; the coach-mode filter must compose with this, not bypass it.
- (To-be-built) toolset-tier gate from TP-030 — same composition rule.
- `internal/config/config.go` — the place a `coach` config stanza lives.
- `internal/intervals/` — current single-athlete request shapes; identify where `athlete_id` is fixed at construction vs. per-call.

## File Scope

Expected files:

- `internal/coach/` (new package) — coach config struct (`Athlete{ID, Name, AllowedTools, DeniedTools, DefaultToolset}`), config-file load + validate, the per-athlete tool-ACL evaluator. Pure logic; no MCP imports.
- `internal/coach/*_test.go` — table-driven for ACL evaluation, including the three-way compose with delete-mode and toolset-tier gates.
- `internal/intervals/coach.go` — `ListAthletes(ctx)` against the upstream coach endpoint; structured `ErrCoachEndpointUnavailable` if the endpoint is missing on the athlete's plan.
- `internal/intervals/coach_test.go` — fixture-based, no network.
- `internal/tools/list_athletes.go` + `_test.go` — registered only when `ICUVISOR_COACH_MODE=on`.
- `internal/tools/select_athlete.go` + `_test.go` — session-scoped state; the session abstraction is whatever the go-sdk exposes — if it has no per-session handle, scope state per-process and document the limitation in `STATUS.md`.
- `internal/mcp/` — registry plumbing: every athlete-scoped tool gets a uniform `athlete_id` arg; the three-gate filter composes; the registry refreshes the tool catalog returned to the client when the selected athlete changes (or documents the MCP-spec limitation that catalogs are per-conversation — PRD §7.4 #7).
- `internal/tools/*.go` — uniform `athlete_id` argument added to all athlete-scoped tools. Argument is **optional** in single-athlete mode (defaults to the only athlete) and **optional** in coach mode (defaults to the selected athlete; the LLM can override per-call).
- `internal/config/config.go` — `coach` stanza loader; `ICUVISOR_COACH_MODE` env-var parsing.
- `docs/coach-mode.md` (new) — config schema, ACL examples (read-only-for-prospect, full-for-active-client), the threat-model summary, the catalog-cache caveat.
- `docs/threat-models/coach-mode.md` (new) — the §7.4 #5 review writeup.
- `README.md` — short pointer to `docs/coach-mode.md` under Quickstart.
- `CHANGELOG.md`.
- `taskplane-tasks/TP-039-coach-mode/STATUS.md`.

## Steps

### Step 1: Threat-model review + endpoint probe

- [ ] Write the threat model: what an LLM-controlled `athlete_id` argument can and cannot do. Specifically, prove that swapping `athlete_id` cannot exfiltrate the coach key, escalate a per-athlete ACL, or operate against an athlete absent from the config roster.
- [ ] Black-box probe the intervals.icu coach-roster endpoint with a real coach key; document path, auth header, response shape, and pagination behavior. If the endpoint is not exposed for free-tier coach accounts, document that gap and proceed with a config-file roster (no `list_athletes` API call — the tool reads from config).
- [ ] Record both writeups in `docs/threat-models/coach-mode.md` + `STATUS.md`.

### Step 2: Config + feature flag

- [ ] `ICUVISOR_COACH_MODE=on|off|auto`. `auto` means "on iff `coach` config stanza is present and non-empty"; default `off`.
- [ ] Config stanza schema:
  ```json
  "coach": {
    "athletes": [
      { "id": "i12345", "label": "Jane (active client)",
        "allowed_tools": ["*"], "denied_tools": ["delete_event", "delete_events_by_date_range"] },
      { "id": "i67890", "label": "Bob (prospect, read-only)",
        "allowed_tools": ["get_*"], "denied_tools": ["*"] }
    ],
    "default_athlete_id": "i12345"
  }
  ```
- [ ] Validate at load: unknown tool names in `allowed_tools`/`denied_tools` fail loudly (typo defense).

### Step 3: Tool registry plumbing

- [ ] `internal/coach.Evaluator(athleteID, toolName) (allowed bool, reason string)` — the third gate.
- [ ] Compose order in the registry: delete-mode → toolset-tier → coach-ACL. Document evaluation order; any single deny is final.
- [ ] All athlete-scoped tools gain a uniform `athlete_id` optional arg with consistent schema description ("Target athlete; defaults to selected athlete in coach mode, or the only athlete otherwise. Format: `i12345` or `12345`.").
- [ ] Per-tool resolver: `athlete_id` is normalized once at request entry; rejected if not present in the coach roster (coach mode) or not the configured single athlete (non-coach mode).

### Step 4: `list_athletes` + `select_athlete`

- [ ] `list_athletes` returns the roster (from upstream if Step 1 confirmed the endpoint; from config otherwise — `_meta.source: "config" | "upstream"`).
- [ ] `select_athlete` updates the per-session default. Returns the previous selection, the new selection, and the tools now allowed for the new athlete. If the go-sdk has no session handle, document and persist per-process state with a `_meta.scope: "process"` note.

### Step 5: Catalog-cache caveat + Tests

- [ ] PRD §7.4 #7 says MCP clients cache the tool catalog per conversation. `select_athlete` may not actually change the catalog the LLM sees mid-conversation. Document this in `docs/coach-mode.md` and surface a `_meta.requires_new_conversation: true` flag on `select_athlete` when the per-athlete ACL would have changed the catalog (this hooks into TP-040 notification work later).
- [ ] Table-driven coverage of ACL composition: every combination of delete-mode × toolset-tier × per-athlete ACL on a representative tool (`delete_event`, `get_athlete_profile`, `add_or_update_event`). Property: any single gate's deny vetoes the tool.
- [ ] End-to-end with a faked intervals client: coach key, two athletes in config (one full-access, one read-only), confirm catalog and request routing.

### Step 6: Documentation

- [ ] `docs/coach-mode.md` covers: enable flag, config schema, ACL examples, catalog-cache caveat, threat-model summary, troubleshooting.
- [ ] README pointer.
- [ ] CHANGELOG.
- [ ] PRD §7.4 #5 marked validated in a separate follow-up PRD edit (out of scope here — open an issue noting the threat model is filed).

## Acceptance Criteria

- `ICUVISOR_COACH_MODE=off` (default) leaves the catalog and behavior unchanged versus today; `list_athletes` and `select_athlete` are not registered.
- `ICUVISOR_COACH_MODE=on` with a populated `coach` config exposes the roster, the selector tool, and the per-athlete-ACL-filtered catalog.
- `athlete_id` is rejected if not in the configured roster (coach mode) or not the configured single athlete (non-coach mode). Rejection message names the cause and is the same for both "wrong format" and "not in roster" cases (avoid roster enumeration via error message).
- ACL composition rule documented and tested: delete-mode AND toolset-tier AND coach-ACL must all allow.
- The coach API key is read only from env / keychain / config-file (TP-036 precedence); the LLM never sees it and cannot read it via any tool.
- Threat-model doc lives at `docs/threat-models/coach-mode.md`.

## Do NOT

- Do not accept the coach API key as a tool argument. The LLM never sees the credential under any circumstance.
- Do not allow `athlete_id` to be a wildcard. One call targets exactly one athlete.
- Do not register `list_athletes` / `select_athlete` in non-coach mode. Registration-time gating, consistent with TP-018.
- Do not silently coerce an unrecognized `athlete_id` to the default. Reject explicitly.
- Do not expose roster membership via differential error messages.
- Do not bypass delete-mode or toolset-tier filters in coach mode — coach ACL composes *with* them, never *over* them.
- Do not introduce per-athlete *credentials* in coach mode in v0.5 — the model is "one coach key, many athletes that key can already see". Per-athlete delegated keys are a future task.

## Documentation

Must update:

- `STATUS.md`
- `docs/coach-mode.md` (new)
- `docs/threat-models/coach-mode.md` (new)
- `README.md` (one-line pointer)
- `CHANGELOG.md`

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-039`, for example: `TP-039 add coach-mode ACL evaluator`.

---

## Amendments

_Add amendments below this line only._

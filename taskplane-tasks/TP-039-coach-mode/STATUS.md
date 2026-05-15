# TP-039-coach-mode: Coach mode + per-athlete tool ACLs — Status

**Current Step:** Step 6: Documentation
**Status:** ✅ Complete
**Last Updated:** 2026-05-15
**Review Level:** 4
**Review Counter:** 21
**Iteration:** 3
**Size:** L

---

### Step 1: Threat-model review + endpoint probe

**Status:** ✅ Complete

- [x] Threat model written (`athlete_id` cannot exfiltrate, escalate, or escape roster)
- [x] Coach-roster endpoint probed; path/auth/shape documented OR gap documented
- [x] Writeup in `docs/threat-models/coach-mode.md`
- [x] R001 revision: mark authenticated coach-key roster probe as blocked/incomplete unless a real coach-scoped key is provided, and phrase config roster as a temporary fallback pending validation

### Step 2: Config + feature flag

**Status:** ✅ Complete

- [x] R003 plan revision: cycle-free ACL validation uses an `internal/toolcatalog` name/pattern boundary and `internal/coach` normalized config types
- [x] R003 plan revision: define feature-flag state machine (`off` default, invalid fail, `auto` non-empty roster, `on` requires roster, `.env` support)
- [x] R003 plan revision: enforce roster validation matrix (ID normalization, duplicates, default selection, deny-overrides-allow ACL semantics, redacted String/log output)
- [x] R004 plan revision: make `internal/toolcatalog` the shared catalog contract via exported canonical tool-name constants and registry/config consumers
- [x] R004 plan revision: keep dependency direction `config -> coach -> toolcatalog`, with config-owned athlete-ID normalization and no `coach -> config` import
- [x] R004 plan revision: validate any present coach stanza even when mode is `off`; omit single-athlete default by filling it; empty `allowed_tools` means deny-all; docs/examples must not use `denied_tools: ["*"]` for read-only
- [x] `ICUVISOR_COACH_MODE=on|off|auto`
- [x] `coach.athletes[]` schema with `allowed_tools` / `denied_tools` / `default_athlete_id`
- [x] Unknown tool names fail loudly
- [x] R006 revision: allow coach-mode `on`/effective `auto` configs to omit top-level `athlete_id` by resolving `Config.AthleteID` from `coach.default_athlete_id`
- [x] R006 revision: add explicit registered-catalog drift test against `toolcatalog.AthleteScopedToolNames()` so ACL validation and registry cannot diverge silently
- [x] R007 revision: when coach mode is effectively on, `Config.AthleteID` must always resolve to `coach.default_athlete_id` regardless of legacy top-level `athlete_id`, with regression coverage

### Step 3: Tool registry plumbing

**Status:** ✅ Complete

- [x] R009 plan revision: compose all three gates in `internal/mcp.safeRegistrar` so coach-denied tools are absent from SDK tools/list, catalog hash inputs, and skip counts
- [x] R009 plan revision: filter `icuvisor_list_advanced_capabilities` against the active-athlete coach ACL so denied tools are not leaked through capability discovery
- [x] R009 plan revision: use concurrency-safe context-scoped athlete routing, bind `get_activities` continuation tokens to the resolved athlete, and use one enumeration-safe public target error
- [x] R009 plan revision: central wrapper adds/removes `athlete_id` for every athlete-scoped tool and tests schema/wrapper drift; resources are gated/routed or explicitly deferred with rationale
- [x] `coach.Evaluator` third gate
- [x] Compose order: delete-mode → toolset-tier → coach-ACL (any deny is final)
- [x] Uniform optional `athlete_id` arg with consistent description
- [x] Per-request normalization + roster check
- [x] Context-scoped target routing strips `athlete_id` before strict tool decoders and makes intervals requests use the resolved athlete
- [x] Registry/request tests cover coach ACL filtering, wrong-roster rejection, single-athlete mismatch rejection, and delete/toolset/coach composition
- [x] R011 revision: direct activity-ID tools must verify upstream activity ownership against the resolved target athlete before reads/writes/deletes
- [x] R011 revision: make `icuvisor_list_advanced_capabilities` derive coach filtering from the authoritative MCP gate, not duplicated app-only `CatalogFilter` wiring
- [x] R011 revision: preserve coach-mode-off catalog compatibility by gating `athlete_id` schema injection/stripping to effective coach mode
- [x] R012 revision: `LinkActivityToEvent` must preflight `event_id` under the resolved target athlete before writing the activity pairing

### Step 4: `list_athletes` + `select_athlete`

**Status:** ✅ Complete

- [x] R014 plan revision: coach mode registers the union of tools allowed by at least one roster athlete, then filters `tools/list`, advanced capabilities, and select responses through one active-athlete visibility helper
- [x] R014 plan revision: session selection store uses SDK session IDs when available, process fallback when not, initializes to `coach.default_athlete_id`, and feeds the Step 3 target resolver
- [x] R014 plan revision: register `list_athletes`/`select_athlete` only in effective coach mode, keep `list_athletes` config-backed only, and return enumeration-safe credential-free errors
- [x] R014 plan revision: `select_athlete` response includes previous/new selection, allowed tools, `_meta.scope`, and `_meta.requires_new_conversation` computed by visible-catalog diff
- [x] `list_athletes` (`_meta.source: "config" | "upstream"`)
- [x] `select_athlete` session/process-scoped state
- [x] `requires_new_conversation` `_meta` flag

### Step 5: Catalog-cache caveat + Tests

**Status:** ✅ Complete

- [x] R017 plan revision: derive `select_athlete.allowed_tools` and `requires_new_conversation` from the authoritative post-gate visible catalog, not coach evaluator alone
- [x] R017 plan revision: document the catalog-cache caveat in `docs/coach-mode.md` now, including new conversation/reconnect guidance and TP-040 future notifications
- [x] R017 plan revision: protocol truth-table tests assert catalog exposure and call-time vetoes for delete-mode, toolset, and coach ACL gates
- [x] R017 plan revision: end-to-end fake intervals coverage proves selected/default/override routing and read-only athlete write/delete denial with enumeration-safe errors
- [x] R018 plan revision: implement/select metadata through an `internal/mcp.safeRegistrar.visibleToolNamesForAthlete` helper over post-registration tools and inject it into `select_athlete` context
- [x] R018 plan revision: add hidden-gate metadata regressions for delete-mode-hidden `delete_event`, core-toolset-hidden full tools, and visible core-read changes
- [x] R018 plan revision: structured JSON tests compare exact visible tool sets across `tools/list`, `select_athlete.allowed_tools`, and `icuvisor_list_advanced_capabilities`, plus two-session isolation
- [x] §7.4 #7 caveat documented
- [x] Composition truth-table coverage
- [x] End-to-end with faked intervals client
- [x] R020 revision: add structured exact `icuvisor_list_advanced_capabilities` row-name assertions for active-athlete post-gate catalogs, including delete-mode-hidden and core-toolset-hidden regressions
- [x] R020 revision: make fake-client write/delete denial use valid arguments and assert the upstream request log is unchanged

### Step 6: Documentation

**Status:** ✅ Complete

- [x] `docs/coach-mode.md`
- [x] README pointer
- [x] CHANGELOG
- [x] Follow-up issue for PRD §7.4 #5 status update

---

## Decisions

- **Per-athlete delegated keys:** out of scope for v0.5. v0.5 ships single-coach-key + many-athletes-it-can-already-see.

## Blockers

- Step 1 authenticated roster probe is blocked: this execution environment has no `INTERVALS_ICU_API_KEY`, `INTERVALS_ICU_ATHLETE_ID`, `ICUVISOR_CONFIG`, or `ICUVISOR_ENV_FILE`, no default config at `~/Library/Application Support/icuvisor/config.json` or `~/.config/icuvisor/config.json`, and no accessible `icuvisor`/`intervals-icu-api-key` OS keychain credential. No real coach-scoped intervals.icu key was provided. Public OpenAPI and unauthenticated probes identified the likely endpoint, but R001 correctly rejected that as insufficient to complete the authenticated coach-key probe requirement.

## Notes

- Step 1 writeup lives at `docs/threat-models/coach-mode.md`.
- Threat model conclusion: `athlete_id` is only a normalized target selector; it cannot exfiltrate credentials, bypass per-athlete ACLs, or escape the local roster if request-time roster checks remain authoritative and compose with delete-mode/toolset gates.
- Endpoint probe conclusion: public OpenAPI documents `GET /api/v1/athlete/{id}/athlete-summary{ext}` as “Summary information for followed athletes” with `SummaryWithCats[]` fields including `athlete_id` and `athlete_name`, but no real coach key was available in the task environment, so TP-039 should implement `list_athletes` from config first (`_meta.source: "config"`) and leave upstream roster support for a later authenticated probe.
- Supervisor steering on 2026-05-15 explicitly treats the authenticated black-box coach-roster probe as an external/operator-deferred validation gate; Step 1 is complete on the documented-gap/fallback basis, not because upstream roster discovery was locally proven.
- Step 2 plan decisions from R003/R004: config parsing stays cycle-free by validating ACL patterns against an `internal/toolcatalog` package rather than importing `internal/tools`; `internal/toolcatalog` owns exported canonical tool-name constants consumed by registry/config to avoid drift. `allowed_tools` is the positive allow list, `denied_tools` is an explicit veto, and deny patterns override allow patterns (`denied_tools: ["*"]` means deny all, not read-only). Config owns athlete-ID normalization and constructs normalized `internal/coach` values; coach does not import config. Any present coach stanza is validated for typo defense even when mode is `off`, while runtime behavior remains single-athlete because effective coach mode is off.
- Step 3 plan decisions from R009: the actual exposed-catalog gate lives in `internal/mcp.safeRegistrar` after delete-mode and toolset checks, not only in `internal/tools`; `icuvisor_list_advanced_capabilities` must consume a coach-filtered catalog; intervals routing must use context/per-call state, never mutate the shared client; pagination tokens must include the resolved canonical athlete; and malformed/unknown/mismatch targets must share one public error.
- Step 3 resource-bypass decision: `icuvisor://athlete-profile` is disabled while coach mode is enabled until a later task adds per-session resource target selection/ACL routing, preventing a resource read from bypassing a denied `get_athlete_profile` tool.
- R011 compatibility decision: `athlete_id` schema injection and stripping are active only when coach mode is effectively on; coach-mode-off tool schemas and strict-decoder behavior remain unchanged.
- Step 4 plan decisions from R014: effective coach mode must register the union of athlete-scoped tools allowed by at least one roster athlete (after delete-mode/toolset), then filter `tools/list` and tool calls by the active session athlete. `list_athletes` remains config-backed (`_meta.source: "config"`) until the operator-deferred upstream roster probe is validated.
- Step 4 implementation uses `coach.SelectionStore` keyed by SDK session ID, with documented process fallback when the SDK session ID is empty (stdio/in-memory transports).
- Step 5 plan decisions from R017/R018: select/catalog metadata must use `internal/mcp.safeRegistrar.visibleToolNamesForAthlete` over the post-registration tool set, the same source used for `tools/list`; tests must parse structured JSON and compare exact tool-name sets, including hidden delete/full-toolset cases and two-session isolation. The cache caveat is documented in `docs/coach-mode.md` during Step 5 and expanded in Step 6.
- Step 6 follow-up issue opened: https://github.com/ricardocabral/icuvisor/issues/13 tracks the separate PRD §7.4 #5 status update.

| 2026-05-15 20:00 | Task started | Runtime V2 lane-runner execution |
| 2026-05-15 20:00 | Step 1 started | Threat-model review + endpoint probe |
| 2026-05-15 20:07 | Review R001 | code Step 1: UNKNOWN |

| 2026-05-15 20:09 | Agent escalate | Blocked on TP-039 Step 1 after code review R001. Reviewer correctly rejected marking the coach-roster endpoint probe complete because the task requires an authenticated black-box probe with a real coa |
| 2026-05-15 20:09 | Worker iter 1 | done in 541s, tools: 51 |
| 2026-05-15 20:09 | Steering | Authenticated coach-roster probe is operator-deferred; proceed with config-backed roster and mark Step 1 complete on documented-gap/fallback basis. |
| 2026-05-15 20:10 | Review R002 | code Step 1: reviewer repeated R001 objection; superseded by supervisor steering to treat gap as complete for TP-039 v0.5. |
| 2026-05-15 20:25 | Review R006 | code Step 2: revise coach-mode configs without top-level athlete_id and strengthen registry/toolcatalog drift test. |
| 2026-05-15 20:31 | Review R007 | code Step 2: revise enabled coach mode so coach.default_athlete_id wins over legacy top-level athlete_id. |
| 2026-05-15 20:12 | Review R002 | code Step 1: UNKNOWN |
| 2026-05-15 20:15 | Review R003 | plan Step 2: REVISE |
| 2026-05-15 20:19 | Review R004 | plan Step 2: REVISE |
| 2026-05-15 20:21 | Review R005 | plan Step 2: APPROVE |
| 2026-05-15 20:32 | Review R006 | code Step 2: REVISE |
| 2026-05-15 20:37 | Review R007 | code Step 2: REVISE |
| 2026-05-15 20:41 | Review R008 | code Step 2: APPROVE |
| 2026-05-15 21:00 | Review R011 | code Step 3: revise activity-ID ownership checks, advanced-capabilities coach filtering source, and coach-mode-off catalog compatibility. |
| 2026-05-15 21:12 | Review R012 | code Step 3: revise link_activity_to_event to preflight event target ownership before PUT. |
| 2026-05-15 20:45 | Review R009 | plan Step 3: REVISE |
| 2026-05-15 20:47 | Review R010 | plan Step 3: APPROVE |
| 2026-05-15 21:04 | Review R011 | code Step 3: REVISE |
| 2026-05-15 21:18 | Review R012 | code Step 3: REVISE |
| 2026-05-15 21:26 | Review R013 | code Step 3: APPROVE |
| 2026-05-15 21:31 | Review R014 | plan Step 4: REVISE |
| 2026-05-15 21:35 | Review R015 | plan Step 4: APPROVE |
| 2026-05-15 21:49 | Review R016 | code Step 4: APPROVE |
| 2026-05-15 21:52 | Review R017 | plan Step 5: REVISE |
| 2026-05-15 21:55 | Review R018 | plan Step 5: REVISE |
| 2026-05-15 21:58 | Review R019 | plan Step 5: APPROVE |

| 2026-05-15 22:13 | Agent reply | Acknowledged. I documented the authenticated roster probe as an operator-deferred validation gap, kept upstream roster support unproven/deferred, and am proceeding with config-backed roster behavior f |
| 2026-05-15 22:13 | Worker iter 2 | killed (wall-clock timeout) in 7472s, tools: 364 |
| 2026-05-15 22:13 | Step 5 started | Catalog-cache caveat + Tests |
| 2026-05-15 22:25 | Review R020 | code Step 5: REVISE |
| 2026-05-15 22:35 | Review R021 | code Step 5: APPROVE |

| 2026-05-15 22:39 | Worker iter 3 | done in 1535s, tools: 100 |
| 2026-05-15 22:39 | Task complete | .DONE created |
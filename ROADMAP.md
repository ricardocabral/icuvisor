# Roadmap

Living document. Phases are scoped and gated, not calendared. icuvisor will not commit to calendar dates pre-launch — each phase is shippable independently. Track current progress in [GitHub Issues](https://github.com/ricardocabral/icuvisor/issues) and [Projects](https://github.com/ricardocabral/icuvisor/projects). Product scope, tool catalog, and Key Results live in the [PRD](docs/prd/PRD-icuvisor.md); this file is the authoritative phasing plan.

## v0.1 — Walking skeleton

**Goal:** end-to-end pipe from binary → MCP → intervals.icu.

- [x] Go module + project layout.
- [x] intervals.icu API client (Basic Auth, retries, structured errors).
- [x] MCP stdio transport wired up via `github.com/modelcontextprotocol/go-sdk`.
- [x] One working tool: `get_athlete_profile`, end-to-end via stdio to Claude Desktop on macOS.
- [x] Manual JSON config (no installer yet).

## v0.2 — Read path

**Goal:** prove response shaping in real conversations before adding writes. Validate that an LLM, given only icuvisor's reads, produces correct training analysis without scale or unit confusion.

- [x] All read-only tools from the catalog (PRD §7.2.C): `get_athlete_profile`, `get_fitness`, `get_best_efforts`, `get_power_curves`, `get_activities`, `get_activity_details`, `get_activity_intervals`, `get_activity_streams`, `get_activity_splits`, `get_activity_messages`, `get_extended_metrics`, `get_training_summary`, `get_wellness_data`, `get_events`, `get_event_by_id`, `get_training_plan`, `get_workout_library`, `get_workouts_in_folder`, `get_custom_items`, `get_custom_item_by_id`.
- [x] Canonical snake_case stream keys across all activities/devices (forum #118); upstream casing differences absorbed at the response boundary.
- [x] `get_extended_metrics` field set per PRD §7.2.C — running dynamics, DFA α1, W' balance, cardiac decoupling, HR drift, aerobic decoupling, zone distributions, IF/VI/polarization, TRIMP/strain/load, L/R balance, RPE/feel/session-RPE, compliance %, device name — gated by upstream availability (PRD §7.4 #4).
- Deferred: `get_planning_parameters` remains out of the registered catalog until intervals.icu exposes athlete-level periodization parameters (ramp-rate %, recovery-week cadence, taper % drop, intensity-distribution preference) through the public API; see `docs/upstream-gaps/periodization-parameters.md`.
- [x] Terse-by-default + `include_full` opt-in; auto-added debug metadata (`fetched_at`, `query_type`) stripped by default, behind `ICUVISOR_DEBUG_METADATA=true`.
- [x] Null-value keys stripped from responses before serialization (wellness in particular — N/A fields are dropped, not emitted as `null`).
- [x] Exhaustive pace-unit enum coverage (`MINS_KM`, `MINS_MILE`, `SECS_100M`, `SECS_500M`, …); unknown units degrade to `_meta.unknown_unit: true` rather than failing the call.
- [x] Distinct sleep fields surfaced separately: manual `sleepQuality` (1–4) and device-imported `sleepScore` (0–100), each with its own in-response scale label.
- [x] Wellness `_meta.provenance` (per bridged field: `source`, `native_scale`, `fetched_at`) + `_meta.stale: true` when the upstream bridge has not refreshed within 24h of the wellness date; raw native sub-fields exposed under `_native.<source>.<field>` (Polar `ans_charge`, `nightly_recharge_status`; Garmin body-battery min/max; Oura raw `sleep_score`).
- [x] `_meta.missing_fields` callout on every read tool that strips nulls — explicit list of which keys were absent for the row, so the LLM declines to infer rather than treating absence as zero.
- [x] `get_event_by_id` upstream-inconsistency handling: structured `unavailable: { reason: "upstream_inconsistency" }` when the detail endpoint 404s on an ID `get_events` just listed.
- [x] In-response scale labels on every subjective field (`feel`, `sleepQuality`, `fatigue`, `mood`, etc.) — not just tool descriptions.
- [x] Disambiguating field names in responses (`calories_burned` not `calories`; `distance_km` / `distance_mi`).
- [x] Server-side pagination for `get_activities`.
- [x] Strava-blocked-activity detection returns structured `unavailable: { reason, workaround }`.
- [x] Per-athlete unit normalization (miles vs km) from `preferred_units`, embedded in field keys / `_meta`.
- [x] Athlete-ID normalization (`i12345` / `12345`).
- [x] Timezone normalization to the athlete's configured TZ.
- [x] `_meta.server_version` in every response.
- [x] Tool-name disambiguation pass on read clusters (`get_activity_details` / `_intervals` / `_streams`); CI guard for new confusable clusters.
- [x] Tool-schema stability rules enforced in CI: additive-only on stable tools; renames/removals require a new tool name.
- [x] Manual JSON config still; stdio only.
- [x] Dogfooded solo, read-only; invited-athlete protocol/template documented and 2–3 athlete validation deferred as maintainer follow-up by TP-016 operator-approved acceptance change.

## v0.3 — Writes with safety gate

**Goal:** ship the write path in a way that an LLM cannot be social-engineered (or self-talked) into destroying data. Validate the env-var safety model end-to-end.

- [x] `ICUVISOR_DELETE_MODE` env var (`safe` default / `full` / `none`) — destructive tools are not _registered_ in modes that forbid them. No per-call `confirm: true` arguments anywhere in the catalog.
- [x] `workout_doc` write-path serializer: structured steps round-trip back to the description-string DSL on upload (intervals.icu rejects structured `workout_doc` on writes). Read → modify → write → read fidelity locked by golden-file tests.
- [x] Write tools: `add_or_update_event` (free-text `description` preserved verbatim, `workout_doc` for structured steps, `tags` supported), `add_activity_message`, `link_activity_to_event` (manual pairing for compliance scoring when auto-pair misses — forum #97), `update_wellness` (full writable field set incl. `injury`, blood pressure, blood glucose, lactate, body fat, `locked`), `update_sport_settings` (FTP, threshold HR/pace, zones; zone-definition overwrites gated by `ICUVISOR_DELETE_MODE` — forum #35).
- [x] Workout-library CRUD: `create_workout`, `update_workout`, `delete_workout` (delete gated by `ICUVISOR_DELETE_MODE`).
- [x] Event delete (`delete_event`, `delete_events_by_date_range`), activity delete, custom-item delete, sport-settings delete, gear delete — all gated by `ICUVISOR_DELETE_MODE`.
- [ ] Past-event deletion guard, orthogonal to `ICUVISOR_DELETE_MODE`: deleting events dated before "today in athlete TZ" almost always destroys planned-vs-actual history (linked activities, compliance scoring, coach annotations). Refuse by default with a structured error pointing the LLM at the historical-edit workflow (update the existing event, don't delete it); allow override only via an explicit env var (e.g. `ICUVISOR_ALLOW_PAST_EVENT_DELETE=true`). Applies to `delete_event` and `delete_events_by_date_range`. Adversarial tests must cover the case where the LLM is talked into "cleaning up old events".
- [x] Custom-item create/update.
- [x] `apply_training_plan`.
- [x] `input_examples` on complex write tools (`add_or_update_event`, `create_workout`, `create_custom_item`, `apply_training_plan`).
- [x] Adversarial test suite: prompts that attempt to talk the server into deleting in `safe` mode must fail by tool-not-found, not by user re-prompt loop.
- [x] Dogfooded against a dedicated test athlete account; no production athletes yet.

## v0.4 — Token efficiency and MCP primitives

**Goal:** validate KR5 (token efficiency) with measured deltas vs both Python references on a shared prompt set. Land the MCP primitives that move long-form content out of the per-session budget.

**Taskplane progress:** TP-030 through TP-034 completed on 2026-05-14. v0.4 implementation is complete; KR5 is partially confirmed in `docs/kr5-benchmark.md` — response-byte targets pass, while the description-token target missed by 58 tokens / 0.53 percentage points and is tracked as `TP-034-KR5-DESC-001`.

- [x] `ICUVISOR_TOOLSET` env var with `core` (default, ~17 tools) and `full` tiers.
- [x] `icuvisor_list_advanced_capabilities` tool lives in `core` for discoverability when an advanced prompt arrives.
- [x] MCP Resources: `icuvisor://workout-syntax`, `icuvisor://event-categories`, `icuvisor://custom-item-schemas`, `icuvisor://athlete-profile`. Long-form schema content moves out of inline tool descriptions.
- [x] MCP Prompts: training analysis, recovery check, weekly planning, race-week taper, coach roster triage.
- [x] Streamable HTTP transport (localhost-bound by default).
- [x] SSE transport decision for remote-client compatibility: **Path B chosen in TP-102.** icuvisor does not add legacy SSE or a generic cloudflared/ngrok recipe for ChatGPT-style remote custom connector UIs. Those UIs require a provider-reachable HTTPS endpoint, while icuvisor remains local-first with stdio and loopback Streamable HTTP. Remote ChatGPT connectors are out of scope until the vNext hosted relay or a future explicit secure-tunnel design; docs now describe this as an intentional safety boundary rather than a transport bug.
- [x] Benchmark harness: run a shared prompt set against icuvisor and Python reference servers; record per-session description tokens and median per-call response bytes. KR5 response-byte targets are confirmed; the description-token gap is documented for follow-up.

## v0.5 — Internal beta

**Goal:** validate KR1 (install success) and the coach use case on real users.

- [x] CLI OS keychain credential storage (macOS Keychain, Windows Credential Manager, libsecret) plus terminal `icuvisor setup` that stores the intervals.icu API key outside plaintext config.
- [ ] Installer/onboarding integration for keychain-backed credentials: GUI/basic setup writes to the same keychain path and never prompts users to place API keys in JSON.
- [ ] Gear read/name-resolution pass prompted by upstream gear feedback: add `get_gear_list`, surface `gear_id` and human `gear_name` in `get_activities` / `get_activity_details` when upstream exposes gear IDs, cache per athlete with an explicit refresh path, and keep `delete_gear` gated as destructive. Resolution is inline-denormalized — every activity row carries the resolved `gear_name` next to `gear_id`, not a separate lookup tool the LLM has to chain.
- [ ] HR-curve and pace-curve siblings to `get_power_curves`: add `get_hr_curves` and `get_pace_curves` so best-effort surfaces are symmetric across the three primary metrics. Same response shape, terse-by-default, same pagination/units conventions. Without these, the LLM is forced to derive curves from streams via the v0.6 analyzers when upstream already exposes them cheaply.
- [ ] Nutrition macros + calories-label clarification in wellness and activity reads: surface upstream nutrition fields (carbs, protein, fat grams) under disambiguated keys, and confirm that `calories_burned` (active) is distinct from any total-calories field upstream returns. Today the read catalog leaves nutrition implicit; the LLM either ignores it or guesses field names that don't exist.
- [ ] Verify null-stripping rule applies to write-tool responses, not only reads: `add_or_update_event`, `update_wellness`, `update_sport_settings`, workout-library writes, and custom-item writes all echo back upstream payloads with many sparse fields. Lock golden-file tests so write responses are as terse as reads.
- [ ] Per-source sleep-score scale labels in `_meta.provenance`: when the bridged source is one of Garmin / Whoop / Oura / Polar, the in-response `native_scale` must reflect that source's actual scale rather than a single canonical 0–100 label. Add fixture coverage for at least two divergent sources so the scale label is asserted, not assumed.
- [ ] Upstream-signal regression pack from the 2026-05 upstream behavior review: Strava numeric/no-`i` empty stubs from Wahoo/MyWhoosh/TrainerRoad sync chains return structured unavailable markers; event detail 404-after-list remains `upstream_inconsistency`; NOTE creates keep accepting date-only tool input while sending upstream's required local datetime payload.
- [ ] NOTE-event discoverability pass: add docs/examples showing `add_or_update_event` with `category: "NOTE"` for nutrition plans, travel logistics, daily reminders, and coach annotations, without adding a separate confusable `add_note` tool unless telemetry shows tool-selection failure.
- [ ] macOS signed installer; manual Claude Desktop / Claude Code config documentation.
- [ ] Onboarding flow (basic — full polish in v1.0): paste API key, autodetect athlete ID + timezone, "Test connection" via `get_athlete_profile`.
- [ ] Coach mode behind a feature flag, with per-athlete granular tool permissions.
- [ ] Post-update notification that tells the user to start a new conversation in their AI client when tool schemas changed.
- [ ] Dogfooded by 5–10 forum-recruited athletes, including at least one coach.

## v0.6 — Analyzers

**Goal:** ship a small, deterministic `analyze_*` / `compute_*` tool family (PRD §7.2.C "Analyzers") that the LLM activates by default instead of writing ad-hoc reduction scripts over `get_*` reads. Validate via benchmark: the same training-analysis prompts must yield correct numbers with ≥40% fewer tokens and zero raw-stream pulls on the trend / distribution / correlation shapes.

- [ ] `analysis_metric` closed enum + rejection-with-hint for unknown metrics. No free-form field arithmetic.
- [ ] MCP Resource `icuvisor://analysis-formulas` — one paragraph per canonical formula (HR drift, Pw:HR decoupling, polarization index, EF, VI, z-score) with cited source. Responses link via `_meta.formula_ref`.
- [ ] Analyzer skeleton: every tool emits `_meta.method`, `_meta.source_tools`, `_meta.n`, `_meta.missing_days`, `_meta.missing_action`, `_meta.insufficient_sample`. Golden-file locked.
- [ ] `analyze_trend`, `analyze_distribution`, `analyze_correlation`, `analyze_efforts_delta`.
- [ ] `compute_zone_time`, `compute_load_balance`, `compute_baseline`, `compute_compliance_rate`.
- [ ] `compute_activity_segment_stats` — the only analyzer that touches raw streams; gated behind the existing stream-key canonicalization tests.
- [ ] `get_fitness_projection` (pulled up from v1.x — forum thread 123739 post #49) ships with the family so projection and analysis land together.
- [ ] Activation-hint pass on every analyzer description: leads with the user-prompt shape that should trigger the tool plus an explicit "do not pull `get_*` rows and reduce them yourself" line.
- [ ] Definition-drift guard: golden-file tests pin the canonical formula for decoupling, drift, polarization, EF, VI. Renaming or redefining is a breaking change, not a silent fix.
- [ ] Toolset placement: family lands in `full` by default; `analyze_trend`, `compute_zone_time`, `compute_baseline` promoted to `core` only after the KR5 benchmark confirms net token savings on the trend / distribution / baseline prompt shapes vs the fetch-and-reduce baseline.
- [ ] Upstream-coverage audit: measure across v0.2 fixture set how often `compute_zone_time` and `compute_load_balance` can use pre-computed per-activity zone times vs falling back to stream math. If stream-math fallback exceeds an agreed threshold, file an intervals.icu API feature request and document the gap in `docs/upstream-gaps/`.
- [ ] Benchmark harness extended (from v0.4): same training-analysis prompt set, with-and-without the analyzer family, measuring tokens and stream-pull counts.

## v1.0 — Public launch

**Goal:** hit KR2 (adoption), KR3 (coverage), KR4 (reliability), and KR6 (client compatibility).

- [ ] Signed installers across platforms:
  - macOS: `.dmg` + Homebrew tap.
  - Windows: `.msi` + Scoop bucket + Winget manifest.
  - Linux: `.deb` + `.rpm` + shell installer.
- [ ] Auto-update via signed releases (opt-out). Post-update notification instructs the user to start a new conversation in their AI client when tool schemas changed, since MCP clients cache the catalog per conversation.
- [ ] DXT bundle for Claude Desktop where supported.
- [ ] Onboarding UI with one-click client config for: Claude Desktop, Claude Code, Claude Cowork, ChatGPT Developer Mode (instructions), Pi.dev, Cursor, Continue, Zed.
- [ ] Documented manual config for any MCP client.
- [ ] Keychain-backed credential path exercised by signed installers and one-click onboarding on all platforms.
- [ ] Opt-in anonymous telemetry (install success, tool call counts; no payloads).
- [ ] Public website at `icuvisor.dev` with download, docs, troubleshooting, and a link to the intervals.icu forum thread.
- [ ] Announcement on the intervals.icu forum thread.

## v1.x — Iterate

- [ ] Local-LLM client recipes (ollmcp, Cline, LM Studio).
- [ ] Diagnostics export button in tray menu.
- [ ] Telemetry-driven response-shape tuning.
- [ ] Strength training and training plan endpoints (depends on PRD assumptions §7.4.3 / §7.4.4).

## vNext — Future (out of scope for v1)

- **Optional hosted relay** (icuvisor cloud, opt-in, BYO key): for mobile-only athletes who can't run a desktop binary. Same code path; the binary runs in our infra and authenticates via a token. Mobile access is a dominant reason athletes pay competing hosted servers, so this may pull forward into v1.x pending PRD §7.4 #8 validation.
- **Strava / TrainingPeaks** companion MCP servers in the same family.
- **Workout templates** library, AI-generated and athlete-curated.
- **Conversation memory** export hooks (Claude Projects integration).
- **Multi-sport / triathlon structured workout files**: surface upstream's triathlon workout-file resources with category (Bike/Run/Swim), metric, and sub-category filters as a dedicated read tool (e.g. `get_triathlon_workout_files`). Today's `workout_doc` DSL is single-discipline, so brick sessions and triathlon plan templates round-trip lossily. Depends on the v0.3 workout-library CRUD and likely some round-trip work in `internal/workoutdoc/` to represent a sequence of discipline-tagged blocks. Worth scoping against the v0.6 analyzer family so multi-sport compliance/zone-time computations don't fork the schema later.
- **Documented self-hosted remote recipe** as an interim before the hosted relay: a `docs/deploy/` recipe for running icuvisor on Fly / Render / a small VM behind a reverse proxy with auth, intended for athletes who want phone access from their AI client today. Explicitly NOT a supported product — same binary, same code path, user-operated. Decision is whether to publish the recipe (cheap, accelerates feedback) or hold the line that mobile access waits for the opt-in relay. Revisit once SSE-transport decision lands.
- **`fill_calendar_from_library`** ("Plan Filler", forum thread 123739 post #24): given a date range, target weekly load (TSS or hours), available hours per weekday, and a workout-library folder filter, assign existing library workouts to days to hit the target. Returns the proposed schedule for review; commit is a separate explicit call. Depends on workout-library CRUD (v0.3) and `apply_training_plan`.

## Out of scope

- Replacing intervals.icu's own UI.
- Becoming a multi-tenant SaaS for primary use.
- Hosting athlete data on our infrastructure outside the future opt-in relay.
- Non-intervals.icu data sources as first-party features (athletes can install other MCP servers alongside icuvisor).
- Mobile-only installs at launch — desktop only for v1; mobile is served via the user's desktop or the future hosted relay.

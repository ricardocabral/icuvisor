# Roadmap

Living document. Phases are scoped and gated, not calendared. icuvisor will not commit to calendar dates pre-launch — each phase is shippable independently. Track current progress in [GitHub Issues](https://github.com/ricardocabral/icuvisor/issues) and [Projects](https://github.com/ricardocabral/icuvisor/projects). Product scope, tool catalog, and Key Results live in the [PRD](docs/prd/PRD-icuvisor.md); this file is the authoritative phasing plan.

## v0.2.0

**Goal:** validate KR1 (install success) and the coach use case on real users.

- [ ] `get_today` daily-digest tool: a single terse-by-default read that returns today's fitness (CTL/ATL/TSB), wellness, completed activities, planned events, and any race or NOTE-category annotation in one call, so the LLM answers "how's today looking?" without chaining `get_fitness` + `get_wellness_data` + `get_activities` + `get_events`. Counts against KR5 (token efficiency): one round trip instead of four. Honours the terse / `include_full` split — the digest stays terse, `include_full` widens per-section detail. Complements the existing "recovery check" MCP prompt rather than replacing it (a prompt is not a tool). Each digest section reuses the response shaping already proven on its source tool so scale labels, unit normalization, and null-stripping do not fork.
- [ ] `_meta.as_of` anchor on time-relative reads: every read whose result depends on "now" — `get_today` plus the current-day paths of `get_wellness_data`, `get_activities`, and `get_events` — carries `_meta.as_of`, the current datetime in the athlete's configured timezone, so the LLM does not mis-reason about partial-day data: morning wellness reflects last night's sleep, accumulated training load is so-far-today, and planned event times may already have passed. Additive `_meta` only; no schema break on stable v0.2 tools.
- [ ] Write tools for completed activities: `update_activity` (rename and/or edit the description of a completed activity) and `set_activity_intervals` (set or correct the interval structure on a completed activity). The current write catalog can annotate (`add_activity_message`) and pair (`link_activity_to_event`) a completed activity but cannot rename or re-describe one, and there is no write counterpart to `get_activity_intervals`. Both gated by `ICUVISOR_DELETE_MODE` like other writes; `set_activity_intervals` overwriting an existing interval set is treated as destructive. Scope `set_activity_intervals` against the v0.6 `_meta.interval_source` / auto-lap disambiguation work so a manually written interval set stays distinguishable from device laps downstream.
- [x] macOS signed installer; manual Claude Desktop / Claude Code config documentation.
- [x] Onboarding flow (basic — full polish in v1.0): paste API key, autodetect athlete ID + timezone, "Test connection" via `get_athlete_profile`.
- [ ] Coach mode behind a feature flag, with per-athlete granular tool permissions.
- [ ] Post-update notification that tells the user to start a new conversation in their AI client when tool schemas changed.
- [ ] Dogfooded by 5–10 forum-recruited athletes, including at least one coach.

## v1.0 — First stable release

**Goal:** hit KR2 (adoption), KR3 (coverage), KR4 (reliability), and KR6 (client compatibility).

- [ ] Signed installers across platforms:
  - macOS: `.dmg` + Homebrew tap.
  - Windows: `.msi` + Scoop bucket + Winget manifest.
  - Linux: `.deb` + `.rpm` + shell installer.
- [ ] Auto-update via signed releases (opt-out). Post-update notification instructs the user to start a new conversation in their AI client when tool schemas changed, since MCP clients cache the catalog per conversation.
- [ ] Onboarding UI with one-click client config for: Claude Desktop, Claude Code, Claude Cowork, ChatGPT Developer Mode (instructions), Pi.dev, Cursor, Continue, Zed.
- [ ] Documented manual config for any MCP client.
- [ ] Keychain-backed credential path exercised by signed installers and one-click onboarding on all platforms.
- [ ] Opt-in anonymous telemetry (install success, tool call counts; no payloads).

## v1.x — Iterate

- [ ] Local-LLM client recipes (ollmcp, Cline, LM Studio).
- [ ] Diagnostics export button in tray menu.
- [ ] Telemetry-driven response-shape tuning.
- [ ] Strength training and training plan endpoints (depends on PRD assumptions §7.4.3 / §7.4.4).

## v2.x 

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

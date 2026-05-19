# PRD — icuvisor

> An open-source, locally-installed MCP connector for [intervals.icu](https://intervals.icu), distributed as a single Go binary. Designed for non-technical amateur athletes who want to talk to their training data through Claude, ChatGPT, Pi, and other mainstream AI tools.

---

## 1. Summary

icuvisor is a free, open-source Model Context Protocol (MCP) server that connects intervals.icu training data to AI assistants. Unlike Python-based alternatives, icuvisor ships as a single self-contained binary with one-click installers so athletes — not engineers — can install it, point Claude or ChatGPT at it, and start asking questions like _"should I increase my FTP?"_ within five minutes.

---

## 2. Background

### Context

Amateur endurance athletes increasingly want to use general-purpose LLMs as a personal coach: analyze recent rides, suggest tomorrow's workout, plan a training block toward a goal event, reflect on wellness trends. intervals.icu is the platform of choice for serious amateurs because it exposes a clean public API covering activities, wellness, events, fitness (CTL/ATL/TSB), and custom fields.

Two paths exist today to bridge intervals.icu and AI:

1. **Open-source Python MCP servers**. Free and capable, but installation commonly requires Python tooling plus hand-edited JSON config with platform-specific absolute paths. The forum thread is dominated by install failures: `spawn uv ENOENT`, hatchling wheel build errors, `.env` confusion, Python version mismatches. The largest MIT option, **`hhopke/intervals-icu-mcp`**, offers a broad tool surface and an easier `uvx intervals-icu-mcp` install path, but still requires the user to install `uv` and hand-edit a JSON file.
2. **icusync.icu** — closed-source, hosted, account-required, opaque pricing. Solves the install problem by hosting everything, but requires trusting a third party with a token to your training data and offers no transparent free tier.

None of these serves the "I just want this to work" athlete who doesn't want to learn `uv` _or_ hand their data to another SaaS.

The broader AI-for-intervals.icu market segments into two buckets athletes already recognize — **"app-style coaches"** (CoachWatts, IntervalCoach, AIEndurance, Intervals.pro, LeCoach, MyTrainPal, PlanWatts) that wrap a model behind a chat UI with bundled credits, and **bespoke prompt frameworks** like the "Section 11" deterministic protocol that ask the athlete to paste data dumps into a generic chatbot (forum thread 123739 posts #7, #8, #16, #36, #59). icuvisor occupies an unoccupied **third bucket**: a protocol/MCP layer that lets any MCP-aware assistant (Claude, ChatGPT, Pi, Cursor, local LLMs) read and write intervals.icu data directly, with the athlete's own model subscription. The positioning is "infrastructure, not a product" — the differentiation is data quality, schema discipline, install simplicity, and BYO-key economics, not a coaching opinion baked into a wrapper.

### Why now

- **MCP momentum**: stable spec, Claude Desktop/Code/Cowork, ChatGPT Developer Mode, Cursor, Pi, Le Chat, and local LLM clients all support MCP servers.
- **Existing Python options have known limits**: Python/uv setup still creates install friction, and open-source issue backlogs target exactly the problems icuvisor is best positioned to solve from day one: model-uncontrollable delete safety, tiered toolsets to cut per-session token cost, in-response scale labels, debug-metadata stripping, and tool-name disambiguation.
- **Distribution gap is now solvable**: Go's single-binary cross-compilation + Homebrew/Scoop/Winget/DXT bundles let us deliver `brew install icuvisor` to a triathlete who has never opened a terminal.
- **Recent intervals.icu API additions** (custom wellness fields, activity messages, structured workout endpoints) make a richer feature set possible than what the original README documents.

---

## 3. Objective

### What and why

Make AI-powered training analysis accessible to **any amateur athlete** with an intervals.icu account, in **under five minutes**, with **zero terminal commands** for the happy path, and **zero recurring cost**. The intervals.icu community deserves a high-quality, open, local-first option that doesn't lock data behind someone else's login.

### How it benefits the company and customers

There is no company. icuvisor is community infrastructure. Benefits to users:

- **Athletes**: free, private (data flows athlete → local binary → AI client; no third-party server in the path), works with whichever AI tool they already pay for.
- **Coaches**: multi-athlete support without each athlete signing up for a SaaS.
- **Developers**: clean Go reference implementation of an MCP server, easy to fork or vendor.

### Alignment with vision and strategy

Vision: _training data should belong to athletes and run on athletes' machines, not vendors' servers_. Strategy: own the install/UX layer (where competitors lose users) while staying API-compatible with intervals.icu's evolving public surface.

### Key Results (12-month, SMART)

- **KR1 — Install success**: ≥95% of new users complete install + first successful tool call within 10 minutes, measured by opt-in anonymous telemetry on the welcome page.
- **KR2 — Adoption**: 2,000 weekly-active installs by month 12 (measured by opt-in update-check pings).
- **KR3 — Coverage**: feature parity with the leading reference surfaces — at least 90% of the deduplicated union of open-source Python MCP capabilities and icusync.icu's advertised tools. Gaps must be deliberate (e.g. dropped on safety grounds), not accidental.
- **KR4 — Reliability**: <1% of tool calls return uncaught errors; p95 latency <2s for read tools (excluding upstream API time).
- **KR5 — Token efficiency**: ≥60% reduction in per-session tool-description tokens vs. `hhopke/intervals-icu-mcp`'s default 58-tool surface (achieved via toolset tiers, description trimming, and MCP Resources for long-form schemas); ≥40% reduction in median per-tool-call response bytes vs. both Python references on the same prompts.
- **KR6 — Client compatibility**: validated working configs for Claude Desktop, Claude Code, Claude Cowork, ChatGPT (Dev Mode), Pi.dev, Cursor, Continue, Zed, and one local-LLM client (ollmcp or Cline).

---

## 4. Market Segments

Markets defined by the job the user is trying to do, not by demographics.

### Primary — "Curious amateur athlete"

> _"I have a Garmin, I log everything to intervals.icu, and I pay for Claude Pro. I want to ask Claude about my training. I am not a developer."_

- Job: get coaching-quality answers from their own data, without learning new tools.
- Constraint: will not run `pip install`, will not edit JSON, will not host a server. Maximum acceptable friction: download a `.dmg`/`.exe`, paste an API key, click "Connect to Claude."
- Estimated size: tens of thousands. The intervals.icu forum thread alone surfaced 50+ such users.

### Secondary — "Coach with a roster"

> _"I have 8–25 athletes on intervals.icu. I want to triage their weeks and plan next week's workouts inside Claude."_

- Job: review multiple athletes from one place; create/edit workouts on each athlete's calendar.
- Constraint: must support athlete-scoped credential delegation safely (issue #88, forum posts #18/#21/#60).

### Tertiary — "Self-experimenting power user / developer"

> _"I want to script analyses, build my own agent, run a local LLM, mix intervals.icu data with MyFitnessPal/Strava."_

- Job: programmatic access, multiple transports, scripting from CLI or notebooks.
- Constraint: needs documented JSON-RPC schema, headless mode, local HTTP transport, ability to vendor as a library.

### Out of scope (initially)

- Non-intervals.icu data sources (Strava direct, TrainingPeaks). The athlete should add other MCP servers alongside icuvisor.
- Mobile-only installs. We will ship for macOS, Windows, Linux; mobile clients connect via the user's desktop or via the optional hosted relay (see §8 Future Versions).

---

## 5. Value Propositions

### Customer jobs / needs

1. _"Analyze my last N activities and tell me what to do next."_
2. _"Plan a training block / taper toward race day X."_
3. _"Push tomorrow's workout to my calendar so it syncs to my Garmin."_
4. _"Reflect on my wellness trends — sleep, HRV, RHR."_
5. _"Coach me on this athlete I'm responsible for."_

### Gains

- **Speed**: 5-minute setup, no documentation deep-dive.
- **Privacy**: API key and data never leave the athlete's machine.
- **Cost**: $0, forever.
- **Choice of AI**: not locked to Claude. Works on whichever assistant the athlete already uses.
- **Up-to-date**: automatic background update of the binary (opt-out), so new intervals.icu API endpoints land without re-following install docs.

### Pains avoided

- Python/uv/hatchling install failures (forum posts #4, #12, #13, #19, #30, #31, #35 + issues #5, #23).
- Conversation-killing context-window blowouts (issue #89, forum #28, #66).
- Wrong scale ranges in LLM context (sleep 1-4, feel 1-5; issues #45, #48, forum #54, #57).
- Trusting a third-party SaaS with intervals.icu credentials (icusync.icu trust model).
- Per-athlete SaaS signup for coaches.
- Timezone drift (issue #49 / forum #49).
- Silent overwriting of athlete/coach free-text workout descriptions by normalization into structured blocks.
- Unit-system mismatch — athlete uses miles but assistant replies in km.
- Confusing "fix didn't land" experiences after server upgrades, caused by MCP clients caching the tool schema per conversation.

---

## 6. Solution

### 6.1 UX / User flows

**Flow A — First-time install (the golden path)**

1. Athlete visits `icuvisor.dev` and clicks the platform-detected download button.
2. macOS: opens signed/notarized `.dmg` → drag to Applications. Windows: signed `.msi`. Linux: `.deb`/`.rpm` or shell installer. Power users: `brew install icuvisor` / `scoop install icuvisor` / `winget install icuvisor`.
3. First launch opens a small native onboarding window (or a localhost page in the default browser):
   - Step 1: "Paste your intervals.icu API key" with a clickable link to `https://intervals.icu/settings` and a screenshot.
   - Step 2: detects athlete ID from the API key — falls back to manual entry, accepting both `i12345` and `12345` (issue #40).
   - Step 3: timezone autodetected from OS, editable (issue #49).
   - Step 4: pick AI client(s). Each option shows a "Set up automatically" button that writes the appropriate config file _and_ a "Show manual config" disclosure for users who prefer it. Supported targets at launch: Claude Desktop, Claude Code, Claude Cowork, ChatGPT (Dev Mode instructions), Pi.dev, Cursor, Continue, Zed.
4. "Test connection" button calls `get_athlete_profile` and shows the athlete's name + FTP. ✅
5. Onboarding closes; a menu-bar / system-tray icon stays running.

**Flow B — Asking a question (the use case)**

User opens Claude Desktop and types _"Analyze my last 10 cycling activities and let me know if I should adjust my FTP."_ Claude calls `get_activities` (terse mode), then `get_activity_intervals` for each, then `get_athlete_profile` for current FTP, then replies. icuvisor's terse-by-default responses keep this under one context window even on free Claude tier (addressing forum #65, #66).

**Flow C — Update**

icuvisor checks `releases.icuvisor.dev` once per day. If a new signed release exists, the tray icon shows a dot; clicking "Update now" replaces the binary and restarts. No terminal commands. Opt-out in settings.

After an update that adds or changes tool arguments, the post-update notification explicitly tells the user to **start a new conversation in their AI client** to pick up the new tool schema. MCP clients (Claude in particular) cache the tool catalog at conversation start, so an in-flight chat will keep using the old schema and report "the fix didn't work."

**Flow D — Coach mode**

Coach pastes a coach-scoped intervals.icu API key. icuvisor lists athletes via `list_athletes`; the coach selects which subset is exposed to tools. The active athlete is passed as a tool argument (`athlete_id`) on every call, with a configurable default. Mirrors issue #88 and forum posts #18/#21/#60.

The coach also picks, **per athlete**, which tools are exposed — e.g. read-only access for a prospective athlete, full read+write for an active client. Granular per-tool permissions are enforced in the server before any intervals.icu call; the LLM never sees disallowed tools in its catalog.

Wireframes will be produced separately; this PRD specifies behavior only.

### 6.2 Key Features

#### A. Distribution

- **Single Go binary**, cross-compiled for macOS (arm64 + amd64, universal), Windows (amd64 + arm64), Linux (amd64 + arm64).
- **Signed and notarized** on macOS; Authenticode-signed on Windows.
- **Installers**: `.dmg`, `.msi`, `.deb`, `.rpm`, plus Homebrew tap, Scoop bucket, Winget manifest.
- **DXT bundle** for Claude Desktop's one-click extension install where supported.
- **Auto-update** with signed-release verification.
- **Reproducible builds** via GoReleaser + GitHub Actions.

#### B. MCP transports

- **stdio** — default; works with all current MCP clients.
- **Streamable HTTP** — bound to `127.0.0.1` by default, optional LAN binding for power users. Required for clients that prefer HTTP (and a future hosted-relay story).
- **No SSE** — deprecated in the MCP spec; not implemented.

#### C. Tool catalog (target launch set)

Union of upstream tool sets, deduplicated, with names harmonized. Each tool ships with a **terse default response** (≤500 tokens typical) and an `include_full: bool` parameter for full payload.

**Athlete & fitness**

- `get_athlete_profile` — FTP, zones, sport settings, thresholds.
- `update_sport_settings` — write FTP, threshold HR, threshold pace, and per-sport zone definitions back to the athlete profile (forum #35 — assistants offered FTP updates that didn't land because there was no write path). Per-sport scoped (`sport: "Ride" | "Run" | …`). Gated by `ICUVISOR_DELETE_MODE` like other destructive writes when used to overwrite zone definitions, since a bad zone overwrite silently miscolours every historical activity.
- `list_athletes`, `select_athlete` — coach mode.
- `get_fitness` — CTL/ATL/TSB trends, taper projections.
- `get_best_efforts` — PRs across sports.
- `get_power_curves` — mean-maximal curves.

**Activities**

- `get_activities` — date-range list; supports `include_unnamed` (issue #67) and pagination.
- `get_activity_details` — single-activity metadata, zones, metrics.
- `get_activity_intervals` — interval splits.
- `get_activity_streams` — time-series (power, HR, altitude, cadence, etc.). **Stream keys are canonicalized** to a single naming convention (snake_case) across activities and devices; the upstream API exposes inconsistent casing (camelCase on some activity types, snake_case on others — forum #118) and the LLM must not have to guess.
- `get_activity_splits` — virtual splits (per-km or per-mile) computed from streams when the activity has no manual laps, so continuous runs/rides are analyzable (forum #25 / #29). Honours `preferred_units`.
- `get_activity_messages` — fetch comments/notes.
- `add_activity_message` — post a comment (forum #99).
- `link_activity_to_event` — pair a completed activity with its planned event so compliance/adherence is computed against the right target (forum #97). Where intervals.icu auto-pairs by date+type, this tool is a manual override for the cases auto-pairing misses.
- `get_extended_metrics` — second-order metrics not exposed on the base activity payload. Target field set (subject to upstream availability per §7.4 #4): running dynamics (GCT, vertical oscillation, stride length, GCT balance), DFA α1, W' balance, core temp, cardiac decoupling (Pw:HR), HR drift %, aerobic decoupling, power-zone distribution, pace-zone time, cadence-by-zone, joules above FTP, intensity factor, variability index, polarization index, TRIMP, strain score, HR/pace/power load, left/right balance, RPE / feel / session-RPE, compliance %, device name (forum #62, #70).
- `get_training_summary` — aggregated volume/TSS/zones.

**Wellness**

- `get_wellness_data` — daily rows. **Includes custom fields** (issue #64, forum #92) and correct scale metadata embedded in the tool description **and the response itself** (`feel` is 1-5, `sleepQuality` is 1-4 — addresses issues #45/#48 and forum #54/#57). In-response labels are required because some MCP clients do not pass tool descriptions back to the LLM at inference time. **Sleep dual-scale handling**: the intervals.icu wellness payload exposes two separate sleep fields with different scales and provenance:
  - `sleepQuality` — integer 1–4 (1 = poor, 4 = great), athlete-entered, subjective.
  - `sleepScore` — integer 0–100, populated by device sync (Garmin Connect, Oura, Whoop, Apple Health) when the device computes a nightly sleep score; absent otherwise.
  - `sleepSecs` — integer seconds slept (separate from both quality fields; do not derive from `sleepScore`).

  Both quality fields are surfaced under distinct keys (no aliasing, no `sleep_rating: 3` collapse) with `_meta.scales` entries:

  ```json
  "_meta": { "scales": { "sleepQuality": "1-4 (athlete-entered, 1=poor 4=great)",
                          "sleepScore":   "0-100 (device-imported nightly score)" } }
  ```

  On `update_wellness`, only `sleepQuality` is writable — `sleepScore` is device-owned and must be rejected with a clear error if submitted (return `field_not_writable: sleepScore (device-managed)` rather than silently dropping). Raw `_native` provider sidecars are bridge-managed and must likewise reject writes with `field_not_writable: _native (bridge-managed)`.

- `update_wellness` — write back the full set of API-accepted fields: subjective scales (`feel`, `fatigue`, `soreness`, `stress`, `mood`, `motivation`, `sleepQuality`, `injury`), body metrics (`weight`, `bodyFat`, `abdomen`), cardiovascular (`restingHR`, `hrv`, `systolic`, `diastolic`), blood/lab (`bloodGlucose`, `lactate`, `spO2`, `vo2max`), respiration, menstrual phase, and the `locked` flag that prevents device sync from silently overwriting manual entries.

**Wellness provenance + freshness** (forum thread 123739, posts #56–#58): intervals.icu collapses upstream-bridged wellness fields (notably `Readiness`) across providers with different native scales — Polar's 1–6 readiness, Garmin's body battery, Oura's readiness — under one field name with no provenance flag, and Polar-bridged values only refresh when the athlete visits intervals.icu's home page in a browser (so an MCP read can return data hours or days stale). icuvisor surfaces both signals so the LLM does not silently compare incompatible scales or reason over stale numbers:

- Every bridged wellness field carries `_meta.provenance` per field: `{ "readiness": { "source": "polar", "native_scale": "1-6", "fetched_at": "<RFC3339>" } }`. `source` is one of `polar | garmin | oura | whoop | apple_health | manual | unknown`.
- Where the upstream exposes raw native sub-fields (Polar's `ans_charge` -10..+10, `sleep_charge`, `nightly_recharge_status` 1–6; Garmin's body-battery min/max; Oura's raw `sleep_score` 1–100), surface them under `_native.<source>.<field>` so the LLM can choose between the normalized and the raw view.
- A wellness row whose `fetched_at` is more than 24h older than the wellness `date` (i.e. the bridge has not refreshed) gets `_meta.stale: true` with a one-line `_meta.stale_reason` (`"polar bridge refresh requires user to open intervals.icu"`).
- When provenance cannot be determined, emit `_meta.provenance.<field>.source: "unknown"` rather than omitting the marker — silence reads as "trusted" to a model.

**Events & workouts**

- `get_events`, `get_event_by_id` — calendar entries.
- `add_or_update_event` — structured workout, race, or note. Returns a **terse** confirmation by default (issue #89). Preserves intervals.icu's distinction between `description` (free text — athlete/coach notes, pacing, nutrition, race countdown) and `workout_doc` (structured steps). On edit, `description` is written through **verbatim** unless the caller explicitly opts into structured normalization; `workout_doc` is the only field that accepts structured-block syntax. Silent normalization of free text is treated as a destructive operation and must not happen by default. Accepts a `tags` array (e.g. `["sweet-spot", "indoor"]`) which round-trips through reads. **Upload-asymmetry note**: intervals.icu emits structured `workout_doc` JSON on reads (an object with `steps[]`, where each step has fields like `duration`, `power` / `hr` / `pace`, optional `reps` for repeat blocks, optional nested `steps[]` for the body of a repeat), but the **write endpoints silently ignore `workout_doc`** — the structured steps must be re-encoded into intervals.icu's plain-text workout DSL and submitted in the `description` field. icuvisor owns this serialization so callers pass structured steps and never see the asymmetry. Implementation requirements:
  - **Tool input** accepts either a structured `steps[]` array _or_ a raw DSL `description` string; both must not be set simultaneously.
  - **On write**: if `steps[]` is provided, the server serializes it to the DSL and submits as `description`; `workout_doc` is not sent. If a free-text `description` is provided, it is submitted **verbatim** with zero normalization (silent reformatting of athlete/coach notes is a destructive op).
  - **DSL serializer** must handle: warmup/cooldown blocks, steady-state intervals (`duration` + single target), repeats (`Nx { ... }`), ramps (start→end target), free-form rest, and target types: % FTP, watts, % threshold HR, bpm, % threshold pace, pace (`mm:ss/km` or `mm:ss/mi` per athlete `preferred_units`), zone references (`Z1`–`Z7`), and RPE. Durations in seconds, minutes, or hours. The canonical encoding is locked behind golden-file tests so subtle reformatting drift (whitespace, repeat syntax, target ordering) cannot regress silently.
  - **Round-trip test**: for every fixture in `testdata/workouts/`, the pipeline `parse(workout_doc) → serialize → submit → re-fetch → re-parse` must yield a structurally identical `steps[]` (modulo documented lossy fields, which must be enumerated in `internal/tools/workout_dsl.go`).
  - **Lossy fields** (target types or step features the DSL cannot round-trip) are surfaced to the caller as a `_meta.lossy_fields: [...]` warning on the write response rather than silently dropped.
- `delete_event`, `delete_events_by_date_range` — destructive. **Gated by `ICUVISOR_DELETE_MODE` env var, not a tool argument** (see §7.2.D — model-controlled `confirm: true` flags are not a credible safety guard).
- `get_training_plan` — fetch plan (forum #70).
- `apply_training_plan` — instantiate a workout-library folder onto the calendar from a start date.
- _Strength training data_ — included if the intervals.icu API exposes it (forum #70).

**Workout library (templates, distinct from calendar events)**

- `get_workout_library`, `get_workouts_in_folder` — read templates by folder.
- `create_workout`, `update_workout`, `delete_workout` — author/edit templates via MCP so a coach can ask the LLM to build a reusable training block without leaving the chat. `delete_workout` is destructive and gated by `ICUVISOR_DELETE_MODE` like event deletion.

**Custom items**

- `get_custom_items`, `get_custom_item_by_id`, `create_custom_item`, `update_custom_item`, `delete_custom_item` — for custom charts/fields/zones. Long-form schema documentation for the inner `content` shape (which varies per `item_type`) lives in an MCP Resource (`icuvisor://custom-item-schemas`), not inline in the tool description (see §7.2.G).

**Planned analyzers (`analyze_*` / `compute_*`) — v0.6 roadmap scope**

The analyzer family is planned roadmap scope, not part of the current generated MCP tool catalog until the v0.6 phase lands. It is a small, deterministic family of derivation tools so the LLM reaches for a documented primitive instead of fetching `get_*` rows and writing an ad-hoc reduction in chat. Every analyzer aggregates from existing reads first and only falls back to stream math when intervals.icu has no windowed view of the field. The split is intentional: `analyze_*` is inferential (baseline, fit, correlation), `compute_*` is deterministic aggregation (sums, counts, time-in-zone).

Design rules that apply to every tool in this family:

- **Closed `analysis_metric` enum.** Inputs accept only known field names (`ctl`, `atl`, `tsb`, `ramp`, `weekly_tss`, `weekly_hours`, `rhr`, `hrv`, `weight`, `sleep_secs`, `sleep_quality`, `feel`, `fatigue`, `pace_at_lt2`, `power_at_lt2`, `if`, `vi`, `np`, `compliance_pct`, …) mirrored from existing read tools. Unknown metric returns a one-line hint pointing at the correct analyzer (e.g. "try `analyze_efforts_delta` for best-effort durations"), never a free-form math input. No ad-hoc field arithmetic — derived metrics enter the enum with a registered formula or they don't ship.
- **Mandatory `_meta.method`.** Every response carries the exact formula used (`"rolling_mean_7d / baseline_28d"`, `"pearson_r over n=42 daily pairs"`) so the LLM explains the calculation rather than narrating the chart from training data.
- **Mandatory `_meta.source_tools`.** Lists which `get_*` reads the analyzer ran, so the LLM and a debugging human can trace the result.
- **Formula registry resource.** Long-form definitions for HR drift, Pw:HR decoupling, polarization index, EF, VI, z-score, etc. live in `icuvisor://analysis-formulas` (one paragraph each, citation to canonical source — Friel / Seiler / Coggan). Responses link via `_meta.formula_ref` rather than restating the math inline. Same pattern as `icuvisor://workout-syntax` in §7.2.G.
- **No silent imputation.** Missing days are skipped and counted: `_meta.missing_days: 3`, `_meta.missing_action: "skip"`. Never forward-fill without an explicit caller opt-in.
- **Boundary-safe defaults.** Minimum sample sizes (`n>=14` for correlation, `n>=7` for baseline) with `_meta.insufficient_sample: true` instead of returning a garbage `r` or slope.
- **Terse by default.** Headline numbers in `summary`; per-bucket `series[]` only when `include_full: true`.
- **Sport and unit normalization** flow through every analyzer (athlete `preferred_units`, configured timezone — §7.2.D rules apply unchanged).
- **Definitions are locked.** Two coaches will disagree on the canonical formula for decoupling or polarization; we publish ours and pin it with golden-file tests. Definition drift is a breaking change, not a silent improvement.

Tool catalog:

- `analyze_trend` — rolling mean, slope, Δ%, week-over-week change for a metric over a window vs a baseline window. Use when the user asks whether something is improving, worsening, or flat (CTL ramp, weekly TSS, pace-at-LT2, weight, HRV).
- `analyze_distribution` — histogram / quantiles / time-in-zone for a numeric field over a window. Use for "how is X spread" / "how polarized was this block".
- `analyze_correlation` — Pearson r, Spearman ρ, n, slope, intercept for two daily or per-activity fields, with optional `lag_days` for lagged correlation (sleep quality → next-day RPE).
- `analyze_efforts_delta` — best-effort durations or distances (5-min power, 20-min power, 5k pace) current window vs baseline window, unit-aware, with Δ%.
- `compute_zone_time` — sum of time per power / HR / pace zone over a window, sport-filtered, with polarization index. Aggregates per-activity zone times from `get_activity_intervals` / `get_extended_metrics` — does not recompute from streams when upstream zone time is present.
- `compute_load_balance` — share of time in Z1+Z2 / Z3 / Z4+ across a window; classifies the block (`polarized` / `pyramidal` / `threshold`).
- `compute_baseline` — rolling baseline (mean, std), current-window value, z-score, and "suppressed" / "elevated" flag for wellness metrics (the HRV-deviation primitive used by HRV4Training-style readiness checks).
- `compute_activity_segment_stats` — within-activity stream math: mean / median / p90 / decoupling / drift / NP / IF over a specified time or distance range inside one activity, pulled from `get_activity_streams`. The only analyzer that touches raw streams by default.
- `compute_compliance_rate` — scheduled vs completed events across a window, mean delta to target, per sport / event type. Reuses `link_activity_to_event` pairings.
- `get_fitness_projection` — forward CTL/ATL/TSB simulation given a hypothetical ramp %, recovery-week cadence, and date horizon. Lives in this family (moved up from v1.x — forum thread 123739 post #49). Returns projected curve plus modeled assumptions in `_meta` so the LLM can explain the result.

**Activation hint pattern.** Tool descriptions lead with the user-prompt shape that should trigger the tool and an explicit "do not roll your own" line, e.g.:

> `analyze_trend` — Use this when the user asks whether a metric is improving, getting worse, or unchanged over a window. Returns rolling mean, slope, and % change against a baseline, with the exact formula in `_meta`. **Do not** pull `get_activities` / `get_fitness` rows and reduce them yourself — this tool is the supported path, applies athlete unit/timezone normalization, and reports sample size and missing-days explicitly.

Combined with the v0.4 MCP Prompt set (`training analysis`, `recovery check`, `weekly planning` — §7.2.G), this keeps the LLM out of the "fetch a hundred activities and write a Python sum in chat" loop.

**Non-goals for this family:**

- No silent imputation of missing days (skip-and-report only).
- No competing physiology models — we aggregate intervals.icu's existing per-activity calculations and do not introduce alternative definitions of TSS, IF, polarization, decoupling, etc.
- No multi-athlete aggregation in v1; coach mode reuses single-athlete analyzers per athlete.
- No free-form field arithmetic from the LLM. If a derived field becomes common, it enters the `analysis_metric` enum with a registered `formula_ref`.

Toolset placement: the family lands in `full` by default; `analyze_trend`, `compute_zone_time`, and `compute_baseline` are promoted to `core` after the KR5 benchmark confirms net token savings vs the fetch-and-reduce baseline.

The generated tool catalog is the source of truth for the current registered tool count. Once the planned v0.6 analyzer family lands, the `core` toolset (see §7.2.D) exposes a curated subset by default; the full surface ships behind an opt-in env var.

#### D. Response shaping (the second differentiator)

- **Terse-by-default**: every read tool returns the smallest useful payload. Heavy fields (streams, raw samples) require explicit opt-in.
- **Null stripping**: keys whose value is JSON `null` are dropped from responses before serialization. Applied at the response-shaping boundary (after upstream decode, before MCP serialization), recursively through nested objects and arrays-of-objects. Wellness rows in particular are dominated by N/A fields when devices haven't synced; omitting them rather than emitting `"hrv": null` cuts payload size and avoids the LLM reasoning over absent measurements as if they were zero. Rules:
  - **Numeric zero, empty string, and `false` are NOT stripped** — they are meaningful values (`feel: 0` is invalid because the scale starts at 1, but `bodyFat: 0` is the upstream's signal for "not measured" and we still surface it explicitly elsewhere — see field-specific notes). Only JSON `null` triggers stripping.
  - **Stripping is applied per top-level row independently** so the response shape across a multi-day wellness window stays consistent (downstream callers can union the keys). A `_meta.fields_present: [...]` list per row is emitted so the LLM can see which measurements were taken without reading every row.
  - **Opt-out** via `include_full: true` so debugging callers can see the raw upstream shape including nulls.
  - **Custom fields** (intervals.icu wellness custom fields) participate in null-stripping the same way; the field is dropped if the day's value is `null`, kept otherwise.
  - **Explicit absence callout** (forum thread 123739 post #49): stripping nulls is necessary but not sufficient — a reference coaching prompt that says "do not guess if zone data is missing" needs an explicit signal of _what_ was missing, not just an absent key. Every read tool that strips nulls emits `_meta.missing_fields: ["hrv", "sleepScore", ...]` per row when stripping actually removed something, so the LLM can decline to infer rather than silently treat absence as zero or as a non-event.
- **No debug cruft**: auto-added fields like `fetched_at` and `query_type` are not in responses by default. They re-appear behind `ICUVISOR_DEBUG_METADATA=true` for troubleshooting. The LLM does not reason over timestamps of when _we_ fetched the data.
- **Server-side pagination** for `get_activities` over long date ranges, with a recommended page size that fits inside Claude free-tier context. Sizing anchor (forum thread 123739 post #44): a full year of one athlete's intervals.icu data — activities, wellness, events — is roughly 150k tokens in the upstream Python's verbose default shape. icuvisor's terse-default shape should compress that to a fraction; default page sizes assume a ~30k-token soft ceiling per tool response so a multi-page "year overview" fits inside a single free-tier conversation, and the `core` toolset's most chatty tools (`get_activities`, `get_wellness_data`, `get_events`) all paginate by default rather than letting an LLM accidentally request a year inline.
- **Scale metadata in tool descriptions AND in the response itself** so the LLM knows `feel` is 1-5, `sleepQuality` is 1-4. The response-level label is mandatory because some MCP clients pass only the response (not the tool description) back to the model at inference time.
- **Disambiguating field names** — emit `calories_burned` rather than `calories`, `distance_km` / `distance_mi` rather than `distance`, etc. Don't make the LLM guess units or direction.
- **Timezone normalization** — all dates rendered in the athlete's configured TZ; tool docstrings mention the convention.
- **Athlete ID normalization** — accept `i12345` or `12345`; emit `i12345` consistently.
- **Strava-imported activity handling** — intervals.icu blocks Strava-synced activities from its public API per Strava's ToS. Tools must detect the blocked state and return a structured `unavailable: { reason: "strava_tos", workaround: "connect device directly to intervals.icu (Garmin, Wahoo, Coros, Suunto, Polar)" }` rather than empty/`N/A` fields the LLM might hallucinate over.
- **Per-athlete unit normalization** — read `preferred_units` (miles vs km) from the athlete profile and render distances/paces in that unit, with the unit name embedded in the field key or `_meta` so the LLM can't drift to its default. Same pattern as the timezone rule.
- **Stream-key canonicalization** — the intervals.icu streams endpoint emits inconsistent key casing across activity types and devices (forum #118: `groundContactTime` on some activities, `ground_contact_time` on others). icuvisor canonicalizes every stream key to snake_case at the response boundary so the LLM (and downstream code) sees one schema. The canonical map lives in code and is covered by tests.
- **Self-explanatory shapes — don't assume a frontier-grade reasoner** (forum thread 123739 post #38): there is a real constituency running local quantized models (~70B and smaller) that are context-bound and weaker at multi-hop inference. Response shapes prefer pre-computed deltas, in-row scale labels, and explicit `_meta` legends over leaving the model to infer them. Every response should be interpretable by a model that does not have the tool description in its prompt at inference time (per the existing in-response scale-labels rule, but applied as a general principle).

#### E. Toolset tiers (token efficiency by default)

The full ~30-tool surface costs meaningful tokens to load into a conversation — every conversation, every model, every client load. icuvisor defaults to exposing a curated **`core`** subset (~17 tools) covering the daily-use path (read activities, fitness, wellness, events; write events, wellness, messages). Power users and coaches opt in to the **`full`** surface via `ICUVISOR_TOOLSET=full` in their MCP client config.

A small `icuvisor_list_advanced_capabilities` tool lives in `core` so the LLM can discover what's hidden and tell the user how to enable it when a prompt requires an advanced tool. This addresses the "tool selection accuracy" failure mode that grows with surface size — smaller models pick the wrong tool less often when the catalog is smaller.

Tool names within a cluster (e.g. `get_activity_details` / `get_activity_intervals` / `get_activity_streams`) must have **distinguishing first sentences in their descriptions**, since name alone won't tell the LLM which access pattern it needs. Confusability is audited at every catalog change.

Complex tools (those whose argument shape isn't obvious from prose — `add_or_update_event`, `create_workout`, `create_custom_item`, `apply_training_plan`) ship with `input_examples` covering the canonical case and one tricky edge. Anthropic reports a 72% → 90% accuracy gain from this single addition.

#### F. Destructive operation safety (env-var gate, not tool args)

Every operation that can permanently destroy data — event delete, activity delete, workout-library delete, gear delete, sport-settings delete, custom-item delete — is gated by an `ICUVISOR_DELETE_MODE` env var read once at startup:

| Mode             | Events            | Activities | Workouts (library)    | Gear | Sport settings | Custom items |
| ---------------- | ----------------- | ---------- | --------------------- | ---- | -------------- | ------------ |
| `safe` (default) | future-dated only | ✗          | future-only or unused | ✓    | ✗              | ✗            |
| `full`           | any date          | ✓          | ✓                     | ✓    | ✓              | ✓            |
| `none`           | ✗                 | ✗          | ✗                     | ✗    | ✗              | ✗            |

Tools that the active mode forbids are **not registered** with the MCP server at all — the LLM cannot see them in its tool catalog, cannot invent a flag to enable them, and cannot be talked into them. A per-call `confirm: true` argument is **not a credible safety guard** because the model controls the argument; if an error message says "set confirm=true to override," the model will. This gate sits outside the model's reach by design. Invalid `ICUVISOR_DELETE_MODE` values fail loudly at startup.

#### G. MCP Resources and Prompts (first-class primitives)

icuvisor ships MCP Resources for long-form, slow-changing content the LLM only occasionally needs, keeping it out of the per-session tool-description budget:

- `icuvisor://workout-syntax` — the intervals.icu structured-workout DSL.
- `icuvisor://event-categories` — the full enum of event categories with descriptions.
- `icuvisor://custom-item-schemas` — the per-`item_type` schema for the `content` field on custom items.
- `icuvisor://athlete-profile` — current athlete profile (auto-refreshing).

It also ships a curated set of MCP Prompts (training analysis, recovery check, weekly planning, race-week taper, coach roster triage) so users on clients that surface prompts get a "what can this thing do?" entrypoint without having to learn the tool catalog.

#### H. Configuration

- All state in a single platform-conventional config dir (`~/Library/Application Support/icuvisor/`, `%APPDATA%\icuvisor\`, `~/.config/icuvisor/`).
- API key stored in OS keychain (macOS Keychain, Windows Credential Manager, libsecret) — not in plain text — fixing a recurring concern that `.env` files leak to backups/repos (forum #35 + Marc's security concern in #61).
- Headless config via CLI flags / env for power users.

#### I. Observability

- Local rotating log file with a "Copy diagnostics" button in the tray menu (eliminates the back-and-forth on forum install threads).
- Opt-in anonymous telemetry: install success/failure, tool call counts (no payloads). Used to measure KR1, KR2, KR4.

### 6.3 Technology

- **Language**: Go 1.23+. Single static binary, cross-compiled via GoReleaser.
- **MCP SDK**: `github.com/modelcontextprotocol/go-sdk` (official) — assumed production-ready for stdio + Streamable HTTP.
- **HTTP client**: stdlib `net/http` + `httpretry` for intervals.icu Basic Auth calls.
- **Onboarding UI**: small embedded webview (Wails or Tauri-equivalent) **or** localhost HTML+HTMX page launched in the default browser. Decision deferred to design spike; localhost-page approach is the safer default for keeping the binary small and avoiding webview signing pain.
- **Tray icon**: `github.com/getlantern/systray` (or equivalent).
- **Build/release**: GitHub Actions + GoReleaser, with macOS notarization via `notarytool` and Windows signing via a hardware token.
- **License**: MIT. We **port from the public intervals.icu API docs, our own black-box testing, and forum/issue insights — not from the GPL Python source** (clean-room, from first principles), with GPL/copyleft implementation code kept out of the MIT-licensed project.

### 6.4 Assumptions (to validate)

**Settled (decisions, not open questions):**

- **License**: MIT.
- **Clean-room implementation**: porting from intervals.icu's public API docs + our own black-box testing, written in Go from first principles. No GPL Python source is read or copied.
- **Auth UX**: athletes paste an intervals.icu API key. No OAuth flow.
- **MCP SDK**: official `github.com/modelcontextprotocol/go-sdk` is treated as production-ready for stdio + Streamable HTTP. No spike or alternative-SDK evaluation.

**Still to validate:**

1. **Auto-update via signed releases is acceptable** to athletes and to the macOS/Windows platforms (notarized binaries can self-update inside the user's home directory). _(Validate during release-pipeline build-out.)_
2. **Token efficiency is achievable** in pure response shaping without an LLM in the middle — KR5 is hit by aggressive default summarization plus opt-in detail. _(Validate by measuring on the 10 most common forum prompts.)_
3. **The intervals.icu API supports strength training and training plan retrieval.** _(Validate during tool-catalog implementation.)_
4. **The full target field set for `get_extended_metrics`** (running dynamics, DFA α1, W' balance, core temp, cardiac decoupling, HR drift, aerobic decoupling, zone distributions, IF/VI/polarization, TRIMP/strain/load variants, L/R balance, RPE/feel/session-RPE, compliance %, device name — see §7.2.C) is exposed by the intervals.icu API rather than computed server-side by competitors. Fields not exposed upstream are dropped from the catalog, not silently zero-filled. _(Validate during tool-catalog implementation; track per-field availability in `testdata/`.)_
5. **Coach-mode credential delegation is safe** when the coach-scoped API key is held only by the local binary and never passed as a tool parameter. _Threat model filed in [`docs/threat-models/coach-mode.md`](../threat-models/coach-mode.md) (TP-039); config-backed per-athlete ACL enforcement shipped via `internal/coach`. Remaining narrower probe: authenticated `list_athletes` against the upstream coach-roster endpoint — currently served from the configured `coach.athletes[]` roster with `_meta.source: "config"` until a real coach key confirms auth, response shape, pagination, and scoping._
6. **Demand**: forum thread (~100 posts, multiple monthly active discussants) suggests a real audience, but we have not surveyed it directly. The competitive signal from icusync.icu is also weaker than its post count suggests — most activity is maintainer support, not latent demand for a free local alternative. _(Validate by pre-launch waitlist on icuvisor.dev; pick a target only once the waitlist is live.)_
7. **MCP tool-schema caching is per-conversation on all target clients.** Implications:
   - Auto-update UX must tell the user to start a new chat (see Flow C).
   - Tool argument changes must be **additive-only** on stable tools — no removals, no renames. Document in `CONTRIBUTING.md`.
   - Every tool response embeds `_meta.server_version` so the LLM can flag a schema mismatch when it sees stale arguments rejected. _(Validate by sweep across Claude Desktop, Claude Code, ChatGPT Dev Mode, Cursor.)_
8. **Mobile access is a dominant reason athletes will pay for a hosted competitor.** Re-evaluate whether the hosted relay (§8 / vNext) is correctly phased or should move earlier as an opt-in optional service. _(Validate during pre-launch waitlist — ask about mobile need explicitly.)_
9. **Token efficiency may not be a strong standalone differentiator.** Competing hosted servers do not appear to suffer obvious context-window problems, so KR5's Python-reference target may not translate directly to the icusync comparison. _(Validate by measuring icusync.icu's response shapes on the same prompt set.)_
10. **Strava-blocked-activity detection** depends on a stable upstream marker. _(Validate by black-box testing against an athlete account with mixed direct/Strava-imported activities.)_
11. **`preferred_units` is exposed on the intervals.icu athlete profile and round-trips through the API.** _(Validate during `get_athlete_profile` implementation.)_
12. **A `core` toolset of ~17 tools covers ≥90% of real prompts** based on the curated lists open competitors have arrived at independently. The remaining ~13 tools (bulk ops, gear, sport settings, curves, histograms, custom items, workout-library writes) are correctly placed behind `ICUVISOR_TOOLSET=full`. _(Validate via opt-in telemetry on tool-call distribution after v0.5 dogfooding.)_
13. **MCP Resources are honored by all target clients.** Verbose schema documentation (workout DSL, custom-item content shapes, event categories) belongs in Resources rather than inline tool descriptions. _(Validate during KR6 client-compatibility sweep — note any client that ignores `resources/list` and fall back to inline docs only for those.)_
14. **`input_examples` is honored by the official Go MCP SDK and surfaces to LLMs.** Anthropic reports 72→90% accuracy on complex argument shapes when examples are present. _(Validate during MCP SDK integration; if not supported, file upstream or use the lower-level SDK API.)_
15. **The `locked` flag on wellness records actually prevents device sync overwrites** as documented, so manual entries via MCP survive the next Garmin/Apple Health/Oura push. _(Validate by writing a wellness record with `locked: true` and triggering a device sync.)_
16. **`get_event_by_id` returns the same event that `get_events` listed.** Upstream behavior reports two unresolved cases where the detail endpoint returns 404 on IDs the list endpoint just returned. Suspected causes (not yet confirmed upstream): (a) recurring-event instances are listed under a parent series ID that the detail endpoint does not accept, requiring resolution to the concrete instance ID; (b) past-dated event IDs may be subject to a separate retention window than future-dated ones; (c) coach-scoped reads may list events the coach key cannot resolve individually. Implementation must (i) try the detail call on the ID returned by `get_events`, (ii) on 404, fall back to scanning the list endpoint over the event's known date range and matching by ID, and (iii) if still not found, return a structured `unavailable: { reason: "upstream_inconsistency", retried: ["detail", "list_scan"] }` rather than a raw 404 the LLM will retry indefinitely. _(Validate during `get_event_by_id` implementation; capture every reproducer in `testdata/events/inconsistent/` with the originating list response so the root-cause investigation has a corpus.)_
17. **Pace-unit enum coverage is exhaustive.** Upstream reports showed running/swim activities failed to deserialize due to missing units. The intervals.icu unit enum (as observed via black-box probing) appears on activity stream metadata, interval target definitions, and best-effort responses. Known members the implementation must handle without crashing:
    - Distance: `M` (meters), `KM`, `MI`, `YD`.
    - Pace: `MINS_KM` (min/km), `MINS_MILE` (min/mi), `SECS_100M` (sec/100m — swim), `SECS_500M` (sec/500m — row).
    - Speed: `KMH`, `MPH`, `MS` (m/s).
    - Time: `SECS`, `MINS`, `HOURS`.
    - Power: `WATTS`, `WKG` (W/kg), `PERCENT_FTP`.
    - Heart rate: `BPM`, `PERCENT_HR`, `PERCENT_LTHR`, `PERCENT_MAX_HR`.
    - Misc: `RPE`, `Z1`…`Z7` (zone refs), `PERCENT`, `KCAL`, `KJ`.

    Treat the unit enum as **load-bearing for activity reads**. Decode into a typed `Unit` enum with an explicit `UnitUnknown` fallback that carries the raw upstream string; return the raw string + `_meta.unknown_unit: <value>` rather than failing the whole tool call. Log unknown values at WARN so we can grow the enum from telemetry. Convert to athlete `preferred_units` at the response boundary (not at decode), so unknown units pass through with their raw label intact. _(Validate by black-box sweep across cycling, running, swimming, rowing, and walking activities at v0.2; treat each new unit discovered after v1.0 as an additive change.)_

18. **Periodization parameters are exposed via the intervals.icu API** (forum thread 123739 posts #28, #30). Athletes/coaches explicitly want read/write access to athlete-level planning parameters distinct from sport settings: ramp-rate %, recovery-week cadence (e.g. "every 4th week"), taper % drop, and intensity-distribution preference (polarized vs pyramidal vs threshold). If the API exposes any subset of these — likely under athlete profile or training-plan endpoints — they ship as a dedicated `get_planning_parameters` / `update_planning_parameters` pair. Fields not exposed upstream are dropped (per §7.4 #4), not silently zero-filled. _(Validate during the read-path build-out; capture per-field availability in `testdata/`. If none are exposed, file an intervals.icu API feature request and document the gap rather than computing them client-side.)_
19. **`workout_doc` upload asymmetry.** intervals.icu emits structured `workout_doc` JSON on reads but rejects it on writes; uploads must serialize back to the description-string DSL. The DSL is line-oriented, indentation-sensitive, and uses `Nx` prefixes for repeat blocks. Implementation owns a single canonical serializer (`internal/tools/workout_dsl.go`) covering all target types from §7.4 #17's enum; the parser must round-trip the serializer's own output bit-identically. _(Validate by read → modify → write → read round-trip on the v0.3 write path; golden-file tests in `testdata/workouts/` lock the canonical encoding. For every workout fixture that icuvisor cannot round-trip, file an issue rather than silently lossy-encode.)_

---

## 8. Release

Phasing — scope, gates, and the v0.1 / v0.5 / v1.0 / v1.x / vNext milestones — lives in [`ROADMAP.md`](../../ROADMAP.md) so plan-of-record edits don't drift across two files. This PRD owns the _what_ and the _why_; the roadmap owns the _when and in what order_.

---

_Document version: 0.1 draft. To be validated against the assumptions in §7.4 before v0.1 spike begins._

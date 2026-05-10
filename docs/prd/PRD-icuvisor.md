# PRD — icuvisor

> An open-source, locally-installed MCP connector for [intervals.icu](https://intervals.icu), distributed as a single Go binary. Designed for non-technical amateur athletes who want to talk to their training data through Claude, ChatGPT, Pi, and other mainstream AI tools.

---

## 1. Summary

icuvisor is a free, open-source Model Context Protocol (MCP) server that connects intervals.icu training data to AI assistants. Unlike the existing Python implementation it draws from, icuvisor ships as a single self-contained binary with one-click installers so athletes — not engineers — can install it, point Claude or ChatGPT at it, and start asking questions like *"should I increase my FTP?"* within five minutes.

---

## 2. Contacts

| Name | Role | Comment |
|---|---|---|
| *TBD* | Product / Maintainer | Owns roadmap, releases, community |
| *TBD* | Lead engineer | Go implementation, MCP transport, intervals.icu API client |
| Marc Vilanova (mvilanova) | Upstream Python reference | Original author of `intervals-mcp-server` — informally consulted; GPLv3 attribution required |
| Andrew Coggan (intervals.icu) | Platform owner | Not contacted — public API only |
| Anthropic / OpenAI / Pi.dev | Integration targets | MCP spec compliance only |

---

## 3. Background

### Context

Amateur endurance athletes increasingly want to use general-purpose LLMs as a personal coach: analyze recent rides, suggest tomorrow's workout, plan a training block toward a goal event, reflect on wellness trends. intervals.icu is the platform of choice for serious amateurs because it exposes a clean public API covering activities, wellness, events, fitness (CTL/ATL/TSB), and custom fields.

Two paths exist today to bridge intervals.icu and AI:

1. **`mvilanova/intervals-mcp-server`** (Python, GPLv3, ~17 tools). Free and capable, but installation requires `git`, Python 3.12, `uv`, and hand-edited JSON config with platform-specific absolute paths. The forum thread is dominated by install failures: `spawn uv ENOENT`, hatchling wheel build errors, `.env` confusion, Python version mismatches.
2. **`hhopke/intervals-icu-mcp`** (Python, MIT, ~58 tools — independent continuation of the eddmann fork). Larger tool surface and actively maintained; install path is `uvx intervals-icu-mcp` plus JSON config, which is materially easier than mvilanova's `uv sync` flow but still requires the user to install `uv` and hand-edit a JSON file. Open issues prioritize token efficiency (toolset tiers, description trimming, debug-metadata stripping) and a non-model-controlled delete safety gate — both of which icuvisor adopts directly.
3. **icusync.icu** — closed-source, hosted, account-required, opaque pricing. Solves the install problem by hosting everything, but requires trusting a third party with a token to your training data and offers no transparent free tier.

None of these serves the "I just want this to work" athlete who doesn't want to learn `uv` *or* hand their data to another SaaS.

### Why now

- **MCP momentum**: stable spec, Claude Desktop/Code/Cowork, ChatGPT Developer Mode, Cursor, Pi, Le Chat, and local LLM clients all support MCP servers.
- **Existing Python options have known limits**: `mvilanova/intervals-mcp-server` (GPLv3) ships ~17 tools with a difficult `uv` install path and its maintainer commented in Nov 2025 that he is *"directing efforts elsewhere presently"*. `hhopke/intervals-icu-mcp` (MIT, ~58 tools) is more actively maintained and easier to install via `uvx`, but its own open-issue backlog targets exactly the problems icuvisor is best positioned to solve from day one: model-uncontrollable delete safety, tiered toolsets to cut per-session token cost, in-response scale labels, debug-metadata stripping, and tool-name disambiguation.
- **Distribution gap is now solvable**: Go's single-binary cross-compilation + Homebrew/Scoop/Winget/DXT bundles let us deliver `brew install icuvisor` to a triathlete who has never opened a terminal.
- **Recent intervals.icu API additions** (custom wellness fields, activity messages, structured workout endpoints) make a richer feature set possible than what the original README documents.

---

## 4. Objective

### What and why

Make AI-powered training analysis accessible to **any amateur athlete** with an intervals.icu account, in **under five minutes**, with **zero terminal commands** for the happy path, and **zero recurring cost**. The intervals.icu community deserves a high-quality, open, local-first option that doesn't lock data behind someone else's login.

### How it benefits the company and customers

There is no company. icuvisor is community infrastructure. Benefits to users:

- **Athletes**: free, private (data flows athlete → local binary → AI client; no third-party server in the path), works with whichever AI tool they already pay for.
- **Coaches**: multi-athlete support without each athlete signing up for a SaaS.
- **Developers**: clean Go reference implementation of an MCP server, easy to fork or vendor.

### Alignment with vision and strategy

Vision: *training data should belong to athletes and run on athletes' machines, not vendors' servers*. Strategy: own the install/UX layer (where competitors lose users) while staying API-compatible with intervals.icu's evolving public surface.

### Key Results (12-month, SMART)

- **KR1 — Install success**: ≥95% of new users complete install + first successful tool call within 10 minutes, measured by opt-in anonymous telemetry on the welcome page.
- **KR2 — Adoption**: 2,000 weekly-active installs by month 12 (measured by opt-in update-check pings).
- **KR3 — Coverage**: feature parity with the leading reference servers — at least 90% of the deduplicated union of `hhopke/intervals-icu-mcp`'s ~58 tools, `mvilanova/intervals-mcp-server`'s ~17 tools, and icusync.icu's ~15 advertised tools. Gaps must be deliberate (e.g. dropped on safety grounds), not accidental.
- **KR4 — Reliability**: <1% of tool calls return uncaught errors; p95 latency <2s for read tools (excluding upstream API time).
- **KR5 — Token efficiency**: ≥60% reduction in per-session tool-description tokens vs. `hhopke/intervals-icu-mcp`'s default 58-tool surface (achieved via toolset tiers, description trimming, and MCP Resources for long-form schemas); ≥40% reduction in median per-tool-call response bytes vs. both Python references on the same prompts.
- **KR6 — Client compatibility**: validated working configs for Claude Desktop, Claude Code, Claude Cowork, ChatGPT (Dev Mode), Pi.dev, Cursor, Continue, Zed, and one local-LLM client (ollmcp or Cline).

---

## 5. Market Segments

Markets defined by the job the user is trying to do, not by demographics.

### Primary — "Curious amateur athlete"

> *"I have a Garmin, I log everything to intervals.icu, and I pay for Claude Pro. I want to ask Claude about my training. I am not a developer."*

- Job: get coaching-quality answers from their own data, without learning new tools.
- Constraint: will not run `pip install`, will not edit JSON, will not host a server. Maximum acceptable friction: download a `.dmg`/`.exe`, paste an API key, click "Connect to Claude."
- Estimated size: tens of thousands. The intervals.icu forum thread alone surfaced 50+ such users.

### Secondary — "Coach with a roster"

> *"I have 8–25 athletes on intervals.icu. I want to triage their weeks and plan next week's workouts inside Claude."*

- Job: review multiple athletes from one place; create/edit workouts on each athlete's calendar.
- Constraint: must support athlete-scoped credential delegation safely (issue #88, forum posts #18/#21/#60).

### Tertiary — "Self-experimenting power user / developer"

> *"I want to script analyses, build my own agent, run a local LLM, mix intervals.icu data with MyFitnessPal/Strava."*

- Job: programmatic access, multiple transports, scripting from CLI or notebooks.
- Constraint: needs documented JSON-RPC schema, headless mode, local HTTP transport, ability to vendor as a library.

### Out of scope (initially)

- Non-intervals.icu data sources (Strava direct, TrainingPeaks). The athlete should add other MCP servers alongside icuvisor.
- Mobile-only installs. We will ship for macOS, Windows, Linux; mobile clients connect via the user's desktop or via the optional hosted relay (see §8 Future Versions).

---

## 6. Value Propositions

### Customer jobs / needs

1. *"Analyze my last N activities and tell me what to do next."*
2. *"Plan a training block / taper toward race day X."*
3. *"Push tomorrow's workout to my calendar so it syncs to my Garmin."*
4. *"Reflect on my wellness trends — sleep, HRV, RHR."*
5. *"Coach me on this athlete I'm responsible for."*

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

### Where we beat competitors (Value Curve sketch)

| Attribute | mvilanova python | hhopke python | icusync.icu | **icuvisor** |
|---|---|---|---|---|
| Install effort | High (uv, Python, JSON) | Medium (uvx + JSON) | Very low (paste URL) | **Very low (installer + token paste)** |
| Cost | Free | Free | Opaque / paid | **Free** |
| Local-only (privacy) | Yes | Yes | No (hosted) | **Yes** |
| Source available | Yes (GPLv3) | Yes (MIT) | No | **Yes (MIT)** |
| Tool count | 17 | 58 | 15 | **~30 (curated, tiered)** |
| Write operations | Yes (verbose) | Yes | Yes | **Yes (terse + structured + library CRUD)** |
| Coach mode | No | No | Yes | **Yes (with per-tool permissions)** |
| Delete safety | confirm flag | env-var gate (planned) | Unknown | **Env-var gate (model-uncontrollable)** |
| Token efficiency by default | Weak | Improving (open issues) | Unknown | **Tiered toolset + Resources + trimmed payloads** |
| MCP Resources / Prompts | No | Yes (2 + 7) | Unknown | **Yes** |
| Multi-client tested | Claude-only | Claude + Cursor + ChatGPT | Claude + ChatGPT | **Claude + ChatGPT + Pi + Cursor + local LLM** |
| Automatic updates | No | No | N/A (hosted) | **Yes (signed, opt-out)** |

---

## 7. Solution

### 7.1 UX / User flows

**Flow A — First-time install (the golden path)**

1. Athlete visits `icuvisor.dev` and clicks the platform-detected download button.
2. macOS: opens signed/notarized `.dmg` → drag to Applications. Windows: signed `.msi`. Linux: `.deb`/`.rpm` or shell installer. Power users: `brew install icuvisor` / `scoop install icuvisor` / `winget install icuvisor`.
3. First launch opens a small native onboarding window (or a localhost page in the default browser):
   - Step 1: "Paste your intervals.icu API key" with a clickable link to `https://intervals.icu/settings` and a screenshot.
   - Step 2: detects athlete ID from the API key — falls back to manual entry, accepting both `i12345` and `12345` (issue #40).
   - Step 3: timezone autodetected from OS, editable (issue #49).
   - Step 4: pick AI client(s). Each option shows a "Set up automatically" button that writes the appropriate config file *and* a "Show manual config" disclosure for users who prefer it. Supported targets at launch: Claude Desktop, Claude Code, Claude Cowork, ChatGPT (Dev Mode instructions), Pi.dev, Cursor, Continue, Zed.
4. "Test connection" button calls `get_athlete_profile` and shows the athlete's name + FTP. ✅
5. Onboarding closes; a menu-bar / system-tray icon stays running.

**Flow B — Asking a question (the use case)**

User opens Claude Desktop and types *"Analyze my last 10 cycling activities and let me know if I should adjust my FTP."* Claude calls `get_activities` (terse mode), then `get_activity_intervals` for each, then `get_athlete_profile` for current FTP, then replies. icuvisor's terse-by-default responses keep this under one context window even on free Claude tier (addressing forum #65, #66).

**Flow C — Update**

icuvisor checks `releases.icuvisor.dev` once per day. If a new signed release exists, the tray icon shows a dot; clicking "Update now" replaces the binary and restarts. No terminal commands. Opt-out in settings.

After an update that adds or changes tool arguments, the post-update notification explicitly tells the user to **start a new conversation in their AI client** to pick up the new tool schema. MCP clients (Claude in particular) cache the tool catalog at conversation start, so an in-flight chat will keep using the old schema and report "the fix didn't work."

**Flow D — Coach mode**

Coach pastes a coach-scoped intervals.icu API key. icuvisor lists athletes via `list_athletes`; the coach selects which subset is exposed to tools. The active athlete is passed as a tool argument (`athlete_id`) on every call, with a configurable default. Mirrors issue #88 and forum posts #18/#21/#60.

The coach also picks, **per athlete**, which tools are exposed — e.g. read-only access for a prospective athlete, full read+write for an active client. Granular per-tool permissions are enforced in the server before any intervals.icu call; the LLM never sees disallowed tools in its catalog.

Wireframes will be produced separately; this PRD specifies behavior only.

### 7.2 Key Features

#### A. Distribution (the differentiator)

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
- `list_athletes`, `select_athlete` — coach mode.
- `get_fitness` — CTL/ATL/TSB trends, taper projections.
- `get_best_efforts` — PRs across sports.
- `get_power_curves` — mean-maximal curves.

**Activities**
- `get_activities` — date-range list; supports `include_unnamed` (issue #67) and pagination.
- `get_activity_details` — single-activity metadata, zones, metrics.
- `get_activity_intervals` — interval splits.
- `get_activity_streams` — time-series (power, HR, altitude, cadence, etc.).
- `get_activity_messages` — fetch comments/notes.
- `add_activity_message` — post a comment (forum #99).
- `get_extended_metrics` — running dynamics, core temp, DFA α1, W' balance (icusync parity).
- `get_training_summary` — aggregated volume/TSS/zones.

**Wellness**
- `get_wellness_data` — daily rows. **Includes custom fields** (issue #64, forum #92) and correct scale metadata embedded in the tool description **and the response itself** (`feel` is 1-5, `sleepQuality` is 1-4 — addresses issues #45/#48 and forum #54/#57). In-response labels are required because some MCP clients do not pass tool descriptions back to the LLM at inference time.
- `update_wellness` — write back the full set of API-accepted fields: subjective scales (`feel`, `fatigue`, `soreness`, `stress`, `mood`, `motivation`, `sleepQuality`, `injury`), body metrics (`weight`, `bodyFat`, `abdomen`), cardiovascular (`restingHR`, `hrv`, `systolic`, `diastolic`), blood/lab (`bloodGlucose`, `lactate`, `spO2`, `vo2max`), respiration, menstrual phase, and the `locked` flag that prevents device sync from silently overwriting manual entries.

**Events & workouts**
- `get_events`, `get_event_by_id` — calendar entries.
- `add_or_update_event` — structured workout, race, or note. Returns a **terse** confirmation by default (issue #89). Preserves intervals.icu's distinction between `description` (free text — athlete/coach notes, pacing, nutrition, race countdown) and `workout_doc` (structured steps). On edit, `description` is written through **verbatim** unless the caller explicitly opts into structured normalization; `workout_doc` is the only field that accepts structured-block syntax. Silent normalization of free text is treated as a destructive operation and must not happen by default. Accepts a `tags` array (e.g. `["sweet-spot", "indoor"]`) which round-trips through reads.
- `delete_event`, `delete_events_by_date_range` — destructive. **Gated by `ICUVISOR_DELETE_MODE` env var, not a tool argument** (see §7.2.D — model-controlled `confirm: true` flags are not a credible safety guard).
- `get_training_plan` — fetch plan (forum #70).
- `apply_training_plan` — instantiate a workout-library folder onto the calendar from a start date.
- *Strength training data* — included if the intervals.icu API exposes it (forum #70).

**Workout library (templates, distinct from calendar events)**
- `get_workout_library`, `get_workouts_in_folder` — read templates by folder.
- `create_workout`, `update_workout`, `delete_workout` — author/edit templates via MCP so a coach can ask the LLM to build a reusable training block without leaving the chat. `delete_workout` is destructive and gated by `ICUVISOR_DELETE_MODE` like event deletion.

**Custom items**
- `get_custom_items`, `get_custom_item_by_id`, `create_custom_item`, `update_custom_item`, `delete_custom_item` — for custom charts/fields/zones. Long-form schema documentation for the inner `content` shape (which varies per `item_type`) lives in an MCP Resource (`icuvisor://custom-item-schemas`), not inline in the tool description (see §7.2.G).

**Total: ~30 tools** at v1.0 — the `core` toolset (see §7.2.D) exposes a curated ~17-tool subset by default; the full surface ships behind an opt-in env var.

#### D. Response shaping (the second differentiator)

- **Terse-by-default**: every read tool returns the smallest useful payload. Heavy fields (streams, raw samples) require explicit opt-in.
- **No debug cruft**: auto-added fields like `fetched_at` and `query_type` are not in responses by default. They re-appear behind `ICUVISOR_DEBUG_METADATA=true` for troubleshooting. The LLM does not reason over timestamps of when *we* fetched the data.
- **Server-side pagination** for `get_activities` over long date ranges, with a recommended page size that fits inside Claude free-tier context.
- **Scale metadata in tool descriptions AND in the response itself** so the LLM knows `feel` is 1-5, `sleepQuality` is 1-4. The response-level label is mandatory because some MCP clients pass only the response (not the tool description) back to the model at inference time.
- **Disambiguating field names** — emit `calories_burned` rather than `calories`, `distance_km` / `distance_mi` rather than `distance`, etc. Don't make the LLM guess units or direction.
- **Timezone normalization** — all dates rendered in the athlete's configured TZ; tool docstrings mention the convention.
- **Athlete ID normalization** — accept `i12345` or `12345`; emit `i12345` consistently.
- **Strava-imported activity handling** — intervals.icu blocks Strava-synced activities from its public API per Strava's ToS. Tools must detect the blocked state and return a structured `unavailable: { reason: "strava_tos", workaround: "connect device directly to intervals.icu (Garmin, Wahoo, Coros, Suunto, Polar)" }` rather than empty/`N/A` fields the LLM might hallucinate over.
- **Per-athlete unit normalization** — read `preferred_units` (miles vs km) from the athlete profile and render distances/paces in that unit, with the unit name embedded in the field key or `_meta` so the LLM can't drift to its default. Same pattern as the timezone rule.

#### E. Toolset tiers (token efficiency by default)

The full ~30-tool surface costs meaningful tokens to load into a conversation — every conversation, every model, every client load. icuvisor defaults to exposing a curated **`core`** subset (~17 tools) covering the daily-use path (read activities, fitness, wellness, events; write events, wellness, messages). Power users and coaches opt in to the **`full`** surface via `ICUVISOR_TOOLSET=full` in their MCP client config.

A small `icuvisor_list_advanced_capabilities` tool lives in `core` so the LLM can discover what's hidden and tell the user how to enable it when a prompt requires an advanced tool. This addresses the "tool selection accuracy" failure mode that grows with surface size — smaller models pick the wrong tool less often when the catalog is smaller.

Tool names within a cluster (e.g. `get_activity_details` / `get_activity_intervals` / `get_activity_streams`) must have **distinguishing first sentences in their descriptions**, since name alone won't tell the LLM which access pattern it needs. Confusability is audited at every catalog change.

Complex tools (those whose argument shape isn't obvious from prose — `add_or_update_event`, `create_workout`, `create_custom_item`, `apply_training_plan`) ship with `input_examples` covering the canonical case and one tricky edge. Anthropic reports a 72% → 90% accuracy gain from this single addition.

#### F. Destructive operation safety (env-var gate, not tool args)

Every operation that can permanently destroy data — event delete, activity delete, workout-library delete, gear delete, sport-settings delete, custom-item delete — is gated by an `ICUVISOR_DELETE_MODE` env var read once at startup:

| Mode | Events | Activities | Workouts (library) | Gear | Sport settings | Custom items |
|---|---|---|---|---|---|---|
| `safe` (default) | future-dated only | ✗ | future-only or unused | ✓ | ✗ | ✗ |
| `full` | any date | ✓ | ✓ | ✓ | ✓ | ✓ |
| `none` | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ |

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

### 7.3 Technology

- **Language**: Go 1.23+. Single static binary, cross-compiled via GoReleaser.
- **MCP SDK**: `github.com/modelcontextprotocol/go-sdk` (official) — assumed production-ready for stdio + Streamable HTTP.
- **HTTP client**: stdlib `net/http` + `httpretry` for intervals.icu Basic Auth calls.
- **Onboarding UI**: small embedded webview (Wails or Tauri-equivalent) **or** localhost HTML+HTMX page launched in the default browser. Decision deferred to design spike; localhost-page approach is the safer default for keeping the binary small and avoiding webview signing pain.
- **Tray icon**: `github.com/getlantern/systray` (or equivalent).
- **Build/release**: GitHub Actions + GoReleaser, with macOS notarization via `notarytool` and Windows signing via a hardware token.
- **License**: MIT. We **port from the public intervals.icu API docs, our own black-box testing, and forum/issue insights — not from the GPL Python source** (clean-room, from first principles), with attribution to mvilanova in the README.

### 7.4 Assumptions (to validate)

**Settled (decisions, not open questions):**

- **License**: MIT.
- **Clean-room implementation**: porting from intervals.icu's public API docs + our own black-box testing, written in Go from first principles. No GPL Python source is read or copied.
- **Auth UX**: athletes paste an intervals.icu API key. No OAuth flow.
- **MCP SDK**: official `github.com/modelcontextprotocol/go-sdk` is treated as production-ready for stdio + Streamable HTTP. No spike or alternative-SDK evaluation.

**Still to validate:**

1. **Auto-update via signed releases is acceptable** to athletes and to the macOS/Windows platforms (notarized binaries can self-update inside the user's home directory). *(Validate during release-pipeline build-out.)*
2. **Token efficiency is achievable** in pure response shaping without an LLM in the middle — KR5 is hit by aggressive default summarization plus opt-in detail. *(Validate by measuring on the 10 most common forum prompts.)*
3. **The intervals.icu API supports strength training and training plan retrieval.** *(Validate during tool-catalog implementation.)*
4. **icusync.icu's "extended metrics"** (DFA α1, W' balance, core temp, running dynamics) are exposed by the intervals.icu API rather than computed server-side by icusync. *(Validate during tool-catalog implementation.)*
5. **Coach-mode credential delegation is safe** when the coach-scoped API key is held only by the local binary and never passed as a tool parameter. *(Threat-model review before coach-mode ships.)*
6. **Demand**: forum thread (~100 posts, multiple monthly active discussants) suggests a real audience, but we have not surveyed it directly. The competitive signal from icusync.icu is also weaker than its post count suggests — most activity is maintainer support, not latent demand for a free local alternative. *(Validate by pre-launch waitlist on icuvisor.dev; pick a target only once the waitlist is live.)*
7. **MCP tool-schema caching is per-conversation on all target clients.** Implications:
   - Auto-update UX must tell the user to start a new chat (see Flow C).
   - Tool argument changes must be **additive-only** on stable tools — no removals, no renames. Document in `CONTRIBUTING.md`.
   - Every tool response embeds `_meta.server_version` so the LLM can flag a schema mismatch when it sees stale arguments rejected. *(Validate by sweep across Claude Desktop, Claude Code, ChatGPT Dev Mode, Cursor.)*
8. **Mobile access is a dominant reason athletes will pay for a hosted competitor.** Re-evaluate whether the hosted relay (§8 / vNext) is correctly phased or should move earlier as an opt-in optional service. *(Validate during pre-launch waitlist — ask about mobile need explicitly.)*
9. **Token efficiency may not be a strong standalone differentiator.** Competing hosted servers do not appear to suffer obvious context-window problems, so KR5's "30% of Python upstream" target wins on the mvilanova comparison but not the icusync comparison. *(Validate by measuring icusync.icu's response shapes alongside mvilanova's on the same prompt set.)*
10. **Strava-blocked-activity detection** depends on a stable upstream marker. *(Validate by black-box testing against an athlete account with mixed direct/Strava-imported activities.)*
11. **`preferred_units` is exposed on the intervals.icu athlete profile and round-trips through the API.** *(Validate during `get_athlete_profile` implementation.)*
12. **A `core` toolset of ~17 tools covers ≥90% of real prompts** based on the curated lists open competitors have arrived at independently. The remaining ~13 tools (bulk ops, gear, sport settings, curves, histograms, custom items, workout-library writes) are correctly placed behind `ICUVISOR_TOOLSET=full`. *(Validate via opt-in telemetry on tool-call distribution after v0.5 dogfooding.)*
13. **MCP Resources are honored by all target clients.** Verbose schema documentation (workout DSL, custom-item content shapes, event categories) belongs in Resources rather than inline tool descriptions. *(Validate during KR6 client-compatibility sweep — note any client that ignores `resources/list` and fall back to inline docs only for those.)*
14. **`input_examples` is honored by the official Go MCP SDK and surfaces to LLMs.** Anthropic reports 72→90% accuracy on complex argument shapes when examples are present. *(Validate during MCP SDK integration; if not supported, file upstream or use the lower-level SDK API.)*
15. **The `locked` flag on wellness records actually prevents device sync overwrites** as documented, so manual entries via MCP survive the next Garmin/Apple Health/Oura push. *(Validate by writing a wellness record with `locked: true` and triggering a device sync.)*

---

## 8. Release

Phasing — scope, gates, and the v0.1 / v0.5 / v1.0 / v1.x / vNext milestones — lives in [`ROADMAP.md`](../../ROADMAP.md) so plan-of-record edits don't drift across two files. This PRD owns the *what* and the *why*; the roadmap owns the *when and in what order*.

---

*Document version: 0.1 draft. To be validated against the assumptions in §7.4 before v0.1 spike begins.*

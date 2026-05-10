# Roadmap

Living document. Phases are scoped and gated, not calendared. icuvisor will not commit to calendar dates pre-launch — each phase is shippable independently. Track current progress in [GitHub Issues](https://github.com/ricardocabral/icuvisor/issues) and [Projects](https://github.com/ricardocabral/icuvisor/projects). Product scope, tool catalog, and Key Results live in the [PRD](docs/prd/PRD-icuvisor.md); this file is the authoritative phasing plan.

## v0.1 — Walking skeleton

**Goal:** end-to-end pipe from binary → MCP → intervals.icu.

- [ ] Go module + project layout.
- [ ] intervals.icu API client (Basic Auth, retries, structured errors).
- [ ] MCP stdio transport wired up via `github.com/modelcontextprotocol/go-sdk`.
- [ ] One working tool: `get_athlete_profile`, end-to-end via stdio to Claude Desktop on macOS.
- [ ] Manual JSON config (no installer yet).

## v0.5 — Internal beta

**Goal:** validate KR1 (install success) and KR5 (token efficiency) on real users.

- [ ] All ~30 launch tools implemented (see PRD §7.2.C), including workout-library CRUD and the expanded `update_wellness` field set with `locked` flag support.
- [ ] Terse-by-default and `include_full` response modes; auto-added debug metadata (`fetched_at`, `query_type`) stripped by default, behind `ICUVISOR_DEBUG_METADATA=true` for troubleshooting.
- [ ] `ICUVISOR_TOOLSET` env var with `core` (default, ~17 tools) and `full` tiers; `icuvisor_list_advanced_capabilities` lives in `core` for discoverability.
- [ ] `ICUVISOR_DELETE_MODE` env var (`safe` default / `full` / `none`) — destructive tools are not *registered* in modes that forbid them, so the LLM cannot see or invoke them. No per-call `confirm: true` arguments.
- [ ] MCP Resources for long-form schema docs: `icuvisor://workout-syntax`, `icuvisor://event-categories`, `icuvisor://custom-item-schemas`, `icuvisor://athlete-profile`.
- [ ] MCP Prompts: training analysis, recovery check, weekly planning, race-week taper, coach roster triage.
- [ ] `input_examples` on complex tools (`add_or_update_event`, `create_workout`, `create_custom_item`, `apply_training_plan`).
- [ ] Tool-name disambiguation audit: every confusable cluster has distinguishing first sentences; CI guard against new clusters without one.
- [ ] In-response scale labels on every subjective field (`feel`, `sleepQuality`, `fatigue`, etc.) — not just tool descriptions, since some clients don't pass descriptions to the LLM.
- [ ] Disambiguating field names in responses (`calories_burned` not `calories`; `distance_km` / `distance_mi`).
- [ ] Server-side pagination for `get_activities`.
- [ ] Strava-blocked-activity detection: tools return a structured `unavailable: { reason: "strava_tos", workaround: ... }` rather than empty fields.
- [ ] Per-athlete unit normalization (miles vs km) read from `preferred_units`, embedded in field names / `_meta`.
- [ ] `add_or_update_event` preserves free-text `description` verbatim; `workout_doc` is the only field that accepts structured-block normalization. Supports `tags`.
- [ ] Tool-schema stability rules enforced in CI: tool argument changes are additive-only on stable tools; renames/removals require a new tool name.
- [ ] `_meta.server_version` embedded in every tool response.
- [ ] Coach mode behind a feature flag, with per-athlete granular tool permissions.
- [ ] OS keychain credential storage (macOS Keychain, Windows Credential Manager, libsecret).
- [ ] Streamable HTTP transport (localhost-bound by default).
- [ ] macOS signed installer; manual Claude Desktop config documentation.
- [ ] Dogfooded by 5–10 forum-recruited athletes, including at least one coach.

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
- [ ] Keychain-based credential storage on all platforms.
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

## Out of scope

- Replacing intervals.icu's own UI.
- Becoming a multi-tenant SaaS for primary use.
- Hosting athlete data on our infrastructure outside the future opt-in relay.
- Non-intervals.icu data sources as first-party features (athletes can install other MCP servers alongside icuvisor).
- Mobile-only installs at launch — desktop only for v1; mobile is served via the user's desktop or the future hosted relay.

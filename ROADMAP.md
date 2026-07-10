# Roadmap

Living document. Phases are scoped and gated, not calendared. Each phase is shippable independently. Completed work belongs in [CHANGELOG.md](CHANGELOG.md); this file tracks only forward-looking scope. Track current progress in [GitHub Issues](https://github.com/ricardocabral/icuvisor/issues) and [Projects](https://github.com/ricardocabral/icuvisor/projects). Product scope, tool catalog, and Key Results live in the [PRD](docs/prd/PRD-icuvisor.md); this file is the authoritative phasing plan.

## v1.x - Local-first stable line

**Goal:** make the local binary feel boring to install, update, diagnose, and connect to mainstream MCP clients while preserving the local-first trust model.

### v1.2 - Package and signing closure

- Windows release trust: Authenticode-sign the MSI or document a durable no-signing decision with explicit SmartScreen expectations.
- Windows package reach: publish and validate the Winget manifest alongside the existing MSI/Scoop path.
- Linux packages: ship `.deb` and `.rpm` packages with libsecret/Secret Service setup guidance and package-manager upgrade behavior.
- macOS package-manager polish: resolve the remaining Gatekeeper story for the Homebrew tarball binary, either by signing/notarizing it or documenting the exact trust tradeoff.
- Installer verification matrix: smoke the credential path, first `icuvisor setup`, upgrade, uninstall, and PATH behavior on macOS, Windows, and Linux release artifacts.

### v1.3 - Update and schema-cache UX

- Signed update channel: add an opt-out update check that verifies signed release metadata before offering an upgrade.
- In-app upgrade path: let the user apply an update without rerunning install docs, reusing the shell/MSI/package-manager path where appropriate.
- Catalog-change notice: turn the existing stale-conversation guidance into an automatic user-visible notice when the running catalog fingerprint changes.
- Client recovery hints: make post-upgrade guidance client-specific where possible, especially for clients that cache MCP schemas per conversation.

### v1.4 - Guided onboarding and client setup

- Onboarding UI: wrap the existing secure setup flow in a non-terminal experience for the happy path.
- One-click client config for Claude Desktop, Claude Code, Claude Cowork, Cursor, Continue, Zed, and any client with a stable local config format.
- Instruction-only setup pages for clients that cannot be configured safely by icuvisor, including ChatGPT Developer Mode and Pi.dev.
- Diagnostics export button in the tray/menu UI, backed by the existing redacted diagnostics command.
- KR6 compatibility sweep across the target client list, including setup, first tool call, stale-schema recovery, and teardown.

### v1.5 - Local LLM and power-user coverage

- Local-LLM recipes for ollmcp, Cline, and LM Studio, with explicit caveats for resource support, context limits, and toolset size.
- Streamable HTTP hardening for local clients that prefer HTTP over stdio, retaining `127.0.0.1` as the default bind.
- Power-user configuration examples for `ICUVISOR_TOOLSET`, `ICUVISOR_DELETE_MODE`, coach mode, and headless environments.
- Client capability notes for MCP Resources, Prompts, input examples, and `_meta` visibility so docs match what each client actually passes to the model.

### v1.6 - Observability and response-shape tuning

- Opt-in anonymous telemetry for install success, update success, tool-call counts, error classes, latency buckets, and catalog/toolset distribution. No payloads.
- Telemetry consent UI and CLI flags, with clear disable and data-inspection paths.
- Response-shape tuning loop: use opt-in aggregate data to adjust page sizes, core/full tool placement, and terse defaults without expanding token cost accidentally.
- Reliability dashboard or maintainer report for KR1, KR2, KR4, KR5, and KR6 tracking.

### v1.7 - Coach and upstream API validation

- Authenticated validation of the sanitized `GET /api/v1/athletes` roster client using a real coach-scoped key through the normal local credential flow; never paste the key into prompts or fixtures.
- Switch `list_athletes` from configured-roster source to upstream source only after auth, response shape, pagination, scoping, and redaction are confirmed.
- Preserve configured-roster ACLs as the local authorization boundary even if upstream roster discovery lands.
- Add regression fixtures for coach roster edge cases discovered during the authenticated probe.

### v1.8 - Remaining first-party intervals.icu coverage

- Training-plan endpoint validation and any missing read/write tools justified by the PRD catalog.
- Strength-training endpoint validation against `docs/upstream-gaps/strength-training.md`; keep current support at best-effort gym notes/simple calendar events until the upstream schema can round-trip.
- Planning-parameter probe for ramp rate, recovery-week cadence, taper target, and intensity-distribution preference. Ship only fields exposed by the upstream API.
- Extended-metrics field audit for fields that are still unproven upstream; drop unsupported fields explicitly rather than synthesizing them silently.

### v1.9 - Analytics evidence and roadmap slice

Forum-derived analytics requests are prioritized by whether icuvisor already has deterministic data access, whether another Taskplane packet owns the work, and whether upstream fields/formulas are proven. Roadmap entries below are implementation maps, not live-feature claims unless marked `existing`.

| Opportunity | User question | Required upstream data | Likely icuvisor surface | Caveats / non-promises | Disposition |
|-------------|---------------|------------------------|-------------------------|------------------------|-------------|
| Durability power curves after kJ | "How does my power curve change after 500/1000 kJ of work?" | Upstream power-curve rows with `after_kj` thresholds and bucket values. | Existing `get_power_curves` durability curves. | Only shown when upstream returns configured/fitted after-kJ curves; missing/unfitted curves are omitted rather than zero-filled. | `existing` via TP-204 |
| Swim stroke, cadence, and SWOLF trends | "Is my swim efficiency improving, and did stroke rate or SWOLF change over the block?" | Confirmed swim stroke count/rate/SWOLF fields from activity rows, streams, intervals, or custom fields; yards-unit support tracked separately. | New follow-up should first audit upstream field names, then add canonical keys and a trend/report surface if data is available. | Do not promise SWOLF or stroke trends until upstream fields are proven; generic cadence/pace support is not equivalent to a swim-efficiency analyzer. | `new follow-up` candidate; TP-215 covers yards pace only |
| Terrain execution analysis for runs | "Did I pace climbs, flats, and descents differently, and where did I fade?" | Activity streams for distance/time plus elevation or grade (`altitude`, `grade_smooth`) and enough samples for segment stats. | Follow-up terrain analyzer that groups stream samples into grade/elevation buckets and reuses `compute_activity_segment_stats` formulas. | Current tools can analyze explicit user-specified segments, but they do not auto-detect terrain sections; grade/elevation streams may be missing or noisy. | `new follow-up` candidate |
| Per-sport performance potential | "Across ride/run/swim, what current thresholds and best curve anchors can I compare safely?" | Athlete profile thresholds plus power, pace, and HR curve anchors per sport. | Existing `get_performance_potential`, with per-sport namespaces and explicit unavailable fields. | No proprietary score, cross-sport ranking, CP estimate, or aerobic-threshold guess; missing sources remain unavailable. | `existing` via TP-208 |
| Mechanical work by configured power zone | "How much mechanical work did I produce in each power zone over this block?" | Activity list, athlete power-zone settings, and canonical recorded power/time streams. | Existing full-toolset `compute_zone_energy`, with timestamp-weighted seconds/kJ, coverage diagnostics, and a pinned formula ref. | Mechanical kJ is not metabolic energy, calorie expenditure, or food calories; missing streams/zones and invalid intervals remain explicit rather than zero-filled or interpolated. | `existing` via TP-236 |
| Aerobic / anaerobic threshold estimation | "Can you estimate my aerobic and anaerobic thresholds from current data?" | Explicit profile thresholds for anaerobic context today; future public deterministic formulas or upstream fields for estimates. | Existing tools report explicit FTP/LTHR/threshold pace and unsupported aerobic threshold; any estimator needs a separate formula/product-review task. | Do not infer medical/physiology thresholds from curve shape or chat math; estimation remains unsupported unless a validated source/formula is approved. | `deferred` |

## v2.x - Planning and ecosystem expansion

**Goal:** expand training-planning automation and companion-server coverage while keeping new hosting or cross-service data retention outside the core binary.

### v2.0 - Planning automation

- `fill_calendar_from_library` ("Plan Filler"): propose workouts from an existing library folder over a date range using target weekly load/hours and weekday availability. Commit remains a separate explicit call. **Future-tool acceptance criteria (not shipped):** before the separate commit, validate each proposed schedule against structured constraints that keep requested session count distinct from athlete-local availability and independent slots; enforce per-session, daily, indoor, mode, sport, and remaining-week time/load caps; account for completed training and protected fixed commitments; and return deterministic violations plus reconciliation totals without silently redistributing deficits.
- Training-plan editing workflows that preserve existing races, notes, unavailable blocks, and user-authored descriptions.
- Workout-template curation: AI-generated and athlete-curated template sets built on top of the existing workout-library tools.
- Plan-preview evaluation that reports load distribution, compliance assumptions, conflicts, and lossy workout fields before writes.

### v2.1 - Multi-sport workout model

- Surface upstream triathlon workout-file resources with category, metric, and sub-category filters.
- Represent discipline-tagged block sequences without forcing them through the single-discipline `workout_doc` shape.
- Round-trip brick sessions and triathlon templates with documented lossy fields, golden fixtures, and analyzer-compatible schema.
- Align multi-sport compliance and zone-time calculations with the analyzer family instead of creating a separate reporting model.

### v2.2 - Plan safety and coaching guardrails

- Science-backed validation tool for generated plans, with transparent rules and citations rather than hidden coaching opinion. This evidence-based layer is separate from v2.0's deterministic placement, availability, and budget validation and from the shipped `masters_plan_review` prompt, which is a strictly read-only, no-rule evidence review rather than an age-aware physiology model.
- Guardrails for ramp rate, recovery weeks, taper shape, intensity distribution, and race-week workload when the required inputs are available.
- Explicit "insufficient evidence" responses when a plan cannot be validated from available athlete data.
- Versioned rule definitions so plan-validation behavior does not drift silently.

### v2.3 - Companion ecosystem

- Strava companion MCP server for direct Strava workflows that are intentionally outside the icuvisor binary.
- TrainingPeaks companion MCP server if demand and API access justify it.
- Conversation-memory export hooks, such as Claude Projects integration, that keep user-owned summaries portable without storing athlete data in icuvisor infrastructure by default.
- Cross-server guidance for users who run icuvisor alongside nutrition, strength, or device-specific MCP servers.

## Out of scope

- Replacing intervals.icu's own UI.
- Becoming a multi-tenant SaaS for primary use.
- Hosting athlete data on our infrastructure outside the optional hosted connector.
- Non-intervals.icu data sources inside the icuvisor binary; athletes can install companion MCP servers alongside icuvisor.
- Native mobile installs; mobile access is served through the user's desktop or the hosted connector where supported.

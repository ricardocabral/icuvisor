# General — Context

**Last Updated:** 2026-05-20
**Status:** Active
**Next Task ID:** TP-134

---

## v0.2 — Read path (TP-007 … TP-016)

Scaffolded 2026-05-11 from PRD §7.2.C–D / §7.4 and ROADMAP.md v0.2.

| ID     | Title                                                                                                       | Depends on     |
| ------ | ----------------------------------------------------------------------------------------------------------- | -------------- |
| TP-007 | Response shaping primitives (terse, null-strip, \_meta, scale labels, TZ, athlete-ID, unit-system plumbing) | TP-002, TP-004 |
| TP-008 | Pace/unit enum + snake_case stream-key canonicalization                                                     | TP-002         |
| TP-009 | Activities read cluster + Strava-blocked detection                                                          | TP-007, TP-008 |
| TP-010 | Fitness / best-efforts / power-curves / training-summary / extended-metrics                                 | TP-007, TP-008 |
| TP-011 | Wellness reads (sleep dual-scale, provenance, staleness, \_native)                                          | TP-007         |
| TP-012 | Events + `get_event_by_id` inconsistency handling + training plan                                           | TP-007         |
| TP-013 | Workout-library + custom-items reads                                                                        | TP-007         |
| TP-014 | Periodization parameters read (or documented upstream gap)                                                  | TP-007         |
| TP-015 | Tool-name disambiguation + CI schema-stability guards                                                       | TP-009…TP-013  |
| TP-016 | v0.2 dogfood (solo + 2–3 invited athletes, read-only)                                                       | TP-009…TP-015  |

---

## v0.3 — Writes with safety gate (TP-018 … TP-029)

Scaffolded 2026-05-13 from PRD §7.2.C / §7.2.D / §7.4 and ROADMAP.md v0.3.

| ID     | Title                                                                          | Depends on                             |
| ------ | ------------------------------------------------------------------------------ | -------------------------------------- |
| TP-018 | `ICUVISOR_DELETE_MODE` safety gate (registration-time filtering, no `confirm`) | TP-007                                 |
| TP-019 | `workout_doc` write-path serializer (DSL round-trip, golden tests)             | TP-013                                 |
| TP-020 | Event write cluster (`add_or_update_event`, `link_activity_to_event`, message) | TP-018, TP-019, TP-012                 |
| TP-021 | `update_wellness` (full writable field set incl. `injury`, BP, lactate, lock)  | TP-018, TP-011                         |
| TP-022 | `update_sport_settings` (FTP/threshold ungated; zones gated)                   | TP-018                                 |
| TP-023 | Workout-library CRUD (`create_workout`, `update_workout`, `delete_workout`)    | TP-018, TP-019, TP-013                 |
| TP-024 | Custom-items create/update                                                     | TP-018, TP-013                         |
| TP-025 | Destructive deletes cluster (event, activity, custom-item, sport, gear)        | TP-018, TP-020, TP-023, TP-024         |
| TP-026 | `apply_training_plan` (dry-run default; replace gated)                         | TP-018, TP-019, TP-020, TP-023, TP-012 |
| TP-027 | `input_examples` on complex write tools (catalog invariant)                    | TP-020…TP-024, TP-026                  |
| TP-028 | Adversarial safety test suite (safe-mode hardening, tool-not-found surrender)  | TP-018, TP-020…TP-026                  |
| TP-029 | v0.3 dogfood (dedicated test-athlete account, write-path validation)           | TP-018…TP-028                          |

---

## v0.4 — Token efficiency and MCP primitives (TP-030 … TP-034)

Scaffolded 2026-05-14 from PRD §7.2.E / §7.2.G / §7.3 / §6 KR5 and ROADMAP.md v0.4.

| ID     | Title                                                                          | Depends on                             |
| ------ | ------------------------------------------------------------------------------ | -------------------------------------- |
| TP-030 | `ICUVISOR_TOOLSET` tiers (`core`/`full`) + `list_advanced_capabilities`        | TP-018                                 |
| TP-031 | MCP Resources (workout-syntax, event-categories, custom-item-schemas, profile) | TP-019, TP-013, TP-012, TP-004         |
| TP-032 | MCP Prompts (training analysis, recovery, planning, taper, coach triage)       | TP-031, TP-010, TP-011, TP-012         |
| TP-033 | Streamable HTTP transport (loopback-bound by default)                          | TP-003                                 |
| TP-034 | KR5 benchmark harness (token efficiency vs Python references)                  | TP-030, TP-031, TP-032, TP-033, TP-029 |

---

## v0.5 — Internal beta (TP-036 … TP-041)

Scaffolded 2026-05-15 from PRD §7.1 / §7.2.A / §7.2.H / §7.4 #5–#12 and ROADMAP.md v0.5.

| ID     | Title                                                             | Depends on                             |
| ------ | ----------------------------------------------------------------- | -------------------------------------- |
| TP-036 | OS keychain credential storage (macOS / Windows / libsecret)      | TP-007                                 |
| TP-037 | macOS signed installer + manual Claude Desktop / Code config docs | TP-036                                 |
| TP-038 | First-run onboarding subcommand (`icuvisor setup`)                | TP-036, TP-004, TP-035                 |
| TP-039 | Coach mode + per-athlete tool ACLs (feature-flagged)              | TP-018, TP-030, TP-002, TP-007, TP-038 |
| TP-040 | Post-update schema-change notification (catalog hash in `_meta`)  | TP-015, TP-007                         |
| TP-041 | v0.5 dogfood (5–10 invited athletes incl. ≥1 coach)       | TP-036…TP-040, TP-029                  |

---

## v0.5 — Internal beta roadmap catch-up (TP-078 … TP-086)

Scaffolded 2026-05-20 from unchecked ROADMAP.md v0.5 entries up to and including the NOTE-event discoverability pass.

| ID     | Title                                                            | Depends on                    |
| ------ | ---------------------------------------------------------------- | ----------------------------- |
| TP-078 | Installer/onboarding integration for keychain-backed credentials | TP-036, TP-038, TP-037        |
| TP-079 | Gear read/name-resolution pass                                   | TP-009, TP-025, TP-030        |
| TP-080 | HR and pace curve siblings to `get_power_curves`                 | TP-010, TP-008, TP-030        |
| TP-081 | Nutrition macros and calories-label clarification                | TP-009, TP-011, TP-007        |
| TP-082 | Null stripping for write-tool responses                          | TP-007, TP-020…TP-024, TP-026 |
| TP-083 | Per-source sleep-score scale labels in wellness provenance       | TP-011                        |
| TP-084 | Upstream-signal regression pack from 2026-05 behavior review     | TP-009, TP-012, TP-075        |
| TP-085 | Concrete Strava-import unavailable workaround text               | TP-009, TP-084                |
| TP-086 | NOTE-event discoverability pass                                  | TP-075                        |

---

## v0.6 — Analyzers (TP-087 … TP-100)

Scaffolded 2026-05-20 from all ROADMAP.md v0.6 analyzer entries.

| ID     | Title                                                                   | Depends on                     |
| ------ | ----------------------------------------------------------------------- | ------------------------------ |
| TP-087 | `analysis_metric` closed enum + unknown-metric hints                    | TP-030, TP-007                 |
| TP-088 | MCP Resource `icuvisor://analysis-formulas`                             | TP-031                         |
| TP-089 | Analyzer skeleton and mandatory `_meta` contract                        | TP-087, TP-088                 |
| TP-090 | Auto-lap disambiguation on `get_activity_intervals`                     | TP-012, TP-089                 |
| TP-091 | `analyze_trend`, `analyze_distribution`, `analyze_correlation`, efforts | TP-087, TP-088, TP-089, TP-090 |
| TP-092 | `get_activity_histogram` single-activity histogram tool                 | TP-087, TP-089, TP-008         |
| TP-093 | `compute_zone_time`, load balance, baseline, compliance                 | TP-087, TP-088, TP-089, TP-090 |
| TP-094 | `compute_activity_segment_stats` raw-stream analyzer                    | TP-008, TP-088, TP-089         |
| TP-095 | `get_fitness_projection` analyzer-family tool                           | TP-010, TP-089                 |
| TP-096 | Activation-hint pass on analyzer descriptions                           | TP-091…TP-095                  |
| TP-097 | Definition-drift guard for canonical formulas                           | TP-088, TP-091, TP-093, TP-094 |
| TP-098 | Analyzer toolset placement and core-promotion gate                      | TP-030, TP-091, TP-093, TP-100 |
| TP-099 | Upstream coverage audit for zone-time/load-balance analyzers            | TP-093                         |
| TP-100 | Extend KR5 benchmark harness for analyzer family                        | TP-034, TP-091…TP-095          |

---

## Distribution and client compatibility (TP-101 … TP-102)

Scaffolded 2026-05-20 from operator-requested Desktop Extension and SSE-transport decision work.

| ID     | Title                                                   | Depends on                    |
| ------ | ------------------------------------------------------- | ----------------------------- |
| TP-101 | ICUvisor Desktop Extension (`.mcpb`) for Claude Desktop | TP-037, TP-078, external spec |
| TP-102 | SSE transport decision for remote-client compatibility  | TP-033, external research     |

---

## Competitive follow-up — 2026-05-29 upstream MCP review (TP-118 … TP-122)

Scaffolded 2026-05-29 from public GitHub issues/PRs/metadata in competing intervals.icu MCP servers reviewed over the prior 7 days. Clean-room rule applies: tasks may use public behavior signals and upstream API/docs, but must not read or copy GPL/copyleft competitor source.

| ID     | Title                                                     | Depends on |
| ------ | --------------------------------------------------------- | ---------- |
| TP-118 | Activity tombstone delete endpoint                         | —          |
| TP-119 | Workout percent FTP regression                             | —          |
| TP-120 | Unit label audit for work, calories, and hydration         | —          |
| TP-121 | Public positioning for structured local-first output       | —          |
| TP-122 | Season planning prompt and context hardening               | —          |

---

## Forum follow-up — 2026-05-29 AI coach thread review (TP-123 … TP-133)

Scaffolded 2026-05-29 from public intervals.icu forum comments over the prior 4 days across IcuSync, Montis, LeCoach, IntervalCoach, and MyTrainPal threads. Clean-room rule applies: tasks may use public product behavior signals and upstream API/docs, but must not read or copy GPL/copyleft competitor source.

| ID     | Title                                           | Depends on |
| ------ | ----------------------------------------------- | ---------- |
| TP-123 | Calendar date resolver and future date anchors  | —          |
| TP-124 | Activity date resolution and detail-routing evals | —        |
| TP-125 | Activity tags and fueling regression/docs pass  | —          |
| TP-126 | Deterministic segment-comparison analyzer workflow | —       |
| TP-127 | Edit-in-place write safety evals                | —          |
| TP-128 | Plan health review prompt                       | —          |
| TP-129 | Readiness fallback guidance for null upstream readiness | —  |
| TP-130 | Token-safe workout library positioning          | —          |
| TP-131 | Workout change preview guidance                 | —          |
| TP-132 | Multiple same-day events regression pack        | —          |
| TP-133 | Gym and strength best-effort support plan       | —          |

---

## Maintenance — Go code audit cleanup (TP-042 … TP-049)

Scaffolded 2026-05-15 from the Go code audit dated 2026-05-15. No version-milestone dependency; land opportunistically before v0.5 dogfood (TP-041).

| ID     | Title                                                                                 | Depends on |
| ------ | ------------------------------------------------------------------------------------- | ---------- |
| TP-042 | Collapse `internal/tools/registry.go` interface-assertion sprawl (audit P0)           | —          |
| TP-043 | Remove process-global state from `internal/response/shaper.go` (audit P1)             | —          |
| TP-044 | Refactor `fetchActivitiesPage` complexity in `internal/tools/get_activities.go` (P1)  | —          |
| TP-045 | Harden `internal/intervals/client.go` `doJSONQuery` (retry split, LimitReader) (P1)   | —          |
| TP-046 | Deduplicate `ProfileClient` interface across tools/resources (P1)                     | —          |
| TP-047 | Consolidate `internal/response/shaper.go` tree-walker + drop marshal round-trip (P1)  | TP-043     |
| TP-048 | `internal/tools/` boilerplate helpers (`DecodeStrict`, `TextResult`) + dead code (P2) | —          |
| TP-049 | Misc Go hygiene cleanups (recover helper, ctx propagation, env in `config`) (P2)      | —          |

---

## Current State

This is the default task area for icuvisor. Tasks that don't belong
to a specific domain area are created here.

Taskplane is configured and ready for task execution. Use `/orch all` for
parallel batch execution or `/orch <path/to/PROMPT.md>` for a single task.

---

## Key Files

| Category | Path                        |
| -------- | --------------------------- |
| Tasks    | `taskplane-tasks/`          |
| Config   | `.pi/taskplane-config.json` |

---

## Technical Debt / Future Work

_Items discovered during task execution are logged here by agents._

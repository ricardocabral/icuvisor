# General — Context

**Last Updated:** 2026-05-13
**Status:** Active
**Next Task ID:** TP-030

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

| ID     | Title                                                                          | Depends on                  |
| ------ | ------------------------------------------------------------------------------ | --------------------------- |
| TP-018 | `ICUVISOR_DELETE_MODE` safety gate (registration-time filtering, no `confirm`) | TP-007                      |
| TP-019 | `workout_doc` write-path serializer (DSL round-trip, golden tests)             | TP-013                      |
| TP-020 | Event write cluster (`add_or_update_event`, `link_activity_to_event`, message) | TP-018, TP-019, TP-012      |
| TP-021 | `update_wellness` (full writable field set incl. `injury`, BP, lactate, lock)  | TP-018, TP-011              |
| TP-022 | `update_sport_settings` (FTP/threshold ungated; zones gated)                   | TP-018                      |
| TP-023 | Workout-library CRUD (`create_workout`, `update_workout`, `delete_workout`)    | TP-018, TP-019, TP-013      |
| TP-024 | Custom-items create/update                                                     | TP-018, TP-013              |
| TP-025 | Destructive deletes cluster (event, activity, custom-item, sport, gear)        | TP-018, TP-020, TP-023, TP-024 |
| TP-026 | `apply_training_plan` (dry-run default; replace gated)                         | TP-018, TP-019, TP-020, TP-023, TP-012 |
| TP-027 | `input_examples` on complex write tools (catalog invariant)                    | TP-020…TP-024, TP-026       |
| TP-028 | Adversarial safety test suite (safe-mode hardening, tool-not-found surrender)  | TP-018, TP-020…TP-026       |
| TP-029 | v0.3 dogfood (dedicated test-athlete account, write-path validation)           | TP-018…TP-028               |

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

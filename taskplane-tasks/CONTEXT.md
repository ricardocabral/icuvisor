# General — Context

**Last Updated:** 2026-05-11
**Status:** Active
**Next Task ID:** TP-017

---

## v0.2 — Read path (TP-007 … TP-016)

Scaffolded 2026-05-11 from PRD §7.2.C–D / §7.4 and ROADMAP.md v0.2.

| ID | Title | Depends on |
|----|-------|------------|
| TP-007 | Response shaping primitives (terse, null-strip, _meta, scale labels, TZ, athlete-ID, unit-system plumbing) | TP-002, TP-004 |
| TP-008 | Pace/unit enum + snake_case stream-key canonicalization | TP-002 |
| TP-009 | Activities read cluster + Strava-blocked detection | TP-007, TP-008 |
| TP-010 | Fitness / best-efforts / power-curves / training-summary / extended-metrics | TP-007, TP-008 |
| TP-011 | Wellness reads (sleep dual-scale, provenance, staleness, _native) | TP-007 |
| TP-012 | Events + `get_event_by_id` inconsistency handling + training plan | TP-007 |
| TP-013 | Workout-library + custom-items reads | TP-007 |
| TP-014 | Periodization parameters read (or documented upstream gap) | TP-007 |
| TP-015 | Tool-name disambiguation + CI schema-stability guards | TP-009…TP-013 |
| TP-016 | v0.2 dogfood (solo + 2–3 invited athletes, read-only) | TP-009…TP-015 |

---

## Current State

This is the default task area for icuvisor. Tasks that don't belong
to a specific domain area are created here.

Taskplane is configured and ready for task execution. Use `/orch all` for
parallel batch execution or `/orch <path/to/PROMPT.md>` for a single task.

---

## Key Files

| Category | Path |
|----------|------|
| Tasks | `taskplane-tasks/` |
| Config | `.pi/taskplane-config.json` |

---

## Technical Debt / Future Work

_Items discovered during task execution are logged here by agents._

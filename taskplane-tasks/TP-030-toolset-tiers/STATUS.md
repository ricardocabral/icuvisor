# TP-030-toolset-tiers: TP-030-toolset-tiers — Status

**Current Step:** Step 2: Per-tool tier membership
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-14
**Review Level:** 2
**Review Counter:** 3
**Iteration:** 1
**Size:** M

---

### Step 1: Tier enum and parsing

**Status:** ✅ Complete

- [x] Enum: `core` (default) and `full`; case-insensitive parsing; unknown/empty → `core`
- [x] Wire `ICUVISOR_TOOLSET` through config loading (`Config.Toolset`, `.env`/env precedence, defensive string rendering)
- [x] Propagate parsed toolset into app startup (`ServerInfo`) and log the resolved tier exactly once without tool names
- [x] Pin Step 1 behavior with tests for parsing, config loading, startup propagation, and minimal logging
- [x] Decide and document the package boundary: extend `internal/safety` vs new `internal/toolset`. Record the choice and rationale in `STATUS.md`

### Step 2: Per-tool tier membership

**Status:** 🟨 In Progress

- [ ] Each tool self-declares its tier (`core` or `full`); default for unmarked tools is `full` (opt-in to `core`)
- [ ] Curate the `core` set to the §7.2.E daily-use path: read activities/fitness/wellness/events, write events/wellness/messages, plus `icuvisor_list_advanced_capabilities`. Target ~17 tools; record the exact list in `STATUS.md`
- [ ] Test matrix: every tool's tier membership is asserted in a table-driven test so catalog drift is caught

### Step 3: Registry filtering composition

**Status:** ⏳ Not started

### Step 4: `icuvisor_list_advanced_capabilities`

**Status:** ⏳ Not started

### Step 5: `_meta` surfacing + docs

**Status:** ⏳ Not started

### Step 6: Verify

**Status:** ⏳ Not started

---

## Reviews

| #   | Type | Step | Verdict | File |
| --- | ---- | ---- | ------- | ---- |
| R001 | plan | 1 | REVISE | `.reviews/R001-plan-step1.md` |
| R002 | plan | 1 | APPROVE | `.reviews/R002-plan-step1.md` |
| R003 | code | 1 | APPROVE | `.reviews/R003-code-step1.md` |

---

## Discoveries

| Discovery | Disposition | Location |
| --------- | ----------- | -------- |
| Package boundary for toolset tiers | Extend `internal/safety` rather than create `internal/toolset`, because TP-018 already centralizes registration-time environment gates and capability decisions there; toolset tiering is an orthogonal registration gate using the same pattern. | Step 1 / `internal/safety` |

---

## Execution Log

| Timestamp  | Action      | Outcome                          |
| ---------- | ----------- | -------------------------------- |
| 2026-05-14 | Task staged | Scaffolded from ROADMAP.md v0.4   |
| 2026-05-14 12:05 | Task started | Runtime V2 lane-runner execution |
| 2026-05-14 12:05 | Step 1 started | Tier enum and parsing |

---

## Blockers

_None_

---

## Notes

- R001 required the Step 1 plan to explicitly include a separate `safety.Toolset` API, config loader plumbing, app startup propagation/logging, and tests before implementation.
- R002 approved the revised Step 1 plan for implementation.
- R003 approved the Step 1 implementation; reviewer verified `go test ./...` and diff checks.
| 2026-05-14 12:40 | Review R003 | code Step 1: APPROVE |

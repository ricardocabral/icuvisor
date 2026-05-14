# TP-030-toolset-tiers: TP-030-toolset-tiers — Status

**Current Step:** Step 1: Tier enum and parsing
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-14
**Review Level:** 2
**Review Counter:** 0
**Iteration:** 1
**Size:** M

---

### Step 1: Tier enum and parsing

**Status:** 🟨 In Progress

- [ ] Enum: `core` (default) and `full`; case-insensitive parsing; unknown/empty → `core`
- [ ] Log the resolved tier once at startup at INFO (count only — never leak tool names that hint roadmap state, consistent with TP-018)
- [x] Decide and document the package boundary: extend `internal/safety` vs new `internal/toolset`. Record the choice and rationale in `STATUS.md`

### Step 2: Per-tool tier membership

**Status:** ⏳ Not started

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

_Reserved for execution notes_

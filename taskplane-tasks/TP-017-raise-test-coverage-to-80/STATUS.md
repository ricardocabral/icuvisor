# TP-017: Raise Go test coverage to 80% — Status

**Current Step:** Step 1: Pick high-value coverage targets
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-13
**Review Level:** 1
**Review Counter:** 0
**Iteration:** 1
**Size:** M

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code
> changes. Workers expand steps when runtime discoveries warrant it — aim for
> 2-5 outcome-level items per step, not exhaustive implementation scripts.

---

### Step 0: Preflight

**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Confirm baseline with `go test ./... -coverprofile=coverage.out -covermode=atomic` and `go tool cover -func=coverage.out | tail -1`

---

### Step 1: Pick high-value coverage targets

**Status:** 🟨 In Progress

> ⚠️ Hydrate: Expand checkboxes when entering this step based on the current per-function coverage report.

- [ ] Generate per-function coverage report and select high-value targets
- [ ] Prioritize behavior-rich low-coverage functions over trivial command paths
- [ ] Record selected target files/functions and rationale here

---

### Step 2: Add focused intervals-package tests

**Status:** ⬜ Not Started

> ⚠️ Hydrate: Expand with specific activity-stream and wellness scenarios after reading existing intervals test helpers.

- [ ] Add activity-stream tests for JSON decoding, raw preservation, query parameters, required IDs, and errors
- [ ] Create `internal/intervals/wellness_test.go` with native-provider and `ListWellness` tests
- [ ] Targeted intervals tests pass

---

### Step 3: Add focused tool/toolcheck tests until the target is met

**Status:** ⬜ Not Started

> ⚠️ Hydrate: Expand based on remaining coverage gap after Step 2.

- [ ] Add extended-metrics tests for uncovered behavior-rich branches
- [ ] Add or extend toolcheck tests only if additional coverage is needed
- [ ] Targeted tests for touched packages pass

---

### Step 4: Testing & Verification

**Status:** ⬜ Not Started

- [ ] Coverage gate passes: `go test -race -count=1 -coverprofile=coverage.txt -covermode=atomic ./...`
- [ ] Total statement coverage is >= 80.0%
- [ ] `make lint` passes
- [ ] `make build` passes
- [ ] Local coverage artifacts removed or intentionally left untracked/ignored

---

### Step 5: Documentation & Delivery

**Status:** ⬜ Not Started

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] Discoveries logged
- [ ] Final coverage percentage recorded

---

## Reviews

| #   | Type | Step | Verdict | File |
| --- | ---- | ---- | ------- | ---- |

---

## Discoveries

| Discovery | Disposition | Location |
| --------- | ----------- | -------- |

---

## Execution Log

| Timestamp  | Action      | Outcome                         |
| ---------- | ----------- | ------------------------------- |
| 2026-05-13 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-13 12:21 | Task started | Runtime V2 lane-runner execution |
| 2026-05-13 12:21 | Step 0 started | Preflight |

---

## Blockers

_None_

---

## Notes

Current observed baseline before staging this task: 76.9% total statement coverage from `go test ./... -coverprofile=coverage.out -covermode=atomic`.

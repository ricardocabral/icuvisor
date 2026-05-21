# TP-055 — Reconcile three documentation conflicts

## Mission

Resolve three documentation conflicts that currently exist between the PRD, ROADMAP, README, and the implemented code. These are not migration artifacts — they predate the website refactor and would silently propagate into the new Hugo site if not fixed first. Each conflict has a concrete remediation; pick the right one per conflict and land it.

The three conflicts (verify each against the current code at the time of execution — do not trust the line numbers in this prompt blindly; they were captured on 2026-05-16):

### Conflict A — Analyzer family ghost

- `docs/prd/PRD-icuvisor.md` §7.2.C (around lines 276–320) declares an analyzer family (`analyze_*`, `compute_*`, `get_fitness_projection`) and a "~39 tools at v1.0" target.
- `README.md` (around lines 35–74) lists ~36 tools, **none** of them analyzers.
- `internal/tools/registry.go` (around lines 117–236) registers **zero** analyzers.
- `ROADMAP.md` does not mention an analyzer phase.

**Decision required:** Either (a) **defer in PRD** — move the analyzer family to "vNext / out of scope for v1" in `docs/prd/PRD-icuvisor.md`, drop the "~39 tools" target to match reality, and explain the rationale in a one-paragraph PRD changelog entry; **or** (b) **add a roadmap phase** — write a `## v0.6 — Analyzers` (or similar) section in `ROADMAP.md` with scope, deliverables, and rough phasing, and keep the PRD as-is.

Recommendation: **(a) defer in PRD**. Rationale: there is no implementation, no taskplane TP for an analyzer family, and no recent ROADMAP signal that this work is imminent. Deferring is the honest reflection of priorities; the family can re-enter the PRD if/when a TP gets queued. The executing agent must record the decision and the reasoning in `STATUS.md`. Get human sign-off before executing if uncertain.

### Conflict B — `get_planning_parameters` ROADMAP contradiction

- `ROADMAP.md` line ~22 lists `get_planning_parameters` as **checked off**.
- `ROADMAP.md` line ~29 lists it as **deferred**.

These cannot both be true. Inspect the tool registry to determine the truth:

- If `get_planning_parameters` **is registered** in `internal/tools/`, keep the checked-off line and **delete** the deferred line.
- If it **is not registered**, keep the deferred line and **delete** the checked-off line.
- If it is registered but the surface differs from the original ROADMAP description, fix both ROADMAP entries to match what the code actually does.

The README must also reflect this. If the tool is registered, it should appear in the README's tool list (and, after TP-051, in the generated catalog). If it is not, no README mention.

### Conflict C — `update_wellness` error contract not surfaced

- `docs/prd/PRD-icuvisor.md` line ~247 specifies that `update_wellness` returns `field_not_writable: sleepScore (device-managed)` when callers attempt to write device-owned fields.
- `README.md` `update_wellness` bullet describes the device-rejection behaviour in prose ("rejecting device-owned `sleepScore`/`_native` fields") but does not surface the `field_not_writable` **error code** an MCP client or test fixture would see.
- The website's reference page for `update_wellness` (landing under TP-052) must show the error contract so end users and integrators understand what to expect on a rejection.

**Decision required:** This is not a contradiction — it is missing user-facing documentation. The fix:

- Verify the actual error code emitted by `update_wellness` against `internal/tools/update_wellness.go`. The PRD's `field_not_writable` may or may not match the implementation's literal error string. **Code wins; update the PRD if needed.**
- Ensure the error contract appears on the website at `/reference/tools/#update_wellness` (the generator from TP-051 should pick this up automatically if the tool's MCP description includes the error in the summary; if not, the generator's `ToolDescriptor` may need an `errors` field — file a TP-051 amendment if so).
- Ensure the same error contract is captured in the README replacement on the website's `/reference/safety-modes/` or `/reference/tools/` page (whichever TP-052 chose for tool error contracts).

PRD anchors: §7.2.C catalog truthfulness, §7.4 #11 (schema/catalog stability), §6 KR5 (token efficiency).

ROADMAP positioning: doc-truth cleanup. No new user-visible capability.

Complexity: Blast radius 2 (PRD, ROADMAP, README, possibly tool registry metadata), Pattern novelty 1, Security 1, Reversibility 2 = 6 → Review Level 2. Size: S.

## Dependencies

- None. This task can land before, during, or after TP-050–054.
- **Recommended order:** TP-055 lands **before** TP-051 and TP-052 so the generator and migration consume corrected content.

## Context to Read First

- `CLAUDE.md`
- `docs/prd/PRD-icuvisor.md` §7.2.C and the line referenced for Conflict C.
- `ROADMAP.md` (the full file — the two `get_planning_parameters` lines are not adjacent).
- `README.md` §"MCP tool catalog".
- `internal/tools/registry.go` — confirm which tools are actually registered.
- `internal/tools/update_wellness.go` — confirm error code literal.
- `CHANGELOG.md`

## File Scope

Expected files:

- `docs/prd/PRD-icuvisor.md` — analyzer-family deferral (or, if (b) chosen, leave alone); possibly `update_wellness` error-code wording.
- `ROADMAP.md` — resolve the `get_planning_parameters` contradiction; possibly add a `v0.6 Analyzers` section (only if Conflict A choice is (b)).
- `README.md` — reflect ROADMAP/PRD truth for `get_planning_parameters` and `update_wellness` error.
- `internal/tools/update_wellness.go` — only if the error literal needs to change to match the documented contract; otherwise leave code alone and align docs to code.
- `CHANGELOG.md` — `[Unreleased]` under "Changed" / "Fixed" as appropriate.
- `taskplane-tasks/TP-055-reconcile-doc-conflicts/STATUS.md`.

Out of scope:

- Implementing the analyzer family (always out of scope for this task, regardless of Conflict A decision).
- Migrating end-user content to the website — TP-052.
- README slim-down — TP-054.

## Steps

### Step 1: Re-verify all three conflicts against the current tree

- [ ] Re-read PRD §7.2.C and confirm the analyzer-family entry and tool-count target.
- [ ] Confirm `internal/tools/registry.go` registers zero analyzers.
- [ ] Confirm both `get_planning_parameters` ROADMAP lines still exist and still contradict.
- [ ] Read `internal/tools/update_wellness.go` and capture the **exact** error literal returned when `sleepScore` is written.
- [ ] Record findings in `STATUS.md` Step 1 with file paths and current line numbers.

### Step 2: Resolve Conflict A (analyzer family)

- [ ] Choose (a) defer in PRD or (b) add roadmap phase. Default: (a). Justify in `STATUS.md`.
- [ ] (a): edit PRD §7.2.C — move analyzer family to a new "Deferred" subsection or to `vNext`; drop the "~39 tools" target to a realistic current number (or remove the count entirely and let the catalog be the truth); add a brief in-PRD note: "Analyzer family deferred; see CHANGELOG `[Unreleased]`."
- [ ] (b): edit `ROADMAP.md` — add a `## v0.6 — Analyzers` section (or appropriate version) listing the family, brief scope, and signal that it depends on the read-side fitness/efforts tools already shipped.

### Step 3: Resolve Conflict B (`get_planning_parameters`)

- [ ] Inspect registry; determine truth.
- [ ] Edit `ROADMAP.md` to remove the contradicting line. Keep one consistent statement.
- [ ] Update `README.md`'s tool list to match. (Note: TP-051 will subsequently replace the hand-list with a generator; this task only ensures the current README is consistent.)

### Step 4: Resolve Conflict C (`update_wellness` error contract)

- [ ] Confirm the error literal in code.
- [ ] If PRD wording does not match code, update the PRD wording. Code wins.
- [ ] Surface the error contract in `README.md`'s `update_wellness` bullet (one short clause) — this is interim until TP-052 puts it on the website.
- [ ] If TP-051 has already landed and the generator's `ToolDescriptor` lacks an `errors` field needed to render this on the website, file an amendment under `taskplane-tasks/TP-051-tool-catalog-generator/PROMPT.md` (append-only under "## Amendments").

### Step 5: CHANGELOG + verification

- [ ] `CHANGELOG.md` `[Unreleased]` entries:
  - Conflict A: "Changed — Analyzer family deferred to vNext; PRD updated." (or "Added — v0.6 Analyzers roadmap phase").
  - Conflict B: "Fixed — `get_planning_parameters` ROADMAP contradiction; entry now matches registry."
  - Conflict C: "Changed — `update_wellness` error contract surfaced in README; PRD wording aligned to code."
- [ ] `make build`, `make test`, `make lint` — these should all still pass; if Step 4 touched `internal/tools/update_wellness.go`, ensure tests still cover the error literal.
- [ ] `git grep -n 'analyze_\|compute_\|get_fitness_projection' README.md docs/prd ROADMAP.md` — confirm no stale references survive Conflict A's resolution.
- [ ] `git grep -n 'get_planning_parameters' ROADMAP.md README.md` — confirm exactly one consistent statement remains in each.

## Acceptance Criteria

- Conflict A resolved with a documented decision: PRD reflects whether the analyzer family is deferred or scheduled, and that decision matches reality.
- Conflict B resolved: `ROADMAP.md` no longer contradicts itself for `get_planning_parameters`, and `README.md` is consistent.
- Conflict C resolved: `update_wellness` error contract appears in the README, matches the code's actual error literal, and the PRD wording is aligned to code.
- `CHANGELOG.md` `[Unreleased]` records each fix.
- `STATUS.md` documents the verification of each conflict against the current tree, the decision made for Conflict A, and the resolution for B and C.
- All existing tests pass.

## Do NOT

- Do not implement the analyzer family.
- Do not change `update_wellness`'s actual behaviour or error literal to make the docs convenient. If docs are wrong, fix docs.
- Do not slim the README in this task — TP-054 owns that. Make minimal edits only to fix the three conflicts.
- Do not migrate any content to the website here — TP-052 owns that.
- Do not retag releases or amend prior CHANGELOG sections — only `[Unreleased]`.
- Do not add emojis (per CLAUDE.md).

## Documentation

Must update:

- `STATUS.md` — per-conflict verification, decision rationale (esp. Conflict A), grep outputs.
- `CHANGELOG.md` `[Unreleased]`.

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-055`, e.g., `TP-055 defer analyzer family in PRD`.

---

## Amendments

_Add amendments below this line only._

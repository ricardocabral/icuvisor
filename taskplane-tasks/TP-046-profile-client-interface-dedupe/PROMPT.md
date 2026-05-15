# TP-046 — Dedupe `ProfileClient` interface across tools and resources (audit P1)

## Mission

Two packages declare an identical `ProfileClient` interface against the same single concrete impl (`*intervals.Client`):

- `internal/tools/get_athlete_profile.go:26`
- `internal/resources/athlete_profile.go:22`

Same method set, same signature, same producer. Renaming the upstream method on `*intervals.Client` would silently require two coordinated edits, and there is no compiler signal that the two declarations are meant to stay in lockstep. Promote the interface to a single home and import it in both consumers.

This was identified in the 2026-05-15 Go audit as a P1 duplication issue. CLAUDE.md anchors: "smaller change, clearer commit" / "prefer editing existing files… sprawl is expensive." No PRD anchor — this is internal-only structural cleanup with no user-visible behaviour.

ROADMAP positioning: maintenance / debt paydown. Independent of any version milestone. Can land before or after TP-042 (registry collapse). If TP-042 lands first the tools-side declaration may become less load-bearing, but the duplication still exists at the type level and this task remains valid — coordinate with the TP-042 author on merge order, do not block on it.

Complexity: Blast radius 1 (two files import a type from a third), Pattern novelty 1 (standard Go interface placement), Security 1 (no credential paths), Reversibility 1 = 4 → Review Level 1. Size: S.

## Dependencies

- None blocking. TP-042 is a coordination touch-point only — both tasks can land in either order. Resolve any trivial textual merge in whichever lands second.

## Context to Read First

- `CLAUDE.md` — "Default to `internal/`", no panic outside main, no sprawl.
- `internal/tools/get_athlete_profile.go` — first `ProfileClient` declaration + its consumer.
- `internal/resources/athlete_profile.go` — second `ProfileClient` declaration + its consumer.
- `internal/intervals/client.go` — the sole concrete impl.
- `internal/tools/get_athlete_profile_test.go` and `internal/resources/athlete_profile_test.go` (if present) — existing fakes that must continue to satisfy the shared interface.

## File Scope

- `internal/clients/profile.go` (new) **OR** an addition to `internal/intervals` — see "Placement choice" below.
- `internal/tools/get_athlete_profile.go` — remove local declaration, import the shared interface.
- `internal/resources/athlete_profile.go` — remove local declaration, import the shared interface.
- `internal/tools/get_athlete_profile_test.go` — verify fake still satisfies (no change expected if method set is identical).
- `internal/resources/athlete_profile_test.go` — same.
- `CHANGELOG.md` — `[Unreleased]` under "Changed".
- `taskplane-tasks/TP-046-profile-client-interface-dedupe/STATUS.md`.

Out of scope: hunting for other duplicated single-method interfaces across `internal/tools/` and `internal/resources/`. If you spot one in passing, file a P2 follow-up; do not bundle it into this PR.

## Placement choice

Two acceptable approaches — pick one and justify in `STATUS.md`:

1. **New `internal/clients` package** holding small cross-consumer interfaces. Default recommendation: this is likely to repeat as more tools and resources share methods, and a neutral home keeps both consumer packages free of cross-imports.
2. **Define on the producer side, in `internal/intervals`.** Standard Go practice is interfaces at the consumer; this is the opposite, but justifiable when there is one impl and ≥2 consumers. Acceptable if you'd rather not introduce a new package for a single type today.

**Default recommendation:** `internal/clients`. The audit explicitly flagged this is likely to recur; a thin shared package costs almost nothing and gives the next duplication a home. Escalate to the producer-side variant only if you have a concrete reason `internal/clients` is wrong here.

Do **not** put the interface in `pkg/` — CLAUDE.md hard rule, internal only.

## Steps

### Step 1: Confirm the duplication is exact

- [ ] Diff the two `ProfileClient` interface declarations. Confirm identical method set, signatures, and doc comments — or normalize trivial differences (whitespace, comment wording) so the merged declaration loses no information.
- [ ] Confirm `*intervals.Client` satisfies the (normalized) interface.
- [ ] Locate every fake / stub used in tests for both consumers; confirm method sets match.

### Step 2: Create the shared declaration

- [ ] Create `internal/clients/profile.go` (or the chosen alternative) with the single `ProfileClient` interface and a doc comment explaining who implements it and who consumes it.
- [ ] Remove the two duplicate declarations.
- [ ] Update imports in both consumer files.

### Step 3: Tests

- [ ] `go build ./...` — fakes that previously satisfied the local interfaces must still satisfy the shared one (since method sets are identical).
- [ ] `make test` and `make test-race`.

### Step 4: Verify

- [ ] `make build`, `make test`, `make test-race`, `make lint`.
- [ ] `grep -rn "type ProfileClient interface" internal/` returns exactly one hit.
- [ ] `git diff --stat` shows the two consumer files shrinking and one new file (or one expanded file in `internal/intervals`).

## Acceptance Criteria

- Exactly one declaration of `ProfileClient` in the repo (verified by `grep -rn "type ProfileClient interface" internal/`).
- Both `internal/tools/get_athlete_profile.go` and `internal/resources/athlete_profile.go` import the shared interface.
- All existing test fakes continue to satisfy the shared interface unchanged.
- `make build` / `make test` / `make test-race` / `make lint` all clean.
- Placement decision (`internal/clients` vs `internal/intervals`) recorded in `STATUS.md` with a one-paragraph justification.
- `CHANGELOG.md` `[Unreleased]` entry under "Changed".

## Do NOT

- Do not promote the interface to `pkg/`. CLAUDE.md hard rule: `internal/` by default.
- Do not change the method set in passing — this task is a placement move, not a redesign. Any signature change must be a separate PR.
- Do not rename `ProfileClient` to something else as part of this task. Renaming + relocating in one diff makes review noisier than it needs to be.
- Do not bundle other interface-dedupe work into this PR. File P2 follow-ups for anything else you spot.
- Do not introduce new dependencies; this is a stdlib-only move.

## Documentation

Must update:

- `STATUS.md` (placement decision + checklist progress)
- `CHANGELOG.md` `[Unreleased]` under "Changed"

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-046`, for example: `TP-046 introduce shared ProfileClient interface`, `TP-046 import shared ProfileClient in tools and resources`.

---

## Amendments

_Add amendments below this line only._

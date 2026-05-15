# TP-047 — Consolidate `internal/response/shaper.go` tree-walkers and drop the marshal round-trip (audit P1 + adjacent P2s)

## Mission

`internal/response/shaper.go:309-417` does ad-hoc JSON tree-walking over `map[string]any` / `[]any` with string-path bookkeeping — `joinPath`, `isDebugPath`, `dropDebugMetadata`, `isProvenancePath`, `isProvenanceFetchedAtPath`. Five near-identical recursive walkers each carry their own path-predicate inline. Combined with `marshalToJSONValue` (~line 110) — the marshal-then-unmarshal "round-trip to `any`" — every shaped response pays an extra encode/decode just so the shaper can manipulate generic maps.

Goal: collapse to either
- (preferred) shape on **typed structs**, using `json.RawMessage` for `include_full` passthrough; or
- (fallback) a **single tree-walker** that takes a visitor function, with each path-class encoded as a small predicate.

Whichever path the implementer picks, eliminate the marshal round-trip on the happy path. Output must remain byte-identical for the same inputs — this is structural cleanup, not a behaviour change.

Absorb these P2 items from the same audit while you're in the file:

- `internal/response/shaper.go:26-37` — `defaultScaleLabels` is data sitting inside the algorithm file. Move to `internal/response/scales.go`.
- `internal/response/shaper.go:166-179` — `shapeWrapperRow` is a near-duplicate of `shapeRow`. Extract a common helper.

This was identified in the 2026-05-15 Go audit as the most expensive single hotspot in `internal/response/`.

PRD anchors: §7.2.E (`_meta` envelope shape — public contract, must not drift), §7.2.D / §7.4 (terse-by-default tool outputs; `include_full` opt-in).

ROADMAP positioning: maintenance / debt paydown ahead of v0.5 dogfood. Independent of any user-visible milestone.

Complexity: Blast radius 2 (every tool that emits `_meta`), Pattern novelty 2 (typed shaping is a small invention if that path is chosen), Security 1, Reversibility 2 = 7 → Review Level 2. Size: M (could grow to L if the typed-shape approach is chosen; pick that only if the diff stays tractable).

## Dependencies

- **TP-043** — remove global state in shaper. Sequenced: land TP-043 first. This task assumes `Options` carries delete-mode and toolset.

No other blocking deps.

## Context to Read First

- `CLAUDE.md` — typed structs over `map[string]any`; no `interface{}` overuse; no panic outside `main`; structured errors.
- `internal/response/shaper.go` — whole file; this is the primary refactor target.
- `internal/response/shaper_test.go` — existing coverage; the safety net for byte-identical output.
- `internal/response/doc.go` — the documented shape contract for `_meta` and `include_full`.
- A representative tool that round-trips through the shaper: `internal/tools/get_activities.go`, `internal/tools/get_fitness.go`.
- `docs/prd/PRD-icuvisor.md` §7.2.D, §7.2.E, §7.4 for the public-contract anchors.

## File Scope

- `internal/response/shaper.go` — primary refactor.
- `internal/response/scales.go` (new) — `defaultScaleLabels` lives here.
- `internal/response/shaper_test.go` — update to match the new internals; do not weaken coverage.
- `internal/response/testdata/` — add golden outputs for ~5 representative tool responses to prove byte-identical behaviour. Create the directory if it does not exist.
- `CHANGELOG.md` — `[Unreleased]` under "Changed" (internal refactor; note that no user-visible output changes).
- `taskplane-tasks/TP-047-shaper-tree-walker-consolidation/STATUS.md`.

Out of scope:
- Changing the public `_meta` shape.
- Changing `include_full` semantics.
- Touching individual tool files beyond what the shaper API forces.
- Introducing a JSONPath library dependency.

## Approach choice

Two acceptable approaches — pick one and justify in `STATUS.md`:

1. **Typed shaping (preferred).** Define typed structs for each shaped envelope; use `json.RawMessage` for `include_full` passthrough so the heavy payload is never decoded. The marshal round-trip disappears entirely on the happy path.
2. **Single visitor-based tree-walker (fallback).** One recursive function that takes a path-predicate visitor; each existing walker becomes a small predicate (`debugPathPredicate`, `provenanceFetchedAtPredicate`, etc.). Still walks `map[string]any` but the five near-duplicates collapse to one. The marshal round-trip can still be removed if the upstream call sites pass the typed value directly to the walker.

**Default recommendation:** start with (1). If the diff balloons past M-size or the typed structs end up mirroring half of `internal/intervals/`, fall back to (2) and document why in `STATUS.md`.

## Steps

### Step 1: Snapshot pre-refactor output

- [ ] Pick ~5 representative tool responses (suggest: `get_activities` terse, `get_activities` `include_full`, `get_fitness`, one wrapper-row tool, one with provenance metadata). Capture their current `_meta`-shaped output as golden fixtures under `internal/response/testdata/`.
- [ ] Commit the fixtures **before** changing shaper code so the diff is auditable.

### Step 2: Pick the approach

- [ ] Decide typed-shape vs single visitor walker. Justify in `STATUS.md` (size of diff, blast radius, fit with `include_full`).
- [ ] If typed-shape: enumerate the envelope structs you intend to introduce and where `json.RawMessage` will sit.
- [ ] If single-walker: sketch the visitor signature and the predicate set you'll keep.

### Step 3: Implement

- [ ] Remove the marshal-then-unmarshal round-trip from `marshalToJSONValue` on the happy path. If a small fallback case remains (e.g., one exotic tool's payload), justify it explicitly in `STATUS.md`.
- [ ] Collapse the five near-identical recursive walkers (`joinPath`, `isDebugPath`, `dropDebugMetadata`, `isProvenancePath`, `isProvenanceFetchedAtPath` and friends) to either typed traversal or a single visitor.
- [ ] Preserve every existing path predicate's semantics. The pre/post snapshot diff is the contract.

### Step 4: Adjacent P2 cleanups

- [ ] Move `defaultScaleLabels` from `shaper.go:26-37` to a new `internal/response/scales.go`. No behaviour change.
- [ ] Extract the common helper shared by `shapeRow` and `shapeWrapperRow`. Both call sites now go through it; the wrapper variant adds only the wrapper-specific bits.

### Step 5: Verify byte-identical output

- [ ] Re-run the snapshot fixtures from Step 1. Diff must be empty.
- [ ] If the diff is non-empty, **stop**: either the refactor changed semantics or the fixtures are wrong. Resolve before continuing.

### Step 6: Build / test / lint

- [ ] `make build`, `make test`, `make test-race`, `make lint`.
- [ ] Eyeball-benchmark on a large response (e.g., `get_activities` with `include_full` on a real account fixture). The refactor must not materially regress wall-clock or allocations — losing the marshal round-trip should make it strictly faster on the happy path.

## Acceptance Criteria

- Marshal-then-unmarshal round-trip eliminated on the happy path. Any remaining narrow case is documented in `STATUS.md` with rationale.
- The five near-duplicate recursive walkers in `shaper.go` are replaced by either typed shaping or a single tree-walker with predicate visitors.
- `defaultScaleLabels` lives in `internal/response/scales.go`.
- `shapeRow` and `shapeWrapperRow` share an extracted helper.
- Pre/post snapshot diff for the chosen fixture set is empty — output is byte-identical for the same inputs.
- `_meta` envelope shape and `include_full` semantics unchanged.
- `make build` / `test` / `test-race` / `lint` all pass.

## Do NOT

- Do not change the `_meta` shape. It is public contract per PRD §7.2.E.
- Do not change `include_full` semantics — terse remains terse, full remains full.
- Do not introduce a JSONPath library or any other new dep to "simplify" path matching.
- Do not regress performance materially. The refactor should be neutral or faster; never slower.
- Do not touch individual tool files beyond what the shaper API change strictly requires.
- Do not weaken existing test coverage — golden fixtures are additive, not a replacement for the existing assertions.

## Documentation

Must update:

- `STATUS.md`
- `CHANGELOG.md` `[Unreleased]` under "Changed" (internal refactor; note no user-visible output change)

`internal/response/doc.go` should only change if the chosen approach changes the package's documented contract — for an internal refactor it should not.

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-047`, e.g. `TP-047 snapshot _meta golden fixtures`, `TP-047 collapse shaper tree-walkers`, `TP-047 move defaultScaleLabels to scales.go`.

---

## Amendments

_Add amendments below this line only._

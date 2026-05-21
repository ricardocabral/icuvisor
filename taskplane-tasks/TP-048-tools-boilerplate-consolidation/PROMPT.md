# TP-048 — Tools-package boilerplate consolidation (audit P2 bundle)

## Mission

Modest, focused consolidation of duplicated boilerplate across the ~40 files in `internal/tools/`, plus a couple of small inline cleanups identified in the 2026-05-15 Go audit. Resist over-abstracting — `CLAUDE.md` prefers explicit and greppable over clever, and "three similar lines is better than a premature abstraction." This task sits at the edge of that line; land only the four changes below and nothing else.

Concretely:

1. **`tools.DecodeStrict[T any](raw json.RawMessage) (T, error)`** — Every tool repeats `decoder := json.NewDecoder(bytes.NewReader(raw)); decoder.DisallowUnknownFields(); if err := decoder.Decode(&req); errors.Is(err, io.EOF) { … }`. See `internal/tools/get_activities.go:188-198` as the exemplar. Extract a single generic helper.
2. **`tools.TextResult(shaped any) Result`** — ~10 handlers hand-build `Result{Content: []Content{{Type: ContentTypeText, Text: …}}, StructuredContent: shaped}`. Collapse to one helper.
3. **Dead code:** `internal/tools/get_activities.go:511` `stringSet` returns `map[string]bool`; per the audit grep, the function has only its definition site. Delete it.
4. **Named struct, not inline:** `internal/tools/get_activities.go:240-264` `validateActivitiesTokenArgs` takes an inline anonymous struct as a parameter — give it a name.
5. **`Requirement` enum:** `internal/tools/registry.go:286-293` defines string constants for capability requirements; convert to a typed enum (typed `int` with `iota`, or a typed string) so `switch` exhaustiveness lints have something to bite on. Keep the wire format stable if it is serialized anywhere.

PRD anchors: §7.2.C catalog stability — tool schemas must not move.

ROADMAP positioning: maintenance / debt paydown. Independent of any version milestone. Friendly to land **after** TP-042 (registry refactor) to avoid trivial merge conflicts, but not blocked by it.

Complexity: Blast radius 2 (touches many tool files, mechanically), Pattern novelty 1 (plain Go), Security 1 (no new credential paths), Reversibility 2 = 6 → Review Level 2. Size: M.

## Dependencies

- None blocking. Easier to land after **TP-042** (registry collapse) so the registry diff is settled when the `Requirement` enum lands; not a hard dependency.

## Context to Read First

- `CLAUDE.md` — especially the "three similar lines is better than a premature abstraction" framing and the "prefer editing existing files" guidance. Read carefully; this task is at the edge of that line.
- `internal/tools/doc.go` — package overview and conventions.
- `internal/tools/registry.go` — `Requirement` constants and how they are consumed.
- 3-4 representative tool files: `internal/tools/get_activities.go` (the exemplar), `internal/tools/get_fitness.go`, `internal/tools/update_wellness.go`, `internal/tools/create_workout.go`. Confirm the decode/result patterns are uniform enough to collapse.
- `internal/toolchecks/schema_stability.go` — the snapshot test that guards catalog stability.

## File Scope

- `internal/tools/decode.go` (new) — `DecodeStrict[T]` + tests in `decode_test.go`.
- `internal/tools/result.go` (new) — `TextResult` + tests in `result_test.go`.
- All ~40 `internal/tools/*.go` tool handler files — replace the two boilerplate clusters with helper calls. Mechanical edits only.
- `internal/tools/registry.go` — `Requirement` typed enum; update call sites within the package.
- `internal/tools/get_activities.go` — remove `stringSet`; name the inline args struct passed to `validateActivitiesTokenArgs`.
- `CHANGELOG.md` — `[Unreleased]` under "Changed".
- `taskplane-tasks/TP-048-tools-boilerplate-consolidation/STATUS.md`.

Out of scope:

- Any tool schema change (catalog must be byte-identical).
- Renaming tool constructors or files.
- Adding new helpers beyond the two above.
- Restructuring how registration happens — that is TP-042's job.
- Unifying request/response struct definitions across tools (these are intentionally per-tool).

## Steps

### Step 1: Helpers + tests

- [ ] Add `internal/tools/decode.go` with `DecodeStrict[T any](raw json.RawMessage) (T, error)`. Behaviour must match the existing inline pattern exactly: `DisallowUnknownFields`, treat `io.EOF` on empty input as "use zero value", wrap other decode errors with a stable prefix.
- [ ] Add `internal/tools/result.go` with `TextResult(shaped any) Result`. Marshalling rules must match the existing inline pattern exactly (same JSON marshaler, same `ContentTypeText` content slot, same `StructuredContent` placement).
- [ ] Table-driven tests for both: zero/empty input, unknown-field rejection, malformed JSON, happy path; for `TextResult`, verify both `Content[0].Text` and `StructuredContent` are populated and shape-equal to a hand-built `Result`.

### Step 2: Mechanical replacement across tool files

- [ ] Replace the decode boilerplate in every `internal/tools/<tool>.go` with `DecodeStrict`. Keep the local `req` variable name; do not rename it.
- [ ] Replace the `Result{Content: …, StructuredContent: …}` boilerplate with `TextResult` everywhere it is exact. **Do not** rewrite handlers that build a non-text content slot or set extra fields — leave those alone.
- [ ] Commit per logical batch (e.g., reads cluster, writes cluster, wellness cluster) so review can land in chunks if needed.

### Step 3: `get_activities.go` cleanups

- [ ] Delete `stringSet`. Confirm with `grep -rn "stringSet" internal/` that it has no callers before removal.
- [ ] Promote the inline anonymous struct passed to `validateActivitiesTokenArgs` to a named type (e.g., `activitiesTokenArgs`) declared near the function. Update the function signature accordingly. No behavioural change.

### Step 4: `Requirement` enum

- [ ] Convert the string constants in `internal/tools/registry.go:286-293` to a named type. Two acceptable shapes; pick one and note the choice in `STATUS.md`:
  - `type Requirement int` with `iota` constants — best for switch exhaustiveness.
  - `type Requirement string` with typed constants — preserves the existing values if any caller is comparing or serialising them.
- [ ] Audit every reference (`grep -rn "Requirement" internal/`). Update call sites.
- [ ] If the value is serialised anywhere (logs, metadata, snapshot), preserve the wire format. The snapshot test will catch silent drift.

### Step 5: Verify

- [ ] `make build`, `make test`, `make test-race`, `make lint`.
- [ ] Schema-stability snapshot tests in `internal/toolchecks/` pass unchanged — tool catalog must be byte-identical.
- [ ] Run the acceptance greps below and confirm the counts collapse to the helper-definition site only.
- [ ] Manual smoke: start the server (stdio), `list_tools`, confirm output unchanged.

## Acceptance Criteria

- `grep -rn "DisallowUnknownFields" internal/tools/ | wc -l` drops to roughly `1` (only `decode.go` defines it).
- `grep -rn "ContentTypeText" internal/tools/ | wc -l` similarly collapses to the helper-definition site (or close to it; non-text handlers may legitimately remain).
- `stringSet` is gone; `grep -rn "stringSet" internal/` returns no hits.
- `validateActivitiesTokenArgs` takes a named struct type, not an inline anonymous one.
- `Requirement` is a named type with typed constants; `grep -rn "Requirement" internal/tools/` shows the type definition plus its consumers, no stringly-typed comparisons remain.
- Tool catalog snapshot is byte-identical to pre-refactor.
- `make build` / `test` / `test-race` / `lint` all clean.

## Do NOT

- Do not introduce additional helpers "while you're in there." Two helpers, two cleanups, one enum — that is the entire surface.
- Do not change any tool schema, name, or `_meta` shape.
- Do not change user-visible error-message wording (the existing wrapping format is part of the LLM-facing contract).
- Do not unify request/response struct definitions across tools — they are intentionally per-tool.
- Do not move or change registration logic — that is TP-042.
- Do not rewrite handlers that build non-text content or set extra `Result` fields just to fit `TextResult` — leave them as hand-built.
- Do not panic anywhere; preserve `%w` error wrapping.

## Documentation

- `STATUS.md`
- `CHANGELOG.md` `[Unreleased]` (under "Changed" — internal refactor, no user-visible behaviour change).

## Git Commit Convention

`TP-048 add DecodeStrict + TextResult helpers`, `TP-048 collapse decode boilerplate in read tools`, etc. One commit per logical batch.

---

## Amendments

_Add amendments below this line only._

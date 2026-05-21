# TP-034 — KR5 benchmark harness (token efficiency vs Python references)

## Mission

Build the benchmark harness that validates KR5: run a shared prompt set against icuvisor and both Python references (`hhopke/intervals-icu-mcp` and a second Python reference), and record per-session tool-description tokens and median per-call response bytes. Confirm or recalibrate the KR5 targets with measured deltas.

Roadmap items (ROADMAP.md v0.4):

- Benchmark harness: run a shared prompt set against icuvisor, `hhopke/intervals-icu-mcp`, and a second Python reference; record per-session description tokens and median per-call response bytes. KR5 targets confirmed or recalibrated.

PRD anchors: §6 KR5 (≥60% reduction in per-session tool-description tokens vs hhopke's 58-tool surface; ≥40% reduction in median per-call response bytes vs both references), §7.4 #8 (token efficiency validated by measurement), §7.4 #9 (KR5 may not be a standalone differentiator — measure honestly).

Complexity: Blast radius 1 (a harness, not a server change), Pattern novelty 3 (cross-server measurement methodology), Security 2 (runs against a real athlete account), Reversibility 1 = 7 → Review Level 2. Size: M/L.

## Dependencies

- **TP-030** — toolset tiers; the `core` tier is the headline KR5 number, so the harness must measure `core` and `full` separately.
- **TP-031** — MCP Resources; description-token savings depend on long-form content having moved to Resources.
- **TP-032** — MCP Prompts; part of the v0.4 surface being measured.
- **TP-033** — Streamable HTTP transport; the harness should be transport-agnostic but at minimum runs over stdio.
- **TP-029** — v0.3 dogfood complete; the full read+write catalog must be in place before the surface is meaningful to benchmark.

## Context to Read First

- `CLAUDE.md`
- `docs/prd/PRD-icuvisor.md` §6 KR5, §7.4 #8/#9, §7.2.E
- `ROADMAP.md` v0.4
- `internal/mcp/` — how to enumerate the registered catalog + descriptions
- The v0.2/v0.3 dogfood prompt sets (TP-016, TP-029 artifacts) — reuse/extend rather than authoring a fresh prompt set
- Install docs for `hhopke/intervals-icu-mcp` (`uvx`) and the second Python reference (`uv sync`)

## File Scope

Expected files:

- `scripts/benchmark/` (or `internal/benchmark/` + a `cmd/` entry — pick one, justify in `STATUS.md`) — the harness
- `scripts/benchmark/prompts/` — the shared prompt set (extend the dogfood sets)
- `scripts/benchmark/testdata/` or `results/` — recorded measurements (committed for reproducibility, redacted of athlete data)
- `docs/` — a short KR5 benchmark methodology + results doc
- `README.md` — pointer to the benchmark doc
- `CHANGELOG.md`
- `taskplane-tasks/TP-034-kr5-benchmark-harness/STATUS.md`

## Steps

### Step 1: Methodology

- [ ] Define the two metrics precisely: (a) per-session tool-description tokens — sum of tokens in all registered tool descriptions + schemas at `tools/list`; (b) median per-call response bytes over the shared prompt set
- [ ] Pin the tokenizer used for (a) and document it; pin the prompt set and the athlete account snapshot for (b)
- [ ] Decide how to handle non-determinism (fixed fixtures vs live account); document in `STATUS.md`. Prefer a frozen account snapshot so runs are reproducible

### Step 2: Shared prompt set

- [ ] Extend the TP-016 / TP-029 dogfood prompts into a shared set that exercises the common user-prompt shapes (the "10 most common common user prompts" from §7.4 #8)
- [ ] The same prompts must be runnable against all three servers — no icuvisor-specific assumptions in the prompt text

### Step 3: icuvisor measurement

- [ ] Measure `core` and `full` tiers separately; `core` is the headline KR5 number
- [ ] Capture description tokens from `tools/list` and response bytes per call

### Step 4: Reference-server measurement

- [ ] Stand up `hhopke/intervals-icu-mcp` and a second Python reference per their install docs; record exact versions in `STATUS.md`
- [ ] Run the same prompt set; capture the same two metrics
- [ ] **GPL boundary:** measuring the second Python reference as a black box (running it, reading its `tools/list` output, timing its responses) is fine. Do **not** read, copy, or transliterate its source into the harness or anywhere in this repo.

### Step 5: Results + KR5 verdict

- [ ] Compute the deltas: icuvisor `core` description tokens vs hhopke's 58-tool surface (target ≥60% reduction); median response bytes vs both references (target ≥40% reduction)
- [ ] Write the methodology + results doc in `docs/`; state plainly whether KR5 targets are confirmed or need recalibration (§7.4 #9 — measure honestly, do not flatter the result)
- [ ] If a target misses, file the gap and a recalibration proposal rather than quietly adjusting the KR

### Step 6: Repeatability

- [ ] The harness is re-runnable with one command; document it
- [ ] Committed results are redacted of any athlete PII

### Step 7: Verify

- [ ] `make test`, `make build`, `make lint`, `go test -race ./...` (for any Go harness code)
- [ ] Re-run the harness end to end; confirm results reproduce within a documented tolerance

## Reference Implementation Policy

- `hhopke/intervals-icu-mcp` (MIT) — run it, measure it, may consult its source. Do not vendor it.
- The second Python reference is measured as a black box only. Do not read, copy, paraphrase, or transliterate its source.

## Acceptance Criteria

- A re-runnable harness measures per-session tool-description tokens and median per-call response bytes.
- icuvisor (`core` and `full`), hhopke and the second Python reference are both measured on the same pinned prompt set; reference-server versions recorded.
- A `docs/` methodology + results doc states the measured deltas and a clear confirmed/recalibrate verdict for KR5.
- Committed results are reproducible within a documented tolerance and redacted of athlete PII.
- README, CHANGELOG updated.

## Do NOT

- Do not read, copy, or transliterate GPL/copyleft source — black-box measurement only.
- Do not cherry-pick prompts to flatter icuvisor's numbers; the prompt set is shared and fixed before measurement.
- Do not commit raw athlete data in the results fixtures.
- Do not silently move the KR5 goalposts — if a target misses, document it and propose recalibration explicitly.

## Documentation

Must update:

- `STATUS.md`
- `docs/` (new KR5 benchmark methodology + results doc)
- `README.md`
- `CHANGELOG.md`

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-034`, for example: `TP-034 define KR5 benchmark methodology`.

---

## Amendments

_Add amendments below this line only._

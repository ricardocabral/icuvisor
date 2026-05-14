# TP-034-kr5-benchmark-harness: TP-034-kr5-benchmark-harness — Status

**Current Step:** Step 5: Results + KR5 verdict
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-14
**Review Level:** 2
**Review Counter:** 10
**Iteration:** 3
**Size:** M/L

---

### Step 1: Methodology

**Status:** ✅ Complete

- [x] Define the two metrics precisely: (a) per-session tool-description tokens — sum of tokens in all registered tool descriptions + schemas at `tools/list`; (b) median per-call response bytes over the shared prompt set
- [x] Pin the tokenizer used for (a) and document it; pin the prompt set and the athlete account snapshot for (b)
- [x] Decide how to handle non-determinism (fixed fixtures vs live account); document in `STATUS.md`. Prefer a frozen account snapshot so runs are reproducible

### Step 2: Shared prompt set

**Status:** ✅ Complete

- [x] Extend the TP-016 / TP-029 dogfood prompts into a shared set that exercises the common forum-prompt shapes (the "10 most common forum prompts" from §7.4 #8)
- [x] The same prompts must be runnable against all three servers — no icuvisor-specific assumptions in the prompt text

### Step 3: icuvisor measurement

**Status:** ✅ Complete

- [x] Measure `core` and `full` tiers separately; `core` is the headline KR5 number
- [x] Capture description tokens from `tools/list` and response bytes per call

### Step 4: Reference-server measurement

**Status:** ✅ Complete

- [x] Stand up `hhopke/intervals-icu-mcp` and `mvilanova/intervals-mcp-server` per their install docs; record exact versions in `STATUS.md`
- [x] Run the same prompt set; capture the same two metrics
- [x] **GPL boundary:** measuring `mvilanova` as a black box (running it, reading its `tools/list` output, timing its responses) is fine. Do **not** read, copy, or transliterate its source into the harness or anywhere in this repo.

### Step 5: Results + KR5 verdict

**Status:** 🟨 In Progress

- [ ] Compute the deltas: icuvisor `core` description tokens vs hhopke's 58-tool surface (target ≥60% reduction); median response bytes vs both references (target ≥40% reduction)
- [ ] Write the methodology + results doc in `docs/`; state plainly whether KR5 targets are confirmed or need recalibration (§7.4 #9 — measure honestly, do not flatter the result)
- [ ] If a target misses, file the gap and a recalibration proposal rather than quietly adjusting the KR

### Step 6: Repeatability

**Status:** ⏳ Not started

- [ ] The harness is re-runnable with one command; document it
- [ ] Committed results are redacted of any athlete PII

### Step 7: Verify

**Status:** ⏳ Not started

- [ ] `make test`, `make build`, `make lint`, `go test -race ./...` (for any Go harness code)
- [ ] Re-run the harness end to end; confirm results reproduce within a documented tolerance

---

## Reviews

| #    | Type | Step | Verdict | File                          |
| ---- | ---- | ---- | ------- | ----------------------------- |
| R001 | plan | 1    | REVISE  | `.reviews/R001-plan-step1.md` |
| R002 | plan | 1    | APPROVE | inline                        |
| R003 | plan | 2    | REVISE  | `.reviews/R003-plan-step2.md` |
| R004 | plan | 2    | APPROVE | inline                        |
| R005 | code | 3    | REVISE  | `.reviews/R005-code-step3.md` |
| R007 | code | 4    | REVISE  | `.reviews/R007-code-step4.md` |
| R008 | code | 4    | REVISE  | `.reviews/R008-code-step4.md` |
| R009 | code | 4    | REVISE  | `.reviews/R009-code-step4.md` |
| R010 | code | 4    | APPROVE | inline                        |

---

## Discoveries

| Discovery | Disposition | Location |
| --------- | ----------- | -------- |

---

## Execution Log

| Timestamp        | Action         | Outcome                                         |
| ---------------- | -------------- | ----------------------------------------------- |
| 2026-05-14       | Task staged    | Scaffolded from ROADMAP.md v0.4                 |
| 2026-05-14 20:28 | Task started   | Runtime V2 lane-runner execution                |
| 2026-05-14 20:28 | Step 1 started | Methodology                                     |
| 2026-05-14 20:30 | Worker iter 1  | done in 87s, tools: 31                          |
| 2026-05-14 20:30 | No progress    | Iteration 1: 0 new checkboxes (1/3 stall limit) |
| 2026-05-14 21:04 | Worker iter 2 | done in 2077s, tools: 96 |
| 2026-05-14 21:04 | Step 4 started | Reference-server measurement |
| 2026-05-14 21:11 | Recovery | Reverted premature Step 4 completion before code review |
| 2026-05-14 21:35 | Step 4 completed | Reference-server measurement approved |
| 2026-05-14 21:35 | Step 5 started | Results and KR5 verdict |

---

## Blockers

_None_

---

## Notes

Step 1: Selected `scripts/benchmark/` rather than `internal/benchmark/` because the harness is an operator/release validation utility, not shipped server code. Metrics are canonical `tools/list` catalog-token counts and median canonical MCP `tools/call` result bytes. Tokenizer pinned to `cl100k_base` via `tiktoken==0.12.0`; prompt set pinned as `kr5-forum-prompts-v1`; frozen redacted account snapshot pinned as `kr5-redacted-test-athlete-v1` with manifest at `scripts/benchmark/testdata/snapshot-manifest.json`. Non-determinism is handled by committed redacted fixtures for reproducibility; live reruns are for recalibration only. R001 revisions removed premature results, pinned a real tokenizer, added call-plan guardrails, and defined snapshot manifest/redaction policy.

Step 2: R003 revisions replaced destructive KR5-10 with a non-destructive coach triage prompt and added `source_prompt_ids`/`prd_anchor` provenance to every scenario.

Step 3: Fixture run measured `icuvisor-core` (17 tools, 4,396 description tokens, 976.5 median audited response bytes) and `icuvisor-full` (38 tools, 9,490 description tokens, 1,154.0 median audited response bytes) in `scripts/benchmark/results/kr5-results.json`. R005 revisions added the harness/results/fixtures to the diff and replaced synthetic icuvisor catalog fixtures with exact `tools/list` output captured from `./bin/icuvisor` for core and full tiers.

Step 4: Stood up `hhopke/intervals-icu-mcp` with `uvx intervals-icu-mcp==2.0.0`, `INTERVALS_ICU_DELETE_MODE=full`, package version `2.0.0`, git tag `v2.0.0` (`d6d8f2b381db0776b0bb6d3ff1081d733bf0ac96`) to capture its 58-tool `tools/list` surface. Stood up `mvilanova/intervals-mcp-server` as a black box from a depth-1 checkout at `12199c61d88f580a885f04921b23dcf7c4524de8`, `uv sync --frozen`, and `uv run mcp run src/intervals_mcp_server/server.py`; installed package metadata reported `intervals-mcp-server==0.1.0` and `tools/list` returned 17 tools. R007/R008/R009 revisions count audited raw response bytes instead of benchmark padding, validate audited redacted byte sizes against committed fixtures, redact athlete IDs from measurement environments, support live `unavailable:<intent>` rows, validate unavailable calls as `isError=true`, preserve exact `.5` medians, and measured `hhopke-intervals-icu-mcp` (58 tools, 10,845 description tokens, 2,063.5 median response bytes) and `mvilanova-intervals-mcp-server` (17 tools, 6,227 description tokens, 1,649.5 median response bytes) in `scripts/benchmark/results/kr5-results.json`.
| 2026-05-14 20:35 | Review R001 | plan Step 1: REVISE |
| 2026-05-14 20:37 | Review R002 | plan Step 1: APPROVE |
| 2026-05-14 20:40 | Review R003 | plan Step 2: REVISE |
| 2026-05-14 20:43 | Review R004 | plan Step 2: APPROVE |
| 2026-05-14 20:54 | Review R005 | code Step 3: UNKNOWN |
| 2026-05-14 21:02 | Review R006 | code Step 3: UNKNOWN |
| 2026-05-14 21:15 | Review R007 | code Step 4: REVISE |
| 2026-05-14 21:22 | Review R008 | code Step 4: REVISE |
| 2026-05-14 21:31 | Review R009 | code Step 4: REVISE |
| 2026-05-14 21:35 | Review R010 | code Step 4: APPROVE |

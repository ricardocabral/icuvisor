# Review R012 — Plan review for Step 3

Verdict: **REVISE**

Reviewed:

- `taskplane-tasks/TP-100-analyzer-benchmark-harness/PROMPT.md`
- `taskplane-tasks/TP-100-analyzer-benchmark-harness/STATUS.md`
- prior Step 2 acceptance review `.reviews/R011-code-step2.md`
- `scripts/benchmark/kr5_benchmark.py`
- `scripts/benchmark/prompts/kr5_shared_prompts.json`
- `scripts/benchmark/testdata/fixtures/icuvisor-analyzer-family.json`
- `docs/kr5-benchmark.md`
- `ROADMAP.md` v0.6 analyzer/core-promotion entry

I also smoke-ran the fixture harness to confirm the current Step 2 implementation can generate v2 output with the real tokenizer:

```bash
python3 scripts/benchmark/kr5_benchmark.py \
  --mode fixtures \
  --prompt-set scripts/benchmark/prompts/kr5_shared_prompts.json \
  --fixture-dir scripts/benchmark/testdata/fixtures \
  --output /tmp/kr5-step3-review.json \
  --generated-at 2026-05-20T00:00:00Z
```

That produced `schema_version: kr5-benchmark-result-v2` and the expected analyzer-family contrast (`analyzers_enabled` response tokens 321 / raw stream pulls 0; `analyzers_disabled` response tokens 2266 / raw stream pulls 3), so there is no harness execution blocker. The plan itself needs tightening before implementation because Step 3 is the point where TP-100 feeds a roadmap gating decision.

## Blocking plan gaps

1. **Define the TP-098 “net-savings” formula before recording the verdict.**

   The current Step 3 plan says to “flag whether `analyze_trend`, `compute_zone_time`, and `compute_baseline` meet net-savings criteria,” but it does not define the criteria. This is easy to misreport because the v2 result contains aggregate mode summaries, while TP-098 is about individual core-promotion candidates. The plan should specify the calculation, for example:

   - use `response_tokens` from the paired prompt rows, not response bytes;
   - compare each enabled candidate prompt against its disabled fetch-and-reduce rows:
     - `KR5-A01`: disabled `get_activities` + `get_activity_streams` vs enabled `analyze_trend`;
     - `KR5-A02`: disabled `get_activity_streams` vs enabled `compute_zone_time`;
     - `KR5-A03`: disabled `get_activities` + `get_fitness` vs enabled `compute_baseline`;
   - subtract the candidate’s incremental catalog-description token cost if the phrase “net” is intended to include core tool-description overhead;
   - state the threshold for “meets” (`> 0` net token savings, or another explicit threshold if the task owner intended one).

   Do not base the TP-098 verdict only on the full `analyzers_enabled` vs `analyzers_disabled` mode aggregate: the enabled catalog contains the whole analyzer family, including tools that are not TP-098 core-promotion candidates, so that aggregate is useful context but not a fair per-candidate promotion test.

2. **Plan the required `docs/kr5-benchmark.md` methodology update, not just a result-table append.**

   The current document still describes a v1 result with four catalog surfaces, `prompt_count: 10`, and response-byte-focused current results. After Step 2, the committed result will be v2 with `benchmark_modes`, `mode_summaries`, `response_tokens`, scoped analyzer prompts, and a synthetic `icuvisor-analyzer-family` fixture. Step 3 should explicitly update the report text so a reader can reproduce and interpret the analyzer comparison:

   - the scope should mention the analyzer-family fixture/modes separately from the four historical KR5 surfaces;
   - the metrics section should define response-token totals and raw-stream pull counts;
   - the running command/result metadata should reflect v2 output and the new generated timestamp;
   - the “Current results” section should preserve the existing KR5 headline comparison while adding a separate analyzer enabled/disabled table and TP-098 candidate table.

3. **Pin the committed-run command and tokenizer expectations in the plan.**

   Step 3 should run the fixture benchmark without `--allow-approx-tokenizer`; approximate output is only acceptable for smoke tests and must not be committed as KR5 evidence. The plan should include the exact command/output path and the chosen `--generated-at` value for `scripts/benchmark/results/kr5-results.json` so the result is reproducible and the docs can cite the same timestamp.

## Non-blocking implementation notes

- Include raw-stream pull deltas both at mode level and, where useful, per prompt. The roadmap target is specifically about reducing fetch-and-reduce behavior, so “3 to 0 LLM-visible raw stream pulls” should be visible in the analyzer section.
- Preserve the wording from Step 1/Step 2 that this is a deterministic call-plan benchmark, not evidence of autonomous model tool selection.
- When updating `STATUS.md`, fix the bookkeeping mismatch noted in R011: the execution log currently records R009 as `REVISE` while the review table correctly records it as `ACCEPT`.

Once the Step 3 plan names the run command and defines the net-savings/reporting methodology, it should be ready to execute.

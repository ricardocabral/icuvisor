# Review R013 — Plan review for Step 3

Verdict: **ACCEPT**

Reviewed:

- `taskplane-tasks/TP-100-analyzer-benchmark-harness/PROMPT.md`
- `taskplane-tasks/TP-100-analyzer-benchmark-harness/STATUS.md`
- prior Step 3 plan review `.reviews/R012-plan-step3.md`
- `scripts/benchmark/kr5_benchmark.py`
- `docs/kr5-benchmark.md`
- `ROADMAP.md` v0.6 analyzer benchmark/core-promotion entries

I also re-ran the planned fixture command to confirm the committed-evidence path is currently executable with the real tokenizer and no `--allow-approx-tokenizer`:

```bash
python3 scripts/benchmark/kr5_benchmark.py \
  --mode fixtures \
  --prompt-set scripts/benchmark/prompts/kr5_shared_prompts.json \
  --fixture-dir scripts/benchmark/testdata/fixtures \
  --output /tmp/kr5-r013-review.json \
  --generated-at 2026-05-20T00:00:00Z
```

The run produced `schema_version: kr5-benchmark-result-v2` and the expected analyzer-family contrast: `analyzers_disabled` response tokens 2266 / raw stream pulls 3 versus `analyzers_enabled` response tokens 321 / raw stream pulls 0.

## Why this plan is ready

The current Step 3 plan addresses the R012 blockers:

- it pins the exact fixture-mode command, output path, fixed timestamp, and real-tokenizer expectation for committed evidence;
- it explicitly updates `docs/kr5-benchmark.md` methodology for v2 fields (`benchmark_modes`, `mode_summaries`, `response_tokens`, scoped analyzer prompts, and raw-stream pull counts) instead of only appending a result row;
- it separates the historical KR5 comparison from the new analyzer-family fixture/mode comparison;
- it defines the TP-098 candidate calculation from paired prompt rows rather than from the full analyzer-family aggregate; and
- it defines a net-savings verdict as gross paired response-token savings minus the candidate analyzer tool's incremental catalog-description cost, with the additional requirement that the enabled candidate row has no LLM-visible raw stream pull.

That is enough for Step 3 to run and record results without overclaiming autonomous model behavior or promoting tools to core in this task.

## Implementation notes

- When computing the per-candidate catalog-description cost, use the same canonical catalog payload/tokenizer conventions as the harness. A small script/snippet is preferable to hand-counting, since the v2 result currently stores aggregate mode description tokens but not per-tool incremental costs.
- In the analyzer section of `docs/kr5-benchmark.md`, include both the TP-098 net-savings table and the roadmap-level analyzer target context: percent response-token reduction and raw-stream-pull reduction for the trend/distribution/correlation shapes. The current plan already requires token deltas and stream-pull counts; making the ≥40%/zero-stream target explicit will make the report easier to audit.
- While updating `STATUS.md`, also fix the existing bookkeeping mismatch noted in R011/R012: the execution log still says R009 was `REVISE`, while the review table correctly says `ACCEPT`.

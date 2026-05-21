# Review R010 — Code review for Step 2

Verdict: **REVISE**

Reviewed changes from `5bfaca8..HEAD`:

- `scripts/benchmark/kr5_benchmark.py`
- `scripts/benchmark/prompts/kr5_shared_prompts.json`
- `scripts/benchmark/testdata/fixtures/icuvisor-analyzer-family.json`
- `taskplane-tasks/TP-100-analyzer-benchmark-harness/STATUS.md`
- `taskplane-tasks/TP-100-analyzer-benchmark-harness/.reviews/R009-plan-step2.md`

Verification run:

- `python3 -m py_compile scripts/benchmark/kr5_benchmark.py`
- `python3 -m json.tool scripts/benchmark/prompts/kr5_shared_prompts.json >/dev/null`
- `python3 -m json.tool scripts/benchmark/testdata/fixtures/icuvisor-analyzer-family.json >/dev/null`
- `python3 scripts/benchmark/kr5_benchmark.py --mode fixtures --prompt-set scripts/benchmark/prompts/kr5_shared_prompts.json --fixture-dir scripts/benchmark/testdata/fixtures --output /tmp/kr5.json --allow-approx-tokenizer --generated-at 2026-05-20T00:00:00Z`

## Blocking issue

### Analyzer-enabled fixture catalog is not the analyzer family catalog

`icuvisor-analyzer-family.json` currently models `analyzers_enabled` by appending only these six tools to the base catalog:

- `analyze_trend`
- `compute_zone_time`
- `compute_baseline`
- `analyze_correlation`
- `compute_compliance_rate`
- `get_activity_histogram`

Those added catalog rows also use stub schemas such as `{"type":"object","additionalProperties":true}` instead of the real tool schemas/descriptions.

This undercounts the enabled-mode catalog overhead and makes the analyzer-family savings result unreliable. The task mission is to compare with and without the analyzer family, and the current registry/tool catalog identifies additional analyzer-family tools such as `analyze_distribution`, `analyze_efforts_delta`, `compute_activity_segment_stats`, `compute_load_balance`, and `get_fitness_projection` (`internal/tools/catalog.go` registers them; `kr5_benchmark.py` also treats `compute_*`, `analyze_*`, `get_activity_histogram`, and `get_fitness_projection` as analyzer-family tools). A user enabling the analyzer family would pay catalog tokens for those tools too, so Step 3 could overstate net token savings if it uses this fixture as-is.

Please make the enabled fixture catalog represent the actual analyzer-enabled catalog: same non-analyzer payloads as disabled mode, plus all registered analyzer-family tools with their real descriptions and schemas. Alternatively, if this benchmark is intentionally only for a smaller “promotion-candidate subset”, rename/scope it that way in the fixture and result semantics; but that would not satisfy the current TP-100 analyzer-family comparison as written.

## Non-blocking notes

- `STATUS.md` has a small bookkeeping mismatch: the review table records R009 as `ACCEPT`, but the execution log line says `Review R009 | plan Step 2: REVISE`. Please correct it while updating status.
- The new `server_scope`, `expected_tools_by_mode`, and `expected_source_tools_by_mode` validation paths are directionally good and the fixture run passes once using the approximate tokenizer.

# Review R016 — Plan review for Step 4

Verdict: **REVISE**

Reviewed:

- `taskplane-tasks/TP-100-analyzer-benchmark-harness/PROMPT.md`
- `taskplane-tasks/TP-100-analyzer-benchmark-harness/STATUS.md`
- prior accepted Step 3 review `.reviews/R013-plan-step3.md`
- `scripts/benchmark/kr5_benchmark.py`
- `scripts/benchmark/results/kr5-results.json`
- `docs/kr5-benchmark.md`
- `CHANGELOG.md`

I also smoke-ran the deterministic fixture command with a temp output and confirmed the current committed result is reproducible:

```bash
python3 scripts/benchmark/kr5_benchmark.py \
  --mode fixtures \
  --prompt-set scripts/benchmark/prompts/kr5_shared_prompts.json \
  --fixture-dir scripts/benchmark/testdata/fixtures \
  --output /tmp/kr5-step4-review.json \
  --generated-at 2026-05-20T00:00:00Z
cmp -s /tmp/kr5-step4-review.json scripts/benchmark/results/kr5-results.json
```

That passed, so there is no obvious execution blocker. The Step 4 plan itself is not yet concrete enough to implement safely.

## Blocking plan gaps

1. **The plan does not specify what regression coverage will be added.**

   Step 4 requires “tests or smoke checks so the harness does not silently break,” but `STATUS.md` only repeats that checkbox. Before coding, pin the exact coverage. A suitable minimal plan would be either:

   - add a stdlib Python unittest module under `scripts/benchmark/` covering the new v2/analyzer invariants; or
   - add a clearly named smoke-check script under `scripts/benchmark/` that runs the fixture benchmark and validates the committed result.

   The coverage should be no-network and fixture-only. It should specifically protect the behavior added in this task, not just check that the Python file imports. Recommended assertions:

   - prompt `server_scope` prevents analyzer-only prompts from being required by legacy fixtures;
   - analyzer disabled mode rejects analyzer-family tools and exposes no analyzer tools in its catalog;
   - enabled/disabled analyzer catalogs keep byte-identical non-analyzer payloads;
   - `source_tool_usage` normalization/count validation catches missing or wrong source tools;
   - `raw_stream_pull_count` counts only top-level `get_activity_streams` aliases and excludes unavailable rows and analyzer `_meta.source_tools`;
   - fixture output has `schema_version: kr5-benchmark-result-v2`, both analyzer modes, `analyzers_disabled` raw-stream pulls `3`, `analyzers_enabled` raw-stream pulls `0`, and lower enabled response-token total.

2. **The plan does not define the exact smoke/quality commands or freshness check.**

   Add an explicit command sequence to Step 4. At minimum:

   ```bash
   python3 -m py_compile scripts/benchmark/kr5_benchmark.py
   python3 scripts/benchmark/kr5_benchmark.py \
     --mode fixtures \
     --prompt-set scripts/benchmark/prompts/kr5_shared_prompts.json \
     --fixture-dir scripts/benchmark/testdata/fixtures \
     --output /tmp/kr5-results-check.json \
     --generated-at 2026-05-20T00:00:00Z
   cmp -s /tmp/kr5-results-check.json scripts/benchmark/results/kr5-results.json
   ```

   If you add a unittest file, include the exact test command, for example:

   ```bash
   python3 -m unittest discover -s scripts/benchmark -p '*_test.py'
   ```

   Run the fixture smoke without `--allow-approx-tokenizer` for result freshness. Approximate-tokenizer output is fine only for local smoke experimentation, not for the committed KR5 evidence.

3. **Clarify Step 4 versus Step 5 quality gates.**

   Step 4 says to run the “full quality gate where applicable,” while Step 5 separately requires `make test`, `make build`, and `make lint`. The plan should avoid ambiguous double-counting by stating one of these approaches:

   - Step 4 runs benchmark-specific checks plus any fast Go gate touched by the change, and Step 5 reruns the full required project gates; or
   - Step 4 runs the full `make test && make build && make lint` gate now, and Step 5 records/reruns it as final verification.

   Either is acceptable, but it needs to be explicit so failures are not silently deferred or lost.

4. **Plan the `CHANGELOG.md` update before implementation.**

   `CHANGELOG.md` currently has no TP-100/KR5 benchmark entry. Step 4 should add a concise `[Unreleased]` bullet, likely under `### Added`, such as extending the KR5 benchmark harness/report with analyzer-enabled versus analyzer-disabled fixture comparisons, response-token metrics, and raw-stream-pull counts. Keep it user-visible but do not claim tool promotion to core; TP-100 only records evidence for TP-098.

## Suggested revised Step 4 plan

- Add no-network benchmark regression coverage in `scripts/benchmark/` for the analyzer-mode validation and v2 result invariants listed above.
- Run `python3 -m unittest discover -s scripts/benchmark -p '*_test.py'` if a unittest file is added.
- Run the deterministic fixture benchmark with fixed `--generated-at 2026-05-20T00:00:00Z` to `/tmp/kr5-results-check.json`, then compare it to `scripts/benchmark/results/kr5-results.json`.
- Run or explicitly defer the full Go quality gates to Step 5, documenting the decision and any failures in `STATUS.md`.
- Add the `[Unreleased]` changelog bullet for the benchmark harness/report visibility.
- Update `STATUS.md` with the exact commands and outcomes.

Once the plan names the concrete tests/smoke commands, the result freshness check, and the changelog placement, it should be ready to execute.

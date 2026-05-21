# Review R017 — Plan review for Step 4

Verdict: **ACCEPT**

Reviewed:

- `taskplane-tasks/TP-100-analyzer-benchmark-harness/PROMPT.md`
- `taskplane-tasks/TP-100-analyzer-benchmark-harness/STATUS.md`
- prior Step 4 plan review `.reviews/R016-plan-step4.md`
- `scripts/benchmark/kr5_benchmark.py`
- `scripts/benchmark/prompts/kr5_shared_prompts.json`
- `scripts/benchmark/results/kr5-results.json`
- `CHANGELOG.md`

The revised Step 4 plan now addresses the blocking gaps from R016: it names benchmark-specific no-network Python unittest coverage, defines the deterministic fixture freshness check against the committed v2 result, explicitly defers full Go project gates to Step 5, and includes the required `[Unreleased]` changelog update without claiming core promotion.

## Implementation notes

- Put the tests in a discoverable stdlib unittest file such as `scripts/benchmark/kr5_benchmark_test.py` so the planned command actually finds them:

  ```bash
  python3 -m unittest discover -s scripts/benchmark -p '*_test.py'
  ```

- Make at least some validation tests negative-path tests, not only assertions over the valid committed fixture. In particular, mutate in-memory copies and assert `BenchmarkError` for:
  - analyzer tools exposed or called in `analyzers_disabled`;
  - non-analyzer catalog payload drift between enabled/disabled modes;
  - missing or wrong `source_tool_usage` counts.

  Positive fixture-output invariant checks are useful, but they will not prove those validators fail closed if the checks are accidentally removed.

- Keep the freshness check exactly deterministic and credential-free, with no `--allow-approx-tokenizer`:

  ```bash
  python3 scripts/benchmark/kr5_benchmark.py \
    --mode fixtures \
    --prompt-set scripts/benchmark/prompts/kr5_shared_prompts.json \
    --fixture-dir scripts/benchmark/testdata/fixtures \
    --output /tmp/kr5-results-check.json \
    --generated-at 2026-05-20T00:00:00Z
  cmp -s /tmp/kr5-results-check.json scripts/benchmark/results/kr5-results.json
  ```

- Record command outcomes in `STATUS.md`. If `tiktoken==0.12.0` is unexpectedly unavailable, do not regenerate committed evidence with the approximate tokenizer; document the environment failure and install the pinned tokenizer or defer as a blocker.

With those execution details followed, the Step 4 plan is ready to implement.

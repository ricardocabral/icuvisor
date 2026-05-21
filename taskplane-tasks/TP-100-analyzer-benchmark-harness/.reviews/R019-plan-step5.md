# Review R019 — Plan review for Step 5

Verdict: **REVISE**

Reviewed:

- `taskplane-tasks/TP-100-analyzer-benchmark-harness/PROMPT.md`
- `taskplane-tasks/TP-100-analyzer-benchmark-harness/STATUS.md`
- `.reviews/R017-plan-step4.md`
- `.reviews/R018-code-step4.md`
- `Makefile`
- `scripts/benchmark/kr5_benchmark_test.py`

The Step 5 section currently repeats the high-level verification checklist, but it does not yet pin the concrete targeted benchmark commands or the failure-recording procedure. Because this task's main changes are in Python benchmark scripts/results and `make test` only runs Go tests, Step 5 needs an explicit targeted check plan before execution.

## Blocking plan gaps

1. **Targeted tests are not specified.**

   Add the exact benchmark-targeted commands to the Step 5 plan/status before running the full project gates. At minimum, rerun the no-network unittest suite added in Step 4:

   ```bash
   python3 -m unittest discover -s scripts/benchmark -p '*_test.py'
   ```

   Since the committed benchmark result is part of the deliverable and is not covered by `make test`, the plan should also rerun the deterministic freshness check unless there is a deliberate reason not to:

   ```bash
   python3 scripts/benchmark/kr5_benchmark.py \
     --mode fixtures \
     --prompt-set scripts/benchmark/prompts/kr5_shared_prompts.json \
     --fixture-dir scripts/benchmark/testdata/fixtures \
     --output /tmp/kr5-results-check.json \
     --generated-at 2026-05-20T00:00:00Z
   cmp -s /tmp/kr5-results-check.json scripts/benchmark/results/kr5-results.json
   ```

   Keep this fixture-only and do not use `--allow-approx-tokenizer` for committed evidence.

2. **Full gate command sequence and failure handling should be explicit.**

   The prompt requires these gates; put them in the Step 5 plan as the commands that will be run:

   ```bash
   make test
   make build
   make lint
   ```

   If any command fails, the plan should say whether the worker will fix the failure immediately or document it in `STATUS.md` with enough evidence to show it is pre-existing and unrelated. Missing local tools, especially `golangci-lint` for `make lint`, should be recorded as an environment/tooling blocker rather than marked as a passing lint gate.

3. **Status updates should capture evidence.**

   Plan to update `STATUS.md` with exact commands and outcomes after execution. While touching status, also correct the bookkeeping inconsistency noted in R018: the execution-history row for R017 says `REVISE`, but both the review table and `.reviews/R017-plan-step4.md` say `ACCEPT`.

Once the Step 5 plan names the targeted Python/freshness checks plus the required `make` gates and status-recording procedure, it should be ready to execute.

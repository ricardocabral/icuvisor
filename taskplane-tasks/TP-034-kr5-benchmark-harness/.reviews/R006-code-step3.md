# Review R006 — Code review for Step 3: icuvisor measurement

Verdict: Request changes

## Findings

1. **Step 3 artifacts are still not in the reviewed commit.** `git diff fd807b9..HEAD --name-only` still only reports `taskplane-tasks/TP-034-kr5-benchmark-harness/STATUS.md`, while the harness, fixtures, and results are untracked (`scripts/benchmark/kr5_benchmark.py`, `scripts/benchmark/results/`, `scripts/benchmark/testdata/`, etc.). A clean checkout of `HEAD` still cannot reproduce the checked-off Step 3 measurements in `STATUS.md:32-33`/`STATUS.md:110`. Add the intended benchmark files to git (and remove/ignore `scripts/benchmark/__pycache__/`) before marking the measurement captured.

2. **The response-byte metric is computed from synthetic padded fixtures, not captured `tools/call` results.** `scripts/benchmark/kr5_benchmark.py:283-286` derives `response_bytes` from the committed `call.result` JSON. The icuvisor fixtures contain generic one-row placeholder results with artificial `redacted_notes` padding, and every icuvisor core call is exactly 1,200 bytes while every full call is exactly 1,460 bytes. That does not demonstrate that per-call bytes were captured from real `tools/call` outputs over the shared prompt call plan. Preserve the actual measured byte count per call (or redacted call fixtures that maintain each real call's byte size with an audit trail) instead of recomputing the metric from synthetic placeholder payloads.

3. **The full-tier catalog result is not reproducible from the documented/example configuration.** `icuvisor-full` reports 38 tools / 9,490 tokens, which only matches `ICUVISOR_TOOLSET=full` plus `ICUVISOR_DELETE_MODE=full`; using the example's safe delete mode gives 31 tools / 8,732 tokens. Step 3 only says “full tier” and `benchmark-config.example.json` uses `ICUVISOR_DELETE_MODE=safe`, so the recorded full measurement omits a required capability dimension. Document the exact env used for each icuvisor measurement (or measure/report full-safe and full-delete separately) and update the config/results metadata accordingly.

## Notes

- I reran the fixture harness; the metric values reproduce, aside from `generated_at`, but that currently only proves determinism of the synthetic fixtures.
- I also compared live `tools/list` from `./bin/icuvisor` with the icuvisor fixtures: core matches, and full matches only when delete mode is `full`.

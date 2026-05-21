# R009 code review — Step 4: Reference-server measurement

Verdict: **REVISE**

## Findings

1. **Median response-byte values are floored for the 32-call prompt set.**  
   `summarize()` casts `statistics.median(response_bytes)` to `int` (`scripts/benchmark/kr5_benchmark.py:425`). With an even number of calls this silently floors `.5` medians, so the committed metrics are not the documented median per-call response bytes: hhopke is 2063.5 but reports 2063, the second Python reference is 1649.5 but reports 1649, and icuvisor-core is 976.5 but reports 976 (`scripts/benchmark/results/kr5-results.json:1`). Please either record the median as a JSON number/float or explicitly define and document a lower-median/rounded-integer policy before Step 5 computes KR5 deltas.

## Checks run

- `git diff 328d16c..HEAD --name-only`
- `git diff 328d16c..HEAD`
- `python3 scripts/benchmark/kr5_benchmark.py --mode fixtures --prompt-set scripts/benchmark/prompts/kr5_shared_prompts.json --fixture-dir scripts/benchmark/testdata/fixtures --output /tmp/kr5-results-review.json --generated-at 2026-05-14T20:00:00Z`
- `cmp -s scripts/benchmark/results/kr5-results.json /tmp/kr5-results-review.json`
- Fixture sanity scripts for server summaries, unavailable rows, redaction envs, and exact median values

# R008 code review — Step 4: Reference-server measurement

Verdict: **REVISE**

## Findings

1. **Live result output can leak the real athlete ID.**  
   `redact_env()` only redacts keys containing `KEY`, `TOKEN`, `SECRET`, or `PASSWORD` (`scripts/benchmark/kr5_benchmark.py:352-367`). The benchmark records `measurement_env` into `kr5-results.json`, and the committed result/fixtures include `INTERVALS_ICU_ATHLETE_ID` unredacted. Today it is `i12345`, but a real live rerun would write the actual athlete identifier to the result file, violating the task redaction requirement and the repo rule against logging raw athlete identifiers. Please redact/omit athlete identifiers (for example keys containing `ATHLETE`, `ATHLETE_ID`, or similar) before writing results, and regenerate fixtures/results.

2. **The audited byte metric is not independently validated.**  
   `audited_response_bytes()` returns `redaction_audit.raw_response_bytes` and only compares it to `redaction_audit.redacted_response_bytes` from the same JSON object (`scripts/benchmark/kr5_benchmark.py:313-342`). It never computes the redacted fixture size after excluding benchmark-only audit/padding, so any typo or arbitrary audit value can change the Step 4 medians while still passing validation. Because the reference-server response-byte results now rely on these audit integers, the harness should validate `redacted_response_bytes` against the committed redacted payload (or store a generated sidecar/audit file with a reproducible stripping rule) before using `raw_response_bytes`.

3. **Live mode cannot reproduce unavailable/error calls for missing reference tools.**  
   The methodology requires explicit unavailable/error results when a server lacks an equivalent tool, but `load_live_measurements()` always sends `item["tool"]` to `tools/call` (`scripts/benchmark/kr5_benchmark.py:237-240`) and `MCPClient.request()` aborts on JSON-RPC errors (`scripts/benchmark/kr5_benchmark.py:133-134`). A live run of the second Python reference with the fixed prompt set will either fail on missing tools or require omitting required intents, while fixture validation expects `unavailable:*` `isError=true` rows. Add a supported config path to synthesize unavailable rows or capture JSON-RPC tool errors as measured error results.

## Checks run

- `git diff 328d16c..HEAD --name-only`
- `git diff 328d16c..HEAD`
- `python3 scripts/benchmark/kr5_benchmark.py --mode fixtures --prompt-set scripts/benchmark/prompts/kr5_shared_prompts.json --fixture-dir scripts/benchmark/testdata/fixtures --output /tmp/kr5-results-review.json --generated-at 2026-05-14T20:00:00Z`
- `cmp -s scripts/benchmark/results/kr5-results.json /tmp/kr5-results-review.json`
- Fixture sanity script checking listed tools, unavailable calls, and audit presence

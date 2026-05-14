# R007 code review — Step 4: Reference-server measurement

Verdict: **REVISE**

## Findings

1. **Response-byte metrics count benchmark-only padding/audit fields, not measured reference responses.**  
   `summarize()` measures `len(canonical_bytes(call.result))` (`scripts/benchmark/kr5_benchmark.py:331-337`), but the reference fixtures put `redaction_audit` and large `redacted_padding` fields inside the counted MCP result object. The committed medians therefore diverge far beyond the documented ±1% byte policy (`docs/kr5-benchmark.md:58`): hhopke raw audit median is `2063.5` bytes while `kr5-results.json` reports `3240`; mvilanova raw audit median is `1734` while results report `2564`. This can materially change the KR5 response-byte verdict (icuvisor core vs mvilanova is ~35% reduction using the audited raw median, not ≥40%). Please either count the audited redacted/raw response byte field with validation, or keep audit/padding metadata outside the measured `result` so canonical fixture bytes match the measured server output.

2. **mvilanova “tool calls” include fabricated successful unavailable calls.**  
   `scripts/benchmark/testdata/fixtures/mvilanova-intervals-mcp-server.json:1` records 13 calls with `tool: "unavailable:*"` even though those names are not in the captured `tools/list`, and their `result.isError` is `false` with normal-looking rows/padding. That is not a black-box `tools/call` measurement and violates the methodology’s “explicit unavailable/error call result” rule (`docs/kr5-benchmark.md:45`). The harness also accepts this because coverage validation only checks intent presence, not whether the tool exists or unavailable calls are errors. Please record actual MCP errors or explicit `isError: true` unavailable results, and add validation/reporting so synthetic unavailable calls cannot be counted as successful response bytes.

3. **Step 4 artifacts are not in `HEAD`.**  
   Per the requested review command, `git diff db17aad..HEAD --name-only` shows only `STATUS.md`; the reference fixture/result/doc changes are still uncommitted in the worktree. If CI/review consumes `HEAD`, the Step 4 measurement artifacts are absent. Please commit/stage the Step 4 files or adjust the review target before marking the step complete.

## Checks run

- `git diff db17aad..HEAD --name-only`
- `git diff db17aad..HEAD`
- `git diff --name-only`
- `python3 scripts/benchmark/kr5_benchmark.py --mode fixtures --prompt-set scripts/benchmark/prompts/kr5_shared_prompts.json --fixture-dir scripts/benchmark/testdata/fixtures --output /tmp/kr5-results-review.json`
- Fixture sanity scripts comparing call tools against captured catalogs and audited raw medians against reported medians

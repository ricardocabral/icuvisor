# Review R005 — Code review for Step 1

Verdict: **REVISE**

Reviewed `git diff e035c36..HEAD` for Step 1. Changed files:

- `scripts/benchmark/kr5_benchmark.py`
- `taskplane-tasks/TP-100-analyzer-benchmark-harness/STATUS.md`

Smoke checks run:

- `python3 -m py_compile scripts/benchmark/kr5_benchmark.py`
- `python3 scripts/benchmark/kr5_benchmark.py --prompt-set scripts/benchmark/prompts/kr5_shared_prompts.json --fixture-dir scripts/benchmark/testdata/fixtures --output /tmp/kr5-test.json --allow-approx-tokenizer --generated-at 2026-01-01T00:00:00Z`

## Blocking findings

1. **Analyzer mode catalog validation only compares non-analyzer tool names, not the catalog payload.**  
   In `validate_analyzer_mode_catalogs` (`scripts/benchmark/kr5_benchmark.py:373-378`), the enabled/disabled fairness check compares only sets of non-analyzer tool names. That lets descriptions or input schemas for shared tools differ between `analyzers_enabled` and `analyzers_disabled` while still passing validation. Since the new v2 output uses per-mode catalog token counts, this can silently attribute non-analyzer description/schema drift to analyzer availability and invalidate the token delta TP-100 is meant to measure. Compare the canonical `catalog_payload` rows for all non-analyzer tools across modes, not just names, and report which shared tool differs.

2. **New response token metrics count benchmark-only `redaction_audit` metadata.**  
   `response_tokens` and `median_response_tokens` are computed from `canonical_json(call.result)` (`scripts/benchmark/kr5_benchmark.py:552` and `:570-572`). Fixture responses commonly include `redaction_audit` fields that the existing byte path explicitly strips/handles as benchmark-only metadata. Counting those audit fields inflates the new token metrics and makes v2 token deltas measure committed fixture scaffolding rather than MCP response payloads. Add a token-measurement helper that mirrors the audited-byte stripping path, or add a fixture audit field for raw token counts and validate it the same way bytes are validated.

## Non-blocking finding

3. **`expected_tools_by_mode` has no unavailable/error escape hatch.**  
   The status design says analyzer-enabled rows should call expected analyzer-family tools unless the fixture explicitly marks an unavailable/error case. Current validation accepts `unavailable:*` rows for intent coverage (`scripts/benchmark/kr5_benchmark.py:407-413`) but then still requires every `expected_tools_by_mode` tool to appear in `calls_by_prompt_mode` (`:432-438`). A valid unavailable fixture will therefore fail expected-tool validation. Either skip expected-tool assertions for prompt/mode pairs with explicit `unavailable:*` error rows, or add an explicit per-prompt/per-mode unavailable marker so fixture authors can represent the intended exception deterministically.

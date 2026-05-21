# Review R009 — Plan review for Step 2

Verdict: **ACCEPT**

Reviewed:

- `taskplane-tasks/TP-100-analyzer-benchmark-harness/PROMPT.md`
- `taskplane-tasks/TP-100-analyzer-benchmark-harness/STATUS.md`
- prior Step 2 reviews `.reviews/R007-plan-step2.md` and `.reviews/R008-plan-step2.md`
- current harness shape in `scripts/benchmark/kr5_benchmark.py`
- current shared prompt set in `scripts/benchmark/prompts/kr5_shared_prompts.json`
- roadmap analyzer/core-promotion requirements for TP-098

The revised Step 2 plan is ready to implement. It addresses the prior blockers:

- `server_scope` is now planned as an explicit prompt-level filter so analyzer-only prompts can live in the shared prompt file without forcing unrelated legacy/reference fixtures to add synthetic analyzer coverage.
- prompt-level `expected_source_tools_by_mode` is now part of the validation plan, which gives the analyzer fixture a real contract for `_meta.source_tools` / fallback source-tool plans instead of merely echoing illustrative rows.
- `KR5-A02` now exercises `compute_zone_time` directly, so TP-100 will produce evidence for all three TP-098 benchmark-gated promotion candidates: `analyze_trend`, `compute_zone_time`, and `compute_baseline`.

## Implementation notes

These are not blockers, but they are worth preserving during coding:

1. Keep the scoped-prompt logic centralized and apply it consistently to required-intent coverage, expected top-level tools, expected source-tool usage, and any per-server call-count/applicability summaries. A scoped-out analyzer prompt should not be treated as missing coverage for legacy fixtures.

2. Make `expected_source_tools_by_mode` count-capable in the committed JSON. The STATUS example uses `{mode: {called_tool: [source tools...]}}`; if duplicate strings encode counts, document that in validation errors. Prefer object rows matching normalized call rows, e.g. `{ "tool": "get_activity_streams", "count": 1 }`, to avoid ambiguity.

3. For the concrete analyzer fixture, make the enabled/disabled top-level call plans explicit in the prompt metadata or fixture rows:
   - `KR5-A01`: `analyze_trend` vs fetch-and-reduce reads.
   - `KR5-A02`: `compute_zone_time` vs `get_activity_streams` raw binning.
   - `KR5-A03`: `compute_baseline` vs wellness/fitness/activity reads.
   - `KR5-A04`: `analyze_correlation` plus `compute_compliance_rate` vs wellness/events/activity reads.
   - `KR5-A05`: `get_activity_histogram` vs `get_activity_streams`.

4. Preserve the Step 1 fairness invariant for `icuvisor-analyzer-family`: byte-identical non-analyzer catalog payloads between `analyzers_enabled` and `analyzers_disabled`, with the enabled catalog adding only analyzer-family tools.

5. For the roadmap savings target, enabled analyzer rows should avoid direct LLM-visible `get_activity_streams` calls wherever the prompt shape is intended to demonstrate avoided fetch-and-reduce behavior. Any internal analyzer source usage should remain in `source_tool_usage`, not in `raw_stream_pull_count`.

With those notes, the Step 2 plan fits the task scope and should unblock implementation.

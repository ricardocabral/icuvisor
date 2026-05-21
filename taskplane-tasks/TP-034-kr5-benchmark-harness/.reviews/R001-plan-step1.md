# Review R001 — Plan review for Step 1: Methodology

Verdict: Request changes

## Findings

1. **The methodology doc asserts results before the harness/results exist.** `docs/kr5-benchmark.md` includes concrete “Current results”, deltas, and a KR5 verdict, but there is no `scripts/benchmark/` harness or committed `scripts/benchmark/results/kr5-results.json` yet. For Step 1 this should stay as methodology with `TBD` result placeholders; factual measurements belong in Step 5 after live/reference runs and fixture generation.

2. **The pinned tokenizer is not a credible token-efficiency measure.** `icuvisor_kr5_regex_v1` ignores whitespace and treats long identifiers as one token, which can skew schema-heavy `tools/list` comparisons versus real MCP client/model tokenization. Pin a real model tokenizer (for example a documented OpenAI/Anthropic-compatible tokenizer with a permissive dependency) or explicitly justify why this approximation is acceptable for KR5 and add a calibration check against a real tokenizer.

3. **The response-byte methodology needs stronger guardrails against cherry-picking.** The doc says the harness executes “pinned MCP tool-call plans” and allows server-specific mappings outside the prompt text. That can make the benchmark measure curated calls rather than “the same prompts” unless the plan defines how mappings are fixed before measurement, how semantic equivalence is verified, and that all calls required to answer each prompt are counted.

4. **The frozen snapshot pin is too vague for reproducibility.** `kr5-redacted-test-athlete-v1` is named, but the plan should define the snapshot manifest: date range, included endpoints/fixtures, redaction rules, byte-size preservation policy, and where the manifest will live. This is needed to satisfy the Step 1 requirement to pin the athlete account snapshot and handle non-determinism.

## Recommendation

Keep the `scripts/benchmark/` choice and the canonical JSON definitions, but revise Step 1 before marking it complete: remove premature results, pin/justify a real tokenizer strategy, document deterministic prompt-to-tool-call mapping rules, and define the frozen snapshot manifest/redaction policy in `STATUS.md` and the methodology doc.

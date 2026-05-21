# Plan Review — TP-096 Step 1

**Verdict:** Approved with clarifications

Step 1 is an audit-only step, and the planned outcomes match the task prompt: enumerate analyzer-family tools, capture the current first sentence, flag missing activation hints / raw-row avoidance language, and check confusing overlap.

## Clarifications to apply while executing

1. **Derive the tool list from the registered catalog, not only file names.** The analyzer family spans multiple naming/grouping patterns and one file can register more than one tool. At minimum the audit should cover:
   - `analyze_trend`
   - `analyze_distribution`
   - `analyze_correlation`
   - `analyze_efforts_delta`
   - `compute_zone_time`
   - `compute_load_balance`
   - `compute_baseline`
   - `compute_compliance_rate`
   - `compute_activity_segment_stats`
   - `get_activity_histogram`
   - `get_fitness_projection`

2. **Treat raw-stream wording carefully.** Most analyzer descriptions should discourage fetching `get_*` rows/streams and reducing them in chat. `compute_activity_segment_stats` is the raw-stream analyzer exception, so the audit should flag it for exception wording rather than applying a generic “never use raw streams” rule.

3. **Capture the audit result in `STATUS.md`.** Step 1 should leave a concrete table or notes under Discoveries/Execution Log with each tool, its current first sentence, and pass/fail notes for activation hint, raw-row/stream avoidance, and confusable wording. This prevents Step 2 from relying on unstated local context.

4. **Use existing helpers/checks where possible.** `internal/toolchecks.FirstDescriptionSentence` and the confusable-name logic are the right reference behavior for “first sentence” and similarity checks; using the same semantics in the audit will make Step 2 tests easier to align.

No code or generated-doc changes are needed in Step 1 unless the worker intentionally advances into Step 2.

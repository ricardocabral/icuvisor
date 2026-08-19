# TP-260 follow-up checklist

This checklist is bounded to findings from the review of batch `20260818T192348`. It is not a production change and does not expand TP-260 into implementation work.

## TP-254 — should fix before the next contract-sensitive release

- [ ] Parameterize or split `rawActivityScaleInt` so `feel` accepts only 1–5 and `icu_rpe` accepts only 1–10.
- [ ] Preserve rejected out-of-range upstream values only in the existing full/raw payload boundary; omit the normalized field and do not infer a replacement.
- [ ] Add list and detail table cases for valid endpoints and values below/above each scale; retain malformed-type and sparse-write coverage.
- [ ] Add a focused review of the generated schema/public scale descriptions after implementation.

**Suggested Taskplane follow-up:** create a small TP-254 follow-up implementation task. This is a bounded correctness fix, not a reason to rewrite activity normalization.

## TP-252 — deferred observation

- [ ] Do not add speculative upstream window query parameters.
- [ ] If upstream documentation or black-box evidence establishes a server-side window contract, add request-query coverage first, then measure transfer/decode/memory cost before changing the local slicing path.

**Suggested Taskplane follow-up:** only create this task when verified upstream contract evidence exists; otherwise retain the current documented local-window behavior.

## No follow-up required from this review

- TP-257: no release finding; retain provenance, availability diagnostics, and no-inference boundaries.
- TP-259: no release finding; retain protocol/deadline coverage and honest local-vs-hosted documentation.

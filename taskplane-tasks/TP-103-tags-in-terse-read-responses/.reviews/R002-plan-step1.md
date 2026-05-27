# Plan Review: Step 1 — Implement event tag read shaping

**Verdict: Approved for implementation**

The revised Step 1 plan in `STATUS.md` addresses the gaps from R001. It now defines tolerant extraction semantics, preserves explicit upstream empty arrays via a copied pointer, accepts the shared `eventRow` impact across event response helpers, and moves the important edge-case tests into Step 1 instead of deferring them.

## Notes for implementation

- Keep tag decoding at the shaping boundary (`event.Raw["tags"]`) unless you add a tolerant custom decoder. A plain typed `[]string` field on `intervals.Event` would risk decode failures or losing the distinction between missing/null/empty.
- The planned `*[]string` row field is the right shape for preserving `tags: []` while omitting missing/null/malformed values under the existing response shaper.
- Because `eventRow` is also used by paths beyond the four named in the prompt (`apply_training_plan` and delete response helpers), treat tags as an inherited shared-row addition there as well, not a regression.
- The targeted tests listed in Step 1 are sufficient. If `internal/intervals/events.go` is modified, include interval decoding tests for null/non-array/mixed tag payloads and raw preservation; otherwise tool-level decode/shape tests are enough.

No plan blockers remain.

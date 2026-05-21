# Plan Review — TP-067 Step 1 (Inventory)

## Verdict: Approved

The updated Step 1 plan now contains the guardrails needed for a safe mechanical split. It records a complete declaration-to-file mapping, explicitly handles the previously missing `DefaultPath`, coach helper methods, TP-049 `DebugMetadata`, and TP-062 `LogValue` additions, and defines the `Load` wrapper/delegation approach so callers and exported signatures remain unchanged.

## What looks good

- The inventory covers all current top-level declarations in `internal/config/config.go` and assigns each to a focused target file.
- The proposed boundaries match the task concerns: load composition, validation, write/atomic file handling, athlete IDs, HTTP bind parsing, dotenv parsing, redaction, and default path resolution.
- Keeping `Load` in `config.go` as the public entry point while delegating composition to an unexported helper in `load.go` satisfies the acceptance criterion without exposing new API surface.
- The validation slicing plan preserves the current order and calls out defaults, precedence, error-string preservation, coach-mode behavior, final URL slash trimming, and final `Config` assembly fields.
- Import/dependency notes are sufficient for keeping extraction commits small and compileable.
- Test split candidates are documented, including redaction, write, HTTP bind, athlete normalization, validation, and load precedence coverage.

## Minor implementation notes

- The plan uses both `validateHTTPBind` and `validateHTTPBindAddress` as names. Pick one during implementation to avoid unnecessary churn; the exact name is not important as long as behavior and error strings stay unchanged.
- `recognizedEnvKey` currently has a narrower `.env` allow-list than `rawFromEnv` can parse from process env. The plan correctly says to preserve the exact recognized key set; do not opportunistically add `ICUVISOR_DEBUG_METADATA` or other keys to `.env` support in this refactor.
- Since `DefaultPath` is moving to `path.go`, keep `config.go` limited to types/defaults/`Load` plus the small config semantics already documented (`Transport.String`, `EffectiveCoachMode`, `CoachModeEnabled`). Avoid letting path or write logic drift back into `config.go`.
- The Step 1 checkbox for actually slicing `validate` remains unchecked in `STATUS.md`. That is fine if the actual code slice is intentionally part of Step 2, but do not mark Step 1 complete until the workflow owner is satisfied that the recorded call-order plan is enough for this checklist item.

## Recommendation

Proceed to Step 2. Keep the implementation mechanical: move declarations without behavior changes first, then split `validate` in the recorded order, running targeted config tests after each extraction.

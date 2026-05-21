# Plan Review — TP-083 Step 2: Apply provenance labels

## Verdict: Approve

The Step 2 plan in `STATUS.md` now addresses the gaps from R004. It names the relevant wellness shaping functions, constrains the change to provenance generation unless missing source evidence is discovered, carries forward the Step 1 source/field/sidecar scale matrix, and explicitly protects canonical response-level scale labels.

## Notes

- The plan correctly keeps provider-native labels under `_meta.provenance.<field>.native_scale` and leaves `_meta.scales.sleepScore` governed by `internal/response/scales.go`.
- The sidecar-aware handling for `sleepScore` and `readiness` is specific enough, including Polar readiness precedence of `nightly_recharge_status` before `ans_charge` and `unknown` source paired with `native_scale: unknown`.
- The stale/provenance preservation criteria are testable: keep `fetched_at`, `stale`, and `stale_reason` behavior intact and run `go test ./internal/tools` before moving to Step 3.
- Deferring expanded divergent-provider fixture coverage to Step 3 is acceptable because Step 2 includes a targeted regression check for existing wellness behavior.

## Non-blocking cleanup

`STATUS.md` still lists R004's verdict as `UNKNOWN` in the Reviews table even though `.reviews/R004-plan-step2.md` says request changes. Consider correcting that row to avoid audit confusion, but it does not block Step 2 implementation.

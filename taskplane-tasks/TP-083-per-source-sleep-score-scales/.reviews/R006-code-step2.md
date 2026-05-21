# Code Review — TP-083 Step 2: Apply provenance labels

## Verdict: Approve

No blocking findings for Step 2. The implementation applies the Step 1 provider/native scale matrix in wellness provenance without changing the canonical response scale labels, and the existing stale/provenance fixture coverage remains green.

## Notes

- `sleepScore` and `readiness` now resolve Garmin, Oura, Polar, and WHOOP native sidecars before falling back to row-level provider evidence.
- Polar readiness keeps the sidecar-specific precedence: `nightly_recharge_status` before `ans_charge`.
- The added test assertion protects `_meta.scales.sleepScore` as the canonical response-level label while `_meta.provenance.sleepScore.native_scale` carries the provider-native label.
- Step 3 should still add the planned divergent-provider fixture coverage, especially direct assertions for Garmin/WHOOP labels and unknown fallback behavior.
- Non-blocking bookkeeping: `STATUS.md` increments the review counter and logs R005, but the Reviews table still omits the R005 row and still lists R004 as `UNKNOWN` even though R004 requested changes.

## Verification

- `git diff --check 5bde733..HEAD` — pass
- `go test ./internal/tools ./internal/intervals` — pass

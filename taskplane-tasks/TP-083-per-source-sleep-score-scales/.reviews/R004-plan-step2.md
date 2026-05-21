# Plan Review — TP-083 Step 2: Apply provenance labels

## Verdict: Request changes

The Step 2 plan in `STATUS.md` is still only the original three checklist items. Step 1 now documents the exact source/field/sidecar matrix, but the Step 2 plan should describe how that matrix will be applied in the wellness response shape before implementation continues. Without that detail, it is easy to regress canonical `_meta.scales`, stale provenance, or the sidecar-specific Polar readiness behavior that Step 1 just fixed.

## Findings

1. **No concrete implementation path is recorded.**
   - Step 2 should name the functions/files that will be changed or verified, especially `internal/tools/get_wellness_data.go` around `wellnessProvenanceEntry`, `wellnessFieldSource`, `wellnessNativeScale`, and `wellnessReadinessNativeScale`.
   - The plan should say whether this step is code-only application of the Step 1 matrix or whether any native extraction changes in `internal/intervals/wellness.go` are still expected.

2. **The plan does not define field/source precedence for application.**
   - Step 1 notes Polar readiness precedence (`nightly_recharge_status` before `ans_charge`) and provider-evidence fallbacks, but Step 2 should explicitly carry that into shaping behavior.
   - Please state that `source: "unknown"` must pair with `native_scale: "unknown"`, and that Garmin/WHOOP/Oura labels are used only when source evidence actually resolves to those providers.

3. **Canonical-vs-native separation needs an implementation guard.**
   - The checklist says to keep response-level canonical scale labels separate, but the plan should explicitly avoid changing `internal/response/scales.go` canonical labels for `_meta.scales.sleepScore`.
   - Step 2 acceptance should include a quick assertion/verification that `_meta.scales.sleepScore` remains the canonical response label while `_meta.provenance.<field>.native_scale` carries the provider-native label.

4. **Stale/provenance preservation should be made testable before Step 3 fixture expansion.**
   - Step 2 says stale behavior remains intact, but does not specify verification. At minimum, run `go test ./internal/tools` after the shaping change and preserve the existing fresh/stale Polar fixture expectations.
   - If Step 2 changes source detection, include a note that `fetched_at`, `stale`, and `stale_reason` must continue using the resolved source without dropping provenance for unknown sources.

## Suggested Step 2 acceptance criteria

Before proceeding to Step 3, expand `STATUS.md` with a short implementation plan such as:

- Apply the Step 1 mapping only inside wellness provenance generation (`internal/tools/get_wellness_data.go`).
- Use sidecar-aware source/scale selection for `sleepScore` and `readiness`, including Polar readiness precedence.
- Preserve generic native units for non-score bridged fields (`sleepSecs`, `avgSleepingHR`, HRV, SpO2, etc.).
- Do not change canonical response-level scale labels in `internal/response/scales.go`.
- Verify with `go test ./internal/tools` that existing stale/provenance behavior remains green.


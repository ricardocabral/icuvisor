# Code Review: TP-011 Step 2 — Implement typed decoding

**Verdict: Revise.**

The new client/tool shape generally follows the Step 2 plan: `sleepQuality`, `sleepScore`, and `sleepSecs` are decoded as separate nullable fields, rows start from raw JSON, and native sidecars are exposed under `_native`. However, the current native-hoisting logic can drop unrecognized nested provider data from terse responses, which conflicts with the raw/custom preservation requirement.

## Findings

### 1. Nested provider objects are deleted even when they contain unclaimed fields

- **Where:** `internal/intervals/wellness.go:130-144`, `internal/tools/get_wellness_data.go:101-105`
- **Severity:** Blocking

`claimNested` appends the whole provider key (`polar`, `garmin`, or `oura`) to `NativeClaimedKeys` as soon as any recognized native field is present. `wellnessRow` then deletes that whole top-level provider object before adding `_native`.

For example, this upstream row:

```json
{
  "id": "2026-05-11",
  "polar": {
    "ans_charge": 6,
    "nightly_recharge_status": 4,
    "bridge_fetched_at": "2026-05-11T08:00:00Z",
    "custom_native_note": "kept by upstream"
  }
}
```

currently becomes terse output with only:

```json
"_native": { "polar": { "ans_charge": 6, "nightly_recharge_status": 4 } }
```

and drops `polar.bridge_fetched_at` / `polar.custom_native_note` unless `include_full:true` is requested. That violates the Step 2 contract to preserve unknown/custom JSON keys and the approved plan note that nested provider containers should only be removed when they contain only claimed native fields; otherwise claimed fields should be removed while unclaimed fields remain.

This is also risky for Step 3 because provider refresh timestamps or source markers may arrive as unrecognized nested fields and should not silently disappear from the normal shaped row path.

**Suggested fix:** Track claimed nested fields separately from claimed top-level keys. When building the terse row, if a provider object contains unclaimed keys, keep a shallow copy of that provider object with only the claimed native fields removed. Only delete the whole provider key when all provider-object keys were claimed. Add focused tests for a nested provider object with both recognized native fields and an extra timestamp/custom field.

## What I checked

- Ran `git diff 29627eef938aafcb47652d93fd3e5ccff24c3110..HEAD --name-only`.
- Ran `git diff 29627eef938aafcb47652d93fd3e5ccff24c3110..HEAD`.
- Read `PROMPT.md`, `STATUS.md`, `internal/intervals/wellness.go`, `internal/tools/get_wellness_data.go`, and `internal/tools/registry.go`.
- Ran `go test ./...` — it passes.

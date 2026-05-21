# Code Review — TP-080 Step 1

Verdict: **APPROVED**

No blocking issues found in the Step 1 changes.

## Review notes

- The new shared curve bucket helper preserves the existing `get_power_curves` observable behavior while making duration- and distance-axis lookup reusable for the HR/pace siblings.
- Power curve response fields remain typed and metric-specific (`duration_seconds`, `watts`, `activity_id`), with terse/full behavior unchanged.
- Intervals client regression coverage now covers HR `secs`, pace `distances`, supplied `type`, and the intentional HR/pace sport-omission behavior.
- The added power curve tests cover the Step 1 contract guardrails: name/description/tier/schema, terse default, `include_full`, defaults, curve spec, and missing-bucket metadata.

## Verification

Ran:

```sh
go test ./internal/intervals ./internal/tools
```

Result: passed.

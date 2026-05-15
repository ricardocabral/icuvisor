# Code Review — Step 2: Pick the approach

**Decision: Approved.**

The only change since `d09f0e2` is `STATUS.md`, and it now records the Step 2 approach clearly enough for implementation to proceed. The status update:

- marks the Step 2 checklist items as complete;
- chooses the fallback single visitor walker;
- explains why full typed shaping was rejected on package-cycle / mirrored-contract / diff-size grounds;
- sketches the visitor/predicate set that will replace the duplicated tree walkers; and
- documents the intended reflection-based `toJSONValue` replacement plus the narrow fallback policy for custom or unsupported values.

I did not find any blocking issues in this Step 2 diff.

## Implementation reminders for Step 3

These are guardrails, not approval blockers:

1. The reflection builder must produce fresh JSON-shaped maps/slices so the shaper does not mutate caller-owned DTOs or `include_full` payloads.
2. Preserve the current `encoding/json` behavior that the shaper implicitly relies on: JSON tags, `omitempty`, embedded fields, nil pointers/interfaces, custom marshalers such as `time.Time`, and `json.RawMessage` if encountered.
3. Keep any marshal/unmarshal fallback narrow and record the exact fallback class in `STATUS.md`; representative tool DTOs covered by the goldens should stay on the new happy path.
4. Treat the Step 1 goldens as the public-contract check before moving to the cleanup steps.

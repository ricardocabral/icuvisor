# Plan Review — TP-236 Step 2

## Verdict: APPROVE

The Step 2 plan is implementation-ready when read with the approved Step 1 contract. The pure function signature, strict zone validation, left-endpoint integration, final-sample behavior, invalid-interval precedence, 60-second gap rule, mismatch/short-input counters, coasting treatment, below-first-zone bucket, exact boundary inclusion, rounding/reconciliation, and J-to-kJ conversion are all deterministic and confined to the two planned analysis files.

The planned table-driven coverage targets the important numerical cases: irregular timestamps, zero-power intervals, exact zone boundaries, missing/misaligned samples, invalid timestamps and power, large gaps, names/below-zone behavior, all-zero work, and unit/rounding reconciliation. Implementation should assert each diagnostic counter and the stated interval-count equations, including precedence when an interval has more than one defect, before running `go test ./internal/analysis -run 'ZoneEnergy'`.

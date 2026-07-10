# Plan Review — TP-228 Step 2

## Verdict: APPROVE

The revised client boundary is implementable and addresses R004: the resolved `RecalcHRZones` bool is encoded only as `recalcHrZones`, apply is a bodyless/queryless PUT with no date-capable signature, and update no longer invokes apply. The plan also requires retry-safe fresh requests/bodies, preserved 422/cancellation behavior, sparse update bodies, and exact local wire coverage for both query values, no implicit apply, and a zero-byte apply body.

Run `go test ./internal/intervals -run 'SportSettings'` after implementation.

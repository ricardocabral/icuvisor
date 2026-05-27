# Plan Review — Step 2

Verdict: **Changes requested before implementation**

The Step 2 checklist is directionally correct, but it is still too high-level for a cross-cutting unit audit. Before coding, make the plan explicit about which surfaces will get new tests, which surfaces are deliberately audit-only, and the exact work/unknown-unit assertions.

## Required plan additions

1. Add concrete extended-metrics regression cases in `internal/tools/get_extended_metrics_test.go` for every raw-joule field that is normalized to kJ: activity `icu_joules_above_ftp`, activity `ss_w_prime`, interval `wbal_start`, interval `wbal_end`, and interval `joules_above_ftp`. Assert both the divided values and `_meta.extended_metric_units` so raw joules cannot be mislabeled as `*_kj` later.
2. Make the unknown-unit coverage explicit. Existing interval and parser tests already cover some behavior; the plan should say whether Step 2 will extend them or record them as existing coverage. Include `response.ToPreferredWithRaw` pass-through assertions for `KJ`, `KCAL`, and an unknown raw token so energy/calorie units are not converted or guessed by preferred-unit shaping.
3. Separate “not relevant / raw-preserved” surfaces from test surfaces. For workout-library reads, decide explicitly whether `joules` / `joules_above_ftp` remain raw/full-only or need additive metadata if surfaced. For custom items, note that content is preserved verbatim and should not parse or relabel embedded units unless a concrete output surface says otherwise. For histogram, note that current metrics are power/HR/pace only unless you find a joule-bearing histogram path.
4. Include an audit command/check in the plan (for example `grep`/targeted file review of `joules`, `KJ`, `KCAL`, `unknown_unit`) and require logging the result in `STATUS.md` discoveries, especially for surfaces that do not receive new tests.
5. Specify targeted verification commands, at minimum `go test ./internal/units ./internal/response ./internal/tools`, or narrower package commands plus the exact relevant test names.

With those details added, the step will be scoped enough to avoid both gaps (missing raw-joule fields) and accidental behavior changes (inventing semantics for unknown/custom/raw upstream units).

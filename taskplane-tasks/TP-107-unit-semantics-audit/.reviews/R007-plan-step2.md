# Plan Review — Step 2

Verdict: **APPROVE**

The updated Step 2 plan addresses the R006 gaps and is scoped tightly enough to implement.

What looks good:

- It names the concrete raw-joule extended-metrics fields to cover: activity `icu_joules_above_ftp`, activity `ss_w_prime`, interval `wbal_start`, interval `wbal_end`, and interval `joules_above_ftp`.
- It requires both value assertions and `_meta.extended_metric_units`, which is the right regression guard against labeling raw joules as kJ without conversion.
- It separates audit-only surfaces from test surfaces for workout-library reads, custom items, and histograms, reducing the risk of inventing semantics for raw/full-only data.
- It includes preferred-unit pass-through coverage for `KJ`, `KCAL`, and unknown raw tokens.
- It includes an explicit grep/review audit and targeted package verification commands.

Implementation notes:

- When extending `get_extended_metrics` tests, make sure `w_prime_balance_end_kj` is asserted explicitly; existing coverage appears to check start balance and interval joules but not the end-balance field.
- If relying on existing `unknown_unit` coverage in `internal/units` or activity-interval tests, log that as a Step 2 discovery so the audit trail explains why no additional tool test was added there.

Proceed with implementation.

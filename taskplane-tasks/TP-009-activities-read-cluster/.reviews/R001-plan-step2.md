# Plan Review — TP-009 Step 2

Decision: **Approve**

## Summary

The Step 2 plan now addresses the contract-level gaps from the previous review. It defines the terse row schema with unit-disambiguated field names, separates terse `fields` usage from `include_full` raw/null preservation, spells out registry/client dependency changes for activity reads plus profile-unit lookup, and adds a focused Step 2 test plan covering pagination, filtering, Strava-blocked shaping, full/raw behavior, units, and registration.

This is sufficient to proceed to implementation of `get_activities`.

## What looks good

- The public row shape is now clear enough to keep terse responses unit-safe (`distance_km`/`distance_mi`, pace/speed suffixes, `calories_burned`, explicit duration seconds) and to avoid leaking ambiguous upstream names outside `include_full`.
- The `include_full` strategy correctly avoids the upstream `fields` query and requires raw-object preservation, which is necessary because the upstream `fields` behavior drops nulls.
- The bounded pagination plan remains explicit about page size limits, over-fetching, same-timestamp ordering, filtered-row cursor advancement, token contents, and invalid/mismatched token errors.
- The dependency plan avoids hidden globals by introducing an activities client interface while continuing to fetch profile units/timezone through `ProfileClient`.
- The Step 2 targeted tests are appropriate for a step-boundary code review rather than deferring all behavior coverage to Step 6.

## Non-blocking implementation notes

- In code/tests, lock down the upstream unit assumptions for each source field used in conversion (`distance`/`icu_distance`, `average_speed`/`max_speed`, elevation). The plan names the output fields but the implementation should make the source-unit mapping explicit so converted metric/imperial fixtures catch mistakes.
- Keep the public input schema descriptions precise for date arguments (`oldest`/`newest` or whatever names are chosen), accepted date/date-time format, default/max `page_size`, and the opaque nature of `next_page_token`.
- Make sure the terse-mode `fields` allowlist includes every marker needed for Strava/hidden detection even if a fixture row looks sparse; otherwise the unavailable-shaping behavior can regress silently.

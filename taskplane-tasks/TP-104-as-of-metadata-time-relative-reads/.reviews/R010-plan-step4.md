# Plan Review R010 — Step 4

**Verdict:** REVISE

I could not find a concrete Step 4 plan beyond the checklist in `STATUS.md`. For this regression-test/changelog step, please document exactly which gaps will be closed before implementation.

The plan should specify:

1. Which tests will cover positive- and negative-offset timezone boundaries. Existing helper coverage has Kiritimati/São Paulo examples, but the plan should state whether Step 4 will rely on that helper coverage or add a range/tool-level positive-offset regression too.
2. Tool-level include/exclude coverage for all current-day range readers: `get_activities`, `get_events`, and `get_wellness_data`. Right now activities has an explicit past-range omit test, while events and wellness have current-day add tests but need a documented past-only/excluded-range regression (or an explicit rationale for covering that only through the shared predicate).
3. A past-only assertion that checks `as_of`, `as_of_date`, and `as_of_weekday` are absent while existing metadata such as `timezone`, pagination/counts, `include_full`, and date-range fields remains intact.
4. The exact `CHANGELOG.md` entry location, preferably an `[Unreleased]` `### Added` bullet describing the additive `_meta.as_of`, `_meta.as_of_date`, `_meta.as_of_weekday`, and `_meta.timezone` behavior for `get_today` plus current-day `get_activities`, `get_events`, and `get_wellness_data`.
5. The targeted test command(s) to run for this step, e.g. `go test ./internal/response ./internal/tools -run 'TestAsOfMetadataInTimezone|TestCurrentDayAsOfMetadataRangePredicate|TestGetActivities.*AsOf|TestGetEvents.*AsOf|TestGetWellnessData.*AsOf'` (adjust the regex to match the final test names).

Once those specifics are added, the step should be low risk and ready to implement.

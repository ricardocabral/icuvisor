# Plan Review R004 — Step 2

**Verdict:** REVISE

I could not find a concrete Step 2 implementation plan beyond the checklist in `STATUS.md`. This step is small, but it has one important race/consistency edge case that should be pinned down before coding.

Please update the Step 2 plan to state:

1. Compute the anchor once in `getTodayHandler`: call the injectable `now()` once, pass that instant to `response.AsOfMetadataInTimezone`, use `asOf.AsOfDate` as the existing `today` query date, and pass the whole metadata object into response shaping. Do not call `now()` separately for `date` and `_meta.as_of*`, or a midnight boundary could produce inconsistent `date` vs `as_of_date`.
2. Extend `getTodayMeta` with JSON fields `as_of`, `as_of_date`, and `as_of_weekday`, while preserving the existing `date`, `timezone`, `activity_window`, `section_counts`, `include_full`, and response-shaper-added `_meta` keys such as `units`.
3. Ensure `_meta.timezone` continues to report the trimmed/defaulted timezone from the helper, not a raw profile/fallback string that may be blank or whitespace.
4. Add/adjust `get_today` tests using `newGetTodayToolWithClock` to assert exact metadata values for the existing São Paulo boundary case, e.g. `as_of`, `as_of_date`, `as_of_weekday`, and `timezone`, and assert existing fetch dates/counts still match the local date.
5. Run targeted tests for at least `go test ./internal/tools -run TestGetToday` before code review.

Once those details are documented, the implementation should be straightforward and low risk.

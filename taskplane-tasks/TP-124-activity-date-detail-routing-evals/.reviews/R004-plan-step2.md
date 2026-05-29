# Review R004 — Plan Step 2

Verdict: Approved

The Step 2 plan matches the task requirements: it adds one race-by-relative-date scenario that forces `get_activities` before downstream detail/interval analysis, adds a second split/rep-by-date scenario that cannot pass with only session summaries, and keeps validation limited to the cookbook eval harness via `make eval-validate`.

Non-blocking implementation notes:

- Encode ordering in the existing scenario fields, not a new unconsumed schema field. The runner only validates tool names, and the judge receives the ordered transcript plus `prompt`, `must_address`, and `anti_patterns`; use those to state that the assistant must resolve the date/activity with `get_activities` before calling `get_activity_details`, `get_activity_intervals`, or `get_activity_splits`.
- For the race scenario, make `get_activities`, `get_activity_details`, and `get_activity_intervals` expected tools, with anti-patterns for claiming no activity/details exist without the date-window lookup or analyzing an unfetched activity.
- For the split/rep scenario, prefer an explicit downstream expected tool (`get_activity_splits` for lap splits, or `get_activity_intervals` for structured reps) plus `get_activities`; add `get_activity_details` as expected or bonus only if the prompt requires detail fields.
- Keep prompts self-contained and redacted per `scripts/eval/README.md`: use relative dates/descriptive references rather than real athlete IDs or private exact data.
- Update `scripts/eval/README.md` only if adding guidance about order-sensitive scenarios; otherwise the scenario JSON change plus `make eval-validate` is sufficient for this step.

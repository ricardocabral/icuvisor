# Plan Review — Step 3

Verdict: Approve with tightenings.

The Step 3 plan is aligned with the task: it limits this step to the public cookbook, requires the examples to show a preview/approval workflow, and explicitly calls out the important distinction between prose `description` and structured `workout_doc`. It does not propose any model-controlled `confirm` argument or safety-gate bypass.

Tighten the plan during implementation:

1. Add a concrete before/after edit example, not only a new-workout draft. The preview should show current vs proposed duration, key intervals, target intensities, load/distance/time deltas, and preserved title/prose/tags/structured steps before asking for approval.
2. Keep `validate_workout` guidance as optional/read-only preflight for uncertain DSL or structured changes. Do not imply it is a server-side write prerequisite.
3. Make the prose-vs-`workout_doc` distinction visible in the sample prompt and/or “good answer”: prose belongs in `description`, structured intervals belong in `workout_doc`, and both can be merged with the sentinel when needed.
4. For docs validation, run `make web-build` if Hugo is available; if not, record the exact limitation. No generated tool-reference update should be needed for a cookbook-only edit.
5. Ensure Step 5 still updates `CHANGELOG.md`; Step 3 itself may remain focused on `web/content/cookbook/build-workouts.md`.

No blocking issues found in the Step 3 plan.

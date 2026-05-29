# Code Review — Step 3

Verdict: REVISE

## Findings

1. `web/content/cookbook/build-workouts.md:48-64` still does not include the concrete before/after edit example requested for Step 3. The new guidance covers a new-workout preview and briefly describes `update_workout`, but the task/plan asked the cookbook to show an existing-template update preview with current vs proposed duration, key intervals, target intensities, load/distance/time deltas, and preserved fields before approval. As written, users only see a “new workout” example where nothing is preserved, so the most risky update workflow remains underspecified.

## Verification

- Reviewed `git diff a1d4a2c17408973eeb50b48fb9e8fc8c9e089309..HEAD --name-only` and full diff.
- Read `web/content/cookbook/build-workouts.md`.
- Ran `make web-build` successfully; Hugo emitted only existing deprecation warnings.

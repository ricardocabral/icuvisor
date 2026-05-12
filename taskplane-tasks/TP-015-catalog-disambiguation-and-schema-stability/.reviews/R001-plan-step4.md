# Plan Review — Step 4: Implement the confusable-names check

Decision: **APPROVE**

## Summary

The Step 4 plan is aligned with the task requirements. It uses the live registry as the source of truth, compares first description sentences within relevant clusters, chooses a concrete similarity family (token Jaccard), plans actionable CI-style output, and includes the required `CONTRIBUTING.md` threshold documentation.

## What looks good

- Reusing the live-registry collection path keeps the check tied to the actual catalog instead of hand-maintained description constants.
- Comparing only first sentences matches PRD §7.2.E and the Step 1 audit goal.
- Skipping singleton clusters avoids noisy failures where there is no real tool-selection ambiguity.
- GitHub annotations and summary output consistent with the schema checker will make CI failures actionable.
- The planned failure message names both tools and tells maintainers to rewrite first sentences around access pattern/payload, which is the right remediation.

## Implementation notes

- Derive cluster membership from current live tool names/prefixes, not only an exact hard-coded v0.2 list. A small explicit domain-alias map is fine for cross-prefix relationships such as calendar events vs. training plans, but new `get_activity_*`, `get_wellness_*`, `get_workout_*`, or similar tools should be picked up automatically.
- Keep the threshold and tokenizer behavior as named constants in reusable code, then document the same threshold in `CONTRIBUTING.md`. Avoid a magic number only in the script.
- Normalize tokens enough for stable scores: lowercase, split punctuation/snake-case-ish tokens, drop empty/one-character tokens, and consider a small stopword set so generic words like `get`, `one`, `by`, and `id` do not dominate short first sentences.
- The checker output should include the two first sentences and the computed similarity score, in addition to the suggested rewrite hint, so maintainers can understand why the pair failed.
- Reuse one first-sentence extraction helper everywhere possible, including handling dotted tokens like `intervals.icu`, so the CI check and the Step 1 audit semantics stay consistent.

The plan is ready for implementation.

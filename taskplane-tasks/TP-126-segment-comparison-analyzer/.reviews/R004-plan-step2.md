# Plan Review R004 — Step 2

**Verdict:** APPROVE with additions before completion.

The Step 2 plan is aligned with TP-126: add a cookbook eval, update the activity-retrospective guidance, optionally tighten the analyzer activation text, and run `make eval-validate` plus `go test ./internal/tools`.

Before marking Step 2 complete, please make the plan/outcome explicit on these points:

1. **Eval scenario should test the full deterministic workflow.** The new scenario should require locating the activity and obtaining total distance (`get_activities` and/or `get_activity_details`) plus `compute_activity_segment_stats` for first `0..10000m` and last `max(total_distance_m-10000,0)..total_distance_m`. Add `get_activity_streams` to `forbidden_tools` or a similarly explicit anti-pattern so the judge catches chat-side raw-stream reduction.
2. **Cookbook prompt should include the bound translation.** Document that “last 10 km” is not a direct tool argument: first compute total activity distance in meters, then call the analyzer on explicit distance bounds. Also mention that pace comparisons use `velocity_smooth` returned in m/s and should be converted in the final answer if the user asks for pace.
3. **If tool activation text changes, update generated/catalog surfaces.** Tightening `compute_activity_segment_stats` should remove the current “maximum/zone-time” mismatch without bloating the description. Because the first sentence feeds catalog summaries, update the related catalog expectations and generated tool data/golden files as needed (`make docs-tools` / gendocs golden) in addition to `internal/tools/catalog_test.go`.
4. **Record the disposition in `STATUS.md`.** Note whether Step 2 stayed docs/eval-only or changed the tool description, and if the description changed, whether `docs/kr5-benchmark.md` was checked and found unaffected or updated.

No blocker to implementing Step 2 with these additions.

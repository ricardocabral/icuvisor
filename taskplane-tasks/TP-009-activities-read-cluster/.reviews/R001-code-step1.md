# Code Review — TP-009 Step 1

Decision: **Changes requested**

## Findings

1. **Blocking — shaped activity responses still do not specify unit-disambiguated field names**
   - `taskplane-tasks/TP-009-activities-read-cluster/STATUS.md:43` and `STATUS.md:44` define the `get_activities` / `get_activity_details` shaped boundaries as keeping `distance` (and other metric fields) rather than explicit unit-bearing names.
   - PRD §7.2.D requires disambiguating field names such as `distance_km` / `distance_mi` instead of `distance`, and the task anchors this work to response shaping and unit normalization. `_meta.units` is useful context, but it does not replace unambiguous response field names for values that change with `preferred_units`.
   - Please update the Step 1 plan to make the public response contract explicit: e.g. which distance, pace/speed, elevation, and split fields are emitted for metric vs imperial preferences (`distance_km`/`distance_mi`, `pace_min_per_km`/`pace_min_per_mi`, etc.), and how the raw upstream names are preserved only under `include_full`. This prevents Step 2+ from implementing a response shape that conflicts with the PRD.

## Notes

- The previous Markdown-structure issue is fixed: the review log rows now live in the `Execution Log`, and `## Blockers` is clean.
- The Step 1 plan now covers the endpoint inventory, bounded client-side `include_unnamed` pagination, token contents/validation, same-timestamp cursor handling, `include_full` null preservation, and stream-key canonicalization.
- There is still an untracked `taskplane-tasks/TP-009-activities-read-cluster/.reviewer-state.json`; keep it untracked if it is local reviewer/tool state.
- `STATUS.md:20` says `plan Step 1: APPROVE`, while the tracked `R001-plan-step1.md` artifact currently says `Decision: Changes requested`. If those artifacts are meant to be authoritative, consider either updating/replacing the stale plan review file or clarifying the later approval in a separate review artifact.

## Verification

- Ran `git diff 7407d3a82f1d828e34c2919206e58b8a2ee27124..HEAD --name-only`.
- Ran `git diff 7407d3a82f1d828e34c2919206e58b8a2ee27124..HEAD`.
- Read `PROMPT.md`, `STATUS.md`, and the changed review artifacts with line numbers.
- No tests were run; this step only changes task planning/status documentation.

# Plan Review: Step 1 — Audit confusable clusters

## Verdict: approved

The revised Step 1 plan is now executable and aligned with the prompt. It enumerates the current high-risk read-tool clusters, defines a safe first-sentence extraction method that avoids the `intervals.icu` period trap, requires before/after evidence in `STATUS.md`, keeps implementation scope to description first sentences, and defers schema/CI work to later steps.

## Checks against Step 1 requirements

- **Cluster enumeration:** covered for the expected activity, event/calendar, workout-library, custom-item, fitness/performance, and wellness areas.
- **First-sentence audit method:** covered; manual extraction with special handling for `intervals.icu` is appropriate for this step.
- **Evidence:** covered; the planned table has enough fields for reviewers to verify changed and unchanged decisions.
- **Scope control:** covered; no tool renames, schema changes, or CI-helper work are planned in Step 1.
- **Verification:** acceptable; `go test ./internal/tools` is the right targeted test command if descriptions change.

## Non-blocking notes for implementation

1. Consider adding `get_athlete_profile` as an explicit singleton/no-op or adjacent fitness-domain boundary in the audit evidence. It is not obviously confusable today, but recording it would make the catalog audit visibly complete.
2. If any first-sentence rewrite is made, treat it as a user-visible catalog change. Either update `CHANGELOG.md` in this step or add a clear `STATUS.md` note that the task-level changelog entry will be made before TP-015 completes.
3. When filling the evidence table, include unchanged tools with concrete reasons, not only “unchanged.” For example, note whether the sentence distinguishes list vs detail, calendar event vs training-plan assignment, metadata vs streams, or summary vs raw/full payload.

# Plan Review: Step 5 Document amendment

APPROVE

The planned documentation amendment is appropriate for the live-probed contract and matches the optional doc called out in `PROMPT.md`: `docs/upstream-gaps/workout-library-create-payload.md`. The status notes have enough specificity for this step: document that workout-library creates require an existing folder, use JSON `type` for the sport/activity type, and that both omitted and explicit `null` `folder_id` were refused upstream.

Implementation guidance for the worker:

- Keep the doc narrowly scoped to the upstream create-payload contract; do not re-open code changes already completed in Steps 2-4.
- Include the accepted minimal shape and rejected permutations from Step 1:
  - accepted: `type: "Ride"` with an existing `folder_id`, optionally with `description`;
  - rejected: `sport` without `type` (`Missing type`);
  - rejected: omitted `folder_id` and `folder_id: null` (`Folder is required`).
- Be explicit that the folder must already exist and be owned by/visible to the target athlete, and that top-level creates were not accepted in the live probe.
- Avoid committing raw athlete-, folder-, or workout-identifying values; use placeholders such as `<existing-folder-id>` or sanitized fixture references.
- Note the scope as create-only unless update semantics were separately probed; do not imply the same constraints apply to every workout-library endpoint beyond the evidence gathered.
- Follow the concise style of `docs/upstream-gaps/event-note-payload.md` unless there is a strong reason for a longer evidence table.

No blocking changes to the plan are needed.

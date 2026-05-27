# Plan Review: Step 1 — Update write-tool wording

**Verdict:** Changes requested before implementation.

The file scope and general sequencing are right, but this step is explicitly a wording-pattern checkpoint and `STATUS.md` does not yet record the final wording pattern to apply. Please capture the exact phrasing (or near-final strings) before editing all tools/snapshots.

## Blocking feedback

1. **Document the final wording pattern per tool.** The implementation must make clear that `description` writes replace the upstream description/DSL field; they are not append-only notes. For updates, also state that omitted fields stay unchanged.
2. **Do not reference TP-109 runtime behavior.** TP-109 is not implemented. Avoid wording that says icuvisor will warn/guard/block description-only workout updates. This task should be wording-only.
3. **Keep the risk specific.** Do not imply every description-only write is destructive. The risk is specifically that, for workout-shaped events/templates, replacing the upstream description/DSL without a `workout_doc` can remove existing structured steps.
4. **Do not reintroduce mutual exclusion.** Preserve the contract that `description` and `workout_doc` may be supplied together, using the merge sentinel to choose placement.
5. **Snapshot verification should include the generator/check owner.** If input-schema descriptions change, regenerate with `go run ./scripts/snapshot_tool_schemas.go`. `go test ./internal/tools` is useful, but it will not by itself prove snapshot freshness; add `go test ./internal/toolchecks` or document why it is deferred.

## Suggested wording pattern

Use concise variants of these patterns:

- `add_or_update_event`: `description` is an optional replacement for the upstream event description/DSL. Omit it to leave unchanged on updates. For WORKOUT updates, supplying `description` without `workout_doc` can replace existing structured steps; include the desired `workout_doc` to preserve/merge structure, and use `<!-- icuvisor:steps -->` to position serialized steps when also supplying prose.
- `create_workout`: creation builds the initial upstream template description/DSL from `description`, `workout_doc`, or both. There is no existing template to preserve, but `description` is still part of the single upstream description/DSL field, not a separate append-only notes channel.
- `update_workout`: sparse fields are omitted to leave unchanged, but a supplied `description` and/or `workout_doc` replaces the template's upstream description/DSL. To preserve structured steps while changing prose, provide the desired `workout_doc` plus prose/sentinel.
- `update_activity`: keep this as a replacement free-text activity description only. Do not imply activity descriptions carry planned workout structure; the current `set_activity_intervals` pointer is appropriate.

Once the final wording is recorded, the Step 1 plan is otherwise acceptable.

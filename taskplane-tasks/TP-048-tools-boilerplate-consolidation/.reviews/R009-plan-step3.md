# R009 Plan Review — Step 3: `get_activities.go` cleanups

Verdict: **APPROVE**

The revised Step 3 plan addresses the blockers from R008:

- It no longer assumes `internal/tools/get_activities.go`'s `stringSet` is dead before implementation. Current grep evidence still shows package-local callers in `get_activities.go` and `get_activity_streams.go`, so replacing those callers before deleting the helper is the right sequence.
- It scopes the `stringSet` acceptance evidence to `internal/tools`; the unrelated helper in `internal/toolchecks/schema_stability.go` is out of scope for this step.
- The named-struct cleanup for `validateActivitiesTokenArgs` remains narrowly scoped and behavior-preserving.

Implementation guardrails:

1. Keep the `stringSet` replacement minimal and local. For example, build the small map inline where membership is checked, or use direct explicit membership logic where that is clearer. Do not introduce a new generic/helper abstraction for this task.
2. Update all current `internal/tools` callers before deleting the helper; otherwise `get_activity_streams.go` will fail to compile because it currently relies on the package-level helper from `get_activities.go`.
3. Preserve the exact JSON tags and pointer-field semantics when naming the supplied-arguments struct for `validateActivitiesTokenArgs`; this struct is used to distinguish omitted fields from zero values in pagination-token validation.
4. After implementation, run at least `go test ./internal/tools` and confirm `grep -rn "stringSet" internal/tools/` returns no hits. The broader `internal/` grep may still find the unrelated schema-stability helper and should not be treated as a Step 3 failure.

No further plan changes are required before implementing Step 3.

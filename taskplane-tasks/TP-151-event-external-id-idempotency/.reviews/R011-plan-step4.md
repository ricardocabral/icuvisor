# Plan Review: TP-151 Step 4 — Refresh schemas, routing, and docs

**Verdict: REVISE**

The Step 4 plan covers the right buckets, and the regenerated `add_or_update_event` schema plus the routing fixture look directionally correct. However, the remaining documentation work is currently too vague for a public write-contract change. This task intentionally adds a new user/LLM-facing idempotency field and changes `apply_training_plan` retry behavior, so “update user docs if affected” should be made explicit before finishing the step.

## Required plan updates

1. **Name the affected user docs and what each should say.** At minimum, plan edits for:
   - `web/content/cookbook/build-workouts.md`: explain when to use optional `external_id` for retry-safe/manual calendar writes, that it is stored upstream and should not contain secrets, blank values are ignored/not a clear operation, and callers should use their own stable namespace rather than provider-owned prefixes such as `strava-` or `hevy-`.
   - `web/content/cookbook/season-and-block-plan.md`: note that `apply_training_plan` now proposes/writes deterministic `icuvisor-plan-v1-...` external IDs for retry review, and repeated applies with the same plan/start/date tuple are safer but still subject to the documented same-day/upstream caveats.
2. **Make the CHANGELOG target explicit.** Add the entry under `[Unreleased]`, not the already-cut `0.1.8` section, and mention both public pieces: `add_or_update_event.external_id` support and deterministic `apply_training_plan` external IDs/protected matching-external-id retries.
3. **Document caveats consistently with Step 1.** The docs/changelog should not overpromise global dedupe. Preserve the conservative contract: best-effort upstream idempotency, same-day preflight, no clear/null support, and external IDs are visible audit metadata in event reads.
4. **Record schema snapshot expectations.** It is fine that only `add_or_update_event.json` changes after running `go run ./scripts/snapshot_tool_schemas.go`; `apply_training_plan.json` is input-schema-only and need not change unless its input schema changes. Add that as an execution note so reviewers do not chase a nonexistent snapshot diff.

The added routing case for explicit external-id event writes is sufficient; I do not see a need for another routing fixture unless the docs add a new ambiguous prompt pattern.

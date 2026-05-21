# R003 Plan Review — Step 3: Verify byte-identical behaviour

Verdict: **APPROVE, with required execution notes**

The Step 3 plan covers the required gates for this mechanical refactor: focused tools tests, full build/test/race/lint, catalog diff, and schema snapshot diff. That is the right verification scope for “no behaviour change.”

Before executing it, make the byte-identity checks precise so they cannot produce false confidence:

1. **Generate schema snapshots into a separate post-refactor directory.**
   - Use the captured baseline at `taskplane-tasks/TP-063-split-get-fitness-one-tool-per-file/pre-refactor/schema_snapshot`.
   - Run `go run ./scripts/snapshot_tool_schemas.go -dir <post-refactor-temp-or-task-dir>` and compare that directory to the baseline, e.g. `diff -ru .../pre-refactor/schema_snapshot <post-dir>`.
   - Do not rely on the script default output directory for this comparison unless you intentionally want to mutate `internal/tools/schema_snapshot` and then verify/clean the worktree afterward.

2. **Catalog diff must use the same generator/environment as the pre-refactor capture.**
   - Compare canonical JSON against `taskplane-tasks/TP-063-split-get-fitness-one-tool-per-file/pre-refactor/tool_catalog.json`.
   - Keep the same toolset/delete-mode/version/timezone assumptions used for the baseline. A raw `icuvisor_list_advanced_capabilities` call is only acceptable if it produces the same catalog shape that was captured pre-refactor; otherwise use the equivalent catalog dump mechanism from Step 2.

3. **Add a formatting/worktree sanity check.**
   - In addition to the planned `make build`, `make test`, `make test-race`, and `make lint`, run `make fmt-check` or `git diff --check`/`gofmt -l` so import splitting does not leave formatting drift.
   - Finish with `git status --short` so generated snapshots or catalog files were not accidentally left modified outside the intended task files.

4. **Treat any diff as a blocker.**
   - If the catalog or schema diff is non-empty, fix the refactor; do not update the pre-refactor baseline for this task.
   - Record the exact commands and pass/fail outcomes in `STATUS.md` so the wrap-up review can verify what was run.

With those notes folded into execution, the Step 3 plan is sufficient to proceed.

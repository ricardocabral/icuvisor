# Plan Review R013 — Step 5: Testing & Verification

**Verdict:** Changes requested

The Step 5 checklist names the required verification gates, but the plan is not yet concrete enough to close this task cleanly. Step 5 is the final test gate and must also address the outstanding audit-trail issue from R012: verification commands need to be recorded in `STATUS.md`, not only checked off.

## Blocking concerns

1. **Targeted tests are not pinned.**  
   “Targeted tests passing” should name the exact affected packages/commands. For this task, at minimum use the targeted command already established in Step 3: `go test ./internal/resources ./internal/mcp`. If any fixes touch other packages during Step 5, add those packages explicitly before the full suite.

2. **Command-result recording is missing from the plan.**  
   R012 found that Step 4 marked full-suite/build/lint complete without recording command outcomes in `STATUS.md`. Step 5 should explicitly require adding timestamped command/result entries for:
   - `go test ./internal/resources ./internal/mcp`
   - `make test`
   - `make build`
   - `make lint`

   If a command fails, record the failure summary and whether it was fixed or pre-existing/unrelated before checking the item complete.

3. **Do not rely silently on Step 4 results.**  
   Step 5 is a separate verification step. Either rerun the commands now, or explicitly document that no relevant files changed since the passing Step 4 run and reference the recorded Step 4 command results. Given R012 says those results were not recorded, rerunning and logging them is the safer path.

## Suggested revised Step 5 checklist

- [ ] Run `git status --short` and confirm the only pending changes are expected task/status/review updates.
- [ ] Run targeted tests: `go test ./internal/resources ./internal/mcp`; record pass/fail in `STATUS.md`.
- [ ] Run full suite: `make test`; record pass/fail in `STATUS.md`.
- [ ] Run build: `make build`; record pass/fail in `STATUS.md`.
- [ ] Run lint: `make lint`; record pass/fail in `STATUS.md`.
- [ ] Fix any failures, or document clearly in `STATUS.md` why they are pre-existing and unrelated.
- [ ] Confirm R012’s missing verification-log finding is resolved before moving to Step 6.

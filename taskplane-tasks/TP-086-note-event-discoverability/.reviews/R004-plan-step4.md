# Review R004 — Step 4 plan

Decision: **Approved with required clarifications.**

The Step 4 delivery checklist is pointed at the right handoff work after Step 2/3 changed both public docs and assistant-facing tool metadata. Execution should make the optional-looking items concrete so the delivery step does not under-verify the changes before moving into the broader Step 5 gate.

## Required clarifications for execution

1. **Treat the changelog update as mandatory.**
   This task changed public docs and `add_or_update_event` metadata/examples, so add a concise `[Unreleased]` entry in `CHANGELOG.md` rather than leaving the checkbox conditional. A good placement is under `### Changed`, e.g. noting that `add_or_update_event` docs/examples now make `category: "NOTE"` discoverable for nutrition plans, travel logistics, daily reminders, and coach annotations.

2. **Run `make test` in Step 4 because code metadata changed.**
   The prompt's Step 4 says to run `make test` if code changed, and this branch changed `internal/tools/add_or_update_event.go` plus the schema snapshot. Step 3's targeted tests are useful but do not satisfy this Step 4 checkbox by themselves. Record the exact command and result in `STATUS.md`; if it fails for an unrelated/pre-existing reason, capture the failing package and summary there.

3. **Do not lose the docs-build evidence.**
   Public Hugo content changed, and Step 3 required `make web-build`. If Step 3 already ran it successfully, Step 4 can reference that result in `STATUS.md`; if not, run it now or explicitly document the attempted command and environment limitation. Avoid marking delivery complete with only implicit verification.

4. **Record the telemetry follow-up narrowly.**
   Add a `STATUS.md` discovery/note along the lines of: after improving examples, only reconsider a separate `add_note` tool if future telemetry/support reports show assistants still fail to select `add_or_update_event` with `category: "NOTE"`. This should remain a future question, not an implementation change or new tool in this task.

5. **Keep the Step 4 commit focused and traceable.**
   The Step 4 commit should include `TP-086` and ideally only delivery bookkeeping such as `CHANGELOG.md`, `STATUS.md`, and this review file. Do not add a separate `add_note` tool, broaden event-write behavior, or hand-edit generated artifacts without rerunning the relevant generator/check.

With those clarifications, the Step 4 plan is sufficient to prepare the task for the full Step 5 verification gate.

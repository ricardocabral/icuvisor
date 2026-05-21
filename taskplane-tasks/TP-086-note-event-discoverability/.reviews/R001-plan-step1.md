# Review R001 — Step 1 plan

Decision: **Approved with minor clarifications.**

The Step 1 audit scope is appropriate for this small discoverability task: it asks the worker to search existing NOTE coverage, determine which docs are generated versus static, and decide whether `add_or_update_event` needs another `input_examples` entry. That is the right information to gather before editing docs or tool metadata.

## Clarifications to carry into execution

1. **Record the exact generation boundary.**
   The generated website reference is currently `web/content/reference/tools.md` plus the `{{< tool-catalog >}}` shortcode fed by `web/data/tools.json`; `make docs-tools` runs `cmd/gendocs`, which currently emits only `ToolDescriptor` fields derived from tool names/descriptions, not input schemas or `input_examples`. Step 1 should explicitly record this so Step 2 does not assume adding schema examples will automatically update the public website reference.

2. **Audit both assistant-facing and user-facing surfaces.**
   Include at least these surfaces in the NOTE search results:
   - `internal/tools/add_or_update_event.go` (`description`, input schema field descriptions, `addOrUpdateEventInputExamples`);
   - `internal/tools/*_test.go` schema/example invariants that may need updates if examples are added;
   - `web/content/reference/tools.md`, `web/data/tools.json`, and the Hugo tool-catalog partials;
   - `README.md`, `web/content/guides/*`, and `web/content/tutorials/*`;
   - existing context docs such as `docs/upstream-gaps/event-note-payload.md` for factual constraints, without broadening this task into PRD/roadmap edits.

3. **Classify the current coverage by required use case.**
   The code already has a NOTE input example for a travel day. Step 1 should still mark which of the required completion-criteria use cases are missing or only implicit: nutrition plans, travel logistics, daily reminders, and coach annotations. That will make Step 2 additive and focused.

4. **Keep Step 1 non-mutating except for status notes.**
   This step should end with `STATUS.md` discoveries/checkboxes updated. Code/docs edits, regeneration, changelog changes, and test runs belong to later steps unless the worker discovers the plan must be amended first.

With those details captured in `STATUS.md`, the audit should produce enough direction for Step 2 without risking unnecessary refactors or a confusable `add_note` tool.

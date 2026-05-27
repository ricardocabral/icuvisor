# Plan Review: Step 3 — Wire command and documentation

**Verdict: Needs revision**

The Step 3 checklist identifies the right outcomes, but it is not yet concrete enough to review as an implementation plan. This step is mostly documentation/wiring, so the plan should name the exact command, doc locations, and operator-facing semantics before edits are made.

## Required clarifications

1. **Choose the command/target and help text**
   - Specify the exact Make target to add, preferably a clear name such as `eval-tool-routing` or the prompt's example `eval-tools`.
   - Add it to `.PHONY` and ensure `make help` shows that it is an opt-in routing smoke eval, distinct from the existing `eval-validate` cookbook target.

2. **Document environment and exit behavior exactly**
   - List required live-run variables: `ICUVISOR_ROUTING_EVAL_PROVIDER=anthropic` and `ANTHROPIC_API_KEY`.
   - List optional variables: `ICUVISOR_ROUTING_EVAL_MODEL` and `ICUVISOR_ROUTING_EVAL_ANTHROPIC_URL`.
   - Explain unset provider behavior: fixture/catalog validation only, provider cases skipped, no provider network call, zero exit.
   - Explain non-zero cases: configured provider errors and routing mismatches.

3. **State the safety guarantees**
   - Documentation must explicitly say the eval sends tool schemas/prompts to the configured model provider but does not execute icuvisor MCP tool handlers.
   - It should also say it does not call intervals.icu; live mode only calls the model provider.
   - Use placeholder API keys only and do not instruct users to paste keys into committed files or STATUS output.

4. **Pick documentation locations**
   - The task requires `CONTRIBUTING.md` or a dedicated docs page. The plan should state which one will be updated.
   - If a new `scripts/toolroutingeval/README.md` is added, also add a short pointer from `CONTRIBUTING.md` or another discoverable developer doc.

5. **Changelog disposition**
   - Since this is developer-visible tooling, plan to add an `[Unreleased]` entry (likely under `### Added`) unless there is a specific reason not to.

Once those details are added to the Step 3 plan, the implementation should be straightforward and low-risk.

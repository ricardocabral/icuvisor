# Plan Review — TP-078 Step 1

**Verdict:** REVISE

## Summary

The revised Step 1 plan is much stronger than R001: it now defines an audit matrix, separates runtime fallback paths from onboarding write paths, includes the key setup/config/credstore/diagnostics code surfaces, and sets an appropriate no-code-change boundary.

However, two gaps remain before this is safe to execute as the source-of-truth audit for a credential-handling task.

## Required changes

1. **Add concrete search commands, not only terms.**
   `STATUS.md:36` lists search terms, but R001 asked for concrete grep/read commands to make the audit reproducible and reduce missed surfaces. Add command-level instructions, for example a repo-wide grep for the listed credential terms plus narrower reads of the code/doc scopes. The exact tool syntax can vary, but the plan should be explicit enough that a second worker would run the same audit.

2. **Expand the docs/assets scope to include non-web client docs and release packaging metadata.**
   `STATUS.md:35` covers the web docs and generic packaging assets, but this repository also has credential-related client documentation outside `web/` (for example `docs/clients/codex-local.md` mentions API-key environment handling), and packaging metadata such as `.goreleaser.yaml` contains install/setup text. Add explicit scope entries for at least:
   - `docs/clients/*`
   - `.goreleaser.yaml` and any other packaging/release installer text found by `find`

## Why this matters

Step 1 is supposed to become the source of truth for Steps 2–4. If the audit only searches the named web docs or relies on informal search terms, it can miss docs-only onboarding paths that still instruct users to place credentials in environment/config contexts, or release/installer copy that shapes the first-run UX.

## What looks good

- The matrix columns are sufficient for distinguishing entrypoint, current behavior, desired credential source, and follow-up action.
- The code scope includes the important storage, loading, redaction, and diagnostics paths.
- The plan correctly classifies process env and legacy JSON/.env support as runtime/power-user fallback unless a later step intentionally changes compatibility.
- The Step 1 no-code-change boundary is correct.

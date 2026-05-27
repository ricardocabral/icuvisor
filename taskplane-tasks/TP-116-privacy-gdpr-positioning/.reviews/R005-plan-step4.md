# Plan Review — Step 4: Testing & Verification

**Verdict: approved to proceed.**

The Step 4 plan matches the scope of the committed changes: documentation/site content plus changelog only. `make web-build` and rendered-link/page-placement checks are the right required verification. Because no Go source, generated tool data, app strings, or runtime files were changed in Steps 2–3, `make test` and `make build` can be skipped as conditional checks, with that rationale recorded in `STATUS.md`.

## Verification guardrails

- Run `make web-build` from the repository root and capture the result in `STATUS.md`.
- Check the rendered site output for the new `/explain/privacy/` page, the explain-index card, and the cross-links from coach mode and HTTP transport. A generated-HTML grep is sufficient if no browser preview is used.
- Include a quick copy sanity check for prohibited overclaims: no “GDPR compliant,” certification, or legal-advice language beyond the explicit non-claim framing.
- If Hugo reports broken `relref`s, warnings, or missing pages, fix them rather than documenting them as unrelated.
- Do not broaden verification into runtime changes; keep any `make test`/`make build` skip tied to the fact that this task only touched docs/changelog/task metadata.

## Notes

If `make web-build` fails because Hugo or the website toolchain is unavailable in the worker environment, document the exact command and failure in `STATUS.md`; otherwise Step 4 should not proceed with unresolved site-build failures.

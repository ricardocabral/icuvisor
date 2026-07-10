# Plan Review — Step 1

## Verdict: REVISE

The revised checklist acknowledges `web/content/guides/coach-mode.md` and `CONTRIBUTING.md`, which addresses R001 in intent, but the Step 1 artifact/file-scope list still names only the two original web pages and the new script. Add both guidance files explicitly to the planned artifacts and require the content contract to cover their corrected wording (or explicitly exclude `CONTRIBUTING.md` with a reason). Otherwise the stated no-false-public-doc claim is not a verifiable plan deliverable.

Also wire `scripts/tests/test_docs_guidance.py` into a CI-executed command. `make test` and the current CI workflow run only Go tests, so a standalone script invoked manually during Step 1/2 will not prevent a later regression. Extend the scoped plan to add a dedicated Make target and invoke it from CI, or add an equivalent CI step. The contract should verify the exact normalization rules for both affected ID-guidance pages (trim whitespace, lowercase only a leading `I`, validate digits, and never add/remove the prefix), the hosted connector route/URL in HTTP troubleshooting, and an explicit prohibition on generic public tunnels.

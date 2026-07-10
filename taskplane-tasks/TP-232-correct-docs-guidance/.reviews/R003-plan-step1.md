# Plan Review — Step 1

## Verdict: REVISE

The expanded artifact list now includes `web/content/guides/coach-mode.md`, `CONTRIBUTING.md`, `Makefile`, and CI wiring, but the content-contract scope is still ambiguous: it says it covers “both affected public guides,” which identifies the two website pages and does not explicitly cover the obsolete normalization statement in `CONTRIBUTING.md`. Step 1 separately plans to correct that statement, so the regression contract must either assert its corrected wording too, or explicitly exclude it with a reason.

Make the contract requirement path-specific: verify the exact prefix-preserving normalization guidance in `web/content/reference/config-file.md`, `web/content/guides/coach-mode.md`, and `CONTRIBUTING.md`; reject the old `normalize(s) ... to i12345` claim; and retain the existing explicit hosted URL/link plus generic-public-tunnel prohibition checks for `web/content/guides/http-transport.md`. This makes the stated “no public doc” criterion mechanically verifiable.

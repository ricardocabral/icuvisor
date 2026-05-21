# Plan Review — Step 2: Docs

Decision: **approved**

The revised Step 2 plan now addresses the prior review items and is specific enough to proceed. It names all seven required `docs/internal-beta/` artifacts, keeps the work artifact-only, calls out the consent/privacy boundaries, references the existing client/install/coach-mode docs instead of duplicating JSON snippets, and ties measurement back to KR1 plus PRD §7.4 #6/#8/#12.

Minor execution notes for the implementation pass:

- Make `README.md` link every other internal-beta doc in execution order, including the empty `findings.md` template, so the acceptance criterion is easy to verify.
- In `onboarding-playbook.md`, keep the Claude Desktop / Claude Code config content as links to `docs/clients/claude-desktop.md` and `docs/clients/claude-code.md`; do not paste JSON blocks into the beta docs.
- Keep `findings.md` strictly empty of examples or sample athlete rows.
- When mentioning tool-call collection, use only names/timestamps/counts/descriptions and keep arguments, payloads, athlete IDs, screenshots with values, transcripts, and raw training data excluded consistently across protocol, measurement, and checklist.
- Leave full verification (`make test`, `make lint`, `make build`) to Step 3 as planned, but do the proposed path/link sanity pass after drafting.

Proceed with Step 2.

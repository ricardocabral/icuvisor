# Plan Review — TP-234 Step 2

## Verdict: REVISE

No implementation plan for Step 2 is present. `STATUS.md` contains only the unchecked outcome checklist; the detailed design recorded for Step 1 is not a file-by-file plan for writing and integrating the guide. There is no proposed guide outline, edit map, or verification sequence to review.

Revise the plan to state:

1. **The new guide's concrete structure and samples.** Map the approved launchd, systemd-user, and current-user Task Scheduler designs into named sections and fenced commands, including same-account interactive `icuvisor setup`, absolute binary-path discovery/validation, literal `--transport http --http-bind 127.0.0.1:8765`, the client-only `/mcp` URL, non-secret logs, lifecycle/recovery/removal commands, and the logged-in credential-store limitation. Explicitly keep API-key values, environment credential directives, `.env`/`--env-file`, and plaintext `api_key` examples out of every service recipe.

2. **Each integration edit.** Specify where `http-transport.md` will link to persistence and replace both implicit foreground HTTP starts and the executable LAN-bind block with loopback-only examples while retaining the unauthenticated-LAN warning as prose. Specify the new guide card in `_index.md` and the exact troubleshooting symptom/link to add for a stopped or failed persistent process. Preserve the hosted HTTPS/OAuth URL and no-public-tunnel guidance for provider-hosted clients, and use `icuvisor` for any connector key/name.

3. **Completion checks and remaining documentation.** Include `make web-build` after the edits and a link/render check for all new `relref`s. State when the required `[Unreleased]` `CHANGELOG.md` entry will be added (in this step or the delivery step), so the mandatory documentation update is not lost.

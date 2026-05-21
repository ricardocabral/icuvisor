# R015 Code Review — Step 5: Guides section

Verdict: APPROVE

No blocking findings.

What I checked:

- Reviewed the changed-file list and full diff from `22ffc03c076a44b4891b824b86dccf9c07faa8de..HEAD`.
- Read the new guide pages under `web/content/guides/` and compared the migrated coach-mode and post-upgrade material against `docs/coach-mode.md` and `docs/post-update.md`.
- Cross-checked the documented env vars, config fields, keychain service/account, HTTP bind defaults, coach-mode values, and ACL rules against `internal/config/`, `internal/credstore/`, `internal/coach/`, and `internal/toolcatalog/`.
- Verified the HTTP transport guide keeps the default loopback bind and includes the required unauthenticated-LAN warning.
- Verified troubleshooting covers the Step 5 required symptoms: Gatekeeper rejection, missing API key, stale schema after upgrade, LAN connection refusal, and Linux keychain/libsecret problems.
- Ran `cd web && hugo --minify --gc`; the site builds successfully with no broken relrefs.

Non-blocking follow-up for the later drift/link sweep:

- As in the connect pages, some env vars appear inside code blocks, tables, or inline examples without being individually linked. The guide pages do link to the CLI/config/safety references where the variables are explained, but Step 7 may want to add one explicit “see CLI reference for environment variables” sentence to `coach-mode.md` and `troubleshooting.md` if the task owner wants a stricter interpretation of the cross-reference guidance.

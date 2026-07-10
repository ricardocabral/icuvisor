# Plan Review — TP-234 Step 3

## Verdict: REVISE

`STATUS.md` provides only Step 3 outcome checkboxes; it does not contain an implementation plan for the new contract test or its Make/CI integration. The earlier Step 1 design mentions intended coverage, but Step 3 needs a concrete file-and-assertion plan before it can be reviewed.

Revise the plan to specify:

1. **Test structure and scope.** Add `scripts/tests/test_http_service_docs.py`, using repository-root-relative paths and failure output consistent with `test_docs_guidance.py`. Parse fenced executable snippets (including shell, PowerShell, and JSON/service-definition content) separately from prose, and check both `persistent-http-service.md` and the edited HTTP transport guide so the foreground/config examples cannot regress.

2. **Positive contract assertions.** Name the exact required guide headings/content for LaunchAgent, systemd user service, and Task Scheduler; the literal server bind and client `/mcp` endpoint; same-account credential-store/setup guidance; per-platform start/status/log/restart/stop/remove lifecycle commands; and the hosted HTTPS/OAuth fallback plus no-public-tunnel guidance. The planned assertions should be precise enough to distinguish executable server binds from the client-only URL and prose warnings.

3. **Negative executable-snippet assertions.** Define checks that reject an API-key assignment or credential environment/service directive (including `INTERVALS_ICU_API_KEY`, dotenv/environment-file sources, and plaintext `api_key`) only in executable samples, while rejecting wildcard (`0.0.0.0`, `::`) or non-loopback HTTP-bind values in every executable sample. This avoids both an incomplete security guard and false failures on explanatory prose.

4. **Integration and verification.** State the `Makefile` edit that runs the new Python test from `docs-guidance-test` alongside the existing guidance contract. Since Ubuntu CI already runs that target, note that no workflow change is needed (or identify one if the plan chooses another invocation), then run both `python3 scripts/tests/test_http_service_docs.py` and `make docs-guidance-test`.

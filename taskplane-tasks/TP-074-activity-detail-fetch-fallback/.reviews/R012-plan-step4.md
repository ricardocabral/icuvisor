# Plan Review — Step 4: Fixtures

Decision: **approve plan**.

The Step 4 plan is appropriate for the current implementation. The original step explicitly allows inline `httptest.Server` setup when the only thing being asserted is HTTP status/sentinel behavior, and the completed Step 1–3 work already exercises the tool-level unavailable shape with fake clients plus inline Strava payloads. Adding JSON files for empty 400/403/404/429/500 responses would likely create unused or low-value fixtures.

## Guardrails for Step 4

- Do not add placeholder JSON fixtures unless a test actually reads and asserts their body shape.
- If you find a missing client-level status-mapping check, prefer a small inline `httptest.Server` or existing client test table over `internal/intervals/testdata/*.json` for empty error bodies.
- Keep the existing inline Strava-blocked payloads close to the tool tests; they are documenting detection behavior, not reusable API fixtures.
- Before marking the step complete, update `STATUS.md` with the explicit decision/rationale: no new fixture files were needed because the error-body content is irrelevant and status/sentinel behavior is covered inline/fake.
- Run at least the targeted tool tests after this cleanup to catch accidental fixture references or dead testdata assumptions.

With those constraints, the plan satisfies the Step 4 intent while avoiding fixture sprawl.

# Plan Review: Step 2 failing test

Result: approved.

The amended Step 2 plan addresses the concerns from R003. It now separates the live-accepted seven-field upstream payload from the actual red test for the defect (`feel` is rejected by intervals.icu), and it includes schema/public-contract coverage so the tool does not keep advertising a field that will be refused upstream.

## What looks good

- The fixture test is scoped correctly: it should document the probed-good `PUT /athlete/{id}/wellness/{date}` shape for `fatigue`, `soreness`, `stress`, `mood`, `motivation`, `sleepQuality`, and `locked`, using the sanitized request/response fixtures.
- The failing behavior is now explicit rejection of submitted `feel`, not silently dropping it from a larger bundle. That avoids false success and protects `_meta.fields_updated` from claiming a write that never happened.
- Requiring zero upstream I/O for `feel` input is the right safety boundary for both the tool handler and the intervals client if `WriteWellnessParams.Feel` remains present during the transition.
- Adding schema/public-contract assertions is important and should make the final fix impossible to leave half-done.

## Implementation notes for the step

- When adding the red tests, also account for existing positive tests that currently use `feel` as a valid write field (`FeelOnlyDoesNotZeroWeight`, response-shaping tests, sparse-body client test, and mixed weight examples). Those do not all need to be rewritten in Step 2, but they will need to be reconciled by the fix or they will block the suite after the new contract lands.
- Keep the public error stable and actionable, ideally matching the existing read-only-field style, e.g. `field_not_writable: feel (not accepted by intervals.icu wellness write)`. The test should assert `PublicErrorMessage`, not an internal wrapped error string.
- The accepted fixture test may pass before the fix; that is fine. The red signal should come from unsupported-`feel` runtime/schema tests.
- Because rejecting `feel` diverges from the current PRD text, make sure the later documentation step records this upstream gap and any public-contract amendment. This does not block Step 2, but it should not be lost.

Proceed with Step 2 as planned.

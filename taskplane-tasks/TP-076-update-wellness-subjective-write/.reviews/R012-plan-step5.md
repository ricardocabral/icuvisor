# Plan Review: Step 5 document amendment

Result: approve.

The Step 5 plan is sufficient. The live probe found an upstream write-shape inconsistency that is worth preserving outside the task status: `feel` is readable on wellness rows but is not accepted by the wellness write endpoint, while the remaining subjective fields tested (`fatigue`, `soreness`, `stress`, `mood`, `motivation`, `sleepQuality`, and `locked`) use the existing camel-case/key names and are accepted. Capturing that in `docs/upstream-gaps/wellness-write-payload.md` matches the prompt's document-amendment requirement and the existing `docs/upstream-gaps/*` pattern.

Execution notes for the document:

- Keep it sanitized: do not include raw athlete IDs, API keys, exact probe dates, raw wellness values beyond synthetic examples, or any other date-identifying live-account data.
- State the probed endpoint shape at a high level (`PUT /api/v1/athlete/{id}/wellness/{YYYY-MM-DD}` with a sparse JSON body) without exposing the real athlete/date.
- Make the read/write distinction explicit: read-side `feel` remains supported, but write-side `feel` should be treated as unsupported and rejected locally before upstream I/O.
- List the accepted subjective write keys, including that `sleepQuality` was accepted as camelCase.
- Document the cleanup limitation carefully: live probing showed the public API did not provide a working way to clear subjective fields or unlock the probe row (`locked:false`/null clears/DELETE variants failed or were ignored), so future probes should avoid setting `locked:true` on fresh rows unless operator cleanup is available.
- Link the sanitized fixtures already captured under `internal/intervals/testdata/wellness/subjective_write_request.json` and `subjective_write_response.json` rather than duplicating raw probe transcripts.

With those details included, Step 5 should give future agents enough context to avoid repeating the live probe and to avoid creating another locked-row cleanup problem.

# TP-260 batch coding lens review

**Scope:** Integrated batch `20260818T192348`, consisting of TP-252, TP-254, TP-257, and TP-259 on `main` at `098fc39`.

**Lenses:**

- **Hickey / simplicity:** separate concepts, data over accidental process, small APIs, and low semantic coupling.
- **Torvalds / directness:** straightforward control flow, honest data structures, explicit failure behavior, and debuggability.
- **Carmack / pragmatism:** measurable value, bounded cost, reliability, security, and avoidance of speculative abstractions.

This is an evidence-based review, not a style ranking. TP-253, TP-255, TP-256, and TP-258 were not executed in this batch and are excluded from implemented-code findings.

## Executive summary

- **Must fix before release:** none found.
- **Should fix:** **F-254-01 (medium, high confidence)** — activity `feel` and native `icu_rpe` are typed as integers and described as 1–5 and 1–10, but the shared normalizer accepts any integral value, including out-of-range values. This is a small, bounded follow-up with a direct correctness payoff.
- **Observation:** **O-252-01** — windowed stream responses are bounded locally, but each request still fetches the complete upstream stream because no verified upstream window query exists. The cost is honestly documented and was not treated as a release defect.
- TP-252, TP-257, and TP-259 have no additional release-blocking finding. The integrated changes preserve raw/terse boundaries, provenance, unavailable-data diagnostics, and no-inference constraints in the reviewed paths.

## Prioritized findings

### F-254-01 — Integral subjective values are not range-validated

- **Lens:** Torvalds / directness; Carmack / pragmatism.
- **Severity:** **Medium — should fix.**
- **Confidence:** High.
- **Evidence:** `internal/intervals/activities.go:151-160` (`rawActivityScaleInt`) accepts any JSON number that is mathematically integral and returns it as an `*int`; it does not know whether the field is `feel` or `icu_rpe`, and it does not enforce a range. `internal/intervals/activities.go:135-138` applies that helper to both fields. `internal/tools/get_activities_row.go:203-214` labels the emitted values as feel 1–5 and RPE 1–10. `internal/tools/activity_commute_rpe_test.go:91-111` covers wrong types and fractional values, but not integral out-of-range values.
- **Behavior at risk:** An upstream `feel: 0`, `feel: 6`, `icu_rpe: 0`, or `icu_rpe: 11` is emitted as a seemingly valid normalized field while the response metadata claims a narrower scale. Consumers may treat the value as a valid athlete rating.
- **Why it matters:** This is a contract/data-quality defect, not a preference. A shared helper hides two different scale contracts, coupling normalization to metadata that it does not enforce. It also makes malformed values look more trustworthy than they are.
- **Smallest remediation:** Keep the raw value available under `include_full`, but make normalization field-specific (or parameterize the helper with the allowed range): accept only feel 1–5 and `icu_rpe` 1–10, otherwise omit the normalized field and retain the existing malformed/raw boundary. Add table cases for lower/upper out-of-range and valid endpoints in list and detail reads.

### O-252-01 — Local windows still pay the complete upstream fetch cost

- **Lens:** Hickey / simplicity; Carmack / pragmatism.
- **Severity:** **Observation** (not a release blocker).
- **Confidence:** High.
- **Evidence:** `internal/tools/get_activity_streams.go:253-255` calls `GetActivityStreams` once with the upstream type list before local shaping. `internal/tools/get_activity_streams.go:609-729` builds a common index mask and bounds the returned `data`, `data2`, samples, and full payload locally. The generated input schema at `internal/tools/get_activity_streams.go:853-854` explicitly says local slicing fetches the complete stream because no verified upstream window query exists. `docs/upstream-gaps/activity-stream-windows.md` records the same limitation; the TP-252 window tests exercise alignment and raw replacement.
- **Behavior/cost:** A narrow window reduces response payload size and downstream token cost, but does not reduce upstream transfer, decode, or temporary memory cost. Very large activities remain expensive even when the caller asks for a small window.
- **Why it matters:** This is a measurable resource trade-off, already visible in the contract. It should not be disguised as server-side query efficiency.
- **Recommendation:** Keep the current explicit local implementation until a verified upstream query contract exists. If that evidence arrives, add request-query coverage and replace the fetch path narrowly; do not send speculative parameters. A future benchmark or diagnostic may quantify the full-fetch cost, but no rewrite is justified by this review.

## Task-by-task review and non-findings

### TP-252 — Windowed sampled activity streams

**Result:** No additional release defect; O-252-01 is the only material cost observation.

- **Simplicity:** Windows are represented as separate time/distance concepts and combined through one common index mask rather than per-channel shifting. `get_activity_streams.go:391-489` keeps requested/effective provenance and intersection semantics together.
- **Directness:** Validation rejects negative, reversed, oversized, and malformed bounds before the client call (`get_activity_streams.go:371-384` and request decoding at `:38-128`). Missing, invalid, non-monotonic, null, or mismatched boundary/channel data produces stable diagnostics instead of interpolation or fabricated zeros (`:410-454`, `:510-520`, `:662-679`).
- **Pragmatism:** `include_full` gates heavy samples, `max_points` is capped at 5,000 and requires full mode (`:237-241`), and bounded raw arrays replace rather than retain unbounded arrays (`:681-704`). The implementation pays the explicit full-fetch cost but keeps response and model-facing payloads bounded.
- **Non-finding:** No speculative upstream window parameters were added. No silent nearest-neighbor interpolation, per-channel alignment, or null-to-zero conversion was found. The focused TP-252 tests passed in the prior task and the local review checks.

### TP-254 — Activity commute and RPE normalization

**Result:** F-254-01 above; otherwise no additional release defect.

- **Simplicity:** `Activity` carries separate nullable `Commute`, `Feel`, and native `RPE` fields (`internal/intervals/activities.go:59-104`) rather than conflating feel and exertion. The write payload remains a separate sparse type (`:40-57`).
- **Directness:** Exact JSON types are required for normalized values; malformed optional values remain in `Raw` and do not make the whole activity unreadable (`:107-140`). Read rows carry source labels and preserve explicit false/zero values (`internal/tools/get_activities_row.go:41-80,203-214`).
- **Pragmatism/security:** `UpdateActivityParams` and its payload contain only verified sparse fields (`activities.go:40-57`); unsupported RPE writes are not advertised or sent. This avoids a speculative write contract and protects unrelated fields from accidental replacement.
- **Non-finding:** No commute inference from names/tags, feel/RPE conversion, null-to-zero normalization, or Strava-specific fabrication was found. The existing malformed-value tests and sparse-write tests provide focused evidence.

### TP-257 — AlphaHRV and custom-field coverage audit

**Result:** No additional release defect.

- **Simplicity:** The implementation does not create an activity-level AlphaHRV alias or join wellness HRV to activity DFA. Only interval `average_dfa_a1` becomes `intervals[].dfa_alpha1`, with explicit source/scope/unit metadata (`internal/tools/get_extended_metrics.go:276-287`). Explicit activity custom fields retain their exact codes and do not become canonical physiological metrics (`internal/tools/get_activities_row.go:84-107`).
- **Directness:** Missing, null, and malformed interval DFA values produce distinct diagnostics and omit the metric rather than guessing (`get_extended_metrics.go:289-316`). Custom null/absent/malformed values are omitted while selected-code provenance remains available; the contract tests cover zero/false preservation and restricted-source precedence (`internal/tools/alpha_hrv_custom_fields_test.go:67-177,180-257`).
- **Pragmatism/safety:** Custom-field discovery is explicit and bounded; no broad default custom-item lookup is added. Provenance marks unit, scale, algorithm, physiology, and source device as not provided where evidence is absent. The evidence ledger and `docs/upstream-gaps/alpha-hrv.md` keep medical/readiness inference out of the API.
- **Non-finding:** No device inference from `device_name`, no wellness-HRV join, no AlphaHRV relabel, and no medical/readiness conclusion was found in the reviewed surfaces.

### TP-259 — Gemini Spark compatibility and client setup

**Result:** No release defect found.

- **MCP correctness:** `internal/mcp/gemini_compatibility_test.go:66-174` exercises JSON-mode Streamable HTTP initialize, `Mcp-Session-Id`, initialized notification, ping, tools/list, a read-only tools/call, and a sanitized tool error. It verifies JSON-RPC IDs/results and content types. The existing listener-backed protocol tests cover the actual `ServeStreamableHTTP` path.
- **Deadlines:** `gemini_compatibility_test.go:67-71,181-208` uses one 5-second deadline-bearing context for every request and response-body read. The core server sets a 5-second `ReadHeaderTimeout`, a 30-minute SDK session timeout, and a 5-second graceful shutdown (`internal/mcp/transport.go:15-20,164-191`). No unbounded compatibility request was introduced.
- **Security boundary:** Local HTTP remains loopback-by-default and intentionally unauthenticated; non-loopback binding emits a warning (`internal/app/wire.go:90-93`). The Gemini page explicitly warns against LAN exposure, API-key placement, and generic tunnels (`web/content/connect/gemini-spark.md:23-39`).
- **Honest docs:** The page says the test is an in-process protocol check, not an end-to-end Gemini product/mobile test, and assigns public HTTPS/OAuth/DCR work to `icuvisor-host` (`gemini-spark.md:7-15,34-39`). The guidance contract test rejects hosted endpoint and unsupported mobile/OAuth claims (`scripts/tests/test_gemini_spark_guidance.py:29-67`).
- **Non-finding:** No claim that core provides hosted reachability, OAuth, DCR, or mobile support was found. The test is intentionally isolated from real accounts, credentials, services, and athlete data.

## Cross-cutting lens assessment

- **Separate concepts:** Activity feel, RPE, commute, interval DFA, wellness HRV, custom fields, and stream boundaries remain distinct. The one exception is the range-validation gap in F-254-01, where one helper serves two scale contracts.
- **Data over accidental process:** Provenance and availability metadata carry source, scope, selection, and failure state. The stream implementation uses a common index mask; the custom-field implementation preserves exact user-selected codes.
- **Explicit failure behavior:** Invalid windows, unavailable/restricted sources, malformed optional values, and unsupported writes return stable user-facing errors or diagnostics. No candidate high-severity silent data-loss path was found.
- **Bounded payloads and cost:** Default terse responses omit raw samples; full stream responses can be capped; custom fields are explicit. TP-252's complete upstream fetch remains the documented, bounded-scope observation above.
- **Security and observability:** Loopback HTTP is the default, LAN binding is warned about, credentials are kept out of client guidance, and errors are sanitized. No credential, private athlete payload, live-service, or competitor-source evidence was used.
- **Abstraction/rewrite discipline:** No speculative upstream parameters, broad AlphaHRV abstraction, or compatibility transport rewrite is warranted by the evidence. F-254-01 can be fixed with one small normalization change and matrix tests.

## Unexecuted batch tasks and exclusions

TP-253, TP-255, TP-256, and TP-258 were not executed in batch `20260818T192348`; they have no shipped implementation to review here. This report also excludes live intervals.icu/Gemini/client smoke tests, credentials and private athlete data, competitor source, and subjective style preferences without behavioral evidence.

## Step 2 validation evidence

- **F-254-01 reproduction:** the focused temporary test `TestReviewRawActivityScaleIntAcceptsIntegralValuesOutsidePublishedScales` passed with `0`, `6`, and `11`; this confirms the current shared helper accepts integral values outside the published feel/RPE scales. The temporary test file was removed after execution. The existing focused suite still passes: `go test ./internal/intervals ./internal/tools -run 'Test(Activity|Raw|GetActivities|UpdateActivity|Commute|RPE)' -count=1`.
- **Generated catalog/schema:** `make docs-tools` completed successfully, and `git diff --exit-code -- web/data/tools.json web/data/tool_schemas.json cmd/gendocs/testdata/tools.golden.json cmd/gendocs/testdata/tool_schemas.golden.json` returned clean. `go test ./cmd/gendocs ./internal/toolchecks -count=1` passed, including deterministic/golden and schema snapshot checks.
- **Public claims:** `python3 scripts/tests/test_gemini_spark_guidance.py` passed (`Gemini Spark guidance contract passed`), and `python3 scripts/eval/run_eval.py --validate` passed for 36 scenarios against the 71-tool catalog. No public claim exceeded the reviewed implementation boundary.

## Verification status

Commands and results are recorded in the packet `STATUS.md`. The final delivery checkpoint records the full Go test suite, race tests, lint, build, generated catalog/schema checks, relevant documentation contracts, and the invariant that this review modified no production implementation files.

## Remediation summary

1. **TP-254 / medium / should fix:** range-check normalized `feel` and `icu_rpe` independently; preserve out-of-range raw values only in full payloads; add boundary/out-of-range tests. See `docs/reviews/TP-260-follow-up-checklist.md`.
2. **TP-252 / observation / deferred:** retain local full-fetch windows until verified upstream window query evidence exists; if it arrives, add a focused client-query implementation and benchmark. See the deferred item in the follow-up checklist.
3. **No remediation task:** TP-257 and TP-259 have no release findings from this review.

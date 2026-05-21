# TP-076-update-wellness-subjective-write — Status

**Current Step:** Step 1: Live probe to isolate the rejection
**Status:** 🚧 Blocked
**Last Updated:** 2026-05-20
**Review Level:** 2
**Review Counter:** 13
**Iteration:** 1
**Size:** M
**Closes:** #7
**Soft-depends on:** TP-072 (closes #8)
**Requires live API access:** YES (.env-dev test athlete)

---

### Step 1: Live probe to isolate the rejection

**Status:** 🚧 Blocked

- [x] Source `.env-dev` safely and run a production-account guard without printing secrets or raw athlete IDs.
- [x] Snapshot the probe-date wellness row before mutation and define exact restore values, with `locked` tested last.
- [x] Run out-of-process probes starting from the production client shape (`PUT /athlete/{id}/wellness/{YYYY-MM-DD}` with sparse body) and bisect subjective fields/combinations.
- [x] Vary date format, method, and athlete scoping only if the production client shape fails before field-specific evidence.
- [x] Record the exact accepted minimal payload and rejected payload; save sanitized request/response fixtures for the implemented shape.
- [ ] Restore the pre-probe row on the test athlete and verify with `get_wellness_data`/direct GET.

### Discoveries

- Step 1 live probe: production-shape `PUT /athlete/{id}/wellness/{YYYY-MM-DD}` with `{"feel":3,"fatigue":2,"soreness":2,"stress":2,"mood":4,"motivation":4,"sleepQuality":3,"locked":true}` returned HTTP 422 because upstream does not recognize wellness write field `feel`; single-field `feel` also returned 422. Single-field `fatigue`, `soreness`, `stress`, `mood`, `motivation`, `sleepQuality`, and `locked` all returned HTTP 200. A bundle omitting `feel` returned HTTP 200 and re-read with those seven fields present. Method/date/scope variants for `feel` did not produce an accepted shape.
- Issue #7 closure path: final delivery commit will include `Closes #7` for merge-time auto-close; direct `gh issue close` is intentionally deferred until after merge because the fix is not on the default branch yet.

### Blockers

- Step 1 cleanup blocker: the live probe intentionally tested `locked` last, but the public API now ignores `locked:false`, ignores `null` clears for subjective fields, rejects `sleepQuality:0`, rejects alternate unlock field names (`lock`, `isLocked`, `icu_lock`, `icuLocked`), rejects/ignores method/date/scope/query unlock variants, and returns 405 for DELETE on the v1 wellness row. The probe row remains locked with synthetic subjective values on the dedicated test athlete; operator intervention through the intervals.icu UI or a documented unlock endpoint is needed before this checkbox can be truthfully completed. Step 4 stdio re-validation reused this already-contaminated locked row for the accepted seven-field path and did not create a new locked row or a new live-account safety gap. Operator steering on 2026-05-18 says to let TP-076 fail rather than perform additional live cleanup attempts.
- Iteration 5 blocker confirmation: no additional live wellness writes, DELETE attempts, endpoint probes, or new probe-row creation were performed because supervisor/operator steering explicitly forbids them; the final unchecked cleanup checkbox cannot be completed truthfully without external manual cleanup.
- Iteration 1 restart blocker confirmation (2026-05-20): resumed at the same first unchecked Step 1 cleanup checkbox. No additional live wellness writes, DELETE attempts, endpoint probes, direct GETs, or new probe-row creation were performed because prior supervisor/operator steering explicitly says to let TP-076 fail; the checkbox remains impossible to complete truthfully without external manual cleanup or documented operator acceptance of the residual test-athlete contamination.

### Step 2: Add a failing test

**Status:** ✅ Complete

- [x] Add an intervals client fixture test for the live-accepted seven-field upstream body and response, without treating it as the red unsupported-`feel` defect test.
- [x] Add failing intervals/tool tests that explicit `feel` input is rejected before upstream I/O with an actionable public error and no partial-success metadata.
- [x] Update schema/public-contract test expectations so `feel` is not advertised as a writable wellness field after the fix.
- [x] Run the targeted wellness tests and confirm the unsupported-`feel` test fails before the fix.
- [x] Extend schema/public-contract tests so `examples` and `input_examples` do not advertise unsupported `feel`.
- [x] Assert unsupported-`feel` validation returns an empty `Result` with no partial-success metadata.

### Step 3: Fix the client / tool

**Status:** ✅ Complete

- [x] Add a stable unsupported-`feel` error at the intervals client boundary so `WriteWellnessParams{Feel: ...}` fails before network I/O.
- [x] Reject submitted `feel` in the update_wellness tool with the explicit public error and no writer call.
- [x] Remove `feel` from writable schema/examples/field metadata while preserving read-side `feel` support and accepted seven-field subjective writes.
- [x] Reconcile existing positive tests that still send `feel` so targeted intervals/tools tests pass.

### Step 4: Build + lint + race + live re-validation

**Status:** ✅ Complete

- [x] Run `make build`, `make test`, `make test-race`, and `make lint` with all checks passing.
- [x] Source `.env-dev` safely and live-validate via stdio MCP that submitted `feel` is rejected with the explicit public error and no mutation.
- [x] Live-validate the accepted write path without creating a new locked row: use a fresh row for the six non-lock subjective fields or the already-contaminated Step 1 row if `locked` must be included.
- [x] Re-read live wellness data and restore all Step 4 overwrite-able fields to the pre-Step-4 snapshot; leave only the pre-existing Step 1 locked-row blocker if API cleanup remains impossible.

### Step 5: Document amendment

**Status:** ✅ Complete

- [x] Document the upstream wellness write payload finding in `docs/upstream-gaps/wellness-write-payload.md`, including unsupported `feel`, accepted subjective fields, and locked cleanup limitation.

### Step 6: Close the issue

**Status:** ✅ Complete

- [x] Update `CHANGELOG.md` under `[Unreleased] → Fixed` with the TP-076 root cause and behavior change.
- [x] Record the issue closure path in `STATUS.md`, using merge auto-close via `Closes #7` because direct closure must wait for merge.
- [x] Commit final delivery with `fix(wellness): repair update_wellness subjective write` and `Closes #7`.

| 2026-05-17 23:24 | Task started | Runtime V2 lane-runner execution |
| 2026-05-17 23:24 | Step 1 started | Live probe to isolate the rejection |
| 2026-05-17 23:27 | Review R001 | plan Step 1: REVISE |
| 2026-05-17 23:28 | Review R002 | plan Step 1: APPROVE |
| 2026-05-18 00:02 | Review R003 | plan Step 2: REVISE |
| 2026-05-18 00:05 | Review R004 | plan Step 2: APPROVE |
| 2026-05-18 00:09 | Review R005 | code Step 2: REVISE |
| 2026-05-18 00:12 | Review R006 | code Step 2: APPROVE |
| 2026-05-18 00:14 | Review R007 | plan Step 3: APPROVE |
| 2026-05-18 00:20 | Review R008 | code Step 3: APPROVE |
| 2026-05-18 00:22 | Review R009 | plan Step 4: REVISE |
| 2026-05-18 00:24 | Review R010 | plan Step 4: APPROVE |
| 2026-05-18 00:32 | Review R011 | code Step 4: APPROVE |
| 2026-05-18 00:34 | Review R012 | plan Step 5: APPROVE |
| 2026-05-18 00:36 | Review R013 | code Step 5: APPROVE |

| 2026-05-17 23:50 | Worker iter 1 | done in 1567s, tools: 39 |
| 2026-05-18 00:04 | Review R003 | plan Step 2: REVISE |
| 2026-05-18 00:06 | Review R004 | plan Step 2: APPROVE |
| 2026-05-18 00:12 | Review R005 | code Step 2: REVISE |
| 2026-05-18 00:17 | Review R006 | code Step 2: APPROVE |
| 2026-05-18 00:19 | Review R007 | plan Step 3: APPROVE |
| 2026-05-18 00:26 | Review R008 | code Step 3: APPROVE |
| 2026-05-18 00:29 | Review R009 | plan Step 4: REVISE |
| 2026-05-18 00:30 | Review R010 | plan Step 4: APPROVE |
| 2026-05-18 00:37 | Review R011 | code Step 4: APPROVE |
| 2026-05-18 00:39 | Review R012 | plan Step 5: APPROVE |
| 2026-05-18 00:42 | Review R013 | code Step 5: APPROVE |

| 2026-05-18 00:45 | Agent escalate | Blocked on TP-076 Step 1 cleanup. Live probe found root cause (`feel` is rejected by upstream wellness write; seven-field bundle without `feel` succeeds), but cleanup is stuck: the probe row is now `l |
| 2026-05-18 00:45 | Worker iter 2 | done in 3273s, tools: 182 |
| 2026-05-18 00:45 | Steering | Hold TP-076 at the Step 1 cleanup blocker. Do not perform further live wellness writes, DELETE attempts, endpoint probing, or new probe-row creation. Await operator decision: manual intervals.icu UI cleanup or explicit acceptance of residual test-athlete contamination as a documented exception. |

| 2026-05-18 00:47 | Exit intercept close | Supervisor directed session close: "let it fail" |
| 2026-05-18 00:47 | Agent reply | Acknowledged. I will hold TP-076 at the Step 1 cleanup blocker, avoid any further live wellness writes/DELETE/probing/new probe rows, preserve the existing evidence, and await operator decision on man |
| 2026-05-18 00:47 | ⚠️ Steering | Hold TP-076 at the cleanup blocker. Do not perform any further live wellness writes, DELETE attempts, endpoint probing, or new probe-row creation. Preserve the current evidence in STATUS.md, including |
| 2026-05-18 00:47 | Worker iter 3 | done in 115s, tools: 10 |
| 2026-05-18 00:47 | No progress | Iteration 3: 0 new checkboxes (1/3 stall limit) |
| 2026-05-18 00:48 | ⚠️ Steering | Operator said "let it fail" for the unresolved Step 1 cleanup blocker; no further live wellness writes, DELETE attempts, endpoint probes, or new probe-row creation will be performed. |

| 2026-05-18 00:48 | Exit intercept close | Supervisor directed session close: "let it fail" |
| 2026-05-18 00:48 | ⚠️ Steering | let it fail |
| 2026-05-18 00:48 | Worker iter 4 | done in 87s, tools: 6 |
| 2026-05-18 00:48 | No progress | Iteration 4: 0 new checkboxes (2/3 stall limit) |
| 2026-05-20 11:42 | Task started | Runtime V2 lane-runner execution |
| 2026-05-20 11:42 | Step 1 started | Live probe to isolate the rejection |
| 2026-05-20 11:44 | Exit intercept close | Supervisor directed session close: "let it fail" |
| 2026-05-20 11:44 | Agent escalate | Blocked on TP-076 Step 1 cleanup after resuming at the first unchecked checkbox. Prior supervisor/operator steering explicitly says to let TP-076 fail and forbids further live wellness writes, DELETE  |
| 2026-05-20 11:44 | Worker iter 1 | done in 108s, tools: 10 |
| 2026-05-20 11:44 | No progress | Iteration 1: 0 new checkboxes (1/3 stall limit) |
| 2026-05-20 11:44 | Paused | User paused at iteration 1 |
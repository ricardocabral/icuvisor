# TP-080: HR and pace curve siblings to get_power_curves — Status

**Current Step:** Step 6: Documentation & Delivery
**Status:** ✅ Complete
**Last Updated:** 2026-05-20
**Review Level:** 2
**Review Counter:** 13
**Iteration:** 1
**Size:** M

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers may expand steps when runtime discoveries warrant it.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Confirm no protected docs are changed without explicit approval

---

### Step 1: Extract reusable curve plumbing
**Status:** ✅ Complete

- [x] Document Step 1 extraction boundary: share date-range validation, curve-spec construction, positive bucket normalization, bucket lookup, activity-id lookup, missing-bucket metadata, and shaped encoding while retaining typed metric-specific response fields.
- [x] Preserve axis differences in the plan and code: power/HR use duration seconds (`secs` / `duration_seconds`), pace uses distance meters (`distances` / `distance_meters`).
- [x] Treat existing HR/pace intervals client methods as the baseline and add endpoint regression coverage for HR `secs`, pace `distances`, supplied `type`, and intentional HR/pace sport omission behavior.
- [x] Keep power-curve behavior unchanged with regression tests for terse default, `include_full`, missing buckets, default sport/date/bucket handling, and unchanged name/schema/tier/field names.

---

### Step 2: Implement HR and pace curve tools
**Status:** ✅ Complete

- [x] Create `get_hr_curves` as a duration-bucket tool using `ListAthleteHRCurves`, `duration_seconds`, `heart_rate_bpm`, shared terse/full shaping, and schema descriptions naming BPM units.
- [x] Create `get_pace_curves` as a distance-bucket tool using `ListAthletePaceCurves`, `distance_meters`, raw elapsed seconds, preferred `pace_seconds_per_km`/`pace_seconds_per_mile` conversion from athlete profile, and unit metadata.
- [x] Add `get_hr_curves` and `get_pace_curves` to the shared `internal/toolcatalog` as athlete-scoped read tools so registry validation, coach ACLs, and schema injection recognize them.
- [x] Register both tools in the full toolset fitness catalog and update tier/catalog/coach-ACL tests for discoverability.
- [x] Regenerate generated tool catalog golden for the newly registered HR/pace tools.
- [x] Add committed schema snapshots for `get_hr_curves` and `get_pace_curves` and verify schema stability.

---

### Step 3: Test curve symmetry
**Status:** ✅ Complete

- [x] Add table-driven tool tests for power/HR/pace shared response shape, terse default/full behavior, missing buckets, and axis-specific bucket fields.
- [x] Add pace-specific tests for metric, imperial, and unknown preferred-unit fallback behavior.
- [x] Run targeted tool/client/catalog/schema tests affected by the curve tools.

---

### Step 4: Docs and full verification
**Status:** ✅ Complete

- [x] Update tool docs/generated reference.
- [x] Update CHANGELOG.md.
- [x] Run `make test`, `make build`, and `make lint`.

---


### Step 5: Testing & Verification
**Status:** ✅ Complete

- [x] Targeted tests passing
- [x] FULL test suite passing: `make test`
- [x] Build passes: `make build`
- [x] Lint passes: `make lint`
- [x] All failures fixed or documented as pre-existing unrelated failures

---

### Step 6: Documentation & Delivery
**Status:** ✅ Complete

- [x] "Must Update" docs modified
- [x] "Check If Affected" docs reviewed
- [x] Discoveries logged
- [x] Final commit includes task ID

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|
| R001 | Plan | 1 | REVISE | .reviews/R001-plan-step1.md |
| R002 | Plan | 1 | APPROVE | inline |
| R003 | Code | 1 | APPROVE | inline |
| R004 | Plan | 2 | REVISE | .reviews/R004-plan-step2.md |
| R005 | Plan | 2 | APPROVE | inline |
| R006 | Code | 2 | REVISE | .reviews/R006-code-step2.md |
| R007 | Code | 2 | APPROVE | inline |
| R008 | Plan | 3 | APPROVE | inline |
| R009 | Code | 3 | APPROVE | inline |
| R010 | Plan | 4 | APPROVE | inline |
| R011 | Code | 4 | APPROVE | inline |
| R012 | Plan | 5 | APPROVE | inline |
| R013 | Code | 5 | APPROVE | inline |

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| HR/pace intervals client methods already existed before this task; missing work was public tools, catalog wiring, and regression artifacts. | Verified and extended with endpoint tests plus public tool registration. | internal/intervals/fitness.go, internal/intervals/fitness_test.go |
| Curve endpoints do not use pagination; terse behavior is controlled by bucket defaults and `include_full` raw payload gating. | Covered via missing-bucket/default/full tests instead of pagination tests. | internal/tools/get_curve_siblings_test.go |

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-20 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-20 12:41 | Task started | Runtime V2 lane-runner execution |
| 2026-05-20 12:41 | Step 0 started | Preflight |
| 2026-05-20 13:37 | Worker iter 1 | done in 3381s, tools: 193 |
| 2026-05-20 13:37 | Task complete | .DONE created |

---

## Blockers

*None*

---

## Notes

Plan revision notes: Step 1 will not add public HR/pace tools; it will refactor shared curve mechanics and add regression coverage. Existing intervals client support already includes `ListAthleteHRCurves`, `ListAthletePaceCurves`, `DurationSeconds`, and `DistanceMeters`; Step 1 client work is limited to tests unless implementation reveals a concrete gap.

Step 2 implementation plan: HR will mirror the power duration-axis contract with `heart_rate_bpm`; pace will use distance-axis buckets and compute preferred pace seconds per km/mile from upstream elapsed seconds over each distance bucket, preserving raw `elapsed_seconds` and activity IDs. Pace defaults to run-style distance buckets from existing best-efforts conventions (400, 1000, 1609, 5000, 10000 meters) with positive/sorted/dedup normalization and `missing_buckets` metadata. If athlete profile lookup fails, the pace tool returns a short user error; if unit preferences are unknown, it falls back to metric via existing profile-unit behavior and reports metric `_meta.units`. Both new tools stay in the full toolset under the fitness catalog group and are added to the shared athlete-scoped tool catalog.

Step 3 test plan: consolidate curve sibling assertions in table-driven tests where practical, then cover pace unit conversion separately because it depends on athlete profile state. No pagination is expected for these upstream curve endpoints; missing-bucket metadata is the applicable terse-boundary check.

Documentation review: Must-update CHANGELOG.md and generated tool reference data were updated. README setup guidance was not affected. PRD scope already covered full-toolset curve surfaces via roadmap catch-up and did not need a product-contract change.
| 2026-05-20 12:47 | Review R001 | plan Step 1: REVISE |
| 2026-05-20 12:48 | Review R002 | plan Step 1: APPROVE |
| 2026-05-20 12:54 | Review R003 | code Step 1: APPROVE |
| 2026-05-20 12:58 | Review R004 | plan Step 2: REVISE |
| 2026-05-20 13:00 | Review R005 | plan Step 2: APPROVE |
| 2026-05-20 13:11 | Review R006 | code Step 2: REVISE |
| 2026-05-20 13:16 | Review R007 | code Step 2: APPROVE |
| 2026-05-20 13:19 | Review R008 | plan Step 3: APPROVE |
| 2026-05-20 13:23 | Review R009 | code Step 3: APPROVE |
| 2026-05-20 13:25 | Review R010 | plan Step 4: APPROVE |
| 2026-05-20 13:29 | Review R011 | code Step 4: APPROVE |
| 2026-05-20 13:30 | Review R012 | plan Step 5: APPROVE |
| 2026-05-20 13:34 | Review R013 | code Step 5: APPROVE |

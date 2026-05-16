# TP-072-update-wellness-missing-write-fields — Status

**Current Step:** Step 4: Build, lint, manual smoke
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-16
**Review Level:** 1
**Review Counter:** 4
**Iteration:** 1
**Size:** S
**Closes:** GitHub #8

---

### Step 1: Add fields to client write struct + payload
**Status:** ✅ Complete

- [x] Verify read-side wellness JSON tags for `spO2`, `vo2max`, `abdomen`, `respiration`, and `menstrualPhase`.
- [x] Add missing fields to `WriteWellnessParams`.
- [x] Extend `writeWellnessBody` to include the new sparse payload keys.

### Step 2: Expose fields in the tool
**Status:** ✅ Complete

- [x] Extend `updateWellnessRequest` with the five new fields.
- [x] Extend `updateWellnessInputSchema` with documented properties for the new fields.
- [x] Add handler validation for numeric ranges and non-empty `menstrualPhase`.
- [x] Map the new request fields into `WriteWellnessParams`.
- [x] Include the new fields in the `fields_updated` echo.

### Step 3: Tests
**Status:** ✅ Complete

- [x] Add table-driven coverage for each new field asserting outbound body keys, `fields_updated`, and validation failures.
- [x] Add a combined-fields test that writes all five new fields at once.

### Step 4: Build, lint, manual smoke
**Status:** ✅ Complete

- [x] Run `make build`, `make test`, `make test-race`, and `make lint` successfully.
- [x] Decide and document whether the optional `.env-dev` manual smoke was run.
  - Manual smoke not run: skipped optional live `.env-dev` write to avoid mutating athlete data in this automated lane; unit/integration tests cover request schema, payload mapping, validation, and response metadata.

### Step 5: Close the GitHub issue
**Status:** ⏳ Pending

| 2026-05-16 20:44 | Task started | Runtime V2 lane-runner execution |
| 2026-05-16 20:44 | Step 1 started | Add fields to client write struct + payload |
| 2026-05-16 20:46 | Review R001 | plan Step 1: APPROVE |
| 2026-05-16 20:50 | Review R002 | plan Step 2: APPROVE |
| 2026-05-16 20:54 | Review R003 | plan Step 3: APPROVE |
| 2026-05-16 20:59 | Review R004 | plan Step 4: APPROVE |

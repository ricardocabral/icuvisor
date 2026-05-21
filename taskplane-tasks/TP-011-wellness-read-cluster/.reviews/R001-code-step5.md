# Code Review: TP-011 Step 5 — Null-stripping integration

**Verdict: Approve.**

No blocking findings.

The Step 5 change adds an end-to-end `get_wellness_data` handler test that exercises the wellness row collection through the shared `response.Shape` null-strip path. The test verifies both standard and custom null fields are stripped in terse mode and reported in row-local `_meta.missing_fields`, while non-null custom fields and scale metadata survive. It also verifies `include_full: true` suppresses missing-field metadata, preserves a null custom field on the shaped row, and exposes the raw upstream payload under `full`.

## Notes

- The test covers the Step 5 contract against the tool boundary rather than only the generic shaper, which is the right integration point for confirming `RowCollections: []string{"wellness"}` is wired correctly.
- Step 6 should still expand this into the planned table-driven fixture suite covering provenance, stale boundaries, provider-native round-trips, and all scale labels. A small additional assertion that `include_full:true` also preserves the standard null field (`hrv`) would make the Step 5 coverage symmetric with the terse-mode checks, but the current implementation path does preserve it and this is not blocking.

## What I checked

- Ran `git diff 9f4e939a3c1df5eae3a06c2705c88e45d94ef590..HEAD --name-only`.
- Ran `git diff 9f4e939a3c1df5eae3a06c2705c88e45d94ef590..HEAD`.
- Read `PROMPT.md`, `STATUS.md`, `internal/tools/get_wellness_data_test.go`, `internal/tools/get_wellness_data.go`, `internal/intervals/wellness.go`, and `internal/response/shaper.go`.
- Ran `go test ./internal/tools -run TestGetWellnessDataNullStrippingAndIncludeFull -v` — passes.
- Ran `go test ./internal/intervals ./internal/tools ./internal/response` — passes.
- Ran `go test ./...` — passes.

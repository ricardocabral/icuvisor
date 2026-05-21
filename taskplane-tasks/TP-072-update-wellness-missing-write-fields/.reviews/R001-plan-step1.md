# Plan Review: Step 1

Status: Approved

No blocking issues found with the Step 1 plan.

I verified `internal/intervals/wellness.go` read-side tags for the target fields:

- `SpO2` -> `json:"spO2"`
- `VO2Max` -> `json:"vo2max"`
- `Abdomen` -> `json:"abdomen"`
- `Respiration` -> `json:"respiration"`
- `MenstrualPhase` -> `json:"menstrualPhase"`

The proposed client-layer changes are correctly scoped: add the five nullable fields to `WriteWellnessParams` and add matching `setSparse` entries in `writeWellnessBody` using the exact upstream keys above. This preserves existing sparse-update semantics and keeps validation/schema/tool exposure for later steps.

Implementation notes:

- Keep `VO2Max` as the Go field name and `"vo2max"` as the payload key, matching the existing read struct.
- Keep `SpO2` casing consistent with the existing `Wellness.SpO2` field and `"spO2"` JSON key.
- Do not add client-side trimming or enum validation for `MenstrualPhase` in Step 1; the current plan correctly leaves public argument validation to Step 2.

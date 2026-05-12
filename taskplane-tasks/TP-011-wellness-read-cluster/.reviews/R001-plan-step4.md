# R001 Plan Review — Step 4: In-response scale labels

Decision: APPROVE

The Step 4 plan covers the required scope:

- It targets the central TP-007 scale-label registry (`internal/response/shaper.go` / `defaultScaleLabels`) rather than adding wellness-specific ad hoc metadata.
- It includes all required subjective wellness fields: `feel`, `sleepQuality`, `fatigue`, `soreness`, `stress`, `mood`, `motivation`, and `injury`.
- It preserves the separate `sleepScore` 0-100 device scale, which is necessary for the TP-011 sleep dual-scale acceptance criteria even though `sleepScore` is not subjective.
- It correctly relies on the existing response shaper path used by `get_wellness_data` (`RowCollections: []string{"wellness"}`), so row-level `_meta.scales` should be generated after null stripping.
- The proposed targeted response/tool test with a row containing every subjective field is the right verification for this step.

Implementation notes:

- Keep the labels concise but explicit about ranges; `sleepQuality` must remain `1-4`, `sleepScore` must remain `0-100`, and all other listed subjective wellness fields should be `1-5`.
- Update existing response-shaper registry tests as well as the wellness tool test so regressions are caught both at the shared registry level and at the shaped `get_wellness_data` response level.
- Ensure the verification checks row-level `_meta.scales` on each wellness row, not only an aggregate/root `_meta.scales` entry.

# R005 Plan Review — Step 2: Shape activity and wellness responses

**Verdict:** APPROVE.

The updated Step 2 plan in `STATUS.md` addresses the R004 blockers. It now fixes the public key mapping, avoids unsupported fields, defines the metadata shape as `_meta.field_semantics`, preserves zero values by using pointer-shaped activity calories, and calls out the legacy wellness key translation/removal behavior.

## Review notes

- The exact mappings are now explicit and aligned with Step 1 evidence:
  - activity `calories` -> `calories_burned`;
  - wellness `kcalConsumed` -> `calories_intake`;
  - wellness `carbohydrates`/`protein`/`fatTotal` -> `carbs_g`/`protein_g`/`fat_g`;
  - no `calories_total` and no activity macro keys without new fixture evidence.
- `_meta.field_semantics` is a reasonable stable contract. It should be safe with the existing response shaper because non-owned `_meta` keys are preserved by `addCommonMeta`, `addStripMeta`, and `addScaleMeta`.
- The plan now accounts for row-level wellness metadata coexisting with provenance/scales/missing-field metadata, which was the main merge-risk in `get_wellness_data`.
- Converting activity `CaloriesBurned` to a pointer is the right way to preserve present zero values while still stripping absent calories.

## Implementation reminders

- Remove legacy upstream wellness nutrition keys from the shaped top-level wellness row before returning normalized keys. In practice, raw upstream names should only remain under `full` when `include_full: true`; do not emit both `kcalConsumed` and `calories_intake` as peer top-level fields.
- Apply the activity metadata labels to both `get_activities` and `get_activity_details`, since both reuse `getActivitiesRow` and expose `calories_burned`.
- Add/merge wellness `field_semantics` even when provenance is absent; avoid depending on the current `addWellnessMeta` early-return path if nutrition semantics need to be present for rows containing those fields.

No further plan changes are required before coding Step 2.

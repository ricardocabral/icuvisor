# Plan Review: TP-082 Step 1 — Audit write response shaping

## Verdict: approved

The revised Step 1 plan is now concrete enough to execute. It addresses the prior review by auditing from the registered `RequirementWrite` catalog, explicitly bringing the prompt-scope omissions (`link_activity_to_event`, `add_activity_message`) into scope, and requiring a per-tool response-shaping matrix in `STATUS.md` before code changes.

## What looks good

- The scope now includes all current registered write tools:
  - `add_or_update_event`
  - `link_activity_to_event`
  - `add_activity_message`
  - `update_wellness`
  - `update_sport_settings`
  - `apply_training_plan`
  - `create_workout`
  - `update_workout`
  - `create_custom_item`
  - `update_custom_item`
- The planned matrix fields are specific and should be sufficient to drive Step 2/3 decisions: upstream echo type, builder, `encodeShaped` include-full mode, row collections, terse null behavior, raw/full null behavior, and test location.
- The evidence-gathering plan is grounded in repository facts: `internal/tools/catalog.go`, `RequirementWrite` declarations, `encodeShaped` calls, and the shared row builders.
- The plan calls out the most likely divergent paths before implementation, especially the hard-coded full shaping in custom-item create/update, wellness wrapper shaping, event-row reuse, and terse confirmation tools.
- It preserves the important sequencing constraint: document intentional exceptions and high-risk paths in `STATUS.md` before adding tests or changing code.

## Minor guidance for execution

- When populating the audit matrix, also note why `RequirementDelete` tools are excluded from this task, if they are excluded. The prior review suggested this as a non-blocking clarification, and it will prevent scope ambiguity later.
- For each tool marked as an intentional exception, make the reason actionable for Step 2; for example, distinguish “no upstream sparse object is returned” from “tool intentionally has no `include_full` input.”
- If the audit discovers a write tool whose response shape comes from a read helper that always forces full mode, flag it as a Step 3 candidate rather than silently accepting it.

No additional plan changes are required before starting the Step 1 audit.

# R001 plan review — Step 1: Assemble the prompt set

Verdict: **REVISE**

The current Step 1 status captures useful inputs (PRD/roadmap/status files read, current registered read-tool list, and `get_planning_parameters` deferral), but it is not yet a reviewable execution plan for assembling `docs/dogfood/v0.2-prompts.md`. Please hydrate the Step 1 plan before proceeding.

## Required revisions

1. **Add a concrete prompt inventory/matrix before writing the doc.**
   - List planned prompt IDs, the PRD §6 job they cover (`#1`, `#2`, `#4`, `#5`; explicitly exclude write-only `#3`), target cluster, and expected tool calls.
   - Include every registered v0.2 read tool listed in STATUS at least once.
   - Make the coverage auditable: activities, fitness, wellness, events/training plan, workout library, custom items, and the deferred/absent periodization case.

2. **Define prompt-file structure and validation metadata.**
   - Each prompt in `docs/dogfood/v0.2-prompts.md` should have more than just user text: include target tools, data prerequisites/placeholders, expected success criteria, and what failure mode it is meant to expose.
   - Use redacted placeholders such as `<ATHLETE_ID>`, `<ACTIVITY_ID>`, `<STRAVA_ACTIVITY_ID>`, `<DUAL_SLEEP_DATE>`, `<EVENT_ID>`, `<FOLDER_ID>`, and `<CUSTOM_ITEM_ID>` instead of real identifiers.

3. **Plan the three adversarial prompts precisely.**
   - Unit-system reasoning: specify a prompt that forces the model to say it lacks a required unit/measurement rather than guessing, and identify the expected safe answer.
   - Strava handling: specify the required data prerequisite (known Strava-imported/blocked activity or list row) and expected `unavailable.reason == "strava_tos"` behavior.
   - Dual sleep scale: specify the required wellness date with both `sleepQuality` and `sleepScore`, and the expected answer must report both scales separately (`sleepQuality` 1–4; `sleepScore` 0–100) without collapsing them.

4. **Avoid accidental heavy-payload/context-window prompts.**
   - For `get_activity_streams`, plan prompts that request explicit stream keys or intentionally validate the terse default; do not default to broad `include_full:true` stream dumps.
   - If any prompt intentionally exercises `include_full`, label it as such and explain why it is safe for Step 2 token-budget measurement.

5. **Make the coach-mode/customer-job #5 coverage explicit.**
   - Since coach-mode write tools and athlete-listing are not shipped in v0.2, define the read-only coach-style prompt using the existing `athlete_id` targeting semantics or redacted invited-athlete context, without implying the server should expose credentials or write operations.

6. **Preserve the upstream-gap decision for periodization.**
   - The prompt set should not expect `get_planning_parameters`, but it should include either a planning-block/taper prompt that uses shipped reads (`get_fitness`, `get_events`, `get_training_plan`, workout library) or a documented note that athlete-level periodization parameters are deferred pending upstream API exposure.

Once those details are added to STATUS or an adjacent plan section, Step 1 should be straightforward to approve.

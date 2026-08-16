# Activity Carbohydrate Intake Write Design

## Context

Issue #58 requests an MCP write path for the athlete-logged carbohydrate intake
already returned by activity reads as `carbs_ingested_g`. The current
`update_activity` tool only accepts `name` and `description`, although the
intervals.icu `PUT /api/v1/activity/{id}` endpoint accepts an `Activity` body
whose integer `carbs_ingested` field represents the same value.

The upstream endpoint does not update Strava activities. icuvisor will describe
that limitation but will continue to rely on the existing short, sanitized write
error path when upstream rejects a particular activity.

## Scope

Extend the existing `update_activity` tool with one optional field:

- `carbs_ingested_g`: athlete-logged carbohydrate consumed during the activity,
  as a whole number of grams greater than or equal to zero.

This is set-only support. Omitting the field leaves the upstream value unchanged.
An explicit JSON `null`, a negative number, a fractional number, or a non-number
is invalid. Zero is valid and must be sent upstream rather than treated as an
omitted value.

The implementation must not make `carbs_used_g` writable. That field is an
upstream estimate, not athlete-logged intake.

## Architecture and data flow

The change follows the existing sparse-update path rather than adding a new MCP
tool:

1. `updateActivityRequest` decodes `carbs_ingested_g` and separately records
   whether the key was supplied.
2. Validation accepts only a non-null integer greater than or equal to zero.
3. The handler maps the public unit-labelled field to
   `intervals.UpdateActivityParams.CarbsIngested` and sets a corresponding
   presence flag.
4. `internal/intervals` builds a sparse JSON payload whose upstream key is
   `carbs_ingested`.
5. The terse response adds `carbs_ingested_g` to `fields_updated`. Existing
   `include_full` behavior continues to return the raw upstream response only
   when explicitly requested.

The existing `RequirementWrite` registration remains appropriate. The update is
not a delete operation and does not require delete-mode gating or a model-supplied
confirmation flag.

## Public contract

The tool description and input schema will state all of the following:

- the value is athlete-logged intake during this activity, in whole grams;
- zero is a logged zero, while omission means unchanged;
- clearing an existing value is not supported;
- `carbs_used_g` is distinct and read-only;
- intervals.icu does not permit this endpoint to update Strava activities.

The generic public write failure remains short and actionable. Internal upstream
details are not exposed to the model.

## Testing

Development will use a red-green-refactor cycle. Tests will cover:

- a positive value reaching the tool client as a sparse update;
- zero reaching the tool client and the HTTP endpoint as `"carbs_ingested": 0`;
- an exact HTTP request body containing `carbs_ingested` and no unrelated fields;
- combination with existing `name` and `description` fields;
- rejection of null, negative, fractional, string, and otherwise empty updates;
- `fields_updated` containing the public name `carbs_ingested_g`;
- schema type, minimum, description, and registration metadata;
- preservation of existing `include_full` behavior and public error sanitization.

After focused tests pass, run formatting, the complete test/check target, and
review the diff for generated documentation and unrelated changes.

## Documentation

Update the PRD activity catalog to document `update_activity` and its supported
sparse fields, update `CHANGELOG.md` under `[Unreleased]`, and regenerate the
website tool reference and schema snapshots through the repository's existing
documentation target.

## Other activity fields audit

The upstream update operation reuses the full `Activity` response schema. Schema
membership alone is not evidence that a field is safely writable: many listed
properties are calculated, device-owned, source identifiers, or outputs that
upstream can ignore or recompute.

Public intervals.icu material supports future investigation of these groups:

- manual activity measurements: `distance`, `coasting_time`, `max_speed`,
  `calories`, `average_heartrate`, `max_heartrate`, and
  `total_elevation_gain`;
- athlete feedback: `icu_rpe` and `feel`, with their exact scale semantics made
  explicit at the MCP boundary;
- activity-specific threshold: `icu_ftp`;
- classifications and equipment: `race`, `trainer`, `commute`, and gear, after
  black-box validation of their exact request shapes and synchronization effects.

Potentially editable schema fields such as tags, strength/swim measurements,
route assignment, time corrections, and recalculation flags need direct public
contract evidence or clean-room black-box validation before exposure. Calculated
carbohydrate use, load, fitness, zones, weather, derived metrics, timestamps,
source/device identifiers, streams, and interval analysis remain out of scope for
generic activity metadata updates.

Those candidates are not bundled into issue #58. They should be proposed as
separate focused issues so validation, units, ranges, update side effects, and
tool-schema token costs can be reviewed independently.

# R007 plan review — Step 3

Verdict: REQUEST CHANGES

The Step 3 plan targets the right failing surfaces (NOTE date serialization and NOTE name validation), but it is too broad/ambiguous in two places that could create regressions outside the live-probed contract.

## Findings

1. **Date serialization scope is broader than the evidence.**
   The plan says to “generalize event date serialization” so NOTE sends an ISO local datetime. Step 1 only proved NOTE create rejects date-only and WORKOUT already expects the suffix. It did not establish that every other category (for example `RACE`, `RACE_B`, etc.) should now be rewritten from `YYYY-MM-DD` to `YYYY-MM-DDT00:00:00`. Please make the implementation plan explicit: either use the minimum scoped change (`WORKOUT` and `NOTE` date-only values get the suffix, other categories unchanged), or add tests/evidence for any broader category-wide change.

2. **NOTE `name` validation should be scoped to the probed operation.**
   STATUS records that “NOTE creates require a non-empty `name`.” The plan says “Enforce ... the NOTE `name` requirement in validation,” which could accidentally reject NOTE updates with `event_id` when the caller only wants to update another field. Unless Step 1 also probed NOTE update semantics, scope the validation to NOTE creates (`event_id` omitted), or add an update probe/test before enforcing it for all NOTE writes.

3. **Include the public validation message in the plan.**
   The current public error summary says to provide date, category, and type for WORKOUT events. After adding NOTE name validation, that message will no longer be actionable for the new failure path. Add updating `invalidAddOrUpdateEventArgumentsMessage` to the Step 3 plan along with the input schema text.

## Suggested adjusted Step 3 checklist

- Serialize date-only NOTE creates as `YYYY-MM-DDT00:00:00`, preserving existing WORKOUT behavior and leaving unprobed categories unchanged unless separately tested.
- Require a non-empty `name` for NOTE creates; do not reject NOTE updates without evidence that upstream requires `name` there too.
- Update the schema and public invalid-arguments message to mention the NOTE create `name` requirement.
- Run the targeted NOTE regression plus existing WORKOUT create/update tests.

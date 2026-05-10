# Code Review — TP-002 Step 1

## Findings

### [P2] Align the planned profile struct with the chosen endpoint schema

`STATUS.md:69-70` selects `GET /api/v1/athlete/{id}` as the v0.1 source for profile-with-sport-settings, but the next row says to use typed `AthleteProfile`/`SportSettings` structs containing fields such as `firstname`, `measurement_preference`, `weight_pref_lb`, `fahrenheit`, and `locale`.

That mixes two different public schemas: `/athlete/{id}` returns `WithSportSettings`, while `/athlete/{id}/profile` returns an `AthleteProfile` wrapper whose stable identity data is nested under `athlete` and does not include the same unit/sport-settings fields. If Step 2 follows the current wording literally, the client can either decode the wrong shape for `/profile` or name/model the `/athlete/{id}` response incorrectly.

Please update the plan to define the main v0.1 response as the `/athlete/{id}` schema (for example `AthleteWithSportSettings`/`WithSportSettings` plus embedded `SportSettings`), and only define a separate `/profile` wrapper type if that lighter endpoint will actually be called.

### [P3] Task state/status is internally inconsistent

`STATUS.md:7` still says `State: Ready`, while `STATUS.md:11-16` marks Step 1 as in progress with all Step 1 checklist items checked. This makes it unclear whether the step is complete, still running, or ready for implementation/review. Please set the task/step status to a consistent state before proceeding to Step 2.

### [P3] Discoveries table will not render as a table

There is a blank line between the table separator and the first row at `STATUS.md:63-66`. In most Markdown renderers, that terminates the table, so the discovery rows render as plain text instead of table rows. Remove the blank line and add a trailing newline at EOF.

## Notes

- No production code changed in this step; review focused on the planning artifact.
- The retry policy recorded in `STATUS.md:71` is reasonable for the v0.1 GET-only client and matches the task constraints.

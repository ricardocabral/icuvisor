# Plan Review — TP-077 Step 3

Verdict: **request changes**

## Findings

1. **Do not plan local validation of folder existence/ownership.**  
   The Step 3 bullet says to require a “non-empty existing `folder_id`” in both the intervals client and tool validation. The code paths being changed (`create_workout` request decoding and `CreateLibraryWorkout`/`writeWorkoutBody`) can validate only that `folder_id` is present after trimming. They cannot prove the folder exists or is owned by the athlete without adding a new list-folders lookup/interface dependency, which would broaden this step and introduce extra I/O. Revise the plan wording to: locally require a non-empty `folder_id`; document that it must be an existing folder owned by the athlete; let upstream/live validation catch non-existent IDs.

2. **Include the public error text in the planned tool changes.**  
   `invalidCreateWorkoutArgumentsMessage` still says `optional folder_id`, which will be wrong once `folder_id` is required. Step 3 should explicitly update that user-facing validation summary (for example, “provide name, sport, folder_id, optional tags...”) in addition to the schema description/required list/examples. Otherwise the tool can reject missing folders with a contradictory public error.

3. **Keep the fix limited to create semantics.**  
   Step 1 confirmed upstream still expects JSON key `type`, not `sport`, and Step 2 already protects that with `create_request.json`. The Step 3 plan should state that the client payload keeps `body["type"] = sport` and only adds create-time folder validation. Do not alter update behavior: update tests currently allow `FolderIDSet` with an empty string for top-level moves, and TP-077 is only about create.

## Recommended amended Step 3 plan

- In `decodeCreateWorkoutRequest`, trim and require non-empty `folder_id`; update the returned/public validation wording to no longer call it optional.
- In `writeWorkoutBody` for `allowSparse == false`, trim and require non-empty `params.FolderID` before building the POST body; send it as `folder_id` and continue sending sport as JSON `type`.
- Update `create_workout` schema `required`, `folder_id` description, and every example/input_example to reflect “required existing folder owned by the athlete.”
- Run the focused red tests from Step 2 (`go test ./internal/tools ./internal/intervals -run 'CreateWorkout|CreateLibraryWorkout'`) and record that they now pass.

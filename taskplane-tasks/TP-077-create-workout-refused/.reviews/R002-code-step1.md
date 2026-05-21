# Code Review — TP-077 Step 1

Verdict: **request changes**

## Findings

1. **Missing probe result for omitted `folder_id` leaves the production contract unresolved.**  
   `STATUS.md:25` documents success with a real folder, rejection for `sport` without `type`, and rejection for explicit `folder_id: null`, but it does not document the separately required case where the `folder_id` key is omitted. That distinction matters because the current production create path omits `folder_id` when the tool caller leaves it blank, while `null` is a different JSON payload. R001 explicitly required probing omitted vs null before the live probe, and Step 1's purpose is to decide whether the next steps should keep top-level omission, require a folder, or model another shape. Please run and record the omitted-key POST result, or if it was already run, add the sanitized status/error summary to `STATUS.md` before using this probe to drive Step 2/3.

2. **The captured response fixture keeps account-specific computed training settings that are not needed for the contract fixture.**  
   `internal/intervals/testdata/workout_library/create_response.json:28-129` includes the full generated `zoneTimes`, `maxWatts`/`minWatts`, `average_watts`, `normalized_power`, and strain score values from the live test athlete. The Step 1 plan/R001 asked for sanitized fixtures and to keep only the payload shape/fields needed for tests. IDs and timestamps were replaced, but these derived values still expose the test account's power-zone profile and make the fixture much noisier than necessary. Please either reduce the fixture to the fields needed by the upcoming test (`id`, `name`, `type`, `folder_id`, `description`, minimal `workout_doc` shape, etc.) or replace the account-derived numeric settings with clearly synthetic values.

## Notes

- `git diff --check aa2f02f..HEAD` passed.

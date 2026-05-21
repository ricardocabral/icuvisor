# Plan Review — Step 3

Decision: **changes requested**

## Blocking finding

The Step 3 plan says to extend `TestGetWorkoutsInFolderFiltersAndPreservesWorkoutDocWithIncludeFull` to assert `description` is present when `include_full: true`, but it does not say to add a non-empty `description` to that test fixture.

The current include-full fixture for the folder-20 workout has no `description` field:

```json
{"id":2,"name":"Sweet Spot","type":"Ride","folder_id":20,...}
```

Because `workoutInFolderRow.Description` has `omitempty`, even the corrected implementation will omit `description` when the upstream workout has no description. A presence assertion would either fail for the wrong reason or force a weak assertion that does not prove the intended behavior.

## Required plan adjustment

Before asserting presence in the include-full path, update the matching workout fixture with a non-empty description, for example:

```json
{"id":2,"name":"Sweet Spot","description":"multi-paragraph coach notes","type":"Ride","folder_id":20,"icu_training_load":70,"moving_time":3600,"target":"POWER","tags":["sweet-spot"],"workout_doc":{"steps":[{"duration":600},{"duration":300}],"name":"raw doc"}}
```

Then assert the exact value is preserved when `include_full: true`, e.g. `row["description"] == "multi-paragraph coach notes"`.

## Validation notes

- Keep the existing terse-default test assertion from Step 1; it already has a non-empty description fixture and should now pass after the Step 2 shaping change.
- Run a targeted test first for quick feedback, then the required broader commands:

```sh
go test ./internal/tools -run 'TestGetWorkoutsInFolder(FiltersAndPreservesWorkoutDocWithIncludeFull|HidesWorkoutDocByDefault)' -count=1
make test
make test-race
```

Once the fixture adjustment is included, the rest of the Step 3 plan is appropriately scoped.

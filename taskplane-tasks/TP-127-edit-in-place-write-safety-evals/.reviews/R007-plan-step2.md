# Plan Review — Step 2: Add eval/adversarial coverage

**Verdict: REVISE before executing Step 2.**

The revised checklist addresses the separate adversarial-doc concern, but two parts remain too ambiguous to lock in the safety contract.

## Findings

1. **The cookbook eval still does not choose a concrete update path.**  
   The plan says “existing event/template” and “update/edit tools,” which leaves the worker free to build an eval that passes while updating the wrong surface. For a prompt about changing tomorrow’s scheduled workout, pin the calendar path explicitly: expected read-before-write tools such as `resolve_calendar_dates` + `get_events`, then `add_or_update_event` with the existing `event_id`. The forbidden set should include the unsafe recreate/delete path (`create_workout`, `delete_workout`, `delete_event`, `delete_events_by_date_range`; consider forbidding `update_workout` too if the scenario is specifically a calendar event rather than a library template). If the intended scenario is a library-template edit instead, say so and pin `get_workout_library`/`get_workouts_in_folder` + `update_workout` instead.

2. **The safe-mode guidance assertion still does not name the actual surface under test.**  
   The checklist says to assert the “actual safe-mode/delete-mode guidance surface,” but the artifacts still point at `internal/tools/*delete*_test.go`, which cannot cover safe-mode unavailability because delete tools are unregistered there. Name the exact target before implementation, e.g. `internal/tools/list_advanced_capabilities_test.go` for the short/actionable server-config-only enablement text and/or `internal/safety/adversarial_test.go` for registration-time absence. Add those files to the Step 2 artifacts/checklist as appropriate.

## Required plan adjustment

- Replace “event/template” with one concrete scenario contract and list exact `expected_tools`/`forbidden_tools`.
- Replace the vague guidance-surface checkbox with the exact test/file surface (`icuvisor_list_advanced_capabilities` output and/or safety registration matrix).
- Make `go test ./internal/safety` unconditional if the Step 2 assertion relies on the safety registration matrix.

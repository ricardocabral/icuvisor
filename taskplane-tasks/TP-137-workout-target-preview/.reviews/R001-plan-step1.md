# Plan Review — Step 1

Result: Needs revision.

The Step 1 checklist is directionally right, but there is no concrete design plan yet to review. Before implementation, please expand Step 1/Discoveries with:

1. **Exact response shape and placement.** Name the field(s), likely under `workout_doc_summary`, and specify compact contents, examples, and omission/null rules. Avoid raw step expansion by default.
2. **Full call-site inventory.** `workout_doc_summary` is shared by `get_events`, `get_event_by_id`, `get_today` annotations, `get_workout_library`, `get_workouts_in_folder`, and write response rows via `workoutToRow`/`eventRow`. The plan currently scopes only some files, so either include these call sites/tests or explicitly justify limiting previews to planned-workout reads.
3. **Profile/threshold selection rules.** Existing handlers already fetch profile but discard thresholds via `toolProfile`; plan how to reuse that fetch without another API call, match sport settings by event/workout sport/type, and handle `indoor_ftp` vs `ftp`.
4. **Supported conversion semantics.** Record formulas and omissions for `% FTP`, HR percent variants, pace threshold percent, zones, ramps/ranges, nested repeats, missing thresholds, and text/non-numeric targets. Pace-percent semantics especially need an explicit convention to avoid misrepresenting targets.
5. **Test target updates.** Add tests covering the shared row paths affected by the chosen scope, not only `get_events`/`get_workout_library`, plus regression assertions that `include_full:false` still omits raw `workout_doc`/`full` payloads.

Once those design decisions are documented in STATUS.md Discoveries (or an amended Step 1 plan), the implementation should be straightforward.

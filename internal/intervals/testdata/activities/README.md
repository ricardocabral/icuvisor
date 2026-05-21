# Activity upstream-signal fixtures

Sanitized fixtures in this directory document upstream edge-case shapes used by regression tests. They must not contain live athlete identifiers, API keys, personal notes, raw GPS tracks, or unredacted activity names.

- `strava_sync_chain_empty_stubs.json` — synthetic reproducer set for numeric/no-`i` activity IDs observed when Strava-origin activities arrive through Wahoo, MyWhoosh, or TrainerRoad sync chains as empty/minimal upstream objects. The stable contract is that icuvisor treats these as Strava-unavailable stubs and returns structured unavailable markers instead of sparse metric rows.

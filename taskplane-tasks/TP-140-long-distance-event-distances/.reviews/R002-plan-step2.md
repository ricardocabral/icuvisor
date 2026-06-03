# Review R002 — Plan Review for Step 2

Result: Approved.

The Step 2 plan aligns with TP-140: add focused regression coverage that 1200 km values are accepted as meter-valued planned distances, preserved on read/write shaping, and not accompanied by any false Icuvisor-generated load claim. Step 1 already found no local max-distance cap, so this should primarily be test coverage unless implementation reveals a hidden cap.

Guidance for implementation:
- Use `1_200_000.0` meters explicitly in tests to avoid km/m ambiguity.
- Cover the tool write boundary (`add_or_update_event`) by asserting `distance_meters` reaches `intervals.WriteEventParams.DistanceMeters` unchanged and the returned row preserves `distance_target_meters` unchanged. Prefer table coverage for create and update if practical, since the step says creating/updating.
- Cover the read boundary (`get_events` / shared `eventRow`) with upstream `distance` and/or `distance_target` set to `1200000`, asserting no truncation in `distance_meters` / `distance_target_meters`.
- For the no-auto-load requirement, assert load fields are absent unless explicitly present upstream; avoid brittle string-only checks where direct field absence checks are available.
- If any `internal/intervals` test or code is added/touched to verify `distance_meters` serializes as upstream `distance_target`, also run `go test ./internal/intervals` in addition to the planned `go test ./internal/tools`.

No blocker to proceed.

# R001 code review — Step 2

Decision: **REVISE**

## Blocking findings

1. **Fallback scan metadata omits/misreports `count` and `truncated` in common miss cases.**  
   In `getEventByIDMeta`, `count` and `truncated` are tagged with `omitempty` (`internal/tools/get_event_by_id.go:56-58`), and `shapeGetEventByIDScanResponse` only sets `truncated` when `len(events) > fallbackEventByIDLimit` (`internal/tools/get_event_by_id.go:192-197`). This breaks the Step 2 contract for structured miss results, which requires `_meta.count` and `_meta.truncated`: a zero-result scan omits `count`, any non-truncated miss omits `truncated:false`, and a real upstream response of exactly the requested cap (`limit=500`) reports/omits `truncated` even though more upstream rows could hide the target. Please make fallback metadata explicit (including `count: 0` and `truncated: false`) and treat an at-cap miss as potentially truncated, e.g. `len(events) >= fallbackEventByIDLimit` when the ID is not found.

## Notes

- `go test ./...` passes locally.
- The detail client route and 404-only fallback flow otherwise match the approved Step 2 plan, and the tool correctly returns a structured non-error `unavailable` envelope on the covered miss path.

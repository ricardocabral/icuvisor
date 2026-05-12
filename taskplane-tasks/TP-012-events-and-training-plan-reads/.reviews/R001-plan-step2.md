# R001 plan review — Step 2

Decision: **APPROVE**

The revised Step 2 plan now pins down the previously blocking contracts: the detail client method and route (`GET /athlete/{athlete_id}/events/{event_id}` with no format suffix), 404-only fallback behavior, deterministic default scan windows via an injected clock, and stable top-level success/miss envelopes. It also preserves the task’s core requirement that a persistent detail/list mismatch returns structured `unavailable` data as a non-error result instead of exposing a raw 404 to the LLM.

## Non-blocking implementation notes

- Make the fallback scan-window decoder explicit in code/tests: require `oldest` and `newest` as a pair, reject mixed/ambiguous `date` plus range inputs unless deliberately supported, and keep the inclusive ±30-day window at 61 days.
- For fallback truncation metadata, consider treating `len(events) >= fallbackLimit` as “could be truncated” on a miss, not only `len(events) > fallbackLimit`, since an upstream-limited page of exactly 500 can still hide the target.
- Match IDs after the same stringification/normalization path used by `intervals.Event.ID`, and scan the full returned slice before applying any response-size cap; the tool returns one event or unavailable, not the scanned list.
- Keep `resolve` defaulting to true only for the fallback list scan, and ensure caller-provided `resolve:false` is respected.
- Reuse `eventRow` for detail and list-scan recovery so `get_event_by_id` cannot drift from `get_events` terse/full rendering.

No further plan changes are required before implementation.

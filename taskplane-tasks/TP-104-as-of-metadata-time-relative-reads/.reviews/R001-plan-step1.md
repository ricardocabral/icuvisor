# Plan Review R001 — Step 1

**Verdict:** Request changes / clarify before implementation.

I only found the high-level Step 1 checklist in `STATUS.md`; there is no concrete implementation plan for the shared helper to review. Before coding, please pin down these details so later steps do not fork behavior across tools:

1. **Helper API and location.** Prefer a small shared helper in `internal/response` (or a clearly shared `internal/tools` file) that reuses the existing timezone-loading behavior from `response.RenderTimeInTimezone` / `RenderDateInTimezone` rather than adding another independent `time.LoadLocation` path. The API should return all four fields from one localized instant: RFC3339 `as_of`, `as_of_date`, `as_of_weekday`, and `timezone`.
2. **Clock contract.** Decide whether the helper accepts `time.Time` or `func() time.Time`. Since Step 3 will need deterministic current-day checks in tools that do not currently inject a clock, the plan should state how those tools will avoid direct, untestable `time.Now()` calls.
3. **Error behavior.** Explicitly preserve “invalid timezone returns an error” and do not silently fall back to UTC. Empty timezone fallback should match current project behavior, but malformed IANA zones must fail.
4. **Tests.** Add table-driven helper tests covering at least: positive-offset date differs from UTC, negative-offset date differs from UTC, weekday derived from the same localized instant, trimmed/empty timezone behavior, and invalid timezone errors.
5. **Integration safety.** If replacing `athleteLocalDate`, keep `get_today`’s existing date semantics unchanged until Step 2 adds the new metadata fields.

Once the plan specifies these points, Step 1 is well-scoped and low risk.

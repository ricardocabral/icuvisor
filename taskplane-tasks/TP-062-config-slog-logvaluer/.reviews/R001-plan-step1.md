# Plan Review — Step 1: Implement `LogValue`

Decision: **changes requested before coding**

The current step outline is directionally correct, but it needs a few implementation details tightened so the change is safe and compiles.

## Findings

1. **Do not implement the literal `slog.GroupValue(slog.Group("config", attrs...).Value)` shape.**
   `slog.GroupValue` takes `...slog.Attr`, not a `slog.Value`, so that expression will not compile. Return the group value directly:

   ```go
   return slog.GroupValue(
       slog.String("api_base_url", c.APIBaseURL),
       // ...
   )
   ```

   The caller's attr key already supplies the outer name; `slog.Info("config", "cfg", cfg)` should produce `cfg: { ... }`, not an accidental nested `cfg.config` group.

2. **Define the athlete-ID redaction decision in the plan.**
   The requested attr name is `default_athlete_id`, but project rules prohibit logging raw athlete identifiers in a way that is hard to scrub. The implementation should explicitly choose one of:
   - a redacted presence marker matching `String()` semantics (`<set>` / `<unset>`), or
   - the canonical `c.AthleteID` only because it is under a stable structured key that operators can scrub.

   In either case, do not log `c.Coach.Athletes`, athlete labels, or roster IDs; keep `coach_athletes_count` as the only coach-roster detail.

3. **Pin the exact attrs and conversions.**
   To avoid accidentally dumping the whole struct, Step 1 should implement only the explicit allowlist from the prompt:
   - `api_base_url`: `c.APIBaseURL`
   - `default_athlete_id`: per the redaction decision above, using the effective/default `c.AthleteID`
   - `http_bind`: `c.HTTPBindAddress`
   - `coach_athletes_count`: `len(c.Coach.Athletes)`
   - `delete_mode`: `c.DeleteMode.String()`
   - `toolset`: `c.Toolset.String()`

   Never include `api_key` as a key or value, and avoid logging nested `Coach` directly.

4. **Resolve the prompt ambiguity before tests.**
   The file-scope note says tests should assert “all other public fields are present”, while the mission names a smaller explicit attr set. Prefer the explicit allowlist above unless the task owner clarifies otherwise; broadening to all exported fields increases the risk of sensitive data appearing in structured logs.

With those adjustments, the plan is small and appropriate for Step 1.

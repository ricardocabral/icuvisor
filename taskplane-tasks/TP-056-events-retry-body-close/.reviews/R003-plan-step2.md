# Review R003 — Plan review for Step 2

Verdict: **changes requested**

The Step 2 plan is aimed at the right remaining issue: `events.go` still has `defer resp.Body.Close()` on the success path inside the `for attempt := 1; ; attempt++` loop. However, the plan needs a little more precision before implementation to avoid either widening scope or reintroducing a close leak on error paths.

## Blocking concerns

1. **Do not implement the success path as `readBody`/`io.LimitReader` unless the task is explicitly amended.**
   The prompt's “Do NOT” section says not to introduce `io.LimitReader` here, and the acceptance criteria require no user-visible behavior change. Calling the package-level `readBody` helper from `events.go` would add the max-response-size behavior to this event write path. For this task, prefer the minimal inline-close pattern around the existing decoder.

2. **The body must be closed even when JSON decoding fails.**
   Avoid this shape:
   ```go
   if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
       return fmt.Errorf("decoding intervals.icu response: %w", err)
   }
   return resp.Body.Close()
   ```
   because it leaks the successful HTTP response body on decode errors. The plan should explicitly use the safe ordering:
   ```go
   decodeErr := json.NewDecoder(resp.Body).Decode(out)
   closeErr := resp.Body.Close()
   if decodeErr != nil {
       return fmt.Errorf("decoding intervals.icu response: %w", decodeErr)
   }
   if closeErr != nil {
       return fmt.Errorf("closing intervals.icu response: %w", closeErr)
   }
   return nil
   ```
   This removes the in-loop defer while preserving closure on every terminal success-response path.

3. **Keep the existing non-2xx retry branch unchanged except for necessary formatting.**
   In the current checkout, that branch already drains and closes before retrying. The fix should not change retry counts, retry eligibility, `Retry-After` handling, or the current “ignore close error when retrying” behavior that mirrors `client.go`.

## Recommended plan revision

- Replace the Step 2 first bullet with: “Replace the success-path `defer resp.Body.Close()` with explicit `decodeErr`/`closeErr` handling so the body is closed before any return, including decode errors; do not call `readBody` or add response-size limiting in this task.”
- Keep the existing retry-branch drain-and-close code intact.
- Optionally add a small regression case for a 200 response with invalid JSON using a counting body, asserting the body is closed when `AddOrUpdateEvent` returns a decode error. This would guard the most likely bad implementation of the defer removal.
- Run the required checks and update `CHANGELOG.md` under `[Unreleased]` / `Fixed` as already planned.

Once these constraints are added, the Step 2 plan is suitably scoped for TP-056.

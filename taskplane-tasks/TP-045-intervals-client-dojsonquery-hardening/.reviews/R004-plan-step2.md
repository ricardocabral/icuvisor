# Plan Review — TP-045 Step 2

**Verdict:** Approved, with implementation cautions.

The Step 2 plan in `STATUS.md` is concrete enough to implement the split. It preserves the existing public `doJSON` / `doJSONQuery` surface, keeps request construction and `User-Agent` handling in a single-attempt helper, moves retry classification into a side-effect-free `shouldRetry`, and assigns attempt budget, `Retry-After`, sleeps, draining, and body closure to the outer loop. The planned 32 MiB `maxResponseBodyBytes` cap via `io.LimitReader(r, maxResponseBodyBytes+1)` and `ErrResponseTooLarge` sentinel directly addresses the unbounded success-path decode hazard.

The recorded retry policy also matches the current `internal/intervals/client.go`: hand-rolled retry, 3 attempts by default, 200ms base delay, 2s max delay, default jitter only for the completely zero `RetryConfig`, retryable `429` / `>=500`, and transport retries only while the context is live and the attempt budget remains.

## Cautions for implementation

1. **Do not let close errors mask the bounded-read error.**  
   The plan says close errors on success will be returned before decode, which is fine when the read succeeded. If `readBody` returns `ErrResponseTooLarge` or another read error and `resp.Body.Close()` also fails, preserve the read error for `errors.Is(err, ErrResponseTooLarge)`; use `errors.Join` or return the read error with close detail only if it does not hide the sentinel.

2. **Keep response ownership local and obvious.**  
   In every branch after `do` returns a non-nil response, close the body before `continue` or `return`. Avoid helper shapes where the body is sometimes closed by the caller and sometimes by the helper; this is the footgun this task is removing.

3. **Ensure `shouldRetry` stays classification-only.**  
   It should answer only whether the `(resp, err)` kind is retryable. Attempt budget, `ctx.Err()`, `Retry-After`, and sleeping should remain in `doJSONQuery`, otherwise the split will drift back toward the current mixed-responsibility loop.

4. **Decode from a closed, bounded buffer.**  
   The target pattern should be: read bounded bytes, close the body, handle read/close errors with intentional precedence, then `json.Unmarshal` or `json.NewDecoder(bytes.NewReader(body))` from that bounded buffer. Do not decode from `resp.Body` on any successful path.

No plan changes are required before coding Step 2.

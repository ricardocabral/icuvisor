# Review R002 — Plan review for Step 1

Verdict: **approved**

The revised Step 1 plan addresses the blocking issues from R001. In particular, it now targets a retriable event write path, instruments response bodies on the client side, and checks that retry response bodies are closed before the next attempt starts instead of only checking the final close count after the method returns.

## Notes for implementation

- Use an update path (`PUT`), either by calling `AddOrUpdateEvent` with a non-empty `EventID` plus required writable fields, or by calling `doJSONBody` directly from the package test. A create path (`POST`) still will not exercise status retries.
- Set retry config explicitly for the 503, 503, 503, 200 sequence, e.g. `RetryConfig{MaxAttempts: 4, BaseDelay: time.Nanosecond, MaxDelay: time.Nanosecond}` with zero jitter.
- The key assertion should run inside the custom `RoundTripper`: on attempt N > 1, verify the body returned for attempt N-1 is already closed before returning the next response.
- Keep the final assertions for total attempts and total close count, but treat them as secondary; they do not by themselves reproduce the deferred-close leak.
- If the test already passes in this checkout because `events.go` currently drains/closes non-2xx retry bodies inline, record that in `STATUS.md` rather than weakening the test.

This is a suitable regression test shape for the leak described by TP-056.

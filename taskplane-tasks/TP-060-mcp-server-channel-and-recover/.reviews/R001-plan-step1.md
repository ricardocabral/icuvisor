# Plan Review — TP-060 Step 1

Verdict: **Request changes**

The current Step 1 plan restates the checklist, but it does not resolve an important behavioral mismatch in the requirement.

## Findings

1. **Buffering alone does not prove or fix “worker returns without sending”.**
   Changing `closed := make(chan error)` to `make(chan error, 1)` prevents a sender from blocking if the receiver is no longer waiting. It does **not** unblock the parent if the worker goroutine exits without sending; `<-closed` will still block forever unless the channel is also closed or another signal is sent. The proposed regression test (“inner goroutine returns without sending”) cannot be satisfied by a buffer-only change.

2. **The plan needs an explicit close-channel strategy if that regression test is kept.**
   If the intended invariant is “parent never hangs when the worker exits before sending,” the implementation plan should include something like `defer close(closed)` inside the worker goroutine, with care that there is only one sender. A buffered channel plus deferred close would let the cancel path’s `<-closed` complete even if a future early return happens before the send.

3. **The test seam is unspecified.**
   `Run` currently obtains a concrete SDK session from `s.server.Connect`, so simulating “worker returns without sending” is not straightforward through the public API. The plan should name a minimal testable seam before implementation, for example a very small unexported helper around the close-wait select, or another focused approach that avoids a broad transport refactor.

## Recommended revision

Update Step 1’s plan to distinguish the two guarantees:

- Keep the required defensive buffer: `closed := make(chan error, 1)`.
- If testing an early return without send, also ensure the worker closes the channel on exit, e.g. a deferred `close(closed)` in the goroutine, and add a focused regression test through a small private helper/seam.
- Use a short timeout only as a test guard, not as the primary synchronization mechanism.

Without this clarification, an implementation may satisfy the literal one-line channel change while leaving the planned regression either impossible to write or not representative of the stated failure mode.

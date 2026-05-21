# Code Review — TP-045 Step 1

**Verdict:** Approved.

No blocking findings. The only code change for this step is the `STATUS.md` planning update, and it now records the Step 1 decisions needed before implementation:

- concrete helper boundaries for `do`, `readBody`, `shouldRetry`, and the outer retry loop;
- a 32 MiB bounded-response cap with `ErrResponseTooLarge`;
- preservation of the existing retry policy and zero-config/partial-config jitter semantics;
- explicit body lifecycle ownership with no `defer resp.Body.Close()` inside the retry loop.

The Step 2 implementation should keep the two cautions from R002 in mind: lock down success close-error precedence in tests/notes, and ensure oversize/malformed-body early returns still close the response body in the same loop iteration.

# Plan Review — Step 1: Statistical test

Result: **approved with minor guardrails**.

The step is appropriately scoped to `internal/intervals/client_test.go` and should be done before the production RNG change so the red/green signal is meaningful.

## Guardrails for the test

- Call the package-local `addJitter` helper directly from `client_test.go`; no HTTP/client retry path is needed for this step.
- Use a delay/ratio with a large enough sample space. For example, `delay := time.Second` and `ratio := defaultJitter` gives a ±200ms nanosecond span, so 1000 samples from the future RNG should almost never collide. Avoid tiny delays such as `time.Nanosecond` or `time.Microsecond`, where the available unique values may be smaller than the `>= 900` threshold or dominated by timer resolution.
- Keep the sampling loop tight and do not insert sleeps; otherwise the current `time.Now().UnixNano()%...` implementation may look artificially good and the required failure-on-main check loses value.
- Count unique returned durations from 1000 calls and fail if `len(unique) < 900`. Consider also asserting every sampled duration stays within `[delay-span, delay+span]` to protect the existing jitter span/formula while this area is being touched.
- Record the failing `go test` invocation/output in `STATUS.md` before Step 2. If the test unexpectedly passes on the current implementation on any platform, tighten the test or reassess before changing production code.

No blockers found for proceeding with Step 1.

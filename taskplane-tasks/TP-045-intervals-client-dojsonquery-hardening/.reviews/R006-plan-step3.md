# Plan Review — TP-045 Step 3

**Verdict:** Approved, with one documentation requirement.

The Step 3 plan is narrowly scoped and addresses the remaining audit item: replacing `normalizeRetryConfig` with an explicit `RetryConfig.WithDefaults` path, updating call sites, and removing the whole-struct zero-value comparison. The recorded semantics are also correct: compute an explicit "all fields unset" boolean before applying defaults so `RetryConfig{}` still gets the default jitter, while partial configs with `Jitter == 0` keep no jitter.

## Required implementation notes

1. **Add an exported-method doc note.**
   `RetryConfig.WithDefaults` will be exported, and this repo has `revive`'s exported-note rule enabled. Add a note starting with `WithDefaults`, e.g. `// WithDefaults returns c with retry defaults applied...`.

2. **Preserve the exact jitter semantics.**  
   Capture `allFieldsUnset := c.MaxAttempts == 0 && c.BaseDelay == 0 && c.MaxDelay == 0 && c.Jitter == 0` before mutating the config. Do not use `c == (RetryConfig{})`, and do not default jitter for partial configs such as `RetryConfig{MaxAttempts: 1}`.

3. **Update direct test/helper uses of the old function.**  
   Besides `NewClient`, `internal/intervals/client_test.go` currently constructs `&Client{retry: normalizeRetryConfig(...)}` directly. Replace that with `RetryConfig{...}.WithDefaults()` or route through `NewClient`; otherwise Step 3 will leave a stale call site or fail to compile.

4. **Keep validation behavior unchanged.**  
   Retain the current normalization rules: non-positive attempts/base/max delays become defaults, negative jitter is clamped to `0`, and only the fully zero config receives `defaultJitter`.

No additional design work is needed before implementing Step 3.

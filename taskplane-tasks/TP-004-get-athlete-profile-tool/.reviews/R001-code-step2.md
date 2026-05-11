# R001 Code Review — Step 2: Implement the typed tool

## Verdict

**Needs changes.** The tool skeleton, registry hook, strict unknown-field rejection, sanitized user errors, and version propagation are mostly in place. I found a few contract/process issues to fix before approving this step.

## Findings

### 1. `arguments: null` bypasses the declared object-only argument contract

**Severity:** Medium  
**File:** `internal/tools/get_athlete_profile.go:125-138`

The input schema declares `type: object` and `additionalProperties: false`, and the Step 2 plan explicitly calls for strict runtime argument validation. However, decoding JSON `null` into `GetAthleteProfileRequest` succeeds in Go and returns the zero-value request. That means a non-object argument is accepted even though only an object with optional `include_full` should be valid.

This matters because the runtime validation is the safety net for forbidden fields/credential mistakes when clients do not honor schema perfectly.

Suggested fix: reject `null` and other non-object JSON before or during decode. For example, after trimming whitespace, accept empty as `{}` for omitted arguments, but require the first byte to be `{` for non-empty input; then keep `DisallowUnknownFields()` and trailing-token rejection.

Add/leave a test in Step 4 for rejected `null`, unknown fields such as `api_key`, and valid `{}` / omitted arguments.

### 2. Mid-call context cancellation is converted into a credential/athlete-ID user error

**Severity:** Medium  
**File:** `internal/tools/get_athlete_profile.go:109-112`

The handler checks `ctx.Err()` before the client call, but if the request context is cancelled while `client.GetAthleteProfile(ctx)` is in flight, the returned cancellation/deadline error is wrapped as:

```text
could not fetch athlete profile; check intervals.icu credentials and athlete ID
```

That is inaccurate and weakens the repository rule to honor cancellation. It can also mislead MCP clients/log readers into thinking credentials are bad when the operation was simply cancelled.

Suggested fix: after the client returns an error, check cancellation before mapping to the public fetch error, e.g. return `ctx.Err()` when non-nil (or use `errors.Is(err, context.Canceled/context.DeadlineExceeded)` if needed), and only use `NewUserError(fetchAthleteProfileMessage, err)` for real fetch failures.

### 3. `Register` can panic on a nil registrar

**Severity:** Low/Medium  
**File:** `internal/tools/registry.go:25-32`

`defaultRegistry.Register` validates the context and profile client, but then calls `registrar.AddTool(...)` without checking whether `registrar` is nil. Since this is a public registry implementation outside `main`, passing a nil registrar can panic, which conflicts with the repository rule to return errors rather than panic outside `main`.

Suggested fix: add a nil guard before the call, returning a short registration error such as `registering get_athlete_profile: missing registrar`.

## Checks run

- `git diff 6a79a13..HEAD --name-only`
- `git diff 6a79a13..HEAD`
- `go test ./...` — passed
- `gofmt -l internal/tools/get_athlete_profile.go internal/tools/registry.go` — no output
- `go vet ./...` — passed

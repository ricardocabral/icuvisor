# Plan Review — TP-068 Step 3

## Verdict

Approved with guardrails. Extracting only the current `terminalPrompter` behavior into `internal/cli/prompt` is the right-sized Step 3 and should be low-risk if it stays mechanical.

## Plan notes / guardrails

- Keep this step limited to the prompter extraction. Do not move setup arg parsing yet, and do not further reshape the setup flow unless the import/type wiring requires it.
- Put the reusable contract in the new package, but keep it minimal: only the three behaviors setup currently uses (`Confirm`, `ReadLine`, `ReadSecret`) unless you intentionally choose equivalent names and update all call sites. Avoid adding generic prompt features.
- Prefer preserving the existing method semantics and byte-level output:
  - `Confirm` writes `prompt + " "`, accepts `y/yes/n/no`, returns the default on blank input, and repeats with `Please answer y or n.` for invalid input.
  - `ReadLine` writes the prompt line, then `> `, and returns trimmed input.
  - `ReadSecret` writes the prompt line, then `> `, uses masked input only for `*os.File`, prints the trailing newline, and keeps the existing non-terminal error behavior.
- Be deliberate about type ownership. Good options are either making `app.SetupPrompter` a type alias to `prompt.Prompter` or changing `SetupOptions.Prompter`/`Options.SetupPrompter` directly to `prompt.Prompter`. Avoid leaving a duplicate app-owned interface unless there is a clear compatibility reason.
- Ensure the default constructors are wired consistently in both default paths:
  - `runSetupCommand` should create the terminal prompter from `opts.Stdin` and command stdout.
  - `RunSetup` defaults should still use stdin when no prompter is injected and should write prompts to setup stdout.
- Exported identifiers in `internal/cli/prompt` need proper Go doc notes, and the package should remain dependency-light; no third-party prompter libraries.
- Keep tests table-driven in the new package. At minimum cover yes/no/default handling, invalid-answer retry text, line prompt formatting/trimming, context cancellation before read, and the non-`*os.File` `ReadSecret` error path. The existing `internal/app` tests should continue to lock setup behavior.

## Checks expected after implementation

- `go test ./internal/cli/prompt ./internal/app`
- `go test ./...` is a useful extra because CLI dispatch also touches setup wiring.

## Blocking findings

None.

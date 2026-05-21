# TP-068-split-setup-and-extract-prompter — Status

**Current Step:** Step 5: Verify
**Status:** ✅ Complete
**Last Updated:** 2026-05-17
**Review Level:** 2
**Review Counter:** 9
**Iteration:** 1
**Size:** M

---

### Step 1: Capture flow regression-gate

**Status:** ✅ Complete

- [x] Verify existing setup coverage and add a focused regression gate for happy-path prompts/output plus invalid-key and keychain-error paths. Verified with `go test ./internal/app`.

### Step 2: Slice `RunSetup`

**Status:** ✅ Complete

- [x] Break `RunSetup` into focused setup-flow helpers without changing prompt/output behavior. Verified RunSetup regression tests with `go test ./internal/app -run 'TestRunSetup'`.
- [x] Run targeted setup tests after the slice. `go test ./internal/app` passed.

### Step 3: Extract `cli/prompt`

**Status:** ✅ Complete

- [x] Create `internal/cli/prompt/` with the minimal reusable prompter interface and terminal implementation. Added table-driven prompt tests; `go test ./internal/cli/prompt` passed.
- [x] Update setup flow/cmd wiring to use the new prompt package without changing prompts. Verified app setup/CLI tests with `go test ./internal/app -run 'TestRunSetup|TestRunCLI|TestSetup'`.
- [x] Run targeted tests for app setup and prompt package. `go test ./internal/app ./internal/cli/prompt` passed.

### Step 4: Split arg parsing

**Status:** ✅ Complete

- [x] Move setup argument parsing and command dispatch into `internal/app/setup_cmd.go` without behavior changes. Verified focused app tests with `go test ./internal/app -run 'TestRunSetup|TestRunCLI|TestParse|TestSetup'`.
- [x] Run targeted app tests after the split. `go test ./internal/app` passed.

### Step 5: Verify

**Status:** ✅ Complete

- [x] Update `CHANGELOG.md` `[Unreleased]` with the internal setup refactor note.
- [x] Run `make build`, `make test`, `make test-race`, and `make lint`. All passed; reran after splitting setup tests.
- [x] Manual smoke: `bin/icuvisor setup --help` works and setup golden path is documented/verified. Reran `bin/icuvisor setup --help`; golden path verified with clean fake keychain via `go test ./internal/app -run TestRunSetupWritesConfigAndVerifiesKeychainRoundTrip -count=1` (no real intervals.icu API key available for live keychain smoke).
- [x] Confirm file sizes with `wc -l internal/app/setup*.go internal/cli/prompt/*.go`. Largest file is `internal/app/setup_test.go` at 294 LOC; source files are 236 LOC or less.

| 2026-05-17 22:17 | Task started | Runtime V2 lane-runner execution |
| 2026-05-17 22:17 | Step 1 started | Capture flow regression-gate |
| 2026-05-17 22:39 | Review R002 | code Step 1: UNKNOWN |
| 2026-05-17 22:41 | Review R003 | code Step 1: UNKNOWN |
| 2026-05-17 22:42 | Review R004 | plan Step 2: APPROVE |
| 2026-05-17 22:46 | Review R005 | code Step 2: APPROVE |
| 2026-05-17 22:48 | Review R006 | plan Step 3: APPROVE |
| 2026-05-17 22:52 | Review R007 | code Step 3: APPROVE |
| 2026-05-17 22:56 | Review R008 | plan Step 4: APPROVE |
| 2026-05-17 22:58 | Review R009 | code Step 4: APPROVE |

| 2026-05-17 23:03 | Worker iter 1 | done in 2753s, tools: 112 |
| 2026-05-17 23:03 | Task complete | .DONE created |
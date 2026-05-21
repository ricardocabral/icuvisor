# R001 Plan Review — Step 1: Backend selection and contract

## Verdict: needs changes before implementation

`PROMPT.md` is clear, but the current Step 1 plan in `STATUS.md` is still only a checklist (`Library: TBD`) rather than a concrete backend/contract decision. For a Level 3 security task, please lock down the exact contract and dependency facts in `STATUS.md` before writing code.

## Required changes

1. **Record the actual backend selection and evidence.**
   - Prefer `the selected OS keyring module` unless a blocking issue is found.
   - Record the module/version and release date. Current module data shows `v0.2.8`, tagged `2026-03-23T12:00:09Z`.
   - Record license review, including transitive deps: `zalando/go-keyring` MIT, `the Windows credential helper module` MIT, `the D-Bus module` BSD-style, `golang.org/x/sys` BSD-style.
   - Record why `99designs/keyring` is not selected: broader surface area and encrypted-file/pass-style fallback risk conflicts with “no key on disk by default.”
   - Record that the selected backend does not require CGO for the target paths.

2. **Make the `Store` interface precise.**
   The prompt shorthand `Store { Get/Set/Delete(account string) (string, error) }` is not a valid/usable contract for `Set`/`Delete`. Define exact signatures before implementation, for example:
   ```go
   type Store interface {
       Get(ctx context.Context, account string) (string, error)
       Set(ctx context.Context, account, secret string) error
       Delete(ctx context.Context, account string) error
   }
   ```
   Including `context.Context` better matches the repo convention that blocking/I/O functions accept `ctx`; the underlying library may not be cancellable, but the wrapper can at least honor cancellation before invoking it.

3. **Define canonical constants, not call-site strings.**
   Step 1 should name the constants that future steps will use, e.g.:
   - `ServiceName = "icuvisor"`
   - `IntervalsAPIKeyAccount = "intervals-icu-api-key"`

   Also state explicitly that there is no per-athlete account split in this task and that coach-mode key layout is deferred to TP-039.

4. **Define error mapping at the wrapper boundary.**
   Add a project-local sentinel such as `credstore.ErrNotFound`, and require all callers to use `errors.Is`.
   - Map `keyring.ErrNotFound` to `credstore.ErrNotFound`.
   - For startup `Get` on Linux when no D-Bus/session keyring is reachable, map the unavailable-backend condition to `credstore.ErrNotFound` so config falls through to env/file as required.
   - Do **not** lose detail for other failures; wrap with `%w` and never include the secret value.
   - Decide whether `Set`/`Delete` should also degrade to `ErrNotFound` on unavailable Linux keyring, or whether they should return an actionable wrapped error. This matters for TP-038 onboarding.

5. **Decide how `.env` participates in credential precedence.**
   The Step 3 prompt lists `explicit INTERVALS_ICU_API_KEY env var > keychain > JSON api_key`, but the current loader also supports a local `.env` file. Step 1 should settle this now so implementation/tests are unambiguous. Recommended: process environment remains highest priority; keychain should beat plaintext `.env`/JSON fallback unless the task explicitly treats `.env` as “explicit env.” Either way, document it and add tests later.

6. **Define the source indicator shape.**
   If `Config.String()` will print a source (`env|keychain|file`), define the field/enum in Step 1, ensure it is `json:"-"`, and decide whether `.env` gets its own source value or is folded into `file`/`env`.

## Non-blocking suggestions

- Keep the default constructor named clearly, e.g. `credstore.OSKeychain() Store`, but make config tests inject `NoopStore`/in-memory stores so they never hit a live keychain.
- Avoid exposing `DeleteAll` from `go-keyring`; this task only needs single-account `Delete` and exposing broader deletion surface is unnecessary.
- Wiring likely belongs in `internal/app`/`config.Options` rather than `cmd/icuvisor/main.go`, since `main` is intentionally thin in this repo.

Once the above decisions are added to `STATUS.md`, the implementation plan for Step 1 is sound and can proceed to code.

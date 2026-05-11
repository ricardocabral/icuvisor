# Plan Review — Step 4: Align code UX with docs if necessary

Verdict: **Approved with minor guidance**

The Step 4 plan matches the prompt: it focuses on only small UX/error-message alignment if docs expose confusing behavior, keeps invalid config failures short/actionable, updates the README pointer, and records the documentation work in `CHANGELOG.md`.

Guidance for implementation:

1. Keep code changes conditional and narrow. The current config/app errors are already mostly actionable and have redaction tests; only change `cmd/icuvisor`, `internal/app`, or `internal/config` if there is a concrete mismatch with the new guide.
2. If any user-facing error text changes, update or add table-driven tests. In particular, preserve the existing guarantees that API keys and raw athlete IDs are not leaked in errors or `Config.String()`.
3. Fix the README quickstart pointer now: it currently says detailed client setup will live in `docs/clients/` later and links to `docs/`. It should point directly to the v0.1 Claude Desktop guide, e.g. `docs/clients/claude-desktop.md`.
4. Add a concise `CHANGELOG.md` entry under `[Unreleased]`, likely in `Added`, for the macOS Claude Desktop manual setup guide and smoke checklist.
5. Do not expand scope into installers, keychain storage, onboarding UI, extra clients, or config behavior changes. The docs use `--config` plus env/JSON inputs that the existing CLI already supports.

No blocking changes are needed before proceeding.

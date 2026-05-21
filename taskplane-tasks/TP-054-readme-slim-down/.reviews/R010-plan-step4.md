# Plan Review R010 — Step 4: Delete migrated `docs/*.md` files

## Verdict

Approved.

The Step 4 plan matches the task prompt: remove only the five migrated end-user docs, remove `docs/install/` only if it becomes empty, and keep `docs/clients/` because `docs/clients/codex-local.md` is still developer content.

## Verification

I reviewed:

- `taskplane-tasks/TP-054-readme-slim-down/PROMPT.md`
- `taskplane-tasks/TP-054-readme-slim-down/STATUS.md`
- current `docs/install/` and `docs/clients/` contents
- current scoped/broader greps for deleted-doc paths outside `taskplane-tasks/**`

Current repository state is compatible with the plan:

- `docs/install/` contains only `macos.md`, so it should be removed after deleting that file.
- `docs/clients/` contains `claude-code.md`, `claude-desktop.md`, and `codex-local.md`; delete only the first two and keep the directory with `codex-local.md`.
- The remaining product grep hits are inside files scheduled for Step 4 deletion, plus the known false-positive link from `docs/coach-mode.md` to the kept `docs/threat-models/coach-mode.md`, which will disappear when `docs/coach-mode.md` is deleted.

## Implementation reminders

- Use a narrow removal command, e.g. delete exactly:
  - `docs/install/macos.md`
  - `docs/clients/claude-desktop.md`
  - `docs/clients/claude-code.md`
  - `docs/coach-mode.md`
  - `docs/post-update.md`
- Do not delete `docs/clients/codex-local.md` or `docs/clients/`.
- Remove `docs/install/` after the file deletion if `rmdir docs/install` succeeds.
- Update `STATUS.md` to mark Step 4 complete and record the directory-retention decision.
- Leave final whole-product greps, builds, tests, lint, Hugo build, and changelog verification for Step 5 as planned.

No blocking changes are needed before implementation.

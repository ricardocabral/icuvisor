# Code Review R011 — Step 4: Delete migrated `docs/*.md` files

## Verdict

Approved.

## Findings

No blocking findings.

## Verification

Reviewed the Step 4 diff from `a218348..HEAD`, including:

- deleted `docs/install/macos.md`
- deleted `docs/clients/claude-desktop.md`
- deleted `docs/clients/claude-code.md`
- deleted `docs/coach-mode.md`
- deleted `docs/post-update.md`
- `taskplane-tasks/TP-054-readme-slim-down/STATUS.md`
- prior review artifact `R010-plan-step4.md`

Checks performed:

- `git diff a218348..HEAD --name-only`
- `git diff a218348..HEAD`
- confirmed `docs/install/` is absent after deleting its only file
- confirmed `docs/clients/` remains and still contains `docs/clients/codex-local.md`
- confirmed the exact five migrated docs requested by Step 4 were removed
- confirmed no product-scope inbound links remain to the deleted doc paths with:
  ```bash
  git grep -nE 'docs/(install/macos|clients/claude-desktop|clients/claude-code|coach-mode|post-update)\.md' -- . ':!taskplane-tasks/**' || true
  git grep -nE '(install/macos|claude-desktop|claude-code|coach-mode|post-update)\.md' -- . ':!taskplane-tasks/**' || true
  ```

The deletion scope matches the task prompt. Step 5 final verification (`make build`, `make test`, `make lint`, Hugo build, and final recorded sweeps/changelog confirmation) remains pending as expected.

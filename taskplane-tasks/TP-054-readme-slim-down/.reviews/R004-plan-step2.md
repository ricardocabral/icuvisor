# Plan Review R004 — Step 2: Inbound link sweep

## Verdict

Needs changes before Step 2 is implemented.

The plan has the right high-level command, but it is still too implicit for the current repository state. A dry run of the requested grep shows hits outside README/docs prose, including a user-facing CLI message and tests, plus historical `taskplane-tasks/**` artifacts that will make a literal whole-repo zero-hit check impossible unless the scope is clarified.

## Findings

### 1. The plan must enumerate hits and actions before editing

Run the sweep first and add a small table to `STATUS.md` with at least: `path`, `current context`, `action`, and `replacement URL / deletion rationale`. The current grep finds product-code/docs hits in:

- `README.md` — will be handled by the Step 3 rewrite.
- `SECURITY.md` — the `docs/install/macos.md` operator-checklist pointer is redundant because the checklist is already duplicated in `SECURITY.md`; do not replace this with the public macOS install page.
- `docs/dogfood/v0.2-findings.md` — update the Claude Desktop pointer to the website connection guide.
- `docs/internal-beta/onboarding-playbook.md` — update the coach-mode pointer to the website guide/explanation as appropriate.
- `internal/app/setup.go` and `internal/app/setup_test.go` — update the setup output and its assertion from `docs/clients/claude-desktop.md` to the new Claude Desktop URL.
- `docs/install/macos.md` internal links — record these as resolved by Step 4 deletion rather than spending time rewriting a file that will be removed.

This matters because Step 2 is not just markdown cleanup; it changes user-visible CLI output and its test expectations.

### 2. Resolve the `taskplane-tasks/**` false-positive problem explicitly

The exact grep also matches many historical task prompts, statuses, and review files, including this task's own `PROMPT.md`/`STATUS.md`. Those are not live inbound product links, and editing historical task artifacts would be noisy and misleading. However, the prompt's Step 5 says to get zero hits “across the whole repo,” which conflicts with the checked-in task artifacts.

Before implementation, update the plan/status with one of these choices:

- use a product-link sweep that excludes task artifacts, for example:
  ```bash
  git grep -nE 'docs/(install/macos|clients/claude-desktop|clients/claude-code|coach-mode|post-update)\.md' -- . ':!taskplane-tasks/**'
  ```
  and record that `taskplane-tasks/**` historical hits are intentionally out of scope; or
- ask the supervisor whether they really want historical task artifacts rewritten.

Do not silently chase old task prompt/review hits as part of the docs cleanup.

### 3. Replacement choices should be context-sensitive

Do not mechanically replace every deleted path with the same site URL. Use the Step 1 destination table:

- Claude Desktop → `https://icuvisor.app/connect/claude-desktop/`
- Claude Code → `https://icuvisor.app/connect/claude-code/`
- Coach-mode setup/how-to → `https://icuvisor.app/guides/coach-mode/`
- Coach-mode conceptual/security/cache caveat → include `https://icuvisor.app/explain/coach-mode/` if the context needs it
- Post-upgrade/schema-change guidance → `https://icuvisor.app/guides/after-upgrade/`
- macOS user install/Gatekeeper/uninstall → `https://icuvisor.app/install/macos/`
- maintainer release-operator checklist → remove the stale link or point to an in-repo developer source that actually contains the checklist; do not point to the public install page if that content was stripped during TP-052.

## Recommended revised Step 2 plan

1. Run the product sweep with the agreed task-artifact scope and paste the output into `STATUS.md`.
2. Add a hit/action table to `STATUS.md`.
3. Update non-deleted product files (`SECURITY.md`, `docs/dogfood/v0.2-findings.md`, `docs/internal-beta/onboarding-playbook.md`, `internal/app/setup.go`, `internal/app/setup_test.go`) with context-appropriate replacements/removals.
4. Mark hits inside files scheduled for Step 3 rewrite or Step 4 deletion as deferred to those steps.
5. Re-run the same scoped grep and record remaining hits with their planned resolving step.

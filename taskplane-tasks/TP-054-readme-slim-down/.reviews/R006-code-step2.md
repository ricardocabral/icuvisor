# Code Review R006 — Step 2: Inbound link sweep

## Verdict

Revise.

## Findings

1. **Inbound link sweep missed relative links to docs scheduled for deletion.**  
   Locations:
   - `docs/internal-beta/README.md:18`
   - `docs/internal-beta/onboarding-playbook.md:8`
   - `docs/internal-beta/onboarding-playbook.md:19`
   - `docs/internal-beta/onboarding-playbook.md:20`
   - `docs/internal-beta/onboarding-playbook.md:30`

   These links point at `../install/macos.md`, `../clients/claude-desktop.md`, `../clients/claude-code.md`, and `../coach-mode.md`, all of which TP-054 plans to delete in Step 4. The Step 2 grep only matched literal `docs/...` paths, so these live product docs were not captured in the hit/action table and would become broken links after deletion. Please update them to the corresponding `https://icuvisor.app/...` destinations (or remove redundant references) and expand the recorded sweep to include relative/bare filename forms, not only `docs/...` literals.

## Notes

- The replacements that were made in `SECURITY.md`, `docs/dogfood/v0.2-findings.md`, `internal/app/setup.go`, and `internal/app/setup_test.go` look appropriate.
- `go test ./internal/app` passes.
- Reproduction command for the missed links:
  ```bash
  git grep -nE '(install/macos|claude-desktop|claude-code|coach-mode|post-update)\.md' -- . ':!taskplane-tasks/**'
  ```

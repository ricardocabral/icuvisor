# Code Review: Step 4 — Exercise every registered MCP tool through Codex prompts

## Verdict

Request changes.

## Findings

### 1. Step 4 checklist marks real intervals.icu validation complete even though it is documented as blocked

- Location: `taskplane-tasks/TP-006-codex-local-mcp-validation/STATUS.md:58` and `STATUS.md:83`
- Severity: Medium

`STATUS.md` checks off “Validate real intervals.icu-backed reads without recording raw personal data,” but the same update records that real intervals.icu-backed `get_athlete_profile` validation “cannot be completed” because `.env` and both required credential variables are unavailable. That makes the Step 4 status internally inconsistent and could incorrectly signal that the live-data acceptance criterion was satisfied.

Please leave that checklist item unchecked, or mark it explicitly as blocked/N/A, while keeping the existing blocker table entry. The tool-dispatch/terse-shape validation through Codex can remain recorded as a pass, but the real intervals.icu read should not be represented as completed until credentials are available and a real-data call succeeds.

## Verification performed

- Ran `git diff 5ce314848a6dcef5bc8778e434440c7ff0d91160..HEAD --name-only`.
- Reviewed the full diff with `git diff 5ce314848a6dcef5bc8778e434440c7ff0d91160..HEAD`.
- Read `PROMPT.md` and the updated `STATUS.md` for task requirements and Step 4 context.
- Read `internal/tools/registry.go` and confirmed the source registry currently exposes only `get_athlete_profile`, matching the Step 4 catalog claim.
- Ran `git diff --check 5ce314848a6dcef5bc8778e434440c7ff0d91160..HEAD`; no whitespace errors were reported.
- Checked the tracked task markdown for direct `INTERVALS_ICU_*=` credential assignments; none were introduced.

## Notes

- No application code changed in this step.
- The redacted Step 4 result table is otherwise aligned with the approved guardrails: it distinguishes Codex MCP dispatch/shape validation from the blocked real intervals.icu validation and does not include raw profile values or secrets.

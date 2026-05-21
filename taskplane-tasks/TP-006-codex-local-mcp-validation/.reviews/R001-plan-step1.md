# Plan Review: Step 1 — Discover current server and Codex CLI behavior

## Verdict

Approved with required clarifications before moving to Step 2.

The Step 1 checklist in `STATUS.md` matches the prompt's scope and is safe: build the binary, identify the launch path, inspect Codex help, discover configuration options, and document a validation plan before touching configuration. However, `STATUS.md` currently only restates the task checklist; it does not yet contain the concrete validation plan that Step 1 is expected to produce.

## Required before Step 1 is marked complete

1. Add a concrete, redacted validation plan to `STATUS.md` after discovery and before any Codex config changes. It should include:
   - the absolute `bin/icuvisor` path selected for MCP stdio launch;
   - the Codex CLI commands/help topics inspected;
   - the discovered MCP configuration mechanism;
   - whether a temporary config/profile is available and how it will be used;
   - fallback behavior if persistent Codex config must be touched, including backup/restore expectations.
2. Keep Step 1 strictly discovery-only with respect to Codex configuration:
   - do not edit persistent Codex config in this step;
   - do not read `.env` yet, except in Step 2 as specified;
   - do not launch an interactive validation session yet unless needed only for non-mutating help output and clearly documented.
3. Record only non-sensitive findings. Paths and command names are fine; API keys, athlete IDs, and raw personal data must not be printed or copied into `STATUS.md`.

## Suggested Step 1 execution details

- Run `make build` and record pass/fail plus the selected absolute path, for example using `pwd`/`realpath` for `bin/icuvisor`.
- Inspect Codex help non-mutatively, starting with `/Users/jusbrasil/Library/pnpm/codex --help` and then only relevant subcommand help shown by that output, such as MCP/config help if available.
- Prefer temporary or isolated config mechanisms discovered from help output over modifying user-level Codex settings.
- If help output is ambiguous, document the ambiguity as a blocker or question in `STATUS.md` instead of guessing and changing config.

## Notes

No issue with the planned file scope for Step 1. Application code should remain untouched during this discovery step unless the build itself reveals a concrete defect, in which case that should be called out separately before expanding scope.

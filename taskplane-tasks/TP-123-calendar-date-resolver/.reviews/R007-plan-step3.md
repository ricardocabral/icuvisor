# Plan Review — Step 3

Verdict: CHANGES REQUESTED

## Findings

### P1 — Include generated tool catalog sync before `make eval-validate`
Step 3 needs eval scenarios to activate `resolve_calendar_dates`, but `make eval-validate` validates tool names against `web/data/tools.json`. The current working tree's `web/data/tools.json` does not contain `resolve_calendar_dates`, so any scenario that adds it to `expected_tools` or `bonus_tools` will fail validation unless the plan explicitly runs the generated docs/catalog step first (for example `make docs-tools`) and commits the generated catalog/reference outputs that are part of the public tool surface.

Please hydrate the Step 3 plan to include this ordering: update activation scenarios, regenerate the web tool catalog/reference docs as needed, then run `make eval-validate`.

### P1 — Pin the eval activation coverage, not just prompt wording
The current Step 3 checklist says to update prompts that mention future weeks or “tomorrow”, but it does not specify which scenarios must exercise the new deterministic anchor or how activation will be scored. The plan should name the affected cookbook scenarios and expected tool changes, at minimum the existing relative-future prompts:

- `CB-PLAN-02` (`12 weeks from today`)
- `CB-TAPER-01` (`9 days from today`)
- the new known-bad weekday/date pairing scenario

For those, make `resolve_calendar_dates` an `expected_tools` entry where deterministic anchoring is required, not only a prose hint or `bonus_tools`, and add `must_address` / `anti_patterns` that require using returned `date` + `weekday` and forbid UTC/client-time/model arithmetic.

### P2 — Make the user guidance name the deterministic tool and offsets
The Claude Project instruction update should explicitly tell the assistant to call `resolve_calendar_dates` before answering date-sensitive planning prompts such as tomorrow, next week, N days from today, or user-supplied weekday/date pairings, and to use athlete-local offsets from that result. Merely saying to use timezone metadata/as_of fields still leaves the model doing the exact date arithmetic this task is intended to avoid.

## Verification

- Read `PROMPT.md` and `STATUS.md`.
- Reviewed existing eval scenario structure and `scripts/eval/run_eval.py --validate` behavior.
- Spot-checked current `web/data/tools.json` and confirmed it is not yet in sync with the new public `resolve_calendar_dates` tool.
- No tests run; this was a plan review.

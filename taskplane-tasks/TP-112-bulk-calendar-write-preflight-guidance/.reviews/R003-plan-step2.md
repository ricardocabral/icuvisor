# R003 plan review — Step 2

Verdict: **APPROVE**

The Step 2 scope is appropriate for a docs-only follow-up to the Step 1 prompt guardrail. Update the cookbook guidance in user-facing language, then record the docs/prompt guardrail in `CHANGELOG.md`.

Approved implementation constraints:

- Update both cookbook pages that match the behavior: `web/content/cookbook/build-workouts.md` for workout/template writes and `web/content/cookbook/season-and-block-plan.md` for bulk calendar scheduling. Keep the wording concise.
- Describe the representative-write pattern without promising a new generic warnings contract: validate/preview one representative structured payload, write one event/template, read it back, inspect current validation diagnostics and existing `_meta.workout_doc_warning` when present plus `workout_doc_summary`/stored description, then continue with the rest.
- Explain the preservation risk in plain language: write tools replace the upstream description/DSL field, so structured steps can be lost if an update sends only prose or omits the intended `workout_doc`.
- Scope the parallel-write caution narrowly to bulk workout/calendar writes when schema wording, warning metadata, or description/`workout_doc` preservation semantics are uncertain. Do not ban all parallel writes.
- Keep examples client-neutral and do not include API keys or private payloads.
- Add a concise `[Unreleased]` changelog entry, likely under `Changed`, for safer bulk workout/calendar write preflight guidance.

No additional changes are needed to `web/content/reference/resources-prompts.md` or `docs/dogfood/v0.3-prompts.md` unless the implementation changes the prompt summary or dogfood instructions beyond the cookbook wording.

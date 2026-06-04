# Plan Review: TP-151 Step 4 — Refresh schemas, routing, and docs

**Verdict: APPROVE**

The revised Step 4 plan addresses the prior R011 gaps. It now names the two affected cookbook pages and the specific external-id contract each must document, records the expected schema-snapshot outcome (`add_or_update_event` changes, `apply_training_plan` unchanged), and pins the `[Unreleased]` changelog entry to both public behaviors: manual `add_or_update_event.external_id` support and deterministic `apply_training_plan` external IDs/retry protection.

No blocking plan issues remain.

## Implementation reminders

- Keep the docs conservative: best-effort upstream idempotency, same-day preflight/protection, no blank/null clear support, and stable caller namespaces that avoid provider-owned prefixes such as `strava-` or `hevy-`.
- Include readback visibility where it fits: event reads now expose upstream `external_id` as audit metadata when present, so docs/changelog wording should not imply the key is secret or hidden.
- Since the tool description/schema changed, record the generated-docs expectation. A temp `cmd/gendocs` run shows no current `web/data/tools.json` diff because the catalog summary first sentence is unchanged, but it is still worth noting this in `STATUS.md` so reviewers do not chase a missing docs-tools diff.
- Leave full `make test` / `make lint` / `make build` to Step 5 as planned.

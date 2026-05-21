# Review R001 — Plan review for Step 1

Decision: **Approved with minor execution guardrails**

I read `PROMPT.md`, `STATUS.md`, and the PRD anchors for `get_extended_metrics` in §7.2.C / §7.4 #4. The revised Step 1 plan now addresses the previously blocking planning gaps and is concrete enough to execute the black-box availability probe.

## What is now sufficient

- **Complete candidate inventory:** The plan enumerates the PRD target field set and proposes stable response keys for each metric, including the subjective fields, compliance, and device metadata.
- **Probe matrix:** The endpoint/query matrix is organized by metric group and covers the likely public API surfaces: activity list/detail, stream metadata/keys, analyzed interval/lap surfaces, fitness/summary, wellness, and event/activity pairing data.
- **Decision semantics:** The `yes`, `conditional`, `computed_not_allowed`, `not_observed`, and `no` states are clear and align with the PRD rule that icuvisor must drop fields that are not directly exposed upstream instead of deriving or zero-filling them.
- **Fixture/evidence standard:** The plan requires sanitized minimal fixtures, endpoint citations, JSON pointers, unit/scale notes, and probe dates for available fields. That is enough to make later tool-field decisions auditable.
- **Privacy/redaction:** The fixture redaction rules call out the important sensitive data classes: API keys, raw athlete/activity IDs, notes/free text, GPS traces, device serials, and unredacted probe output.
- **Representative samples:** The sample plan correctly recognizes that these metrics are conditional on sport/device/input data and asks for runs with dynamics, rides with L/R balance, subjective entries, paired events, and advanced-sensor examples where available.

## Required guardrails while executing

1. **Do not let missing credentials weaken the conclusion.** If live black-box probes cannot be run, do not mark fields as definitive `no` based only on lack of local credentials or lack of observed samples. Use `not_observed` with explicit limitations unless public API documentation conclusively documents absence and the checked surface is recorded.
2. **For every `yes`/`conditional`, include the exact fixture and JSON pointer.** The later implementation should only expose fields that can be traced back to those rows.
3. **Treat locally derivable metrics as excluded unless returned directly by intervals.icu.** This is especially important for decoupling, HR drift, IF/VI, TRIMP, load/strain variants, and zone distributions, which can be tempting to compute from streams but are out of scope for this task.
4. **Record final results in both `availability.md` and `STATUS.md`.** `availability.md` should be the durable evidence table; `STATUS.md` should summarize discoveries and any probe limitations.

## Suggested final `availability.md` shape

Use one row per PRD candidate metric with columns like:

- PRD candidate metric;
- proposed response key;
- availability status;
- endpoint/query variants checked;
- upstream JSON pointer/key;
- unit or scale;
- fixture path;
- probe date;
- notes/limitations.

With the guardrails above, the plan is ready to execute Step 1.

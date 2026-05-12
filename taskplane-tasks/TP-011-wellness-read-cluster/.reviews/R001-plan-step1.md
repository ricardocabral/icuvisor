# Plan Review: TP-011 Step 1 — Map the wellness payload

**Verdict: Approved to proceed.**

The revised `STATUS.md` now contains a concrete Step 1 plan rather than only the prompt checklist. It defines the mapping artifact schema, clean-room evidence boundaries, black-box fixture/probe strategy, provider/provenance decision rules, and custom-field handling. That is sufficient to begin the Step 1 discovery work.

## What is now adequate

- The mapping table schema has the right fields for downstream Step 2/3 work: upstream key/path, observed type/nullability, category, planned response path, provider, native scale, and evidence.
- Clean-room constraints are explicit, including the GPL exclusion and permitted evidence sources.
- The probe/fixture plan targets the important acceptance cases: Polar fresh/stale, Garmin body battery, Oura raw sleep score, manual-only, custom fields, and null-heavy rows.
- Provenance rules correctly avoid inventing source attribution and require `unknown`/`unknown` where evidence is insufficient.
- Custom fields are explicitly preserved and included in null-stripping rather than being dropped or misclassified.
- The plan keeps Step 1 discovery/documentation-focused and does not prematurely implement the decoder/tool.

## Follow-up items during Step 1 execution

These are not blockers, but they should be reflected in the eventual Step 1 findings before marking the step complete:

1. **Record the concrete evidence sources used.** Include the exact public API doc references and the relevant intervals.icu forum thread/post references from the prompt, plus whether any black-box probe was actually possible. If live probing is unavailable, record that as uncertainty/blocker instead of implying full coverage.

2. **Add the actual field catalog table.** The current plan defines the schema, but Step 1 is not complete until every observed/documented wellness key is listed with category, scale/unit, source, planned response path, and uncertainty.

3. **Explicitly enumerate subjective scales.** Make sure the Step 1 findings call out the full set needed later by Step 4: `feel`, `sleepQuality`, `fatigue`, `soreness`, `stress`, `mood`, `motivation`, and `injury`, with ranges or clearly marked uncertainty.

4. **Be careful with the stale fixture wording.** The plan says the stale Polar case has `fetched_at` more than 24h after the wellness date boundary. During execution, define the staleness comparison unambiguously in the findings so Step 3 tests do not invert the boundary or treat exactly 24h as stale.

5. **Avoid depending on a wellness typed client for probing.** Since the typed wellness client is not implemented until later, Step 1 probing should use raw/sanitized HTTP payloads or fixtures as evidence, then feed the resulting catalog into the typed decoder design.

## Conclusion

Proceed with Step 1. The next review should focus on whether the completed `STATUS.md` findings contain a real, evidence-backed wellness field catalog and clearly label any upstream gaps or synthetic assumptions.

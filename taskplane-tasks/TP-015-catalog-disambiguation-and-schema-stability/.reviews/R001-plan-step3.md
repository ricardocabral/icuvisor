# Plan Review — Step 3: Implement the CI schema-stability check

Decision: **APPROVE**

## Summary

The revised plan addresses the prior blocking concerns. It now separates the two necessary checks:

- snapshot freshness/canonicalization against the current tree's committed snapshots; and
- additive-only stability against a pre-PR baseline snapshot directory.

It also explicitly fails missing baseline tools, accepts genuinely new tool names only as additions with committed snapshots, and requires GitHub Actions annotations plus a Markdown step summary.

## What looks good

- The baseline/current comparison model is now capable of catching the important failure mode: a PR that removes or renames an argument and regenerates snapshots in the same change.
- Treating every baseline tool as required in the current registry closes the tool-rename loophole.
- Reusing shared snapshot/check logic between the generator, checker, and tests should reduce drift in canonical JSON behavior.
- The planned property-level rules are appropriately conservative: existing properties must remain with the same schema, and new properties must be optional.
- The plan includes actionable CI output with tool/property context and file paths, which satisfies the requirement for more than a bare non-zero exit.
- Compile-time interface assertions around the fake all-tools catalog client are a good guard against future registry additions being accidentally omitted from snapshot generation.

## Non-blocking implementation notes

- Make the CLI contract concrete in code/help text, especially how CI supplies the baseline directory. A required explicit `-baseline-dir` for stability mode is safer than trying to infer Git state silently.
- Keep the freshness check and stability check independently reportable so failures make it clear whether snapshots are stale or a breaking schema change was introduced.
- When comparing root schema invariants, fail additions to `required`; allowing an existing required property to become optional is likely non-breaking, but changing the property schema itself should remain a failure as planned.
- Avoid placing reusable helper code in the `scripts` package only; it should be importable by tests without invoking `go run`.

The plan is ready for implementation.

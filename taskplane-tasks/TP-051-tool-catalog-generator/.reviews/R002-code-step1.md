# R002 Code Review — Step 1: Design `ToolDescriptor`

**Verdict:** Request changes.

I reviewed `git diff c9700ab..HEAD`, read the changed files, and ran:

- `go test ./internal/tools` — passes
- `golangci-lint run ./internal/tools` — fails

## Blocking findings

1. **Lint fails because the new catalog helpers are currently unused.**  
   `internal/tools/catalog.go:18` adds `sortToolDescriptors` and `internal/tools/catalog.go:27` adds `toolSummary`, but Step 1 does not call them yet. `golangci-lint run ./internal/tools` reports both as `unused`, so the repository is not lint-clean after this step. Either defer these helpers until `Catalog()`/the generator uses them, add focused tests/call sites that exercise them, or make Step 1 include enough implementation to use them.

2. **`STATUS.md` misrecords the previous review outcome.**  
   `taskplane-tasks/TP-051-tool-catalog-generator/.reviews/R001-plan-step1.md` says `Verdict: Request changes before implementation`, but `taskplane-tasks/TP-051-tool-catalog-generator/STATUS.md:51` records `Review R001 | plan Step 1: APPROVE`. That makes the task audit trail incorrect. Update the status entry to reflect the actual R001 result and then add a separate R002 entry for this review.

## Non-blocking note

- `toolSummary` is close to the requested first-sentence extraction and handles `intervals.icu`, but it duplicates the existing `firstDescriptionSentence` helper in `list_advanced_capabilities.go`. When Step 2 wires this into `Catalog()`, prefer consolidating to one tested helper so summary generation and `icuvisor_list_advanced_capabilities` cannot diverge.

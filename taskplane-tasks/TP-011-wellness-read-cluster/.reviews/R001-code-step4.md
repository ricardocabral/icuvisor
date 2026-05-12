# Code Review: TP-011 Step 4 — In-response scale labels

**Verdict: Approve.**

No blocking findings.

The change expands the central response-shaper scale registry with the missing TP-011 wellness subjective fields (`soreness`, `stress`, `motivation`, `injury`) while preserving the existing `sleepQuality` 1-4 and `sleepScore` 0-100 labels. The updated shaper test now exercises all registered wellness subjective fields plus the dual sleep-score labels in a shaped response, and the implementation continues to use the shared `_meta.scales` assembly path rather than adding wellness-specific metadata.

## Notes

- The Step 4 plan asked for row-level verification through the wellness row shaping path. I confirmed from `shapeWrapperRow`/`shapeRows` that `Options{RowCollections: []string{"wellness"}}` shapes each row with `shapeRow(..., includeCommonMeta=false)`, so row-level `_meta.scales` is added when those fields are present. The current test coverage is still at the generic shaper/root-object level, not a `get_wellness_data` tool fixture; Step 6 should add that end-to-end wellness fixture assertion as planned.

## What I checked

- Ran `git diff 3a331d5bc80ca232340acf7ab50fe9766293d905..HEAD --name-only`.
- Ran `git diff 3a331d5bc80ca232340acf7ab50fe9766293d905..HEAD`.
- Read `PROMPT.md`, `STATUS.md`, `internal/response/shaper.go`, `internal/response/shaper_test.go`, and `internal/tools/get_wellness_data.go` for row-shaping context.
- Reviewed the relevant PRD wellness/response-shaping scale metadata requirements.
- Ran `go test ./internal/response ./internal/tools` — passes.

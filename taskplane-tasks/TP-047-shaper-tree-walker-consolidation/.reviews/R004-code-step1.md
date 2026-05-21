# Code Review — Step 1: Snapshot pre-refactor output

**Decision: Approved.**

The Step 1 snapshot infrastructure now provides a deterministic safety net for the upcoming shaper refactor. The R003 blockers are addressed: catalog runtime metadata is reset via parent and per-subtest cleanup, and the representative activity/fitness cases use typed DTO inputs with JSON tags/`omitempty` so the current marshal/tag behavior is locked into the golden outputs.

## Verification

- `go test ./internal/response -count=1` — passes
- `go test ./internal/response -shuffle=on -count=1` — passes
- `go test ./internal/response -run TestShapeGoldenSnapshots -count=1` — passes
- `go test ./...` — passes

## Notes

- The fixture set covers the requested surfaces: `get_activities` terse/full, `get_fitness`, wrapper row collections, full-payload null preservation, caller `_meta` preservation/overwrite behavior, row-level missing-field metadata, and provenance `fetched_at` preservation while ordinary debug-looking fields are stripped.
- The golden comparison uses explicit canonical indented JSON bytes and an opt-in `UPDATE_RESPONSE_GOLDENS=1` regeneration path, which is appropriate for Step 5's byte-identical pre/post check.

No blocking findings for Step 1.

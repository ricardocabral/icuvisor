# Code Review — TP-007 Step 1

Verdict: **Not approved yet**. Step 1 now records some high-level choices, but the design still leaves key response-shaping contract points underspecified and one choice conflicts with the null-stripping requirements.

## Blocking findings

1. **`omitempty` is still unresolved and the current pipeline would hide required missing-field data.**
   - `STATUS.md` says the response package will convert structs through JSON marshal/unmarshal semantics so tags and `omitempty` stay authoritative.
   - That means `omitempty` fields are removed before the null-stripper runs. Existing `get_athlete_profile` response structs already use `omitempty`, including zero-value scalar fields and pointer fields.
   - This conflicts with the PRD/task requirement that only JSON `null` is stripped and that `0`, `""`, and `false` are preserved. It also prevents accurate `_meta.missing_fields` because the shaper cannot tell whether a missing key was omitted by `omitempty`, intentionally absent from a terse schema, or originally `null`.
   - Please make an explicit design decision: no `omitempty` on shaped nullable/value fields, expected-field metadata, reflection-based shaping, or another concrete mechanism that lets the shaper distinguish nulls from meaningful zero/empty/false values.

2. **The response chokepoint/API is not concrete enough to guarantee `_meta.server_version` everywhere.**
   - The design mentions “one response-package chokepoint” but does not name the function(s), inputs, outputs, or ownership of MCP `Content` vs `StructuredContent`.
   - Existing tools currently marshal text and assign structured content in-handler. Step 2+ needs a specific API all read tools must call, such as a builder that returns both JSON text content and the shaped structured value, with version/debug/unit/scales options.
   - Without this, downstream implementations can easily bypass server-version injection or produce text/structured content that diverge.

3. **Per-row metadata semantics are underspecified.**
   - The design says “row metadata is assembled after stripping,” but does not define what counts as a row for a single object, an array response, or a wrapper object containing rows.
   - It also does not specify how nested object/array nulls are represented in `_meta.missing_fields`, whether paths are dotted or local keys, whether `fields_present` includes `_meta`, or how ordering is made deterministic.
   - These details are required for the Step 2 tests and for consistent multi-row wellness/activity shapes.

4. **`include_full: true` cannot currently promise raw nulls.**
   - `STATUS.md` says `include_full: true` bypasses stripping to preserve raw typed nulls.
   - With the selected JSON round-trip plus authoritative `omitempty`, null pointer fields tagged `omitempty` will still be absent, not present as JSON null. This is the same root issue as finding #1, but it specifically breaks the explicit debug/inspection opt-out behavior.

5. **Review/status bookkeeping is inconsistent.**
   - `STATUS.md` still has `State: Pending` and Step 1 marked `In Progress` even though all Step 1 checklist items are checked.
   - It also appends `| 2026-05-11 12:18 | Review R001 | plan Step 1: APPROVE |` under `## Blockers`, outside a table, while the committed `R001-plan-step1.md` verdict is “Not approved yet.”
   - Please correct the status/audit trail so reviewers and future recovery runs do not treat an unapproved design as approved.

## Non-blocking notes

- The per-tool JSON tag decision for canonical field names is the right direction. The design should still spell out how tools choose between dynamic unit-specific keys (`distance_km` vs `distance_mi`, pace keys) when the key cannot be a single static struct tag.
- Please explicitly confirm that `config.NormalizeAthleteID` remains the central normalizer and that any helper in `internal/response` will call/configure around it rather than adding a second normalization path.


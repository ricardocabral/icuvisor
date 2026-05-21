# Plan Review — Step 3: Implement

**Decision: APPROVE.**

The updated Step 3 plan addresses the contract-critical gaps from R008. It now explicitly requires fresh ownership of JSON containers, defines the fast-path/fallback scope for the reflection converter, calls out `json.RawMessage` and marshaler handling, adds focused converter/walker tests, and makes fallback accounting auditable in `STATUS.md`. That is enough guardrail coverage to proceed with implementation.

## What looks good

- The single visitor walker remains the right M-sized approach for this repository: it keeps the API surface and call sites stable while consolidating the duplicate traversal logic.
- The `toJSONValue` ownership contract is now clear: every map/slice/array returned by the converter must be freshly allocated, including nested `include_full` raw payloads. This preserves the mutation isolation that the old marshal/unmarshal round-trip provided.
- The plan recognizes the important `encoding/json` edge cases: `json.Marshaler`, `encoding.TextMarshaler`, `json.RawMessage`, unsupported values, and numeric byte-equivalence.
- The Step 3-specific tests are now explicit enough to catch regressions beyond the five golden fixtures, especially for tag handling, `omitempty`, raw/fallback behavior, no-mutation, provenance/debug paths, null stripping, and scale collection.
- Fallback accounting is now part of the plan, so it should be possible to prove that normal golden fixtures do not silently rely on the old whole-response round-trip.

## Non-blocking implementation reminders

- Include embedded/anonymous struct field behavior in the converter tests, or explicitly route that case through a narrow fallback. The status says embedded fields will be honored; the tests should lock that down because `encoding/json` conflict resolution is subtle.
- Keep any marshal fallback per-value only. Do not reintroduce a whole-response marshal/unmarshal path under a helper name.
- When replacing the walkers, preserve the existing path formatting exactly (`nested.key`, `array[0].field`, root array paths if encountered), because missing-field golden output depends on it.
- Be careful with caller-owned `_meta` maps in `include_full` mode: even if null stripping is disabled, later metadata functions still mutate the shaped map, so the converter deep-copy is required before shaping starts.

Proceed to Step 3 implementation.

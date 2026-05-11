# Plan Review — TP-007 Step 4

Verdict: **Not approved yet**. The Step 4 section in `STATUS.md` only repeats the prompt checklist. This step is small, but it still needs a concrete plan because scale metadata must integrate with the existing `response.Shape` chokepoint, row-collection semantics, and `_meta` merge rules established in Steps 2-3.

## Blocking findings

1. **No concrete registry/API is defined.**
   - The plan does not name the registry location, exported helper(s), or how callers/tests will inspect labels.
   - A package-level mutable map would be easy to accidentally mutate from tests or future code. The plan should define whether the registry is private with lookup/clone helpers, constants, or an injectable option.

2. **Integration with `Shape` and row collections is unspecified.**
   - Step 4 requires `_meta.scales` for every registered field present in a response row, but the current shaper has multiple row contexts: a single object wrapper, wrapper-level objects, and named `Options.RowCollections` whose elements are shaped with common metadata disabled.
   - The plan must state exactly where scale labels are added for single-row responses and for row collections, and whether wrapper/root metadata gets `_meta.scales` only when registered fields exist on the wrapper itself.

3. **Ordering relative to null stripping and `include_full` is not defined.**
   - Scale detection should run after terse null stripping so labels are emitted only for fields the LLM actually sees.
   - The plan also needs to define `include_full: true`: since null stripping is skipped but common metadata still runs, registered keys present with JSON `null` should either receive labels or be explicitly excluded. This should be intentional and tested.

4. **The initial scale set is undefined.**
   - PRD/roadmap require in-response labels for subjective fields and explicitly call out `feel`, `sleepQuality`, and `sleepScore`; the roadmap also mentions `fatigue`, `mood`, etc.
   - The plan should list the exact field-name-to-label strings to seed now, and distinguish this registry plumbing from the out-of-scope wellness tool/provenance work. Avoid inventing labels for fields whose scales are not confirmed by the authoritative docs.

5. **`_meta.scales` merge/collision behavior is missing.**
   - Existing `_meta` can already contain `fields_present`, `missing_fields`, and `server_version`.
   - The plan should say whether computed `scales` overwrites an existing `_meta.scales`, merges with it, or rejects non-object values. Since `internal/response` owns metadata assembly, reserving and overwriting `scales` with the computed registry result is likely safest, while preserving other `_meta` keys.

6. **Tests are not planned at the right boundaries.**
   - Step 7 has a global scale-registry test item, but this step should still define focused tests before coding.
   - Needed cases include: no `_meta.scales` when no registered field is present; labels only for present registered fields; no labels for unknown/custom fields; terse mode omits labels for stripped-null registered fields; include-full behavior for null registered fields; row-collection rows get independent scales; existing `_meta` is preserved except for the reserved `scales` key; and returned registry maps cannot mutate the central registry if a clone helper is exposed.

## Required additions before approval

Please update `STATUS.md` with a Step 4 implementation plan covering at least:

- The registry representation and API in `internal/response` (`ScaleLabel`, `ScaleLabelsForRow`, registry clone, or equivalent), including mutability guarantees.
- The exact seeded labels and their source in PRD/roadmap, plus any intentionally deferred labels.
- How `response.Shape` calls the scale helper for single rows and named row collections.
- The order of operations relative to null stripping, debug metadata, strip metadata, and common metadata.
- `include_full` semantics for registered fields whose value is `null`.
- `_meta.scales` merge/overwrite behavior and preservation of other `_meta` keys.
- Focused tests for the cases above.

Once these decisions are recorded, Step 4 should be straightforward to implement and review.

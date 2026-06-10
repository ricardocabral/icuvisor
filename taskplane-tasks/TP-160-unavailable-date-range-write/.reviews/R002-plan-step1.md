# Plan Review: Step 1 — Design the range-write contract

Verdict: REVISE

The updated Step 1 contract is much closer and resolves most of R001: it chooses a dedicated tool, gives a bounded date/category surface, identifies catalog/coach ACL surfaces, and defines per-day writes around the existing single-date event client. Before implementation, please tighten these contract points:

1. **Idempotency key is under-specified/unsafe.** The proposed generated `external_id` hashes only normalized category/date/name, but the writable request also includes `description`. A later call with the same date/category/name and a different description would match the generated external_id and be skipped, even though it is not an identical retry. Either include all fields that define the intended marker in the idempotency fingerprint, or explicitly define matching-external-id-but-different-fields as a conflict/non-skip case.

2. **Response `status` needs explicit mixed/all-skipped semantics.** The draft always shows `status: "created"`, but the contract also allows all days skipped, mixed created/skipped ranges, and conflicts. Define the status enum now (for example `created`, `partial`, `skipped`) so tests do not encode misleading confirmations.

3. **Partial failure semantics should be stated.** Because implementation will loop one upstream write per day, a mid-range upstream error can leave earlier days created. The contract should say whether the tool returns an error with no structured counts, returns partial success metadata, or attempts no rollback. This is important for retry guidance and idempotency tests.

4. **Category aliases should be kept intentionally narrow or justified.** `VACATION`/`TIME_OFF` are clearly time-off aliases, but `TRAVEL` and `AWAY` are broader and are not documented upstream categories in this repo. If they remain, note that they intentionally normalize to `HOLIDAY`; otherwise remove them to avoid surprising holiday markers.

Also add to the Step 2 checklist that a new write tool must satisfy the existing input-example/schema stability tests (`examples` + `input_examples`, catalog tier expectations), not just schema snapshots/hash.

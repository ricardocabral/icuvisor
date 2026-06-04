# Plan Review: Step 1 — Design diagnostic contract

Verdict: **Changes requested**

## Findings

1. **Description baseline omits the catalog hash required by the task.**
   The mission explicitly asks for the loaded version/hash to be visible in the tool description so stale-catalog diagnosis works when clients hide `_meta`. The plan intentionally embeds only version/toolset/delete-mode and says the catalog hash will not be in the description. That leaves no visible baseline hash to compare with the live `catalog_hash` response, so same-version catalog/schema changes, dev builds, or regenerated schemas may not be diagnosable without `_meta`.

2. **Mismatch detection is underspecified.**
   The response includes `status` and `action`, but the plan does not define what the tool compares against. A no-argument tool cannot automatically know the client's stale visible description/hash. The contract should either:
   - return a neutral `status` such as `compare_visible_description` with instructions to compare visible description fields to response fields, or
   - accept an explicit optional baseline value from the user/client, if that is intended.

3. **Self-referential hash problem needs a concrete contract, not just omission.**
   It is valid that embedding the exact `hashToolCatalog` result in a description that is itself hashed is circular. The plan should define an alternative that still satisfies the requirement, for example a clearly named `description_catalog_fingerprint` computed with the diagnostic tool's dynamic hash token normalized/excluded, or a baseline hash over the catalog before injecting the diagnostic description. The response can then expose both the live full `catalog_hash` and the comparable description fingerprint/hash.

## Required plan updates before implementation

- Define the visible description fields exactly, including some comparable catalog hash/fingerprint value, not only version/toolset/delete-mode.
- Define response field semantics and mismatch/action wording without implying the server can detect stale client state it cannot observe.
- Add tests to the plan for same-version catalog/schema drift where only the visible baseline hash/fingerprint changes.

The privacy boundary and no-network/no-athlete/no-secret direction are sound.

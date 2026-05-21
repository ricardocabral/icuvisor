# Plan Review: Step 1 — Define the typed structs

**Decision: approved.**

The updated Step 1 plan addresses the blocking concerns from R001 and aligns with the prompt: write-body structs live in `internal/intervals`, shaped response summaries live in `internal/tools`, sparse update semantics are called out explicitly, and opaque `full` payloads are no longer planned as `map[string]any` response fields.

## Implementation notes to carry into Step 1

- Make the `full` representation decision concrete in code, preferably `json.RawMessage`, rather than leaving multiple opaque representations in play.
- Keep the write request unexported unless tests or another package require otherwise. Pointer fields are important for update semantics:
  - send `folder_id:""` when `FolderIDSet` is true and the trimmed value is empty;
  - send `tags:[]` when `TagsSet` is true with an empty slice;
  - reject nil descriptions when `DescriptionSet` is true, but preserve the existing allowance for an explicit empty description string.
- Use typed summary fields with the existing JSON keys only. The intended shapes should cover:
  - workout doc summary: `present`, `step_count`, `name`, `top_level_keys`;
  - training plan summary: `id`, `name`, `description`, `folder_id`, `type`, `category`, `child_count`, `workout_count`, `top_level_keys`.
- It is fine for helper code to inspect upstream `map[string]any` values internally while normalizing summaries, but those maps should not remain as request/response body fields in the three audited files.

With those constraints, proceed to implement Step 1.

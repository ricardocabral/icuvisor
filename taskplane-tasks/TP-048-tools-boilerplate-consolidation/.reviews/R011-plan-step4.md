# R011 Plan Review — Step 4: `Requirement` enum

Verdict: **REVISE**

The Step 4 status/checklist has not made the required enum-shape decision yet, and that decision matters in this tree.

Current audit notes:

- `internal/tools/registry.go` already declares `type Requirement string` with typed `RequirementRead`, `RequirementWrite`, and `RequirementDelete` constants. If this is the intended shape, the Step 4 plan should say so explicitly and avoid unnecessary churn.
- Requirement values are serialized into a user-visible/tool response via `internal/tools/list_advanced_capabilities.go` (`advancedCapabilityRow.Requirement` and `toolRequirement`). The wire strings `"read"`, `"write"`, and `"delete"` must remain unchanged.
- References are not limited to `internal/tools`; `internal/mcp/protocol_test.go` and `internal/safety/adversarial_test.go` also use `tools.Requirement`. The audit command in the plan should cover at least `internal/`, not only `internal/tools/`.
- There is still a zero-value/default path (`tool.Requirement == ""` in `toolRequirement`) that treats omitted requirements as read. If the acceptance criterion is “no stringly-typed comparisons remain,” the revised plan should preserve this behavior using a typed/default sentinel or a central helper, not a raw string literal, and should not require setting `RequirementRead` on every read tool unless snapshot parity is confirmed.

Recommended revised plan:

1. Choose and record `type Requirement string` in `STATUS.md`. This is the safest option because the value is already serialized and current tests compare the public strings.
2. Keep `RequirementRead = "read"`, `RequirementWrite = "write"`, and `RequirementDelete = "delete"` exactly. Do not switch to `int`/`iota` unless the plan also adds a string conversion/marshalling boundary that preserves the current output.
3. Audit all references with `grep -rn "Requirement" internal/`, including `internal/mcp` and `internal/safety` tests.
4. Preserve the existing zero-value behavior for tools that omit `Requirement` (they are read tools). If cleaning up `tool.Requirement == ""`, use a typed private sentinel or a small central function; do not change registration semantics.
5. Verify with targeted tests for the affected packages (`go test ./internal/tools ./internal/mcp ./internal/safety`) plus the schema-stability check in Step 5.

Once the enum-shape decision and zero-value handling are documented, this step should be small and safe to implement.

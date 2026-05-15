# R012 Plan Review — Step 4: `Requirement` enum

Verdict: **APPROVE**

The revised Step 4 plan now addresses the R011 blockers:

- The enum shape decision is recorded: keep `type Requirement string` with exact public values `"read"`, `"write"`, and `"delete"`.
- The plan acknowledges that requirement values are serialized by `icuvisor_list_advanced_capabilities`, so wire format must remain stable.
- The audit scope is correctly widened to `grep -rn "Requirement" internal/`, including `internal/mcp` and `internal/safety` tests.
- The zero-value/default-read behavior is explicitly called out and should be preserved without raw string comparisons.

Implementation guidance for this step:

1. Do not switch to `int`/`iota`; the typed string enum is already the safe shape for this repository.
2. Keep the existing constant values byte-for-byte. Any user-visible `requirement` field emitted by `list_advanced_capabilities` must remain unchanged.
3. Prefer a tiny central normalization point, e.g. a private zero sentinel or an `EffectiveRequirement`/`requirementOrRead` helper that returns `RequirementRead` for the zero value. Use that from `RequiresWrite`, `RequiresDelete`, and `toolRequirement` if needed.
4. Avoid a noisy mass edit that sets `RequirementRead` on every read tool unless tests prove it is necessary; omitted requirement currently means read.
5. Run at least `go test ./internal/tools ./internal/mcp ./internal/safety` after the change, with full verification deferred to Step 5.

With those guardrails, the plan is small, behavior-preserving, and within TP-048 scope.

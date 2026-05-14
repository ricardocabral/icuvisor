# Plan Review R002 — Step 1: Resource registration plumbing

**Verdict: Approved.**

I read `PROMPT.md`, the updated `STATUS.md`, the prior R001 review, and the current MCP server wiring. The revised Step 1 plan now covers the required SDK integration details and is safe to implement.

## What is good

- The plan explicitly uses `(*mcp.Server).AddResource(...)` instead of custom JSON-RPC handlers, so the Go SDK remains responsible for `resources/list`, `resources/read`, capability advertisement, pagination, and not-found behavior.
- The proposed `internal/resources` boundary keeps domain resource definitions decoupled from SDK types, matching the existing tool registry pattern.
- Startup validation is called out for URI shape, duplicates, metadata fields, MIME type, and nil readers/handlers, with failures returning `NewServer` errors rather than panics.
- The plan wires resources before sessions initialize, which is necessary for `initialize` to advertise the resources capability.
- Error handling is addressed: log internal details, return short/safe client-facing errors, and preserve SDK `ResourceNotFoundError` behavior for unknown resources.
- The Step 1 content contract is clear enough for plumbing tests: one text `ResourceContents` item with URI, MIME type, and text populated.
- Static vs. dynamic decisions for all four required resources are now recorded in `STATUS.md`.
- The proposed protocol-level tests cover the important first-integration risks.
- The canonical SDK API link is recorded in `STATUS.md`.

## Implementation notes

These are not blockers, but please keep them in mind while coding:

1. `AddResource` can panic on invalid URIs, and duplicate resource URIs are silently replaced by the SDK feature set. The registrar should validate and track duplicates before calling `AddResource`, and it is still worth wrapping SDK registration with `defer recover` as the tool registrar does.
2. Keep Step 1 to plumbing plus test resources. Do not register placeholder production resources from `internal/app` until the later steps can satisfy the derivation/golden/cache requirements for the real resources.
3. For sanitized read failures, return a generic protocol error message without leaking the original error string; log the detailed error with structured fields. Unknown/unregistered URIs should remain the SDK `CodeResourceNotFound` path.
4. Prefer a separate `ResourceRegistry` option rather than overloading the existing tool `Registry`, so `internal/app` can pass the default resource registry later without coupling tools and resources.

With those cautions, the plan is ready for implementation.

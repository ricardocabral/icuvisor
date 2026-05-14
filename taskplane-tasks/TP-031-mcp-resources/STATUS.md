# TP-031-mcp-resources: TP-031-mcp-resources — Status

**Current Step:** Step 1: Resource registration plumbing
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-14
**Review Level:** 2
**Review Counter:** 3
**Iteration:** 1
**Size:** M

---

### Step 1: Resource registration plumbing

**Status:** ✅ Complete

- [x] Wire `resources/list` and `resources/read` into the MCP server via the Go SDK
- [x] Define a small internal interface so each resource is one greppable registration, mirroring the tool registry pattern
- [x] Decide static vs dynamic per resource; document in `STATUS.md`

### Step 2: `icuvisor://workout-syntax`

**Status:** ⏳ Not started

### Step 3: `icuvisor://event-categories`

**Status:** ⏳ Not started

### Step 4: `icuvisor://custom-item-schemas`

**Status:** ⏳ Not started

### Step 5: `icuvisor://athlete-profile`

**Status:** ⏳ Not started

### Step 6: Trim inline tool descriptions

**Status:** ⏳ Not started

### Step 7: Verify

**Status:** ⏳ Not started

---

## Reviews

| #   | Type | Step | Verdict | File |
| --- | ---- | ---- | ------- | ---- |
| R001 | Plan | 1 | REVISE | .reviews/R001-plan-step1.md |
| R002 | Plan | 1 | APPROVE | inline |
| R003 | Code | 1 | APPROVE | inline |

---

## Discoveries

| Discovery | Disposition | Location |
| --------- | ----------- | -------- |

---

## Execution Log

| Timestamp  | Action      | Outcome                          |
| ---------- | ----------- | -------------------------------- |
| 2026-05-14 | Task staged | Scaffolded from ROADMAP.md v0.4   |
| 2026-05-14 14:09 | Task started | Runtime V2 lane-runner execution |
| 2026-05-14 14:09 | Step 1 started | Resource registration plumbing |

---

## Blockers

_None_

---

## Notes

### Step 1 plan

- SDK reference consulted: https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk/mcp#Server.AddResource (go-sdk v1.4.1 resource registration API).
- Use the SDK Resource API only: register resources with `(*mcp.Server).AddResource(...)` so the SDK owns `resources/list`, `resources/read`, capabilities, pagination, and list-change behavior; do not add custom JSON-RPC handlers.
- Define a registry boundary that mirrors the tool registry while keeping domain resource definitions out of SDK types. `internal/resources` will expose small resource definitions/registries; `internal/mcp` will validate and convert them to SDK resource registrations.
- Validate registry entries before server startup completes: absolute `icuvisor://...` URI, no duplicate URIs, non-empty name/title/description/MIME type, and non-nil read handlers/content readers. Invalid registration returns a `NewServer` error, never a panic.
- Wire the registry through MCP server construction options so `internal/app` can pass the default resource registry alongside the tool registry. Add resources before sessions initialize so `initialize` advertises the resources capability.
- Handler failures are logged and returned as short, safe client-facing errors. Unknown/unregistered resources keep the SDK not-found protocol behavior.
- Step 1 content contract: read handlers return one text `ResourceContents` item with populated `URI`, `MIMEType`, and `Text`; per-resource steps may override MIME type, with long-form docs defaulting to `text/markdown`.
- Static/dynamic decisions: `icuvisor://workout-syntax` is static/derived from `internal/workoutdoc`; `icuvisor://event-categories` is static from the same enum/source as event tools; `icuvisor://custom-item-schemas` is static/derived from custom-item validation/schema sources; `icuvisor://athlete-profile` is dynamic cached content with TTL/staleness policy finalized in Step 5.
- Protocol tests for Step 1 will use in-memory MCP client helpers to assert initialize advertises resources when configured, list returns metadata, read dispatches with URI/MIME/text, invalid/duplicate registrations fail server construction, and unknown reads return the SDK not-found protocol error.

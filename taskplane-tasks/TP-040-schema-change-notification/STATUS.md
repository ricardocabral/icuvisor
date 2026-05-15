# TP-040-schema-change-notification: Post-update schema-change notification — Status

**Current Step:** Step 3: Tests
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-15
**Review Level:** 2
**Review Counter:** 10
**Iteration:** 1
**Size:** S

---

### Step 1: Catalog hash

**Status:** ✅ Complete

- [x] SHA-256 over canonical sorted records for the exposed registered catalog after toolset/capability filtering, including name, tool description, input schema, and any advertised output schema
- [x] Store catalog hash on `internal/mcp.Server` and expose `Server.CatalogHash() string` computed once from the actual tools passed to the SDK
- [x] Determinism + sensitivity fixture tests (registration order, nested map order, filtering, add / remove / rename / description-edit, tool-description edit, output-schema edit if advertised)
- [x] R003: run gofmt on changed Go files so lint passes
- [x] R003: remove trailing whitespace from committed review artifacts

### Step 2: `_meta` injector

**Status:** ✅ Complete

- [x] Add concurrency-safe runtime catalog metadata in `internal/response`, set by `internal/mcp.NewServer` after `Server.CatalogHash()` is computed, with test reset/set hooks, a deterministic no-server default catalog hash for direct tool tests, and no hash in tool descriptions or schemas
- [x] `catalog_hash` on every response via response-owned `_meta`, overwriting any caller-provided schema-change keys to prevent spoofed metadata
- [x] Audit all direct JSON response paths (`StructuredContent: payload` / `StructuredContent: response` / direct `json.Marshal(...)`) and convert bypasses including `list_advanced_capabilities`, `update_sport_settings`, and `update_wellness` so every tool response uses the common metadata injector consistently
- [x] First-seen hash/version tracking with an atomic/mutex-protected current snapshot and documented per-process fallback caveat because no SDK session handle is available at the response shaper boundary
- [x] `schema_changed` block populated on divergence with previous/current versions, previous/current hashes, and a testable `schemaChangeMessage(previousVersion, currentVersion)` template

### Step 3: Tests

**Status:** ✅ Complete

- [x] Hash determinism + sensitivity
- [x] Injector session-start / steady / simulated-change
- [x] Tool golden files unaffected

### Step 4: Documentation

**Status:** ⏳ Not started

- [ ] `docs/post-update.md`
- [ ] CHANGELOG `[Unreleased]`

---

## Decisions

- **Description-only changes hash:** counted as schema change (descriptions are part of the LLM-facing contract).

## Notes

_Add notes as work progresses._

| 2026-05-15 14:35 | Task started | Runtime V2 lane-runner execution |
| 2026-05-15 14:35 | Step 1 started | Catalog hash |
| 2026-05-15 14:37 | Review R001 | plan Step 1: UNKNOWN |
| 2026-05-15 14:40 | Review R002 | plan Step 1: APPROVE |
| 2026-05-15 14:47 | Review R003 | code Step 1: UNKNOWN |
| 2026-05-15 14:52 | Review R004 | code Step 1: APPROVE |
| 2026-05-15 14:57 | Review R005 | plan Step 2: REVISE |
| 2026-05-15 15:00 | Review R006 | plan Step 2: REVISE |
| 2026-05-15 15:02 | Review R007 | plan Step 2: APPROVE |
| 2026-05-15 15:13 | Review R008 | code Step 2: APPROVE |
| 2026-05-15 15:15 | Review R009 | plan Step 3: APPROVE |
| 2026-05-15 15:19 | Review R010 | code Step 3: APPROVE |

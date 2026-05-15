# TP-040-schema-change-notification: Post-update schema-change notification — Status

**Current Step:** Step 1: Catalog hash
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-15
**Review Level:** 2
**Review Counter:** 3
**Iteration:** 1
**Size:** S

---

### Step 1: Catalog hash

**Status:** 🟨 In Progress

- [x] SHA-256 over canonical sorted records for the exposed registered catalog after toolset/capability filtering, including name, tool description, input schema, and any advertised output schema
- [x] Store catalog hash on `internal/mcp.Server` and expose `Server.CatalogHash() string` computed once from the actual tools passed to the SDK
- [x] Determinism + sensitivity fixture tests (registration order, nested map order, filtering, add / remove / rename / description-edit, tool-description edit, output-schema edit if advertised)
- [ ] R003: run gofmt on changed Go files so lint passes
- [ ] R003: remove trailing whitespace from committed review artifacts

### Step 2: `_meta` injector

**Status:** ⏳ Not started

- [ ] `catalog_hash` on every response
- [ ] First-seen-per-session tracking (per-process fallback caveat)
- [ ] `schema_changed` block populated on divergence

### Step 3: Tests

**Status:** ⏳ Not started

- [ ] Hash determinism + sensitivity
- [ ] Injector session-start / steady / simulated-change
- [ ] Tool golden files unaffected

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

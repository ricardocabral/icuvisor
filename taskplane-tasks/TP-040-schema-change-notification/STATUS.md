# TP-040-schema-change-notification: Post-update schema-change notification — Status

**Current Step:** Step 1: Catalog hash
**Status:** ⏳ Not started
**Last Updated:** 2026-05-15
**Review Level:** 2
**Review Counter:** 0
**Iteration:** 0
**Size:** S

---

### Step 1: Catalog hash

**Status:** ⏳ Not started

- [ ] SHA-256 over sorted `(name, schema-json)` pairs
- [ ] Determinism + sensitivity fixture tests (add / remove / rename / description-edit)

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

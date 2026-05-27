# TP-116: Privacy and GDPR positioning — Status

**Current Step:** Step 1: Define privacy claims and boundaries
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-27
**Review Level:** 1
**Review Counter:** 0
**Iteration:** 1
**Size:** S

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers may expand steps when runtime discoveries warrant it.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Existing privacy, security, coach-mode, and HTTP transport claims reviewed

---

### Step 1: Define privacy claims and boundaries
**Status:** 🟨 In Progress

- [ ] Substantiated privacy/security claims inventoried
- [ ] Explicit non-claims defined: no legal advice/certification, AI-client caveat, upstream data relationship caveat
- [ ] Dedicated page versus existing-page update decision made

**Plan-review checkpoint**

---

### Step 2: Add privacy-conscious user-facing copy
**Status:** ⬜ Not Started

- [ ] Local trust boundary, credential storage, HTTP bind defaults, and coach-mode credential handling documented
- [ ] EU/GDPR-conscious language framed as design posture/questions, not certification
- [ ] Homepage/README pointer added only if useful

---

### Step 3: Link from relevant docs and update changelog
**Status:** ⬜ Not Started

- [ ] Privacy positioning linked from local-first, coach-mode, HTTP transport, and/or safety docs where relevant
- [ ] `SECURITY.md` remains authoritative for security policy
- [ ] `CHANGELOG.md` updated under `[Unreleased]` if appropriate

---

### Step 4: Testing & Verification
**Status:** ⬜ Not Started

- [ ] Docs/site build passing: `make web-build`
- [ ] Rendered links and page placement checked
- [ ] FULL test suite run if non-doc/generated app files are touched: `make test`
- [ ] Build passes if app strings or generated assets are touched: `make build`
- [ ] All failures fixed or documented as pre-existing unrelated failures

---

### Step 5: Documentation & Delivery
**Status:** ⬜ Not Started

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] Discoveries logged
- [ ] Final commit includes task ID

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-27 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-27 18:13 | Task started | Runtime V2 lane-runner execution |
| 2026-05-27 18:13 | Step 0 started | Preflight |

---

## Blockers

*None*

---

## Notes

*Reserved for execution notes*

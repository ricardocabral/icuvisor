# TP-162: Codex Streamable HTTP JSON-RPC smoke coverage — Status

**Current Step:** Step 1: Add Streamable HTTP JSON-RPC handshake smoke tests
**Status:** 🟡 In Progress
**Last Updated:** 2026-06-10
**Review Level:** 2
**Review Counter:** 0
**Iteration:** 1
**Size:** M

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied

---

### Step 1: Add Streamable HTTP JSON-RPC handshake smoke tests
**Status:** 🟨 In Progress

- [ ] Initialize response envelope test added
- [ ] Ping response envelope test added
- [ ] Codex-like HTTP headers covered without external process
- [ ] Targeted MCP tests pass

---

### Step 2: Fix transport/protocol behavior only if tests fail
**Status:** ⬜ Not Started

- [ ] Transport/server fixes applied if needed
- [ ] Stdio and loopback defaults preserved
- [ ] Protocol errors remain short/actionable
- [ ] Targeted MCP tests pass

---

### Step 3: Testing & Verification
**Status:** ⬜ Not Started

- [ ] FULL test suite passing
- [ ] Integration tests (if applicable)
- [ ] All failures fixed
- [ ] Build passes

---

### Step 4: Documentation & Delivery
**Status:** ⬜ Not Started

- [ ] `docs/clients/codex-local.md` reviewed/updated if affected
- [ ] `CHANGELOG.md` updated
- [ ] Discoveries logged

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
| 2026-06-09 | Task staged | PROMPT.md and STATUS.md created |
| 2026-06-10 12:17 | Task started | Runtime V2 lane-runner execution |
| 2026-06-10 12:17 | Step 0 started | Preflight |
| 2026-06-10 12:17 | Step 0 completed | Required paths and Go module dependencies verified |
| 2026-06-10 12:17 | Step 1 started | Streamable HTTP JSON-RPC smoke tests |

---

## Blockers

*None*

---

## Notes

Public signal: Montis forum #512 and #516-518 described Codex handshake failure until JSON-RPC 2.0 wrapping was strict.

# TP-041-v05-dogfood-validation: v0.5 dogfood prep (artifacts only) — Status

**Current Step:** Step 1: Diagnostics subcommand
**Status:** ⏳ Not started
**Last Updated:** 2026-05-15
**Review Level:** 2
**Review Counter:** 0
**Iteration:** 0
**Size:** M

---

### Step 1: Diagnostics subcommand

**Status:** ⏳ Not started

- [ ] `icuvisor diagnostics` prints version, catalog hash, config source, mode env vars, OS/Go runtime, last-N tool-call names+timestamps
- [ ] No-secret-leakage test (fixture-injected key, stdout grep)
- [ ] TP-035 `--help` golden file updated

### Step 2: Docs

**Status:** ⏳ Not started

- [ ] `protocol.md` — recruitment + consent + eligibility + 5–10 cap + 14-day time-box
- [ ] `onboarding-playbook.md` — operator terminal recipe + coach variant + troubleshooting
- [ ] `measurement.md` — KR1 / §7.4 #6/#8/#12 measurement procedure + table template
- [ ] `exit-interview.md` — 8–12 question end-of-beta script
- [ ] `findings.md` — empty template only
- [ ] `checklist.md` — single-page operator checklist

### Step 3: Cross-check + verify

**Status:** ⏳ Not started

- [ ] `README.md` index links all five docs in execution order
- [ ] Cold-read of `checklist.md` is self-sufficient
- [ ] `make test` / `make lint` / `make build` clean

---

## Decisions

- **Scope:** artifacts only. Recruitment, cohort execution, synthesis are maintainer-run after this task closes.
- **`findings.md`:** empty template — no fabricated example data.

## Notes

_Add notes as work progresses._

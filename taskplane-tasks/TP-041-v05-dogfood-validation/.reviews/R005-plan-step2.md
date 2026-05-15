# Plan Review — Step 2: Docs

Decision: **changes requested**

I do not see a concrete Step 2 implementation plan beyond the checklist in `STATUS.md`, so this is not yet approvable as a plan. Before drafting, please spell out the document-by-document structure and cross-reference strategy so the docs satisfy the prompt without drifting into beta execution or duplicated setup content.

Required plan fixes:

1. **Account for all required files.** The current Step 2 status lists six docs and omits `docs/internal-beta/README.md`, while the prompt requires seven files total. Step 3 also says “all five docs,” which is inconsistent with the prompt. The plan should name all seven outputs: `README.md`, `protocol.md`, `onboarding-playbook.md`, `measurement.md`, `exit-interview.md`, `findings.md`, and `checklist.md`.

2. **Define each document’s skeleton before writing prose.** Keep the files under ~150 lines as requested, but specify the headings/tables/checklists each file will contain. In particular:
   - `protocol.md`: recruitment script, consent statement, eligibility filters, cohort cap 5–10, 14-day time-box, and explicit non-execution boundary.
   - `onboarding-playbook.md`: operator terminal recipe from DMG install through first tool call, coach-mode variant, troubleshooting path to `icuvisor diagnostics`, and release/update exercise as instructions only.
   - `measurement.md`: KR1 and PRD §7.4 #6/#8/#12 procedure plus the table schema.
   - `exit-interview.md`: 8–12 questions covering coach mode, schema-change notification clarity, willingness to recommend, and daily-use blockers.
   - `findings.md`: empty template only, with no fabricated rows or example athlete data.
   - `checklist.md`: one-page Recruit → Onboard → Run → Synthesize checklist with links.
   - `README.md`: execution-order index linking every doc.

3. **Cross-reference setup docs without duplicating JSON snippets.** The prompt says the onboarding playbook should include the Claude Desktop / Code snippets by cross-reference and “do not duplicate — link.” The plan should point to the existing `docs/clients/claude-desktop.md` and `docs/clients/claude-code.md` (and any relevant install doc) instead of pasting their JSON blocks into `docs/internal-beta/`.

4. **Consent/privacy wording needs to be deliberate.** The plan should state that `protocol.md` will say exactly what the maintainer receives: install-to-first-call timing, top tool-call names/timestamps or descriptions, mobile-need answer, qualitative notes, blockers, diagnostics output when voluntarily shared, and no raw training data, API keys, athlete IDs, payloads, arguments, screenshots with values, or transcripts. It also needs revocation instructions.

5. **Measurement scope must stay manual and artifact-only.** Because opt-in telemetry and cohort execution are out of scope, the plan should make KR1 / §7.4 #12 collection manual for this beta and avoid promising automatic telemetry. It should also explicitly ask the mobile-need question for §7.4 #8 and demand/willingness-to-recommend signal for §7.4 #6.

6. **Coach-mode coverage should be concrete but non-invasive.** The onboarding and interview docs should reference TP-039 roster/ACL behavior and coach-mode config flow, but must not ask coaches to share roster names, athlete IDs, or raw athlete data. The plan should specify redacted categories only.

7. **Link hygiene and status updates should be included.** The plan should include updating `STATUS.md` checkboxes/notes as docs are created and then doing a local link/path sanity pass before Step 3. If any docs reference PRD anchors, prefer stable section labels plus assumption numbers (for example, `PRD §7.4 #12`) so the maintainer can find them later.

Suggested plan shape:

- Create `docs/internal-beta/` and draft the seven files with short, checklist/table-oriented sections.
- Reuse links to `docs/install/macos.md`, `docs/clients/claude-desktop.md`, `docs/clients/claude-code.md`, and `docs/coach-mode.md` rather than copying setup JSON.
- Add a single shared measurement table schema in `measurement.md`, then copy only the empty header/table skeleton into `findings.md`.
- Ensure `README.md` and `checklist.md` link to all other internal-beta docs in execution order.
- Update `STATUS.md` after drafting to reflect each completed file and leave verification commands for Step 3.

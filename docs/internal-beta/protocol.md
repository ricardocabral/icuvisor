# v0.5 internal beta protocol

Purpose: recruit a small, consented beta group to validate v0.5 prep against PRD §4 KR1, §5 target users, and §7.4 #6/#8/#12. TP-041 only creates this protocol; the maintainer runs it later.

## Cohort boundary

- Target size: 5-10 participants total.
- Recruitment window: 14 calendar days from the maintainer's first post/DM.
- Segments: self-coached endurance athletes first; include 1-3 coaches only if TP-039 coach-mode setup is ready to support them.
- Out of scope during prep: posting notices, inviting people, collecting data, filling findings, or changing PRD assumptions.

## Eligibility screen

Ask each candidate:

| Filter        | Eligible answer                                                                                      |
| ------------- | ---------------------------------------------------------------------------------------------------- |
| OS            | macOS user willing to install a signed DMG.                                                          |
| AI client     | Uses or can install Claude Desktop or Claude Code.                                                   |
| intervals.icu | Has their own intervals.icu account and API key.                                                     |
| Coach mode    | Athlete-only is fine; coach candidates must manage a small roster and accept redacted feedback only. |
| Mobile need   | Can answer whether mobile/tray access would change daily use (PRD §7.4 #8).                          |
| Availability  | Can complete install + first call and a short exit interview inside the beta window.                 |

## Recruitment script

> I'm preparing a small v0.5 icuvisor internal beta for intervals.icu users. The goal is to see whether a local MCP server can be installed and used quickly with Claude Desktop/Code, and whether the workflow is useful enough for daily training questions. The beta is limited to 5-10 people and will run for about two weeks after recruitment closes. You keep your own data and API key locally; I only ask for timing, tool-call names/descriptions, blockers, and qualitative feedback. Would you be willing to try the signed macOS build and do a short exit interview?

## Consent statement

Before onboarding, send and receive explicit agreement to:

- What the maintainer may receive: install-to-first-call timing, top tool-call names/timestamps or descriptions, mobile-need answer, qualitative notes, blockers filed, and `icuvisor diagnostics` output only when voluntarily shared.
- What the maintainer must not receive: raw training data, API keys, athlete IDs, tool arguments, tool payloads, screenshots containing values, full transcripts, or coach roster names/IDs.
- Storage: feedback is summarized in `findings.md` using participant labels like `P01`; no identifying athlete values are recorded.
- Revocation: a participant can stop at any time by replying "withdraw"; remove any unpublished notes tied to that participant and stop follow-up.

## Enrollment record

Track only this minimal roster outside the repo while running the beta:

| Participant label | Segment | OS/client | Coach-mode? | Consent received? | Onboarded? |
| ----------------- | ------- | --------- | ----------- | ----------------- | ---------- |
| P01               |         |           |             |                   |            |

Do not commit the live roster or contact details to the repository.

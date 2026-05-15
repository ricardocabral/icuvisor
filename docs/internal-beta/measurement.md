# v0.5 internal beta measurement plan

Collect evidence manually. TP-041 does not add telemetry, run the cohort, or populate findings.

## PRD anchors

| Anchor | What this beta measures | Manual procedure |
| --- | --- | --- |
| KR1 (§4) | Time from install start to first successful MCP tool call. | Operator starts a timer at DMG install step 1 and stops after first successful icuvisor tool response. |
| PRD §7.4 #6 | Demand / willingness to keep using and recommend. | Ask exit-interview recommendation and daily-use questions; record a short signal. |
| PRD §7.4 #8 | Whether mobile/tray access is needed for daily usefulness. | Ask the mobile-need question during screening and exit interview. |
| PRD §7.4 #12 | Whether schema-change/catalog guidance is understandable. | During update exercise or discussion, ask if catalog-hash/schema-change guidance was clear. |

## What to collect

- Participant label only (`P01`, `P02`); no names, athlete IDs, or account IDs in repo files.
- Segment: `self-coached`, `coach`, or `other-qualified`.
- Install-to-first-call minutes for KR1.
- Top 5 tool-call names with timestamps/descriptions only; no arguments or payloads.
- Mobile-need answer for PRD §7.4 #8.
- Willingness-to-recommend / demand signal for PRD §7.4 #6.
- Schema-change notification clarity for PRD §7.4 #12.
- Free-text surprises with values redacted.
- Blockers filed or follow-up issue IDs.

## Table schema

Copy this table into [findings.md](findings.md) and fill one row per participant during the run.

| Participant | Segment | Client | Coach mode? | Install-to-first-call minutes (KR1) | Top 5 tool calls (names/timestamps only) | Mobile need (§7.4 #8) | Recommend/demand (§7.4 #6) | Schema-change clarity (§7.4 #12) | Surprises (redacted) | Blockers filed |
| --- | --- | --- | --- | ---: | --- | --- | --- | --- | --- | --- |

## Operator notes

- If a participant shares diagnostics, store only the redacted `icuvisor diagnostics` output or summarize it.
- If a participant accidentally sends sensitive values, remove them from notes before copying into findings.
- If a blocker requires product work, file or link an issue after the beta; do not expand TP-041 scope.

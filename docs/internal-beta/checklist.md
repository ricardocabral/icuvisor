# v0.5 internal beta operator checklist

One-page runbook. Start here, then open linked docs only when a step needs detail.

## Recruit

- [ ] Open the [protocol](protocol.md); set a 14-day recruitment window and a 5-10 participant cap.
- [ ] Send the recruitment script to eligible self-coached athletes and optional coach-mode candidates.
- [ ] Screen OS, MCP client, intervals.icu account/API-key readiness, coach-mode fit, mobile-need signal, and availability.
- [ ] Capture explicit consent; do not store names, athlete IDs, API keys, raw data, or contact details in this repo.

## Onboard

- [ ] Open the [onboarding playbook](onboarding-playbook.md).
- [ ] Guide DMG install, `icuvisor setup`, linked Claude Desktop/Code config, client restart, and first-call verification.
- [ ] For coaches, use the coach-mode variant and record only redacted categories.
- [ ] Record KR1 install-to-first-call minutes in [findings](findings.md).

## Run

- [ ] Use [measurement](measurement.md) to log top tool-call names/timestamps/descriptions only.
- [ ] Ask mobile-need (§7.4 #8), demand/recommendation (§7.4 #6), and schema-change clarity (§7.4 #12) questions.
- [ ] For support, request `icuvisor diagnostics` output only when voluntarily shared.
- [ ] File blockers after removing secrets, athlete IDs, payloads, arguments, and raw training values.

## Synthesize

- [ ] Run the [exit interview](exit-interview.md) for each completed or withdrawn participant.
- [ ] Fill [findings](findings.md) with one anonymous row per participant.
- [ ] Summarize KR1, PRD §7.4 #6/#8/#12 signals, top blockers, and follow-up issues.
- [ ] Verify the redaction checklist before sharing findings.

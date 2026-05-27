# Plan Review — Step 3: Link from relevant docs and update changelog

**Verdict: approved to proceed.**

The Step 3 plan is small and consistent with the approved privacy boundaries: add contextual incoming links to the new privacy page, keep `SECURITY.md` authoritative, and add a concise `[Unreleased]` changelog note if docs changed. This should complete discoverability without changing runtime behavior or reopening TP-113 homepage/local-first positioning work.

## Implementation guardrails

- Prefer narrow cross-links over new privacy/security prose. Good targets are `coach-mode.md` and `http-transport.md`; a one-line link from `local-first.md` is acceptable only if it does not rewrite the TP-113-owned positioning.
- Do not claim GDPR compliance/certification or provide legal advice. Keep language as privacy posture / due-diligence framing.
- Leave `SECURITY.md` as the authority for vulnerability reporting, release integrity, and hardening details; link to it rather than duplicating policy text.
- Avoid homepage and broad README changes unless there is an explicit supervisor instruction, because the task is proceeding under the constrained no-TP-113-overlap pass.
- If linking from `safety-modes.md`, keep it clearly tied to write/delete trust boundaries; do not force a privacy link where it reads unrelated.
- Add a short `CHANGELOG.md` entry under `[Unreleased]` (likely `Added`) noting the privacy posture documentation and contextual links.

## Notes

Step 3 should be limited to docs/linkage and changelog. Verification belongs in Step 4 (`make web-build` and rendered-link checks).

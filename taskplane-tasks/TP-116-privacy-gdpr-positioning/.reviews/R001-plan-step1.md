# Plan Review — Step 1: Define privacy claims and boundaries

**Verdict: request changes before proceeding.**

The Step 1 checkpoint is not ready for approval: `STATUS.md` contains only the generic checklist, with no source-backed claim inventory, non-claim boundaries, or page/linking decision. The task explicitly lists `STATUS.md` discoveries/plan notes as the Step 1 artifact.

## Blocking gaps

1. **Dependency TP-113 is not satisfied.**
   - `PROMPT.md` says TP-113 should land first to avoid overlapping homepage/privacy copy.
   - `taskplane-tasks/TP-113-local-first-positioning-refresh/STATUS.md` is still `Ready for Execution` / `Not Started`.
   - Step 0 marks “Dependencies satisfied” as complete, but that is contradicted by the dependency status. Either block this task until TP-113 lands or document explicit human approval to proceed and constrain this plan to avoid overlapping copy.

2. **No Step 1 artifact was produced in `STATUS.md`.**
   - The discoveries table is empty and notes are reserved.
   - The plan does not inventory substantiated claims from `SECURITY.md`, existing explain pages, HTTP transport docs, README/homepage, or code/config behavior.
   - Without this, Step 2 risks introducing unverified privacy/GDPR language.

3. **The required non-claims are not concretely defined.**
   - The plan should explicitly record wording boundaries such as: not legal advice, not certified/GDPR-compliant, selected AI client may process prompt/tool-result content, and intervals.icu remains governed by its own terms/privacy role.
   - These need to be documented before drafting user-facing copy.

4. **No page strategy decision is recorded.**
   - Step 1 must decide whether to add `web/content/explain/privacy.md` or strengthen existing local-first/coach-mode pages.
   - The decision should include the intended link points and how to avoid duplicating `SECURITY.md`.

## Required revision

Update `STATUS.md` before implementation with:

- A source-backed inventory of allowed claims and file locations.
- Explicit prohibited/non-claim language boundaries.
- Dependency disposition for TP-113.
- A clear page/linking strategy for Step 2/3.

After that, resubmit the Step 1 plan review.

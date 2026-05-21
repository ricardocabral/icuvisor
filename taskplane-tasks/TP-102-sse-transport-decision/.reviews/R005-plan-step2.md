# Review R005 — Plan review for Step 2

**Verdict:** REVISE

The Step 2 plan is not concrete enough yet. `STATUS.md` still only contains the original high-level Step 2 checklist, and I do not see a worker-specific plan that says how the product decision will be made, recorded, and approved. Because this step is the product decision/approval boundary for a transport/security change, the plan needs more detail before execution.

## Required changes

1. **State the proposed decision and decision basis up front.**
   - Step 1 recommends **Path B** with medium-high confidence. The Step 2 plan should say whether it intends to accept that recommendation or seek an operator override.
   - Include the short rationale that will be recorded: current MCP/OpenAI guidance accepts Streamable HTTP, the missing capability is secure public reachability rather than legacy SSE, and generic tunnels expose an unauthenticated personal intervals.icu control surface.

2. **Define the approval/decision-record mechanism.**
   - Specify where the final decision will be recorded in `STATUS.md` (for example, a new “Step 2 decision record” note plus an execution-log entry).
   - Specify what counts as operator approval if approval is required: approver identity/source, timestamp, and exact approved path.
   - If no additional operator approval is needed for Path B because it aligns with the existing PRD “No SSE” contract, say that explicitly. If Path A is chosen, require explicit operator approval before any protected product-doc changes.

3. **Make the PRD/roadmap conflict check explicit for both paths.**
   - For **Path B**, the plan should note that it appears consistent with the PRD “No SSE” language, so Step 2 should not edit the PRD; later steps should document the limitation and update roadmap/changelog status as required.
   - For **Path A**, the plan must treat this as a product-scope change that conflicts with current PRD language and must prepare an explicit PRD/roadmap update proposal before implementation.

4. **Keep Step 2 scoped to decision/approval, not implementation.**
   - The plan should state that Step 2 will only update `STATUS.md` with the decision and approval/conflict assessment.
   - Web docs, `CHANGELOG.md`, `ROADMAP.md`, and any transport code should remain for Step 3/4 unless the operator explicitly asks otherwise.

5. **Define acceptance criteria for the step.**
   At minimum, Step 2 should be considered complete only when `STATUS.md` contains:
   - chosen path A or B;
   - concise rationale linked to Step 1 evidence;
   - PRD/roadmap conflict assessment;
   - approval status/approver or an explicit “not required” justification;
   - next-step instructions for Step 3.

## Non-blocking suggestion

For traceability, use a small decision table in `STATUS.md` with columns like `Decision`, `Rationale`, `PRD impact`, `Approval`, and `Step 3 implication`. That will make later code/doc reviews easier and avoid re-litigating the research findings.

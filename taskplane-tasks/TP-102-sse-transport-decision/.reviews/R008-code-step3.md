# Review R008 — Code review for Step 3

**Verdict:** REVISE

The new ChatGPT documentation text is directionally aligned with Path B, but Step 3 cannot be approved while the task status records rejected reviews as approved and the required troubleshooting doc coverage remains incomplete.

## Findings

1. **Blocking — Review bookkeeping is false for R007 (and still false for R005/R006).**  
   `taskplane-tasks/TP-102-sse-transport-decision/STATUS.md:96-98` records R005, R006, and the newly added R007 as `APPROVE`, and `STATUS.md:120-122` logs those outcomes as approvals. The artifacts say otherwise: `.reviews/R005-plan-step2.md:3`, `.reviews/R006-code-step2.md:3`, and `.reviews/R007-plan-step3.md:3` all say `**Verdict:** REVISE`. This is not just historical noise: the Step 3 plan review artifact added in this diff explicitly rejected the plan, but the status table claims it approved execution. Correct the review table/execution log to match the artifacts, or add subsequent approving review artifacts and record those accurately before closing Step 3.

2. **Blocking — Step 3 does not address the requested HTTP troubleshooting/security documentation.**  
   The only user-facing doc changed is `web/content/connect/chatgpt.md:51-57`. The Step 3 plan review required `web/content/guides/http-transport.md` to get troubleshooting/security language so remote-connector failures are not mistaken for an HTTP transport bug, but that file is unchanged and its troubleshooting table still only covers local/LAN connection symptoms. Add a short troubleshooting row/section explaining that ChatGPT-style remote connector UIs cannot reach `127.0.0.1`, require a provider-reachable HTTPS endpoint, and are intentionally unsupported until the hosted relay or a future secure-tunnel design.

## Notes

- The new `chatgpt.md` section correctly states the reachability limitation, safety rationale, no legacy SSE/generic tunnel recipe, and the vNext/future secure-tunnel path.
- No Go code changed, so I did not run build/test/lint. I reviewed `git diff 76ca085..HEAD`, `PROMPT.md`, `STATUS.md`, the changed ChatGPT doc, the HTTP transport guide, and the relevant review artifacts.

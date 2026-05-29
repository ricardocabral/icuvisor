# Review R009 — Plan review for Step 3

**Verdict:** APPROVE

The revised Step 3 plan is safe to proceed. It keeps the docs work focused on the plan-health workflow, makes the Step 2 prompt/MCP carry-over check explicit, adds the affected prompt-reference and cookbook index pages, and includes the key caveats from the prompt contract: deload interpretation, race-date handling, incomplete wellness/readiness data, and no opaque scoring/writes.

## Notes

- When implementing, also check whether `web/content/cookbook/prompt-library.md` needs a short copy-paste prompt for plan-health review. This is not a blocker because the main cookbook/reference targets are now named, but it is an obvious affected page in the cookbook section.
- The targeted verification list is appropriate: `go test ./internal/prompts ./internal/mcp` plus `make web-build` covers the known prompt-list blast radius and the changed docs pages. Full `make test`, lint, and build can remain Step 4 gates.

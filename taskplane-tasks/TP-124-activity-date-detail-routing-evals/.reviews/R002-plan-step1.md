# Review R002 — Plan Step 1

Verdict: Approved

The revised Step 1 plan addresses the prior blocker. It now includes the split-routing surface (`get_activity_splits`, implemented in `internal/tools/get_activity_streams.go`) in addition to activities, details, intervals, cookbook/prompt guidance, and eval scenarios.

The plan is appropriate for a discovery-only mapping step because it:

- covers the date lookup entry point (`get_activities`) and downstream ID-based tools (`get_activity_details`, `get_activity_intervals`, `get_activity_splits`);
- includes cookbook prompts, prompt testdata, and existing eval scenarios, which are the likely routing-hint surfaces for “last Sunday” and split/rep prompts;
- keeps changes limited to recording gaps and chosen changes in `STATUS.md` before implementation in later steps;
- retains the targeted test command: `go test ./internal/tools ./internal/prompts`.

Non-blocking reminder: when logging Discoveries, keep the categories separate as noted in `STATUS.md` Notes: athlete-local date resolution, list→ID detail/interval/splits routing, and split-vs-interval wording.

# Plan Review: Step 1 — Add `weekly_review` prompt registration

Verdict: **Needs revision before implementation**

The updated STATUS addresses the main R001/R002 planning gaps: Step 1 now includes `registry.go`, non-golden `catalog_test.go` updates, the targeted `go test ./internal/prompts` command, and fallback guidance for full-toolset helpers. One completion-criteria gap remains.

## Required revision

1. **Add wellness/staleness to the Step 1 prompt plan.**
   The task completion criteria require the prompt guidance to cover wellness staleness, but the planned weekly-review tool sequence currently lists profile/timezone, fitness/summary, activities, events/training plan, analyzer/compute tools, and fallback guidance without `get_wellness_data` or an explicit staleness/provenance instruction. Because the prompt text is authored in Step 1, this cannot be deferred to Step 2 golden coverage.

   Please revise the Step 1 plan to include:
   - `get_wellness_data` in the advertised tool set/sequence, and
   - an instruction to check `_meta.stale`, missing/provenance warnings, and not infer readiness/sleep/HRV when wellness data is stale or absent.

## Minor clarification before coding

- Name the compliance helper exactly as `compute_compliance_rate` if it is part of the prompt. The current “compute ... compliance or available equivalents” wording is a bit loose, and exact tool names matter in prompt catalogs and golden tests.

After these changes, the plan should be ready to implement.

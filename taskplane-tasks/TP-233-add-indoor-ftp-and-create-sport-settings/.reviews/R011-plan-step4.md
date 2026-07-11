# Plan Review — TP-233 Step 4

## Verdict: REVISE

The revised plan covers the generator goldens, public create/update boundaries, changelog, and uncached `cmd/gendocs` verification from R010. It still omits the existing PRD statements that call themselves the current generated catalog count. The registry now deterministically produces 70 tools (30 `core`, 40 `full`), while the PRD still states 69/39 at line 318 and 36 full tools at lines 342 and 420. Leaving those untouched makes the public contract self-contradictory after this step.

Revise the PRD work item to update every current-catalog count statement to 70 total, 30 `core`, and 40 additional `full` tools, alongside adding `create_sport_settings` and the indoor-FTP semantics. Then regenerate both website JSON files and both `cmd/gendocs/testdata/*.golden.json` fixtures from the same registry output and run the already planned uncached generator/catalog/toolcheck tests.

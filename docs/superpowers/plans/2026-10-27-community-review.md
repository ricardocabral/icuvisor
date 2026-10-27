# Community review improvements implementation plan

User authorization: implement all findings from the October 27 review except item 2 (Garmin-origin data availability); use parallel subagents and worktrees. This document records implementation choices under that authorization, not a new approval gate.

Base: 1b45c0f8. Integration worktree: /Users/jusbrasil/.codex/worktrees/a511/icuvisor; branch feat/community-review-improvements. Original checkout must remain unchanged.

## Global constraints
- No Garmin-origin availability, workaround, integration, or policy work.
- Clean-room: public API specification and black-box behavior only; no competitor implementation source, patches, or copied tests.
- Stdlib and existing dependencies; terse default, raw data only include_full; additive stable argument schemas; explicit absence, units and provenance.
- Existing capability/delete-mode and per-athlete authorization apply to every new operation. No credentials in tools or output. No live athlete writes for testing.
- Implement concrete behavior from the review; do not invent physiology models, automatic coaching, or a new autonomous calendar planner. Existing roadmap boundaries remain.
- Tests exercise real production behavior via fixture-backed interfaces/HTTP servers; record red/green evidence. Focused tests per lane, full/race/lint/build/generated-catalog checks on integration.
- Root owns central registration/catalog wiring, generated schemas/docs, CHANGELOG.md, README.md, PRD/roadmap integration, and website files. Workers report needed wiring and never write those files unless specifically assigned.
- Each worker uses ONLY its assigned native worktree, creates its own conventional branch and commits, does not delegate, and writes its report under /private/tmp/icuvisor-review-20261027/.

## Task 1: Activity comments and writable metadata (worker activity)
Fix deleted timestamp decoding while retaining compatibility of existing boolean response if appropriate and exposing deletion timestamp explicitly. Add edit/delete own activity-comment operations with activity membership, upstream authorization, existing safety requirements, recovery evidence, and no cross-athlete leaks. Extend update_activity with selected publicly verified writable fields (tags, commute, feel/RPE, subtype, distance as supported); preserve missing/zero/clearing semantics and avoid arbitrary field passthrough. Validate manual-vs-file time behavior before exposing time corrections. Tests for live+deleted notes, null/absent/deleted timestamp, wrong activity message, no chat, already deleted, 403/404, sparse writes/clearing/units. Files: internal/intervals activity/messages client/model files, internal/tools activity message/update_activity files and tests. Root wires tools.

## Task 2: Calendar constraints and write evidence (worker calendar)
Expose training_availability, can_train_sports, max_training_time in terse event reads and reused planning context, preserve empty/zero/omitted values. Verify planned duration consistency after writes (text-only WeightTraining collapsing to warm-up is key case), with explicit mismatch/unavailable evidence rather than false success; preserve existing write-ID/retry contracts. Add focused regressions for calendar constraints, interrupted/repeated writes, newer user edits when retrying where applicable, and readback duration. Harden existing plan/prompt boundaries against ignored explicit restrictions, illness/planned state mismatch, and requested counts vs slots without building deferred autonomous planner or physiology policy. Files: internal/intervals/events*, internal/tools event/planning files, internal/planning and relevant prompt tests/text if necessary. Root owns catalog and published docs.

## Task 3: Curves and semantic API drift (worker curves)
Add convenient bounded season comparison to existing curve surface while preserving required oldest/newest stable arguments (or separate additive tool if necessary), current date-range and after-kJ behavior. Expose compact effort dates and correct W/kg/activity attribution when public upstream fields support it; include upstream model numbers only when proven, not locally invented estimates. Test athlete-local boundaries, absent seasons, best W/kg activity differing from peak watts, long durations, fractional model values and missing data. Current season context exists, SEASON_START generic events work; no documented /seasons route. Also extend existing openapidiff to deterministic property-level additions/removals/type changes in existing schemas (avoid huge semantic diff framework), with fixture tests and maintained documentation. Files: internal/intervals/fitness*, internal/tools curve files, scripts/openapidiff*, docs/maintenance-openapi-endpoint-diff.md. Root wires any new tool.

## Task 4: Shared workout-library discovery (next available worker)
Validate and implement public GET /api/v1/folder-search and POST /api/v1/athlete/{id}/folders/{folderId}/share-with-me with typed bounded inputs, concise summaries, explicit read/write capability and authorization. Sharing is an explicit operation separate from search; no automatic side effects. Inspect exact public API required fields and response. Unit tests include limits/pagination, upstream errors, wrong athlete, capability gates, terse/full. Root approves product scope by reflecting user's explicit authorization in PRD/catalog docs and handles registration.

## Task 5: Wellness and workout regression coverage (next available worker)
Inspect and improve HRV baseline comparability for provider/device changes without mixing RMSSD and SDNN or inventing conversions. Expose gaps/continuity caveats where provenance actually exists; don't infer device identity. Add meaningful regressions for explicit source changes, stale/absent data, subjective vs wearable conflict visibility. Verify repeated-workout evidence carries activity ID/date and mixed sports refuse comparison; GPS pair regression and target priority/repeat boundaries already pass, add only uncovered meaningful cases. Root coordinates any overlapping files from Tasks 1-3.

## Task 6: Website/cookbook/client setup (root)
Promote own-assistant/local control, race splits and dated comparisons, clarify full-toolset requirements, improve client setup/capability comparison and missing ChatGPT-option troubleshooting with current official sources. Correct generic cross-platform signed claim to self-contained binary; keep platform-specific signing docs accurate. Avoid unsupported speed/unlimited/hosted pricing claims. Add concrete evidence/constraints guidance to existing planning recipes and plan-health prompts. No redesign, no Garmin changes.

## Task 7: Integration and verification (root)
Review every lane's diff and evidence, cherry-pick commits onto integration branch, wire new tools using existing safety/coach/client interfaces, update PRD catalog/roadmap pointers where scope changes, README/CHANGELOG and generated schema/catalog artifacts. Run gofmt/goimports, make check, docs guidance/evals, Hugo build and targeted rendered browser QA for copy changes. Independent final whole-branch review, fix findings, leave reviewable branch/worktrees without merge/push/publish.

## Preflight conflict scan
| Tasks | Shared surface | Decision |
|---|---|---|
| 1, 2, 3, 4 | registry/catalog/generated docs | Root owns wiring and generation after integration. |
| 1, 5 | wellness/subjective schema provenance | Task 1 owns activity writes, Task 5 owns wellness baseline reads. |
| 2, 6 | prompts/plan docs | Task 2 reports behavioral contract needs; root edits published recipes. |
| 3, 4 | library and fitness client interfaces | Separate client files; root reconciles concrete-client catalog wiring. |
| all | changelog/README/PRD/roadmap | Root owns combined updates. |
| all | scope versus future planner/physiology | Existing read/prompt guards and tests now; preserve deferred autonomous planning scope. |

Every task agrees with its behavioral tests and assigned files. User explicitly chose parallel worktrees, overriding sequential-dispatch process defaults. User explicitly authorized all non-Garmin findings; scope decisions are documented here and in final report rather than reopening approval ceremony.

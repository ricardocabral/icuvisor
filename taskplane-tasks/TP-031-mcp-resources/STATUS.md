# TP-031-mcp-resources: TP-031-mcp-resources — Status

**Current Step:** Step 3: `icuvisor://event-categories`
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-14
**Review Level:** 2
**Review Counter:** 9
**Iteration:** 1
**Size:** M

---

### Step 1: Resource registration plumbing

**Status:** ✅ Complete

- [x] Wire `resources/list` and `resources/read` into the MCP server via the Go SDK
- [x] Define a small internal interface so each resource is one greppable registration, mirroring the tool registry pattern
- [x] Decide static vs dynamic per resource; document in `STATUS.md`

### Step 2: `icuvisor://workout-syntax`

**Status:** ✅ Complete

- [x] Register `icuvisor://workout-syntax` in the default resource registry
- [x] Content derived from the `internal/workoutdoc` grammar — do not hand-author a second copy that can drift
- [x] Covers every step/target type the serializer supports; a test asserts coverage parity with `workoutdoc`
- [x] R006: Make the parity source non-self-referential by driving serializer-supported forms/units from shared `workoutdoc` data used by docs/tests
- [x] R006: Fix resource-handler lint issue from `fmt.Errorf(genericResourceErrorMessage)`

### Step 3: `icuvisor://event-categories`

**Status:** 🟨 In Progress

- [x] Register `icuvisor://event-categories` in the default resource registry
- [x] Full event-category enum with one-line descriptions, sourced from the same enum the event tools use
- [x] Static content; golden-file locked

### Step 4: `icuvisor://custom-item-schemas`

**Status:** ⏳ Not started

### Step 5: `icuvisor://athlete-profile`

**Status:** ⏳ Not started

### Step 6: Trim inline tool descriptions

**Status:** ⏳ Not started

### Step 7: Verify

**Status:** ⏳ Not started

---

## Reviews

| #   | Type | Step | Verdict | File |
| --- | ---- | ---- | ------- | ---- |
| R001 | Plan | 1 | REVISE | .reviews/R001-plan-step1.md |
| R002 | Plan | 1 | APPROVE | inline |
| R003 | Code | 1 | APPROVE | inline |
| R004 | Plan | 2 | REVISE | .reviews/R004-plan-step2.md |
| R005 | Plan | 2 | APPROVE | inline |
| R006 | Code | 2 | REVISE | .reviews/R006-code-step2.md |
| R007 | Code | 2 | APPROVE | inline |
| R008 | Plan | 3 | REVISE | .reviews/R008-plan-step3.md |
| R009 | Plan | 3 | APPROVE | inline |

---

## Discoveries

| Discovery | Disposition | Location |
| --------- | ----------- | -------- |
| `workoutdoc` has implicit serializer grammar but no exported syntax descriptor yet; pace text is supported only as non-ramp text ending in `Pace`, while text targets cannot be ramps. | Add a `workoutdoc` syntax descriptor in Step 2 and generate the resource from it. | `internal/workoutdoc/{types,serialize,parse}.go` |
| Event tools currently trim/pass through arbitrary category strings and do not share an enum descriptor; Intervals public OpenAPI `GET https://intervals.icu/api/v1/docs` exposes `Event`/`EventEx.category` enum values `WORKOUT`, `RACE_A`, `RACE_B`, `RACE_C`, `NOTE`, `PLAN`, `HOLIDAY`, `SICK`, `INJURED`, `SET_EFTP`, `FITNESS_DAYS`, `SEASON_START`, `TARGET`, `SET_FITNESS`. | Add a shared descriptor but do not turn it into validation; preserve pass-through/custom category behavior. | `internal/intervals/events.go`; public OpenAPI |

---

## Execution Log

| Timestamp  | Action      | Outcome                          |
| ---------- | ----------- | -------------------------------- |
| 2026-05-14 | Task staged | Scaffolded from ROADMAP.md v0.4   |
| 2026-05-14 14:09 | Task started | Runtime V2 lane-runner execution |
| 2026-05-14 14:09 | Step 1 started | Resource registration plumbing |

---

## Blockers

_None_

---

## Notes

### Step 1 plan

- SDK reference consulted: https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk/mcp#Server.AddResource (go-sdk v1.4.1 resource registration API).
- Use the SDK Resource API only: register resources with `(*mcp.Server).AddResource(...)` so the SDK owns `resources/list`, `resources/read`, capabilities, pagination, and list-change behavior; do not add custom JSON-RPC handlers.
- Define a registry boundary that mirrors the tool registry while keeping domain resource definitions out of SDK types. `internal/resources` will expose small resource definitions/registries; `internal/mcp` will validate and convert them to SDK resource registrations.
- Validate registry entries before server startup completes: absolute `icuvisor://...` URI, no duplicate URIs, non-empty name/title/description/MIME type, and non-nil read handlers/content readers. Invalid registration returns a `NewServer` error, never a panic.
- Wire the registry through MCP server construction options so `internal/app` can pass the default resource registry alongside the tool registry. Add resources before sessions initialize so `initialize` advertises the resources capability.
- Handler failures are logged and returned as short, safe client-facing errors. Unknown/unregistered resources keep the SDK not-found protocol behavior.
- Step 1 content contract: read handlers return one text `ResourceContents` item with populated `URI`, `MIMEType`, and `Text`; per-resource steps may override MIME type, with long-form docs defaulting to `text/markdown`.
- Static/dynamic decisions: `icuvisor://workout-syntax` is static/derived from `internal/workoutdoc`; `icuvisor://event-categories` is static from the same enum/source as event tools; `icuvisor://custom-item-schemas` is static/derived from custom-item validation/schema sources; `icuvisor://athlete-profile` is dynamic cached content with TTL/staleness policy finalized in Step 5.
- Protocol tests for Step 1 will use in-memory MCP client helpers to assert initialize advertises resources when configured, list returns metadata, read dispatches with URI/MIME/text, invalid/duplicate registrations fail server construction, and unknown reads return the SDK not-found protocol error.

### Step 2 plan

- Add a small exported `internal/workoutdoc` syntax/spec descriptor as the single source for resource docs and parity tests. The descriptor will list step forms, target families, supported units/aliases, examples, and limitations that match `Serialize`/`Parse`; `internal/resources` will generate Markdown from this descriptor instead of storing a standalone Markdown grammar.
- Add `internal/resources/registry.go` with `NewRegistry`/default registry construction and one greppable `WorkoutSyntaxResource` registration. Wire `internal/app` to pass this registry to `mcp.NewServer` so normal server runs advertise the resource.
- Resource contract: URI `icuvisor://workout-syntax`, name `workout_syntax`, title `Workout syntax`, short description for the intervals.icu structured-workout DSL, MIME type `text/markdown`, static/no-network read handler that checks context cancellation, one text content item with URI/MIME/text populated.
- Markdown content will document serializer coverage: duration and distance steps (`mtr`, `km`, `mi` canonical output), repeat blocks and no nested repeats, free-ride steps, ramps using `start`/`end`, optional cadence (`rpm`), power targets (`%FTP`, watts, power zones, scalar/range), HR targets (`%HR`, `%LTHR`, bpm, HR zones, scalar/range), pace targets (percent threshold pace, pace zones, `PACE` numeric/text handling as actually supported), and RPE scalar/range.
- Markdown content will also document limitations enforced by the serializer: one primary target per step, ramp requires a primary target and cannot use text targets, freeride cannot combine with ramp, repeat blocks cannot also carry simple-step fields, repeat blocks require reps/children, and simple steps require duration or distance.
- Coverage parity tests will table-drive representative `workoutdoc.Step` values through `workoutdoc.Serialize`, compare their generated DSL snippets to descriptor examples, and assert each documented feature key is rendered into the Markdown. New serializer-supported families/units should require adding a descriptor entry and corresponding expected Markdown.
- Add deterministic tests: a golden/snapshot-style generated Markdown test under `internal/resources/testdata`, plus protocol/registry assertions that `resources/list` exposes `icuvisor://workout-syntax` and `resources/read` returns the generated Markdown as `text/markdown`.
- Planned file layout: `internal/workoutdoc/syntax.go` for descriptors and generated examples, `internal/resources/registry.go`, `internal/resources/workout_syntax.go`, `internal/resources/workout_syntax_test.go`, `internal/resources/testdata/workout_syntax.md`, and small app/protocol test updates. README/tool-description trimming remains Step 6; CHANGELOG will be updated once resources are documented in final steps.

### R006 revision notes

- Code review found the first syntax descriptor was still self-referential; revise by moving supported unit/alias matrices into exported `workoutdoc` syntax data and using those matrices in serializer formatting plus resource tests.
- Code review also noted a lint failure in Step 1 resource error handling; replace `fmt.Errorf(genericResourceErrorMessage)` with a non-format error.

### Step 3 plan

- Add a shared event-category descriptor in `internal/intervals` (for example `EventCategories() []EventCategory`) containing the public OpenAPI `Event`/`EventEx.category` enum and one-line descriptions. Event-facing code and the resource will consume this descriptor; generated Markdown will not be the source of truth.
- Upstream evidence: public intervals.icu OpenAPI document at `https://intervals.icu/api/v1/docs`, `components.schemas.Event.properties.category` and `EventEx.properties.category`, fetched during planning without consulting GPL sources. Scope is the calendar event category enum including fitness-model calendar categories (`SET_EFTP`, `FITNESS_DAYS`, `SET_FITNESS`); `WithCourses.category` and `CategorySummary.category` are different schemas and out of scope.
- Preserve current tool behavior: `get_events` filters and `add_or_update_event` writes still trim/pass through caller category values and preserve upstream/custom values in responses. The descriptor is documentation/schema metadata only and must not reject athlete/account-specific category strings.
- Resource contract: URI `icuvisor://event-categories`, name `event_categories`, title `Event categories`, short description, MIME type `text/markdown`, static/no-network read handler that checks context cancellation, deterministic OpenAPI order, and one text content item with URI/MIME/text populated.
- Register `EventCategoriesResource()` in `resources.NewRegistry()` alongside `WorkoutSyntaxResource()` so normal server runs expose it in `resources/list` and `resources/read`.
- Tests: golden-lock generated Markdown; assert every descriptor entry appears exactly once with a non-empty one-line description; assert event tool schemas/descriptions reference the shared descriptor/resource without enum validation; add registry/protocol assertions for list/read URI and MIME type.
- Step 6 boundary: avoid broad tool-description trimming and README updates here; only minimal wording/schema metadata changes needed to point category docs at the shared descriptor/resource.

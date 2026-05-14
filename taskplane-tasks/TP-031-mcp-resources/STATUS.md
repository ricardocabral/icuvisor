# TP-031-mcp-resources: TP-031-mcp-resources — Status

**Current Step:** Step 5: `icuvisor://athlete-profile`
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-14
**Review Level:** 2
**Review Counter:** 15
**Iteration:** 2
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

**Status:** ✅ Complete

- [x] Register `icuvisor://event-categories` in the default resource registry
- [x] Full event-category enum with one-line descriptions, sourced from the same enum the event tools use
- [x] Static content; golden-file locked
- [x] R010: Update public event write schema examples to use documented race categories and add a guard against descriptor/example drift

### Step 4: `icuvisor://custom-item-schemas`

**Status:** ✅ Complete

- [x] Register `icuvisor://custom-item-schemas` in the default resource registry
- [x] Per-`item_type` schema for the `content` field (chart/field/stream/panel/zones)
- [x] Reuses the schema samples the custom-item reads/writes already validate against — single source of truth
- [x] Golden-file locked
- [x] R014: Remove library panic from static custom-item sample construction
- [x] R014: Render and test per-`item_type` schemas via concrete samples or explicit aliases

### Step 5: `icuvisor://athlete-profile`

**Status:** 🟨 In Progress

- [ ] Add a shared athlete-profile shaper used by both `get_athlete_profile` and the resource
- [ ] Register `icuvisor://athlete-profile` as a dynamic cached resource with documented TTL/staleness behavior
- [ ] Cover resource list/read, cache refresh, context cancellation, and shape parity with focused tests

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
| R010 | Code | 3 | REVISE | .reviews/R010-code-step3.md |
| R011 | Code | 3 | APPROVE | inline |
| R012 | Plan | 4 | REVISE | .reviews/R012-plan-step4.md |
| R013 | Plan | 4 | APPROVE | inline |
| R014 | Code | 4 | REVISE | .reviews/R014-code-step4.md |

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
| 2026-05-14 16:09 | Worker iter 1 | killed (wall-clock timeout) in 7200s, tools: 229 |
| 2026-05-14 16:09 | Step 5 started | `icuvisor://athlete-profile` |

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

### R010 revision notes

- Public `add_or_update_event` input examples should not advertise `RACE`, because the documented upstream enum now points users to priority-specific `RACE_A`/`RACE_B`/`RACE_C` values. Keep runtime pass-through behavior, but make public examples and schema guards align with `intervals.EventCategoryValues()`.

### Step 4 plan

- Move reusable custom-item content schema primitives out of `internal/tools` into a small shared package (planned `internal/customitemschemas`): schema inference from sample `content`, value validation, missing-key detection, JSON kind/path rendering, and static descriptor/sample metadata. Update create/update custom-item write validation to call that package so live validation and the resource share the same inference/validation machinery.
- Keep the resource static/golden-locked: the package will include representative static `content` samples for documentation, while live create/update continues to fetch readable custom items for the target athlete/item and validate against those live samples. The resource is general guidance and must explicitly say it is not a validation allow-list or replacement for live readable-schema validation.
- Document item-type families and concrete known upstream values from existing tool descriptions, dogfood fixtures, and black-box/read-side samples: chart/table/trace (`FITNESS_CHART`, `FITNESS_TABLE`, `TRACE_CHART`, `ACTIVITY_CHART`, `ACTIVITY_HISTOGRAM`, `ACTIVITY_HEATMAP`, `ACTIVITY_MAP`), fields/streams (`INPUT_FIELD`, `ACTIVITY_FIELD`, `INTERVAL_FIELD`, `ACTIVITY_STREAM`), panels (`ACTIVITY_PANEL`), and zones (`ZONES`). Unknown/custom upstream item types remain pass-through and are not rejected by the documentation descriptor.
- Resource representation: generated Markdown will render one section per family with item types, one-line notes, representative sample `content` JSON, and inferred schema paths/kinds generated by the same shared inference code used by write validation tests.
- Resource contract: URI `icuvisor://custom-item-schemas`, name `custom_item_schemas`, title `Custom item schemas`, short description, MIME type `text/markdown`, static/no-network read handler that honors context cancellation, one text content item with URI/MIME/text populated, and wording that live writes still validate against readable custom-item schemas.
- Register `CustomItemSchemasResource()` in `resources.NewRegistry()` alongside workout syntax and event categories so normal server runs expose it in `resources/list` and `resources/read`.
- Tests: golden-lock generated Markdown; assert every descriptor family/item type/sample and inferred path appears in the resource; add registry/read/cancellation/protocol coverage; update custom-item write validation tests or add focused tests to prove existing rejection behavior and detail-sample fallback still work after moving inference code.
- Step 6 boundary: defer broad custom-item tool-description trimming and README updates; only minimal wording changes are allowed if needed to point metadata at the resource or compile against the shared package.

### R014 revision notes

- Replace `mustSample` panic in `internal/customitemschemas/descriptors.go` with plain Go literals or an error-returning descriptor path.
- Change descriptor/resource output so each documented `item_type` has its own concrete sample/schema subsection or explicitly declares an alias, with tests enforcing coverage.

### Step 5 plan

- Factor the existing `get_athlete_profile` response structs and shaping helpers into a small shared package so the tool and resource cannot drift on unit/timezone/`_meta` behavior.
- Resource contract: URI `icuvisor://athlete-profile`, name `athlete_profile`, title `Athlete profile`, MIME type `application/json`, dynamic read handler that returns the default terse shaped profile (`include_full=false`) as one text resource content item.
- Refresh/staleness policy: cache one shaped profile per resource instance for 15 minutes; concurrent reads share the same cache; after TTL expiry the next read refreshes via the configured intervals client; failed refreshes return the short safe profile-fetch error and do not perform retry loops or background polling.
- Context cancellation is honored before acquiring/refreshing the cache and while calling the upstream client. No unbounded upstream calls: each `resources/read` causes at most one `GetAthleteProfile` call, and cached reads cause zero calls.
- Wire `resources.NewRegistryWithOptions(client, ResourceOptions{Version, TimezoneFallback, DebugMetadata})` from `internal/app`; keep `resources.NewRegistry()` for static test/default use by accepting nil client only when callers intentionally want static resources.
- Tests: shared shaper parity with current tool outputs, dynamic resource cache hit/expiry behavior, canceled context behavior, registry/protocol list/read coverage for all four resources.

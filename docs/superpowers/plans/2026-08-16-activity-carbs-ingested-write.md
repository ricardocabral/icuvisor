# Activity Carbohydrate Intake Write Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let `update_activity` set a completed non-Strava activity's athlete-logged carbohydrate intake as a sparse, validated whole-gram update.

**Architecture:** Extend the existing MCP request and `intervals.UpdateActivityParams` together with presence-tracked carbohydrate fields so explicit zero survives while omission remains unchanged. Serialize only upstream `carbs_ingested`, preserve terse/full response behavior, and update generated contracts and product documentation.

**Tech Stack:** Go 1.25, `encoding/json`, `net/http/httptest`, the existing MCP tool registry and response shaper, Make-generated tool data.

## Global Constraints

- Public MCP field: `carbs_ingested_g`; upstream JSON field: `carbs_ingested`.
- Accept only JSON integers from 0 through 2147483647; reject null, fractional, string, negative, and larger values.
- Omission means unchanged; explicit zero is sent; clearing is unsupported.
- Keep `carbs_used_g` read-only and distinct from athlete-logged intake.
- Keep `RequirementWrite`: visible in safe/full modes, hidden when writes are disabled, without a `confirm` argument.
- State that intervals.icu does not update Strava activities through this endpoint.
- Do not claim live persistence without an authenticated PUT-then-GET smoke test on a disposable non-Strava activity.
- Add no dependency and copy no GPL/copyleft source.

---

### Task 1: End-to-end sparse carbohydrate update

**Files:**
- Modify: `internal/intervals/activities_test.go`
- Modify: `internal/intervals/activities.go`
- Modify: `internal/tools/update_activity_test.go`
- Modify: `internal/tools/update_activity.go`

**Interfaces:**
- Produces: `intervals.UpdateActivityParams{CarbsIngested int, CarbsIngestedSet bool}`.
- Produces: sparse `updateActivityPayload.CarbsIngested *int` serialized as `carbs_ingested`.
- Produces: `updateActivityRequest.CarbsIngestedG *int`, `carbsIngestedGProvided bool`, and response field name `carbs_ingested_g`.

- [ ] **Step 1: Write failing client transport tests**

Add table-driven cases around `Client.UpdateActivity` that decode the real HTTP body and verify both a positive value and explicit zero:

```go
for _, tc := range []struct {
	name string
	grams int
}{
	{name: "positive grams", grams: 90},
	{name: "logged zero", grams: 0},
} {
	t.Run(tc.name, func(t *testing.T) {
		// httptest handler decodes the body into map[string]any.
		if got, ok := decoded["carbs_ingested"]; !ok || got != float64(tc.grams) {
			t.Fatalf("carbs_ingested = %#v, present %v, want %d", got, ok, tc.grams)
		}
		if len(decoded) != 1 { t.Fatalf("body = %#v, want carbs_ingested only", decoded) }
		_, err := client.UpdateActivity(context.Background(), UpdateActivityParams{
			ActivityID: "a1", CarbsIngested: tc.grams, CarbsIngestedSet: true,
		})
		if err != nil { t.Fatalf("UpdateActivity() error = %v", err) }
	})
}
```

Add validation cases for `-1`, `2147483648`, and no presence flags.

- [ ] **Step 2: Write failing MCP tool tests**

Add table-driven handler cases for `90` and `0` and assert `CarbsIngestedSet`, the value, and `fields_updated`:

```go
for _, tc := range []struct {
	name string
	raw  string
	want int
}{
	{name: "positive grams", raw: `{"activity_id":"a1","carbs_ingested_g":90}`, want: 90},
	{name: "logged zero", raw: `{"activity_id":"a1","carbs_ingested_g":0}`, want: 0},
} {
	t.Run(tc.name, func(t *testing.T) {
		client := &fakeActivityUpdaterClient{activity: decodeActivity(t, `{"id":"a1"}`)}
		tool := newUpdateActivityTool(client, client, "test", false)
		result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(tc.raw)})
		if err != nil { t.Fatalf("Handler() error = %v", err) }
		call := client.calls[0]
		if !call.CarbsIngestedSet || call.CarbsIngested != tc.want { t.Fatalf("call = %#v", call) }
		fields := resultMap(t, result)["fields_updated"].([]any)
		if len(fields) != 1 || fields[0] != "carbs_ingested_g" { t.Fatalf("fields_updated = %#v", fields) }
	})
}
```

Extend bad-argument cases with `null`, `-1`, `2147483648`, `1.5`, and `"90"`. Extend registration assertions to require integer type, `minimum: 0`, `maximum: 2147483647`, set-only and Strava wording, and no `carbs_used_g` or `confirm` property.

- [ ] **Step 3: Run focused tests and verify RED**

Run: `go test ./internal/intervals ./internal/tools -run '^TestUpdateActivity'`

Expected: compile/test failures because the carbohydrate request, parameter, and payload fields do not exist.

- [ ] **Step 4: Implement minimal intervals transport**

In `internal/intervals/activities.go`, add:

```go
const maxActivityCarbsIngested = 1<<31 - 1

type UpdateActivityParams struct {
	// existing fields
	CarbsIngested    int
	CarbsIngestedSet bool
}

type updateActivityPayload struct {
	// existing fields
	CarbsIngested *int `json:"carbs_ingested,omitempty"`
}
```

Include `CarbsIngestedSet` in the at-least-one-field check. Reject values outside `0..maxActivityCarbsIngested`. When present, assign a local copy's address to the payload so zero is serialized. Update comments and errors to name `carbs_ingested`.

- [ ] **Step 5: Implement minimal tool contract**

In `internal/tools/update_activity.go`, add:

```go
const maxActivityCarbsIngestedG = 1<<31 - 1

type updateActivityRequest struct {
	// existing exported JSON fields
	CarbsIngestedG *int `json:"carbs_ingested_g,omitempty"`
	// existing presence flags
	carbsIngestedGProvided bool
}
```

Set the presence flag from `rawObjectFields`; reject a provided nil pointer and values outside `0..maxActivityCarbsIngestedG`; include it in the at-least-one-field condition; map it to `intervals.UpdateActivityParams`; and append `carbs_ingested_g` in `updateActivityFieldsUpdated`.

Update the tool/interface comments, descriptions, both validation messages, and output schema. Add the input property:

```go
"carbs_ingested_g": map[string]any{
	"type": "integer",
	"minimum": 0,
	"maximum": maxActivityCarbsIngestedG,
	"description": "Optional athlete-logged carbohydrate consumed during this activity, in whole grams (0-2147483647). Zero is a logged zero; omit to leave unchanged. Clearing is not supported. Distinct from read-only carbs_used_g. Intervals.icu does not update Strava activities through this endpoint.",
},
```

- [ ] **Step 6: Format and verify GREEN**

Run:

```bash
gofmt -w internal/intervals/activities.go internal/intervals/activities_test.go internal/tools/update_activity.go internal/tools/update_activity_test.go
go test ./internal/intervals ./internal/tools -run '^TestUpdateActivity'
```

Expected: PASS.

- [ ] **Step 7: Commit the behavior**

```bash
git add internal/intervals/activities.go internal/intervals/activities_test.go internal/tools/update_activity.go internal/tools/update_activity_test.go
git commit -m "feat: support activity carbohydrate intake updates"
```

### Task 2: Product documentation and generated contracts

**Files:**
- Modify: `docs/prd/PRD-icuvisor.md`
- Modify: `CHANGELOG.md`
- Regenerate: `internal/tools/schema_snapshot/update_activity.json`
- Regenerate: `web/data/tools.json`
- Regenerate: `web/data/tool_schemas.json`
- Regenerate: `cmd/gendocs/testdata/tools.golden.json`
- Regenerate: `cmd/gendocs/testdata/tool_schemas.golden.json`

**Interfaces:**
- Consumes: the final registered `update_activity` description and input schema.
- Produces: authored and generated documentation matching the live registry.

- [ ] **Step 1: Update authored documentation**

Add `update_activity` under PRD §6.2 Activities, documenting sparse `name`, `description`, and set-only `carbs_ingested_g`, zero/omission semantics, read-only `carbs_used_g`, range, and Strava limitation. Under `[Unreleased]`, add an `### Added` entry describing the validated activity-intake write.

- [ ] **Step 2: Regenerate schemas and website tool data**

Run:

```bash
go run ./scripts/snapshot_tool_schemas.go
make docs-tools
```

Expected: generated `update_activity` schemas contain `carbs_ingested_g`; no unrelated schema files change unless canonical generation legitimately rewrites them.

- [ ] **Step 3: Verify generated-contract tests**

Run: `go test ./internal/toolchecks ./cmd/gendocs`

Expected: PASS with no stale snapshot or golden failures.

- [ ] **Step 4: Review and commit documentation**

Run: `git diff --check && git diff --stat && git status --short`. Inspect every generated file, then commit only expected files:

```bash
git add CHANGELOG.md docs/prd/PRD-icuvisor.md internal/tools/schema_snapshot/update_activity.json web/data/tools.json web/data/tool_schemas.json cmd/gendocs/testdata/tools.golden.json cmd/gendocs/testdata/tool_schemas.golden.json
git commit -m "docs: document activity carbohydrate intake writes"
```

### Task 3: Full verification and upstream-smoke status

**Files:**
- Modify only if verification exposes a defect in the files above.

**Interfaces:**
- Consumes: Tasks 1 and 2.
- Produces: fresh repository evidence and an explicit upstream persistence-test status.

- [ ] **Step 1: Run repository checks**

Run: `make check`

Expected: exit 0 with tests, formatting, lint, schema, and documentation checks passing.

- [ ] **Step 2: Run the race suite required by project guidance**

Run: `make test-race`

Expected: exit 0 with no race reports.

- [ ] **Step 3: Review final scope**

Run: `git status --short && git diff be2b006^ --check && git diff be2b006^ --stat`.

Confirm only the two design commits, implementation, tests, authored docs, plan, and generated contracts changed.

- [ ] **Step 4: Record manual upstream verification status**

If an authorized disposable non-Strava activity is available, PUT a temporary `carbs_ingested` integer, GET it back, verify persistence, and restore the prior known value. Otherwise report exactly: `Authenticated non-Strava PUT→GET persistence smoke test not run; OpenAPI request shape and local transport are verified.`

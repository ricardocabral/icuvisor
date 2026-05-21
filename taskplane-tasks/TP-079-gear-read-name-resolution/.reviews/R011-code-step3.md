# Code Review — TP-079 Step 3

**Verdict:** REVISE

The implementation generally follows the approved Step 3 plan: activity list/detail rows now carry `gear_id`, `gear_name`, and explicit `gear_resolution` states; resolution uses the registry-scoped `gearListCache`; non-context gear lookup failures degrade to `lookup_unavailable`; and the targeted tools test suite plus full suite pass locally. I found one pagination edge that can silently drop the new gear behavior for accepted terse continuation tokens.

## Findings

### 1. Existing terse `next_page_token`s can suppress `gear_id` on subsequent pages

**Severity:** P2  
**Location:** `internal/tools/get_activities_cursor.go:239-247`, with the new required field in `internal/tools/get_activities.go:25-30`

`terseActivityFields` now correctly includes `gear_id`, but `newPageCursor` overwrites the freshly seeded field list whenever a continuation token is supplied:

```go
if !args.IncludeFull {
    cursor.token.Fields = append([]string(nil), terseActivityFields...)
}
if token != nil {
    ...
    cursor.token.Fields = append([]string(nil), token.Fields...)
}
```

Because v1 tokens issued before this change, and the existing `identicalTimestampStallToken` fixture, are still accepted and do not contain `gear_id`, continuing such a page will call `ListActivities` without requesting `gear_id`. With `applyListParams`/real upstream field filtering, the returned activities then have empty `Activity.GearID`, `resolveActivityGear` skips the gear lookup, and the page omits `gear_id`/`gear_name`/`gear_resolution` even when upstream gear data exists. That violates Step 3's “request `gear_id` in terse list fields” and makes gear resolution inconsistent across pages during token rollover.

Please either merge mandatory current terse fields into token fields when `include_full=false` (at least ensure `gear_id` is present), or bump/reject old token versions. Add a regression test using an accepted old-style token without `gear_id` plus an activity containing `gear_id`, asserting that the follow-up request still includes/resolves gear.

## Tests run

- `git diff 54960e21a6da9a0a84efe3ebf19fa427bede3e99..HEAD --name-only`
- `git diff 54960e21a6da9a0a84efe3ebf19fa427bede3e99..HEAD`
- `go test ./internal/tools`
- `go test ./...`

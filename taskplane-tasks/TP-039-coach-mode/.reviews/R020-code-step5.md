# R020 code review — Step 5: Catalog-cache caveat + Tests

Verdict: **REVISE**

## Findings

1. **Step 5's advanced-capabilities catalog assertion is still missing.**  
   R018/R019 required structured, exact visible-catalog comparisons across `tools/list`, `select_athlete.allowed_tools`, and `icuvisor_list_advanced_capabilities`. The new two-session test only compares `select_athlete.allowed_tools` to `tools/list` (see `internal/mcp/protocol_test.go:991-1006`), and the existing advanced-capabilities coverage remains a substring-only non-leak check for `get_power_curves` (`internal/mcp/protocol_test.go:1235-1254`). That does not prove `icuvisor_list_advanced_capabilities` is using the same active-athlete/post-gate catalog as `tools/list`/`select_athlete`, nor does it cover the hidden delete-mode/full-toolset cases called out in the Step 5 plan. Please add structured JSON parsing of the advanced-capabilities response and exact row-name assertions for the selected athlete, including the delete-mode-hidden and core-toolset-hidden regressions, before marking the R018 exact-comparison checkbox complete.

2. **The fake-client E2E does not explicitly prove denied writes/deletes make no upstream call.**  
   The read-only denial loop asserts the enumeration-safe error (`internal/mcp/protocol_test.go:1153-1164`), which is useful, but the R019 implementation reminder asked for an explicit assertion that the fake intervals client was not called for denied write/delete paths. Because the calls use empty arguments, a regression that reaches handler decoding could fail for the wrong reason without proving a valid denied write/delete would be stopped before any intervals request. Add a request counter or route log around the denied calls (ideally with syntactically valid write/delete arguments) and assert it is unchanged.

## What passed

- `go test ./internal/mcp ./internal/tools ./internal/coach`
- `go test ./...`
- `go test -race ./internal/mcp`

The `safeRegistrar.visibleToolNamesForAthlete` injection into `select_athlete` looks like the right production direction; the remaining blockers are Step 5 coverage/verification gaps rather than compile failures.

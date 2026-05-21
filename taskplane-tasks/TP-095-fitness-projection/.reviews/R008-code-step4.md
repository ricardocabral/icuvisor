# R008 Code Review — Step 4: Docs and verification

**Verdict:** Request changes

## Findings

### 1. Website catalog data is still stale, so the public reference will not list the new tool

- **Severity:** High
- **File:** `web/data/tools.json:187`

`web/content/reference/tools.md` uses the `{{< tool-catalog >}}` shortcode, which is backed by `web/data/tools.json`. Step 4 updated the human-readable assumptions note and the gendocs golden fixture, but did not regenerate/check in `web/data/tools.json`. The current data jumps from `get_fitness` directly to `get_hr_curves`, so the rendered website catalog will still omit `get_fitness_projection`.

I verified with a temp output that the generator would add the missing entry:

```sh
tmp=$(mktemp)
go run ./cmd/gendocs --out "$tmp"
diff -u web/data/tools.json "$tmp" | grep -C 4 get_fitness_projection
```

Please run `make docs-tools` (or the equivalent generator command) and commit the resulting `web/data/tools.json` update.

### 2. The full-series response still drops explicit zero training loads

- **Severity:** Medium
- **File:** `internal/tools/get_fitness_projection.go:81`

The unresolved R006 blocker remains: `fitnessProjectionPoint.TrainingLoad` is still tagged `json:"training_load,omitempty"`. Valid projections can contain a real zero load (`planned_daily_loads[].training_load: 0` or `recovery_week_load_pct: 0`), but `include_full:true` responses will omit `training_load` for those days. That makes an explicit rest/zero-load plan indistinguishable from missing data in the public tool contract.

Remove `omitempty` or otherwise encode zero explicitly, and add coverage/golden output for a zero planned load or 0% recovery week.

### 3. Horizon and recovery-cadence contracts are still inconsistent across schema, decoder, and user-facing error text

- **Severity:** Medium
- **File:** `internal/tools/get_fitness_projection.go:19`, `internal/tools/get_fitness_projection.go:241`, `internal/tools/get_fitness_projection.go:329`

Step 4 did not resolve the R006 schema/decoder mismatch. The public error says callers must provide “exactly one `horizon_date` or `horizon_days`”, but the decoder accepts neither and silently defaults to 42 days. Similarly, the JSON Schema advertises `recovery_week_cadence` as any integer from 0 to 12, while the decoder rejects `1` and only allows `0` or `2-12`.

Please choose one contract and align all three surfaces: schema/descriptions, decoder behavior, and user-facing invalid-argument text. Add invalid/default tests for omitted horizon and `recovery_week_cadence: 1`.

### 4. STATUS still records rejected reviews as approvals and marks verification complete without accurate history

- **Severity:** Low
- **File:** `taskplane-tasks/TP-095-fitness-projection/STATUS.md:54`, `taskplane-tasks/TP-095-fitness-projection/STATUS.md:82`, `taskplane-tasks/TP-095-fitness-projection/STATUS.md:115`

The Reviews table is empty, while the Notes section records R006 and R007 as `APPROVE`. The checked-in review files say R006 was `Request changes` and R007 was `Needs changes / not yet reviewable`. Step 4 also checks “Run full quality gate” without recording the commands/outcomes in the execution log.

Please move the review history into the Reviews table with the actual verdicts, keep unresolved blockers/discoveries visible, and record the quality-gate commands and outcomes accurately.

## Verification

Ran:

```sh
go test ./cmd/gendocs ./internal/safety ./internal/analysis ./internal/tools ./internal/toolcatalog
make test
make build
make lint
```

All passed. The stale `web/data/tools.json` issue is not caught by those gates; the temporary gendocs diff above shows the missing generated docs artifact.

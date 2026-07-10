# Plan Review — TP-228 Step 3

## Verdict: REVISE

The request/schema migration, strict legacy-field rejection, truthful metadata, generated artifacts, and tool coverage are sufficiently specified. One blocking implementation detail remains: the stated schema-stability “allowance in schema-stability tests” cannot approve the real removal in CI. CI invokes `CheckSchemaStability` in `internal/toolchecks/schema_stability.go` against the pre-PR snapshot, where removing `update_sport_settings.effective_date` will still produce `property-removed`.

Revise the plan to name and describe the production, narrowly scoped approval mechanism (including modifying `internal/toolchecks/schema_stability.go`, or an equivalent CI-consumed policy): it must suppress only the `property-removed` failure for exactly `update_sport_settings.effective_date`, carry a durable TP-228/safety-correction rationale, and leave all other tool/property removals unchanged. Add tests that exercise the real mechanism: the approved removal passes, while a different property on that tool and `effective_date` on another tool still fail. Then regenerate artifacts and run the stated targeted tests.

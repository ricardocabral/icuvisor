# Code Review: Step 2 — Add additive meta to interval reads

## Verdict

Request changes.

The additive `_meta.interval_source` / `_meta.auto_lap_suspected` wiring is in the right response path and targeted tests pass, but the shared classifier has two contract issues that can misclassify interval sources before Step 3 analyzers start relying on it.

## Findings

1. **Non-generic names can still be classified as device laps because empty sibling fields count as generic.**
   `internal/analysis/interval_source.go:187-194` treats a row as generic if *any* of `Name`, `Type`, or `Label` is generic. Since `isGenericLapText("")` returns true, a row named `Climb 1`, `Set A`, or another custom non-structured/non-generic label with empty `type`/`label` still passes the generic-row gate and can become `device_laps` if the distances are uniform. The Step 1 contract says only generic rows should become device laps, and non-generic labels should stay `unknown` unless there is explicit source evidence. Consider requiring every non-empty text field to be generic, or otherwise making empty fields neutral rather than sufficient.

2. **Explicit auto-lap markers with boolean or `auto` values are ignored.**
   `internal/analysis/interval_source.go:148-156` only classifies explicit lap-source keys when the marker value string contains `autolap`, `devicelap`, `device`, or `lap`. `anyMarkerString` returns an empty string for booleans, so common raw shapes such as `{"auto_lap": true}` are ignored. Likewise `{"lap_type":"auto"}` is ignored because `auto` alone is not accepted. The task explicitly included raw marker inspection for candidate keys like `auto_lap`/`lap_type`; these should classify as device laps when true/auto unless structured evidence has already won.

## Tests run

- `go test ./internal/analysis ./internal/tools`

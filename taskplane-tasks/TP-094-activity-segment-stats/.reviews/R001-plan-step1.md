# Plan review — Step 1: Define segment selection and metrics

**Verdict: Changes requested**

The plan has a good foundation: exactly-one segment selector, closed enums, terse-by-default `include_full`, and validation for missing/ambiguous/out-of-range ranges. Before implementation, a few contract details need to be tightened so the tool does not ship with ambiguous schema or formula behavior.

## Blocking issues

1. **`ftp_watts` is required by IF but missing from the request shape.**  
   STATUS line 34 says `if` uses caller-provided `ftp_watts`, but line 32 only defines `activity_id`, segment fields, `stats[]`, metric, and `include_full`. Add `ftp_watts` to the planned schema and validation: required when `stats` contains `if`, positive finite number, ignored/rejected otherwise by an explicit rule.

2. **The metric enum is not aligned with canonical stream keys.**  
   The proposed enum includes `speed_mps` and `pace_seconds_per_km` (line 32), but the current stream canonicalizer exposes `velocity_smooth` for upstream speed and has no canonical `speed_mps` or pace stream. Since this task must stay behind stream-key canonicalization, either expose canonical stream metrics (`watts`, `heart_rate`, `cadence`, `velocity_smooth`, `distance`, `time`) or explicitly define `speed_mps`/`pace_seconds_per_km` as derived metrics with their required canonical source stream(s) and unit conversion. Add tests for this mapping.

3. **`stats[]` is ambiguous with the analyzer `_meta` contract.**  
   Multiple stats can have different required streams, formulas, sample counts, and insufficiency states, but `analysis.AnalyzerMeta` currently has a single `formula_ref`, `n`, and `insufficient_sample`. Define one of these before implementation:
   - restrict the schema to a single `stat`, or
   - keep `stats[]` but return per-stat status/method/formula/sample metadata while defining what top-level mandatory `_meta` means.

4. **Formula refs for NP/IF are unresolved.**  
   Step 1 explicitly asks to define formula refs for decoupling/drift/NP/IF. The plan correctly references HR drift and Pw:HR, but for NP/IF it says there are no resource refs and method text will be placed in `_meta.method` (line 34). That leaves the Step 1 checkbox only partially satisfied. Decide explicitly whether to add new formula refs, leave `formula_ref` empty for NP/IF, or represent refs per result. Do not reuse EF/VI refs for IF/NP unless the formula resource is intentionally expanded.

5. **NP/IF must be time-based, not sample-count-based.**  
   The proposed “at least 30 power samples” minimum (line 34) assumes 1 Hz samples. Streams have a time channel, so the 30-second rolling NP calculation should be specified against elapsed seconds using the `time` stream, with validation for sufficient 30-second coverage and enough finite power samples. This matters especially for irregular sample cadence and distance-selected segments.

6. **Split-half formulas need stricter eligibility semantics.**  
   A minimum sample count of 2 for decoupling/drift (line 34) means one sample per half, which is too weak for a formula intended to compare segment halves. Define per-half minimums, require positive denominators (`avg_hr`, and `avg_power`/ratios for Pw:HR), and specify that halves are split by elapsed time rather than array index when cadence is uneven.

## Non-blocking clarifications to add

- Replace “derived stats ignore or validate incompatible metric values” with a deterministic rule. Prefer rejecting incompatible combinations; silent ignore will confuse LLM callers.
- Define whether segment bounds are inclusive on both ends intentionally. `start <= axis <= end` can double-count boundary samples across adjacent calls; that may be acceptable, but it should be explicit.
- State what `include_full:true` includes for this analyzer. Terse mode must not include raw samples; full mode should probably include only the sliced/calculation inputs needed for audit, not all activity streams.

Once these are addressed in the Step 1 notes, the plan should be ready for implementation.

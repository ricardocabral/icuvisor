# Plan Review R001 — Step 3: `get_extended_metrics`

Verdict: **APPROVE**

The revised Step 3 plan is now implementable and aligns with the TP-010 prompt, PRD §7.2.C/§7.4 #4, and the Step 1 availability table. It keeps the tool fidelity-focused by using only upstream-exposed activity, interval, and power-vs-HR fields; explicitly drops unavailable/computed fields; avoids raw streams; and gates raw upstream payloads behind `include_full:true`.

## What is covered well

- Defines a concrete `ExtendedMetricsClient` contract and registry gating so the new tool follows the existing greppable registry pattern.
- Specifies the request shape and top-level response shape, including optional `intervals`, optional `full`, and `_meta`.
- Separates activity-level metrics from interval-level metrics, including distinct W' balance start/end keys instead of implying a single scalar.
- Requires J → kJ conversion for `joules_above_ftp` and W' balance fields, with tests expected to assert that boundary conversion.
- Uses the existing profile/shaping path (`toolProfile`, `response.Shape`, `UnitSystem`) so `_meta.server_version`, `_meta.units`, scale metadata, and null stripping remain consistent with earlier tools.
- Defines partial behavior for optional sources: activity detail is required; interval and power-vs-HR `ErrNotFound`/`ErrUnauthorized` can return activity-level metrics with `_meta.partial` and short unavailable-source names.
- Keeps device-identifying details terse by default by emitting only `device_name` unless `include_full:true` is explicitly requested.

## Implementation notes

- When shaping the response, also preserve the existing activity-read convention for Strava-blocked/imported activities (`strava_imported` plus the standard `unavailable.reason: strava_tos` marker where applicable). This follows the repository MCP convention and the existing `get_activities` / `get_activity_details` behavior.
- Keep `pw_hr` tied to direct upstream `icu_power_hr` / `powerHr` evidence from the availability table; do not derive it locally from power and HR streams.
- For `_meta.extended_metric_units`, use existing canonical `units.Unit` values only where the enum has an appropriate value; leave genuinely unitless/score fields documented by scale/description rather than inventing ad hoc unit tokens.

No further plan changes are required before implementation.

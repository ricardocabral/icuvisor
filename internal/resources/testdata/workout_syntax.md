# Workout syntax

This resource documents the Intervals.icu structured-workout DSL emitted by icuvisor. Examples are generated from `internal/workoutdoc` structured steps with `workoutdoc.Serialize`, so the resource follows the serializer rather than a separate hand-authored grammar.

## General form

- Simple steps begin with `- ` and include an optional description, a duration or distance, then at most one primary target plus optional cadence.
- Repeat blocks use an `Nx` header and two-space-indented child steps.
- Numeric ranges use `low-high`; zones use `Zlow-Zhigh`.

## Supported features

### Duration steps

A simple step needs a positive duration or a distance. Durations serialize as h/m/s tokens.

- `duration_percent_ftp`: Duration step with a percent-FTP power target.

```text
- Warmup 10m 55%
```


### Distance steps

Distance steps serialize with canonical mtr, km, or mi suffixes.

- `distance_mtr`: Meter distance canonicalizes to mtr.

```text
- Stride 400mtr 120%
```

- `distance_km`: Kilometer distance canonicalizes to km.

```text
- Tempo 5km 92-96% Pace
```

- `distance_mi`: Mile distance canonicalizes to mi.

```text
- Cooldown 1mi freeride
```


### Repeat blocks

Repeat blocks use an Nx header and indented child steps.

- `repeat_block`: Two child steps repeated three times.

```text
Main set 3x
  - Hard 2m 105-115% 95-105rpm
  - Easy 1m freeride
```


### Free-ride steps

Freeride is a primary target and cannot be combined with another primary target.

- `freeride`: Open target free ride.

```text
- Open 5m freeride
```


### Ramp targets

Ramp steps use start and end target bounds and serialize with a ramp prefix.

- `power_ramp`: Power ramp from one percent-FTP target to another.

```text
- Build 8m ramp 70-95%
```


### Cadence targets

Cadence is an optional secondary target in rpm and may be scalar or range.

- `cadence_range`: Cadence range appended after the primary target.

```text
- Spin 3m 60% 100-110rpm
```


### Power targets

Power targets support percent FTP, watts, power zones, scalar values, and ranges.

- `power_percent`: Percent FTP scalar.

```text
- Endurance 10m 75%
```

- `power_percent_range`: Percent FTP range.

```text
- Sweet spot 10m 88-94%
```

- `power_watts`: Watts scalar.

```text
- Erg 5m 250w
```

- `power_zone`: Power zone range.

```text
- Zone work 15m Z2-Z3
```


### Heart-rate targets

Heart-rate targets support percent max HR, percent LTHR, bpm, HR zones, scalar values, and ranges.

- `hr_percent`: Percent max HR scalar.

```text
- Aerobic 10m 80% HR
```

- `hr_lthr`: Percent LTHR range.

```text
- Threshold HR 10m 95-99% LTHR
```

- `hr_bpm`: BPM scalar.

```text
- Cap 5m 150bpm
```

- `hr_zone`: HR zone range.

```text
- Zone HR 10m Z2-Z3 HR
```


### Pace targets

Pace targets support percent threshold pace, pace zones, numeric PACE values, and non-ramp text pace labels.

- `pace_percent`: Percent threshold pace scalar.

```text
- Cruise 10m 95% Pace
```

- `pace_zone`: Pace zone range.

```text
- Pace zone 10m Z2-Z3 Pace
```

- `pace_numeric`: Numeric PACE unit as currently serialized.

```text
- Numeric pace 5m 5 Pace
```

- `pace_text`: Text pace label.

```text
- Marathon 20m Marathon Pace
```


### RPE targets

RPE targets support scalar values and ranges.

- `rpe_scalar`: RPE scalar.

```text
- Steady 10m RPE 6
```

- `rpe_range`: RPE range.

```text
- Progression 10m RPE 6-8
```

## Limitations

- `one_primary_target`: Each simple step can contain only one primary target among power, heart rate, pace, RPE, or freeride.
- `ramp_requires_numeric_target`: Ramp requires a power, heart-rate, pace, or RPE target with start and end bounds; text targets cannot be used for ramps.
- `freeride_not_ramp`: Freeride cannot be combined with ramp or another primary target.
- `repeat_fields`: Repeat blocks require reps greater than zero and child steps, cannot be nested, and cannot also carry simple-step fields.
- `simple_step_duration_or_distance`: Simple steps require a positive duration or a supported distance.

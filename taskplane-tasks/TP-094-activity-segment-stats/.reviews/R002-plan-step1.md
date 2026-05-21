# Plan review — Step 1: Define segment selection and metrics

**Verdict: Changes requested**

The revised plan addresses the R001 blockers well: it moves to a single `stat`, adds `ftp_watts`, aligns scalar metrics to canonical stream keys, rejects incompatible `metric`/`ftp_watts` combinations, uses existing formula refs only where they exist, and makes NP/IF elapsed-time based rather than sample-count based.

One formula/validation detail still needs to be fixed before implementation.

## Blocking issue

1. **NP/IF must not require positive power samples.**  
   The plan says NP/IF require “finite positive power samples.” Zero-watt samples are valid cycling power data and must be included in normalized power calculations; excluding or rejecting them would inflate NP/IF and make coasting/stop-start segments unusable. Tighten the eligibility rule to require finite, non-negative watts samples, allow zero values, and reject only negative/non-finite values (or document an explicit skip policy for non-finite samples). If a positive denominator is needed, apply that to the final FTP for IF and any average denominator, not to every power sample.

## Non-blocking clarifications

- Define `_meta.n` precisely per stat before coding: for mean/median/p90 it can be the number of finite sliced metric samples; for decoupling/drift it should be the number of paired finite samples used across both halves; for NP/IF it should be either the number of finite watts samples in the segment or the number of valid 30-second rolling windows. Pick one and test it so `insufficient_sample` is stable.
- For irregular stream cadence, specify whether the 30-second NP rolling average is time-weighted over elapsed seconds, a simple average of samples whose timestamps fall in the rolling window, or requires approximately regular cadence. The current “elapsed seconds rather than array count” direction is correct, but the implementation needs one deterministic rule.

Once the positive-power requirement is corrected, the Step 1 design is ready to implement.

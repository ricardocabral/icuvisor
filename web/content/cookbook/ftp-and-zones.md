---
title: "FTP and zones review"
description: "Decide whether a threshold has actually moved, then update training zones to match."
weight: 50
---

"Should I bump my FTP?" is a recurring question — and the wrong answer in either direction costs you weeks of training at the wrong intensity. This recipe makes the assistant weigh real best-effort evidence before recommending a change, and only updates zones with your sign-off.

## When to use this

- After a ramp test, 20-minute test, or a breakthrough effort.
- Every 4-6 weeks during a build, as a routine threshold check.
- When workouts feel consistently too easy or too hard for their prescribed zone.

## The recipe

```text
Help me decide whether my FTP or threshold has actually moved. Use icuvisor
with my intervals.icu data.

1. Read my athlete profile: current FTP, threshold heart rate, threshold pace,
   and zones for each sport.
2. Pull my best efforts and my power curve (and pace curve, if I run) for the
   last 90 days.
3. Compare those against the prior 90 days to see which durations changed.
4. If I name a ramp or test activity, read its intervals to extract a fresh
   threshold estimate from that effort.

Then:
- Tell me whether the evidence supports raising, holding, or lowering each
  threshold, and by roughly how much.
- Show the specific durations where I actually gained or lost.
- If the evidence is thin or noisy, say so — do not force a change.

If we agree on a new value, update my cycling zones to match — but show me the
exact before/after first and wait for my explicit OK.
```

## What icuvisor does

| Step | Tool | Why |
| --- | --- | --- |
| 1 | [`get_athlete_profile`]({{< relref "/reference/tools#get_athlete_profile" >}}) | Current thresholds and zone definitions. |
| 2 | [`get_best_efforts`]({{< relref "/reference/tools#get_best_efforts" >}}), [`get_power_curves`]({{< relref "/reference/tools#get_power_curves" >}}), [`get_pace_curves`]({{< relref "/reference/tools#get_pace_curves" >}}) | Peak outputs by duration. |
| 3 | [`analyze_efforts_delta`]({{< relref "/reference/tools#analyze_efforts_delta" >}}) | How best efforts shifted versus a baseline window. |
| 4 | [`get_activity_intervals`]({{< relref "/reference/tools#get_activity_intervals" >}}) | Reads a named test activity for a fresh estimate. |
| 5 | [`update_sport_settings`]({{< relref "/reference/tools#update_sport_settings" >}}) | Applies new zones — only after you confirm. |

## A good answer looks like

> **Cycling FTP — evidence supports a small raise.**
>
> Current FTP: 282 W (`get_athlete_profile`).
>
> Versus the prior 90 days (`analyze_efforts_delta`): your 5-min power is up 14 W, 8-min up 11 W, and 20-min up 9 W (to 297 W). Sprint and 1-min power are flat. The gains cluster around threshold and above — consistent with a real FTP move, not just freshness.
>
> **Recommendation:** Raise FTP from 282 W to roughly 290 W — a conservative read of the 20-min gain (about 95% of 297 W minus a margin). Holding at 282 W would now under-prescribe your threshold work.
>
> **Proposed zone change** — say the word and I will apply it with `update_sport_settings`:
> | | Current | Proposed |
> | --- | --- | --- |
> | FTP | 282 W | 290 W |
> | Z2 top | 207 W | 213 W |
> | Z4 (threshold) | 268-296 W | 276-305 W |

## Variations

- **Running:** swap to threshold pace and heart rate — "review my threshold pace from the last 90 days".
- **Ramp test extraction:** "Find my most recent ramp or threshold test — make one `get_activities` call over the last 120 days and scan names for 'ramp', 'FTP test', or 'threshold'. Read FTP, threshold HR, and any VO2max estimate from that activity only; if none is found after one scan, say so rather than re-querying."
- **Heart-rate athletes:** ask for [`get_hr_curves`]({{< relref "/reference/tools#get_hr_curves" >}}) and a threshold-HR check instead of power.

## Why this prompt works

- **Evidence before action.** Asking for best-effort deltas across durations stops the assistant from rubber-stamping an FTP bump just because you asked.
- **"Do not force a change."** Freshness inflates short tests. This line lets the assistant say "hold" — often the right call.
- **Confirm-before-write.** Zone updates are a [gated write]({{< relref "/reference/safety-modes" >}}). Requiring a before/after preview means you never get a silent zone change.

# Issue 64: remaining stream and split failures

## Evidence

On 2026-09-30, a synthetic TCX run uploaded to the `.env-dev` test account reproduced the full v1.7.1 failure pattern against the real Intervals.icu API. Activity `i192029828` has 2,858 samples, cumulative distance from 5 to 7,400 m, and 24 paused seconds. All scalar streams have `data2:null`. Direct unfiltered/filtered/windowed stream reads withhold samples; splits are empty with `paused_samples_present` and `insufficient_split_coverage`. Controls pass: HR histogram n=2857, pace histogram n=2833, segment mean 0–120 seconds n=121. The recording is synthetic, not a copy of the reporter's activity.

## Implementation

1. Capture the upstream stream response in `internal/tools/testdata/activity_streams/issue64_nonzero_paused.json`. Exercise the real HTTP client and tool handlers in `get_activity_streams_test.go`. Assert direct reads, filtering, full responses, windows, and downsampling work while all three controls retain their values/counts. Run these tests before changes and confirm the reported failures.
2. In `get_activity_streams.go`, distinguish absent scalar `data2:null` from a null paired channel. Preserve rejection for latitude/longitude and declared array channels. Gate the window's secondary null check by paired-channel presence. Keep malformed/nonfinite/mismatched pairs and primary null samples rejected.
3. In `get_activity_splits.go`, remove the global zero-origin prerequisite. Emit only fully covered absolute km/mi/100m boundaries and preserve their cumulative indices. Report omitted initial splits; do not extrapolate the origin, shift distance, fabricate a partial tail, or remove elapsed pause time. Pin durations and numbering in `get_activity_splits_test.go`, including multiple omitted initial splits and no fully covered split.
4. Update the changelog and upstream-gap/fixture documentation. Run focused regressions, formatting, lint, and the full race suite. Repeat the live calls through an MCP client against the synthetic test activity. Review the diff and push a topic branch/PR only after verification.

## Review focus

Scalar null secondary fields; declared paired nulls; malformed/null-element/mismatched GPS pairs; bounded and sampled alignment; omitted initial boundaries with paused elapsed time and truthful numbering.

## Verification

The captured-response regressions failed before production changes, with the reported scalar-null withholding and empty splits. After the fix, focused tests and `GOTOOLCHAIN=go1.26.6 make check` passed (formatting, vet, lint, full race suite). An independent review found no blocking issues.

Live MCP SDK client/server sessions backed by the actual upstream API passed all nine calls on both synthetic recordings: nonzero-origin `i192029828` returned six splits indexed 2–7; zero-origin control `i192034341` returned seven splits indexed 1–7. Both retained the 24-second elapsed pause in kilometre 4 and histogram/segment control counts 2857, 2833, and 121. The test recordings remain available for validation; neither is the reporter's original private activity.

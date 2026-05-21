# R005 Code Review — Step 1: Define histogram contract

**Verdict:** REQUEST CHANGES

The R004 `watts` stream-key mismatch is fixed. I found one remaining blocking ambiguity in the contract that can lead to incorrect pace/power/HR zone labels when zone boundaries are filtered or sorted.

## Blocking finding

1. **Configured-zone sorting does not explicitly preserve boundary/name pairs.**
   - Location: `taskplane-tasks/TP-092-activity-histogram/STATUS.md:135`
   - The contract says to filter non-finite boundaries, sort boundaries ascending, then use `ZoneNames[i]` for `boundary[i]`. That leaves it ambiguous whether names are carried with their original boundary during filtering/sorting or re-indexed after sorting.
   - This matters especially for pace zones, which may arrive in descending numeric order before conversion/sorting. An implementation that sorts only the numeric boundary slice and then indexes the original names by the sorted index would silently mislabel buckets while still matching the written contract.
   - Please update the contract to say boundaries and names are treated as pairs before filtering/conversion/sorting: drop the name with a dropped boundary, sort pairs by emitted boundary value, then label each bucket from the sorted pair's name (falling back only when that paired name is missing/empty). Extra names without boundaries remain ignored.

## Non-blocking notes

- `STATUS.md:131`/`:133` still leaves `_meta.bucket_method` unclear for unavailable results where no fixed-width edges can be computed. Consider explicitly stating whether unavailable payloads omit `bucket_method`, set it to `null`, or report the zone method if zones were selected before stream validation.
- The execution-log rows remain appended under `## Notes` (`STATUS.md:142-145`) rather than the `## Execution Log` table. This is STATUS hygiene, not a contract blocker.

Tests were not run; this step only changes task/status documentation.

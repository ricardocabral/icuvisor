# Review R008 — Plan Review for Step 3: Autodetect + verify

Verdict: **REVISE**

I could not find a Step 3 implementation plan to approve. `STATUS.md` still contains the completed Step 2 plan, but the Step 3 section only lists checklist items. Before coding, please add a concrete Step 3 plan to `STATUS.md` and resubmit.

## Required plan details

1. **Profile fetch abstraction and endpoint.**
   - Specify the injected dependency `RunSetup` will use for tests (for example, a small `SetupProfileClient`/function on `SetupOptions`) so Step 3 can be tested without network access.
   - Do not route through the MCP tool.
   - Do not use the existing `intervals.NewClient(...).GetAthleteProfile` unchanged for autodetection: that constructor requires a configured athlete ID and `GetAthleteProfile` calls `/athlete/{id}`, while this task verifies/discovers via `/athlete/0/profile`.

2. **Verified data model passed forward to Step 4.**
   - State what normalized setup result will be produced after Step 3: canonical `athlete_id`, confirmed `timezone`, display name, FTP for the success/test message, and API key retained only in memory until Step 4.
   - Keep the sequence explicit: no keychain or config writes occur until after successful online verification, or until `--offline` has collected and validated manual fields.

3. **Auth/network error handling.**
   - Map 401/403 using `errors.Is(err, intervals.ErrUnauthorized)` (or the equivalent sentinel) to the exact user-facing message from the prompt, then exit non-zero with no writes.
   - Treat transport/network failures separately: print the “Could not reach intervals.icu. Nothing was written…” guidance and do not write unless `--offline` was explicitly supplied.
   - Ensure returned errors do not include the API key or raw secret-bearing args.

4. **Offline mode behavior.**
   - Plan the full offline path: skip HTTP entirely, prompt for athlete ID manually, normalize via `config.NormalizeAthleteID`, prompt for timezone, validate it, and then pass those values to Step 4.
   - Clarify that offline mode cannot display connected athlete name/FTP because it intentionally skipped verification.

5. **Timezone confirmation/validation.**
   - Specify how `time.Local` is converted into a saveable timezone. Be careful with Go’s common `time.Local.String() == "Local"`; the persisted config should be an IANA zone accepted by `time.LoadLocation`, not an ambiguous placeholder.
   - Confirm the re-prompt loop for overrides: invalid pasted zones should not fall through to Step 4.

6. **FTP selection for the success message.**
   - Define which sport setting supplies the displayed FTP when multiple sport settings are present, and how the message behaves when no positive FTP exists.

7. **Tests for Step 3.**
   - Add table-driven cases for: successful online profile normalization; profile ID without `i` prefix; 401/403 no writes; network failure no writes; `--offline` skips the profile client and validates manual ID/timezone; invalid timezone re-prompts or errors; and no-FTP profile success message.

Once these design points are in `STATUS.md`, the Step 3 plan should be reviewable.

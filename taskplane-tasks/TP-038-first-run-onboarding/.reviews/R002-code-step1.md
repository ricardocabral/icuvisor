# R002 Code Review — Step 1: UX script

## Verdict: changes requested

Step 1 now has a useful first draft, but it still leaves several required UX decisions ambiguous or inconsistent with the task/R001 feedback. I would not move to implementation until these are tightened in `STATUS.md`.

## Findings

1. **Existing config-file behavior conflicts with the task.**  
   `STATUS.md` scripts `A config file already exists ... Overwrite? [y/N]`, but the task says to refuse clobbering an existing config file without `--force`, and R001 asked to script that branch explicitly. Please change the transcript to a non-interactive refusal unless `--force` is supplied, and add the `--force` success wording if you want to document it.

2. **Timezone override/error branches are missing.**  
   The happy path asks `Detected timezone: Europe/Madrid. Use this? [Y/n]`, but there is no transcript for answering `n`, entering a custom IANA name, invalid-zone rejection, and retry. Step 1/R001 specifically asked for the override path and invalid IANA handling so later tests can pin exact copy.

3. **No fallback for undetectable IANA timezone.**  
   The script assumes detection yields `Europe/Madrid`. R001 called out that `time.Local` may only surface `Local`; `STATUS.md` should include the fallback prompt, e.g. `Could not detect an IANA timezone. Enter one like Europe/Madrid:`.

4. **Credential-store wording is still platform-ambiguous.**  
   The generic v0.5 flow says `OS keychain`, then the success line says `macOS Keychain`. Either mark this script as the macOS manual-sweep copy or define platform-specific wording for macOS Keychain / Windows Credential Manager / Linux Secret Service so the implementation does not bake macOS copy into all platforms.

5. **Autodetect client gap is not documented.**  
   R001 requested a note that profile validation must use `/athlete/0/profile` before an athlete ID is known, while the existing client path may be athlete-ID based. `STATUS.md` still does not mention this, which leaves a trap for Step 3 implementation.

6. **Masking/testability note is incomplete.**  
   The note confirms `golang.org/x/term` and no fancy prompt library, but it does not record the planned injected password-reader abstraction for tests. Since `term.ReadPassword` requires a terminal FD, add this now to avoid brittle fake-stdin tests.

7. **Review history line is inaccurate.**  
   `STATUS.md` says `Review R001 | plan Step 1: APPROVE`, but `R001-plan-step1.md` has `Verdict: changes requested`. Please correct the status note so future reviewers can trust the task history.

## What looks good

- The happy-path copy includes masked key input, credential check/autodetect, canonical athlete ID, FTP, save location, final test connection, and the Claude Desktop next-step link.
- The 401/403 and network-failure messages are short and actionable, and the network branch clearly says nothing was written.
- The masking decision avoids a new prompt framework and calls out that `golang.org/x/term` is not yet in `go.mod`.

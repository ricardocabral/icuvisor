# R001 Plan Review — Step 1: UX script

## Verdict: changes requested

`STATUS.md` does not yet contain the Step 1 deliverable. It only has the checklist and two runtime notes, so there is no UX script to approve. Before coding starts, paste the actual terminal transcript(s) into `STATUS.md` and explicitly record the password-masking approach.

## Required fixes before approval

1. **Add the prompt sequence to `STATUS.md`.** Include the happy path end-to-end, not just the sample from the prompt. It should show:
   - welcome + settings URL;
   - masked API-key input;
   - initial credential check/autodetect result with canonical `i12345` athlete ID;
   - timezone detection, confirmation, and override path;
   - keychain + config save locations;
   - final saved-configuration test connection showing athlete name + FTP;
   - next-step pointer to `docs/clients/claude-desktop.md`.

2. **Script the important branches, not only the happy path.** At minimum add short transcript snippets for:
   - existing keychain key: `An API key is already stored. Overwrite? [y/N]` with default No;
   - existing config file: refuse without `--force`;
   - 401/403: `API key not accepted by intervals.icu. Double-check the key on https://intervals.icu/settings.` and no writes;
   - network failure: no writes, with a clear note that `--offline` is available and writes blindly;
   - offline mode: make the risk obvious in the copy;
   - timezone override with invalid IANA zone and retry.

3. **Confirm `golang.org/x/term` details precisely.** The status note should say the implementation will use `golang.org/x/term.ReadPassword` and no prompt framework. Also note that `golang.org/x/term` is a new direct dependency, so it must be added intentionally and license-compatible. For testability, plan for an injected password-reader abstraction; `ReadPassword` needs a terminal file descriptor and should not make fake-stdin tests brittle.

4. **Avoid promising an IANA timezone if the OS lookup cannot provide one.** The task says to read from `time.Local`, but Go often reports the local zone as `Local`, which is not the desired persisted IANA value. The UX script should include a fallback such as `Could not detect an IANA timezone. Enter one like Europe/Madrid:` rather than assuming `Europe/Madrid` is always available.

5. **Be explicit about platform-specific storage wording.** The sample says `macOS Keychain`; either limit the script note to the macOS manual sweep or define copy that varies by OS (`macOS Keychain`, Windows Credential Manager, Linux Secret Service/keyring). The task is security-sensitive, so the user-facing copy should accurately name where the secret was stored.

6. **Call out the current client gap for autodetect.** Existing `intervals.NewClient` requires an athlete ID and `GetAthleteProfile` calls `/athlete/{id}`, while the task requires validating the pasted key against `/athlete/0/profile` before the athlete ID is known. The Step 1 notes should acknowledge this so the later plan does not accidentally require a dummy athlete ID or call the MCP tool.

## Minor suggestions

- Include exact capitalization/punctuation for prompts now; tests can then assert stable output without churn.
- Keep error messages short and actionable, with no raw internal error strings in the transcript shown to athletes.
- Show that the config file message never mentions `api_key` and that only non-secret fields are written.


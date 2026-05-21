# Plan Review — Step 3: Add a repeatable local smoke checklist

Verdict: **Approved with minor guidance**

The planned Step 3 scope matches the prompt: it covers `icuvisor version`, `make build`, Claude Desktop discovery/callability for `get_athlete_profile`, an anonymized success response shape, and the note that unit tests are offline while manual smoke needs real intervals.icu credentials.

Minor guidance for implementation:

1. Put the detailed checklist in `docs/clients/claude-desktop.md` near the setup/troubleshooting sections, and keep `STATUS.md` limited to progress/results. If adding only one checklist, the client guide is the most durable operator-facing location.
2. Make the checklist copy/pasteable but secret-safe:
   - do not instruct users to print `.env` or config file contents containing the API key;
   - use placeholder paths/IDs only;
   - explicitly say not to paste API keys or real athlete IDs into Claude/chat.
3. Ensure the expected response shape mirrors the current tool contract without real personal data. A good sample should include placeholders for `athlete_id`, `timezone`, `units`, `sport_settings`, and `_meta.server_version` / `_meta.include_full`, and avoid suggesting that `include_full` is required for the smoke test.
4. Repeat the restart/new-chat requirement in the checklist before the Claude Desktop verification, because MCP tool schemas can be cached per conversation.
5. Distinguish local/offline checks from the networked check: `make build` and `./bin/icuvisor version` should not require credentials; the Claude `get_athlete_profile` call does require a real intervals.icu account/API key.

No blocking changes are needed before proceeding.

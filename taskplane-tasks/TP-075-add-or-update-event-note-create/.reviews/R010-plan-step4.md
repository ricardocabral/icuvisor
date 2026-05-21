# R010 plan review — Step 4

Verdict: REQUEST CHANGES

The Step 4 plan has the right broad gates (`make build`, `make test`, `make test-race`, `make lint`, plus live MCP validation and cleanup), but it is too underspecified for the live re-validation acceptance criteria and could pass without actually exercising the repaired NOTE path.

Blocking changes to the plan:

- Specify the live MCP payload shape: use a NOTE create with a date-only `start_date_local` and a non-empty `name` (and no `type`, unless intentionally testing the optional field). If the payload already includes `T00:00:00`, the live run does not prove the tool fixed the date-only NOTE serialization defect.
- Explicitly verify the created event with `get_events` for the chosen date before deletion, not just by trusting the `add_or_update_event` response.
- Spell out cleanup mechanics and verification: capture the returned event ID, delete that exact event, then re-run `get_events` for the date and confirm the unique test name is absent. If using the MCP `delete_event` tool, include any required `.env-dev` delete-mode configuration; otherwise state the fallback cleanup method.
- Use a unique test name/date marker for this run so the create and cleanup checks cannot be confused with prior probe data or another live validation attempt.

Non-blocking implementation notes:

- Run live validation against the binary produced by `make build` so the smoke test covers the artifact being validated.
- Keep credentials sourced from `.env-dev` out of command output/history as much as practical, and do not paste raw live responses containing athlete/account metadata into tracked files.

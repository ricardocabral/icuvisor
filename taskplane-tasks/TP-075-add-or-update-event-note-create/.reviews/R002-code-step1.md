# Review R002 — Step 1 code review

Decision: **Request changes.**

The live-probe outcome is useful and the accepted request/response fixtures are structurally valid JSON, but Step 1 is not quite ready to approve because the committed artifact does not fully satisfy the agreed sanitization/documentation requirements.

## Findings

1. **Redact the captured calendar ID in the response fixture.**  
   `internal/intervals/testdata/events/note_create_response.json:7` keeps the live `calendar_id` as `1`. The prior review's fixture guidance required redacting calendar IDs along with athlete/event IDs. Even if this value is commonly `1`, it is live-account metadata in a committed probe capture. Replace it with a neutral placeholder such as `"CALENDAR_ID_PLACEHOLDER"` (or remove/null it if the fixture does not need the field), and re-check the fixture for any other account-specific IDs before continuing.

2. **Document the full R001 probe matrix or mark it incomplete.**  
   `taskplane-tasks/TP-075-add-or-update-event-note-create/STATUS.md:19-23` records date format, `type`, description, and name-required discoveries, but it does not state the result of the R001-required category-casing check (`NOTE` vs observed UI/API casing) or the name-only / description-only / name+description matrix. Since Step 1 is meant to isolate the upstream contract, please either add sanitized outcomes for those axes or leave the relevant checkbox incomplete until they are probed.

## Notes

- I ran the required diffs and validated both new JSON fixtures with `python -m json.tool`.
- I did not find committed scratch probe files.

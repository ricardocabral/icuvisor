# Plan Review — TP-077 Step 1

Verdict: **approved with two required amendments before running the live probe**.

## Required amendments

1. **Probe omitted `folder_id` separately from `folder_id: null`.**
   The current implementation omits `folder_id` when the caller does not provide one; `null` is a different upstream contract and would require a different client representation. The permutation matrix should include all three cases where practical:
   - real existing folder ID,
   - `folder_id` key omitted,
   - explicit `"folder_id": null`.

2. **Make credential-safe probing explicit.**
   Since this step sources `.env-dev` and may use curl, the plan should explicitly avoid putting API keys in shell history, process args, logs, or committed artifacts. Prefer a throwaway scratch Go probe that reads env vars and only prints sanitized status/body summaries, or a curl config/stdin approach; run with `set +x`; do not use command-line `-u`/headers containing the raw key if they can appear in process listings.

## Strong recommendations

- Record every successfully created workout ID in a non-committed scratch note and clean it up immediately after the relevant observation. Use `ICUVISOR_DELETE_MODE=full` only if cleaning up through MCP, or use the same direct API path/UI, then verify with `get_workout_library` / `get_workouts_in_folder`.
- Sanitize fixtures before writing them under `internal/intervals/testdata/workout_library/`: replace raw workout IDs, folder IDs, athlete IDs, timestamps/dates, and any account-specific names with stable synthetic values. Keep only the payload shape and fields needed for tests.
- Capture rejected permutation outcomes in `STATUS.md` or a temporary scratch note as status code + sanitized error summary. The committed fixtures only need the accepted request/response, but the decision trail should survive enough to justify the chosen fix.
- Ensure the direct POST exactly matches the production client endpoint and headers where relevant (`/athlete/{athlete}/workouts`, JSON content type, normal user agent/auth), so the probe isolates the payload contract rather than a transport difference.

With those amendments, Step 1 is appropriately scoped and aligned with the task requirements.

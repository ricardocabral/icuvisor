# TP-005 — Status

**Issue:** v0.1 — manual smoke docs
**State:** Ready

## Step 1: Plan the manual config and smoke test

**Status:** ⬜ Not started

- [ ] Identify exact v0.1 config inputs
- [ ] Check local `.env` availability for `INTERVALS_ICU_ATHLETE_ID` and `INTERVALS_ICU_API_KEY`; record only availability, not values
- [ ] Identify macOS Claude Desktop config file path and JSON shape
- [ ] Use placeholders only for secrets/IDs
- [ ] Write smoke-test plan in STATUS.md

## Step 2: Write manual setup documentation

**Status:** ⬜ Not started

- [ ] Document local build/install for v0.1
- [ ] Document intervals.icu API key acquisition
- [ ] Document config/env inputs
- [ ] Provide Claude Desktop macOS JSON config example with placeholders
- [ ] Explain MCP schema caching/new chat requirement
- [ ] Include troubleshooting for common startup/auth/config errors

## Step 3: Add a repeatable local smoke checklist

**Status:** ⬜ Not started

- [ ] Checklist for `icuvisor version`
- [ ] Checklist for `make build`
- [ ] Checklist for Claude Desktop tool listing/callability
- [ ] Expected anonymized `get_athlete_profile` response shape
- [ ] Note manual smoke requires a real intervals.icu account/API key

## Step 4: Align code UX with docs if necessary

**Status:** ⬜ Not started

- [ ] Tighten confusing user-facing errors without leaking secrets
- [ ] Ensure invalid config failures are short/actionable
- [ ] Point README quickstart to detailed client guide
- [ ] Update `CHANGELOG.md`

## Step 5: Verify v0.1 gate

**Status:** ⬜ Not started

- [ ] Run `make build`
- [ ] Run `make test`
- [ ] Run `make lint` if available
- [ ] Perform manual Claude Desktop smoke test if credentials are available, or record remaining human verification
- [ ] Confirm every v0.1 roadmap checkbox is represented in TP-001 through TP-005

## Discoveries

| Date | Finding | Impact |
| ---- | ------- | ------ |

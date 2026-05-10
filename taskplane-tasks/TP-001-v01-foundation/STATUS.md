# TP-001 — Status

**Issue:** v0.1 — foundation
**State:** Ready

## Step 1: Capture the foundation plan in STATUS.md

**Status:** ⬜ Not started

- [ ] Inspect current module/layout/Makefile/README/CI
- [ ] Decide minimal CLI shape for v0.1
- [ ] Decide internal package boundaries
- [ ] Write plan before changing source files

## Step 2: Implement the CLI and version foundation

**Status:** ⬜ Not started

- [ ] Keep `icuvisor version` working
- [ ] Delegate default startup from thin `main` to internal package
- [ ] Pass build version to lower layers
- [ ] Return errors from internal packages; handle exit in `main`

## Step 3: Implement minimal manual config loading

**Status:** ⬜ Not started

- [ ] Define typed v0.1 config inputs
- [ ] Load config from manual JSON and/or env with tested precedence
- [ ] Support/document safe local `.env` loading for `INTERVALS_ICU_ATHLETE_ID` and `INTERVALS_ICU_API_KEY` without printing secrets
- [ ] Normalize athlete IDs centrally
- [ ] Do not write API keys to disk
- [ ] Never log or echo API keys

## Step 4: Add tests for foundation behavior

**Status:** ⬜ Not started

- [ ] Table-driven tests for athlete-ID normalization
- [ ] Table-driven tests for config loading/validation/defaults/redaction
- [ ] Tests for short actionable invalid/missing config errors

## Step 5: Verify and document

**Status:** ⬜ Not started

- [ ] Run `go fmt ./...`
- [ ] Run `make test`
- [ ] Run `make build`
- [ ] Run `make lint` if available
- [ ] Update `CHANGELOG.md`

## Discoveries

| Date | Finding | Impact |
| ---- | ------- | ------ |

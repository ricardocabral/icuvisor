# Plan Review — TP-067 Step 1 (Inventory)

## Verdict: Needs changes before implementation

The current Step 1 entry in `STATUS.md` only repeats the prompt checkboxes. It does not yet contain the inventory or the proposed declaration-to-file mapping that this step is supposed to produce. For this task, that inventory is the guardrail against accidental behavior changes during the split, so I would not start moving code until it is recorded explicitly.

## Required plan fixes

1. **Record a complete top-level declaration inventory in `STATUS.md`.**
   The inventory should cover every current top-level declaration in `internal/config/config.go`, including the TP-049/TP-062-era additions that are already present (`DebugMetadata`, `ParseDebugMetadata`, `Config.LogValue`). At minimum, assign these declarations to target files:

   - `config.go`: env/default constants, `Transport`, transport constants, `Transport.String`, `Config`, `APIKeySource`, API key source constants, `Options`, `WriteOptions`, and the public `Load` entry point if you keep a thin wrapper there.
   - `load.go`: file/env composition (`readJSONConfig`, `processEnv`, `rawFromEnv`, `rawConfig.merge`, `shouldSet`, `warnLegacyAPIKey`, possibly `fileConfig` and `rawConfig`).
   - `validate.go`: `validate` plus the planned per-field validators; probably `ParseDebugMetadata` unless you choose a more specific file and document why.
   - `write.go`: `Write`, `writeConfigFile`, `writeFileConfig`.
   - `athlete.go`: `NormalizeAthleteID`, `NormalizeAthleteIDForDisplay`.
   - `httpbind.go`: `ValidateHTTPBindAddress`, `NormalizeHTTPBindAddress`, `HTTPBindAddressIsLoopback`, `splitHTTPBindAddress`.
   - `dotenv.go`: `readDotEnv`, `cleanEnvValue`, `recognizedEnvKey`.
   - `redaction.go`: `Config.String`, `Config.LogValue`.
   - **Explicitly decide where `DefaultPath` lives.** It is an exported default-path helper, and the prompt calls default-path resolution out as a separate concern, but the suggested layout does not name a `path.go`. Either keep it in `config.go` with a short rationale or add a focused `path.go`; do not leave it implicit.
   - **Explicitly decide where `EffectiveCoachMode` / `CoachModeEnabled` live.** They are small exported config behavior helpers and are not named in the suggested layout. Keeping them in `config.go` is probably fine, but record it.

2. **Clarify how `Load` satisfies the layout.**
   The prompt says `config.go` should hold types, defaults, and the public `Load` entry point, while `load.go` should hold composition. The plan should say whether `Load` remains in `config.go` as a thin wrapper around an unexported loader in `load.go`, or whether you are intentionally putting the exported `Load` in `load.go`. Without this, the mechanical split can accidentally miss the acceptance criterion that `config.go` holds the public entry point.

3. **Make the validation slicing plan precise.**
   Step 1 should name the sub-validators and their call order so Step 2 can be mechanical. Preserve the current validation order and error strings:

   1. API key presence/source handling
   2. coach mode parse
   3. coach config validation
   4. athlete ID resolution depending on effective coach mode
   5. timezone default/load
   6. API base URL default/validation/trim
   7. HTTP timeout default/parse
   8. transport default/parse
   9. HTTP bind default/normalization
   10. final `Config` assembly, including `DeleteMode`, `Toolset`, `DebugMetadata`, `CoachMode`, and `Coach`

   Be careful not to “improve” behavior while slicing: no changed defaults, no changed precedence, no changed error messages, and no added .env/env key support unless already present. For example, `recognizedEnvKey` currently has its exact behavior; this refactor should not change it opportunistically.

4. **Account for imports and package-level dependencies in the inventory.**
   Moving redaction to `redaction.go` requires `log/slog` there for `LogValue`; moving HTTP bind helpers requires `net/netip` and `strconv`; moving `.env` requires `bufio`; etc. This does not need to be elaborate, but the inventory should identify any declarations whose move affects imports so each extraction commit remains small and compileable.

5. **Note test split candidates now, even if done later.**
   Step 1 does not need to edit tests, but the inventory should flag likely test destinations (`athlete_test.go`, `httpbind_test.go`, `write_test.go`, `redaction_test.go`, `validate_test.go`/`load_test.go`) so Step 2 does not accidentally mix broad test churn into source extraction commits.

## What looks good

- The task is correctly scoped to the `internal/config` package and preserves exported names/signatures.
- The intended file boundaries generally match the existing responsibility seams in `config.go`.
- The existing test suite already has coverage for athlete normalization, precedence, HTTP bind normalization, redaction, keychain precedence, coach validation, and write/load round trips, which should make the refactor safe once the inventory is explicit.

## Recommendation

Update `STATUS.md` with the complete declaration mapping and the exact validator breakdown before starting Step 2. After that, the implementation should be a straightforward mechanical split with tests after each extraction.

# TP-067-split-config-by-concern — Status

**Current Step:** Step 3: Verify
**Status:** ✅ Complete
**Last Updated:** 2026-05-17
**Review Level:** 1
**Review Counter:** 3
**Iteration:** 1
**Size:** S

---

### Step 1: Inventory

**Status:** ✅ Complete

- [x] List every top-level decl in `config.go` and assign each to a target file. Record in `STATUS.md`.
- [x] Slice `validate` into per-field sub-validators (e.g., `validateFTP`, `validateHTTPBind`) — same total logic.
- [x] Address plan review: record complete declaration-to-file mapping, including `DefaultPath`, coach helpers, TP-049/TP-062 declarations, imports, and test split candidates.
- [x] Address plan review: record exact `Load` placement and validation sub-validator call order before code movement.

#### Load placement and validation slicing plan

`config.go` keeps the exported `Load(ctx context.Context, opts Options) (Config, error)` signature as the public entry point required by acceptance criteria. Its body will delegate to an unexported helper in `load.go` (for example `load(ctx, opts)`) that contains the current file + `.env` + process env + keychain + flag composition logic. This keeps callers unchanged while making `config.go` types/defaults/public entry point only.

`validate(raw rawConfig) (Config, error)` remains the single validation entry point in `validate.go`. It will call sub-validators in this exact current order, preserving defaults, precedence, trimming, and error strings:

1. `validateAPIKey(raw)` trims `raw.apiKey`; returns the existing missing-key error with `EnvAPIKey`, `credstore.ServiceName`, and `credstore.IntervalsAPIKeyAccount`.
2. `validateCoachMode(raw)` calls `coach.ParseMode(raw.coachMode)`.
3. `validateCoachConfig(raw, coachMode)` copies `*raw.coach` when present, then calls `coach.ValidateConfig(rawCoach, coachMode, NormalizeAthleteID)`.
4. `validateAthleteID(raw, coachMode, coachConfig)` uses `coachConfig.DefaultAthleteID` when `coach.EffectiveMode(coachMode, coachConfig) == coach.ModeOn`; otherwise calls `NormalizeAthleteID(raw.athleteID)`.
5. `validateTimezone(raw)` applies `DefaultTimezone`, calls `time.LoadLocation`, and returns the existing invalid timezone message.
6. `validateAPIBaseURL(raw)` applies `DefaultAPIBaseURL`, parses with `url.Parse`, requires absolute `http`/`https`, returns the existing invalid URL message, and leaves final slash trimming to final assembly.
7. `validateHTTPTimeout(raw)` applies `DefaultHTTPTimeout`, parses positive durations with `time.ParseDuration`, and returns the existing invalid timeout message.
8. `validateTransport(raw)` applies `TransportStdio`, lowercases explicit values, and preserves the existing stdio/http error.
9. `validateHTTPBind(raw)` applies `DefaultHTTPBindAddress`, then calls `NormalizeHTTPBindAddress` and returns its existing errors unchanged.
10. `buildConfig(raw, apiKey, athleteID, loc, baseURL, timeout, transport, httpBindAddress, coachMode, coachConfig)` assembles `Config` with `APIKeySource`, `DeleteMode`, `Toolset`, `DebugMetadata`, `CoachMode`, and `Coach` exactly as today.

#### Top-level declaration inventory and target files

| Declaration(s)                                                                                                                                                                                                                                                             | Target file    | Import / dependency notes                                                                                                                               |
| -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `EnvAPIKey`, `EnvAthleteID`, `EnvConfigPath`, `EnvTimezone`, `EnvAPIBaseURL`, `EnvHTTPTimeout`, `EnvTransport`, `EnvHTTPBind`, `EnvDotEnvPath`, `EnvDebugMetadata`, `EnvCoachMode`; `DefaultAPIBaseURL`, `DefaultTimezone`, `DefaultHTTPTimeout`, `DefaultHTTPBindAddress` | `config.go`    | `time` for `DefaultHTTPTimeout`; constants remain with public types/defaults.                                                                           |
| `Transport`, `TransportStdio`, `TransportHTTP`, `Transport.String`                                                                                                                                                                                                         | `config.go`    | No extra imports.                                                                                                                                       |
| `Config`, `APIKeySource`, `APIKeySourceEnv`, `APIKeySourceKeychain`, `APIKeySourceFile`, `Options`, `WriteOptions`                                                                                                                                                         | `config.go`    | `time`, `coach`, `credstore`, `safety`.                                                                                                                 |
| `Load`                                                                                                                                                                                                                                                                     | `config.go`    | Public entry point remains in `config.go` as a thin wrapper; composition moves to `load.go`.                                                            |
| unexported loader helper created from current `Load` body                                                                                                                                                                                                                  | `load.go`      | `context`, `errors`, `fmt`, `log/slog`, `os`, `strings`, `credstore`; called by the public `Load` wrapper.                                              |
| `fileConfig`, `rawConfig`, `readJSONConfig`, `processEnv`, `rawFromEnv`, `rawConfig.merge`, `shouldSet`, `warnLegacyAPIKey`                                                                                                                                                | `load.go`      | `encoding/json`, `errors`, `fmt`, `log/slog`, `os`, `strings`, `safety`; `fileConfig` kept with JSON read; `rawConfig` kept with merge/env composition. |
| `DefaultPath`                                                                                                                                                                                                                                                              | `path.go`      | `fmt`, `os`, `path/filepath`; exported default-path helper gets a focused default-path concern file instead of staying implicit in `config.go`.         |
| `Write`, `writeConfigFile`, `writeFileConfig`                                                                                                                                                                                                                              | `write.go`     | `context`, `encoding/json`, `errors`, `fmt`, `os`, `path/filepath`, `strings`, `time`; keeps atomic write and write-only JSON shape together.           |
| `NormalizeAthleteID`, `NormalizeAthleteIDForDisplay`                                                                                                                                                                                                                       | `athlete.go`   | `errors`, `strings`, `unicode`.                                                                                                                         |
| `Config.String`, `Config.LogValue`                                                                                                                                                                                                                                         | `redaction.go` | `fmt`, `log/slog`; includes TP-062 `LogValue` with `String` so redaction stays centralized.                                                             |
| `Config.EffectiveCoachMode`, `Config.CoachModeEnabled`                                                                                                                                                                                                                     | `config.go`    | `coach`; small behavior helpers remain beside `Config` because they are public-ish config semantics rather than a separate concern.                     |
| `validate` plus new sub-validators (`validateAPIKey`, `validateCoachMode`, `validateCoachConfig`, `validateAthleteID`, `validateTimezone`, `validateAPIBaseURL`, `validateHTTPTimeout`, `validateTransport`, `validateHTTPBindAddress`, `buildConfig`)                     | `validate.go`  | `errors`, `fmt`, `net/url`, `strings`, `time`, `coach`, `credstore`, `safety`; preserves current order and error strings.                               |
| `ParseDebugMetadata`                                                                                                                                                                                                                                                       | `validate.go`  | `strings`; debug metadata is validation/parsing of raw config into `Config`.                                                                            |
| `ValidateHTTPBindAddress`, `NormalizeHTTPBindAddress`, `HTTPBindAddressIsLoopback`, `splitHTTPBindAddress`                                                                                                                                                                 | `httpbind.go`  | `errors`, `net/netip`, `strconv`, `strings`.                                                                                                            |
| `readDotEnv`, `cleanEnvValue`, `recognizedEnvKey`                                                                                                                                                                                                                          | `dotenv.go`    | `bufio`, `context`, `errors`, `fmt`, `os`, `strings`, `safety`; preserve exact recognized key set.                                                      |

#### Test split candidates

- `athlete_test.go`: `TestNormalizeAthleteIDForDisplay`, `TestNormalizeAthleteID`.
- `load_test.go`: load precedence/defaults, debug metadata from env, explicit/default `.env`, config path, config file errors, keychain precedence, legacy API key warning.
- `dotenv_test.go`: `.env` absent/fill behavior and recognized-key edge cases if split from load tests.
- `validate_test.go`: validation errors, coach mode/config validation, `ParseDebugMetadata`, transport selection defaults.
- `httpbind_test.go`: `TestLoadTransportAndHTTPBindSelection`, `TestValidateHTTPBindAddress`.
- `write_test.go`: write non-secret fields, round trip, clobber behavior.
- `redaction_test.go`: `TestConfigLogValueStructuredAndRedacted`, `TestConfigStringRedactsSecret`.

### Step 2: Mechanical split

**Status:** ✅ Complete

- [x] Extract default path resolution into `path.go`, keeping `DefaultPath` exported and running targeted config tests.
- [x] Extract athlete ID helpers into `athlete.go`, keeping exported names/signatures and running targeted config tests.
- [x] Extract redaction helpers into `redaction.go`, keeping `Config.String`/`LogValue` behavior and running targeted config tests.
- [x] Extract HTTP bind helpers into `httpbind.go`, keeping exported names/signatures and running targeted config tests.
- [x] Extract `.env` parsing helpers into `dotenv.go`, preserving recognized-key behavior and running targeted config tests.
- [x] Extract write helpers into `write.go`, preserving atomic write behavior and running targeted config tests.
- [x] Extract load composition helpers into `load.go`, leaving `Load` as the public `config.go` wrapper and running targeted config tests.
- [x] Extract validation helpers into `validate.go`, preserving call order/error strings and running targeted config tests.

### Step 3: Verify

**Status:** ✅ Complete

- [x] `make build` / `test` / `test-race` / `lint`.
- [x] Round-trip test: load a fixture config, write it, reload it — output must be byte-identical to pre-refactor.
- [x] `wc -l internal/config/*.go` — each file ≤ ~250 LOC.

## Discoveries

| Date | Area | Discovery |
| ---- | ---- | --------- |

## Blockers

| Date | Blocker | Attempts |
| ---- | ------- | -------- |

## Execution Log

| Date             | Event          | Notes                            |
| ---------------- | -------------- | -------------------------------- |
| 2026-05-17 21:55 | Task started   | Runtime V2 lane-runner execution |
| 2026-05-17 21:55 | Step 1 started | Inventory                        |
| 2026-05-17 21:57 | Review R001    | plan Step 1: UNKNOWN             |
| 2026-05-17 22:01 | Review R002    | plan Step 1: APPROVE             |
| 2026-05-17 22:04 | Review R003    | plan Step 2: APPROVE             |
| 2026-05-17 22:17 | Worker iter 1 | done in 1326s, tools: 105 |
| 2026-05-17 22:17 | Task complete | .DONE created |

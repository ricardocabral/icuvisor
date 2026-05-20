# Changelog

All notable changes to icuvisor are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Added shared analyzer response scaffolding so planned analyzer tools consistently emit mandatory `_meta.method`, `_meta.source_tools`, `_meta.n`, `_meta.missing_days`, `_meta.missing_action`, and `_meta.insufficient_sample` fields.
- `icuvisor://analysis-formulas` MCP Resource documents canonical analyzer formula refs for HR drift, Pw:HR decoupling, polarization index, EF, VI, and z-score.
- `get_hr_curves` and `get_pace_curves` expose upstream-computed heart-rate and pace curves as full-toolset siblings to `get_power_curves`, with terse/default buckets, raw `include_full` payloads, and athlete-preferred pace units.
- Added NOTE-category examples for `add_or_update_event` so assistants and users can discover nutrition plans, travel logistics, daily reminders, and coach annotations without a separate note tool.
- Added reusable closed `analysis_metric` enum helpers for planned analyzer tools, including canonical schema values, conservative aliases, source metadata, and concise unknown-metric hints.
- `get_gear_list` lists intervals.icu gear IDs/names in the full toolset with a manual refresh path for the per-athlete gear cache.
- Activity reads now surface `gear_id`, resolved `gear_name`, and explicit `gear_resolution` statuses when upstream exposes gear IDs.
- Activity and wellness reads now disambiguate nutrition fields: activity `calories` is exposed as `calories_burned`, while wellness intake/macros use `calories_intake`, `carbs_g`, `protein_g`, and `fat_g` with `_meta.field_semantics` labels.
- Homebrew install path: releases now produce a `darwin_universal` tarball and publish a Homebrew formula to the [`ricardocabral/homebrew-icuvisor`](https://github.com/ricardocabral/homebrew-icuvisor) tap. Install with `brew install ricardocabral/icuvisor/icuvisor`. The macOS tarball binary is not yet codesigned independently of the DMG, so first run may require dismissing Gatekeeper; signing the tarball binary is tracked as a follow-up.

### Changed

- `get_wellness_data` provenance now reports provider-native sleep/readiness `native_scale` labels for Garmin, WHOOP, Oura, and Polar, while unresolved sources report `unknown` instead of a guessed device scale.
- Write-tool echo responses now strip upstream null keys by default, including custom-item create/update responses, while preserving meaningful zero, false, and empty-string values.
- Documented the TP-102 remote-connector decision: icuvisor will not add legacy SSE or generic public-tunnel recipes for ChatGPT-style remote custom connector UIs before the hosted relay; local stdio and loopback Streamable HTTP remain the supported ChatGPT paths.
- `icuvisor setup` now writes a non-secret `credential_ref` to generated config files so users and docs can see the OS keychain service/account while the API key remains outside JSON. Setup stores and verifies the keychain credential before writing the config file, so keychain failures do not leave a fresh onboarding config behind.

## [0.0.2] - 2026-05-19

### Changed

- `icuvisor setup` always prompts for the athlete ID and requires the `i` prefix (for example `i12345`). intervals.icu does not expose the ID through the API, so the previous `/athlete/0/profile` autodetect could silently succeed with an empty ID; the API key and athlete ID are now verified together against `/athlete/{id}`.
- The server loads the platform default config (`os.UserConfigDir()/icuvisor/config.json`) when neither `--config` nor `ICUVISOR_CONFIG` is set, so a fresh `icuvisor setup` works without needing an env file or explicit `--config` path.
- Startup logs drop the noisy "env file not found" line on healthy runs, log when the API key was loaded from the OS keychain, and include the active athlete ID and key source on the "server starting" line.

## [0.0.1] - 2026-05-18

### Added

- Initial public release.

[Unreleased]: https://github.com/ricardocabral/icuvisor/compare/v0.0.2...HEAD
[0.0.2]: https://github.com/ricardocabral/icuvisor/compare/v0.0.1...v0.0.2
[0.0.1]: https://github.com/ricardocabral/icuvisor/releases/tag/v0.0.1

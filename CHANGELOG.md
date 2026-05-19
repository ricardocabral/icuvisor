# Changelog

All notable changes to icuvisor are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Homebrew install path: releases now produce a `darwin_universal` tarball and publish a Homebrew formula to the [`ricardocabral/homebrew-icuvisor`](https://github.com/ricardocabral/homebrew-icuvisor) tap. Install with `brew install ricardocabral/icuvisor/icuvisor`. The macOS tarball binary is not yet codesigned independently of the DMG, so first run may require dismissing Gatekeeper; signing the tarball binary is tracked as a follow-up.

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

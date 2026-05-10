# Changelog

All notable changes to icuvisor are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- v0.1 foundation CLI with thin `main`, internal app startup wiring, build-version propagation, and `icuvisor version` support.
- Manual v0.1 config loader for JSON/env/`.env` inputs with centralized athlete-ID normalization and secret redaction.
- Foundation tests covering CLI delegation, config precedence/defaults, validation errors, redaction, and athlete-ID normalization.
- Initial repository scaffolding: Go module, Makefile, GoReleaser config, GitHub Actions CI/release pipelines, golangci-lint config, issue/PR templates, CODEOWNERS.
- Project documentation: README, CONTRIBUTING, CODE_OF_CONDUCT, SECURITY, ROADMAP, CHANGELOG.
- PRD for v1.0 (`docs/prd/PRD-icuvisor.md`).

[Unreleased]: https://github.com/ricardocabral/icuvisor/compare/HEAD...HEAD

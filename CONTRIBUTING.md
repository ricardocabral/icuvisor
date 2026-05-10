# Contributing to icuvisor

Thank you for considering a contribution. This project is built for amateur athletes, and every fix, doc tweak, or new tool helps the whole community.

By participating in this project you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md).

## Ways to contribute

- **Report a bug** — open a [bug report](https://github.com/ricardocabral/icuvisor/issues/new?template=bug_report.yml).
- **Request a feature** — open a [feature request](https://github.com/ricardocabral/icuvisor/issues/new?template=feature_request.yml).
- **Improve docs** — typos, clarifications, and client setup guides are very welcome.
- **Send code** — see "Submitting a change" below.
- **Test a release candidate** — comment on the relevant tracking issue.

For larger work, please open a discussion or issue first so we can agree on scope before you spend time coding.

## Development setup

Prerequisites:

- Go 1.23 or newer.
- [`golangci-lint`](https://golangci-lint.run) (matches `.golangci.yml`).
- (optional) [`goreleaser`](https://goreleaser.com) for release dry-runs.

Clone and verify your environment:

```bash
git clone https://github.com/ricardocabral/icuvisor.git
cd icuvisor
make build
make test
make lint
```

## Submitting a change

1. **Fork** the repository and create a topic branch from `main`:
   ```bash
   git checkout -b feat/short-description
   ```
2. **Code**:
   - Keep changes focused. One logical change per PR.
   - Run `go fmt ./...` and `make lint` before pushing. CI will block on either.
   - Add or update tests for behaviour changes. Aim for table-driven tests in Go.
   - Don't introduce abstractions beyond what the change requires.
3. **Commit** using [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/):
   ```
   feat(tools): add get_power_curves tool
   fix(client): retry on 429 with exponential backoff
   docs(readme): document the brew install path
   ```
   Prefixes we use: `feat`, `fix`, `perf`, `refactor`, `docs`, `test`, `ci`, `build`, `chore`.
4. **Update** `CHANGELOG.md` under `[Unreleased]` for user-visible changes.
5. **Open a PR** against `main`. Fill in the template. Link related issues.
6. **CI must pass** before review.

## Code style

- Idiomatic Go: follow [Effective Go](https://go.dev/doc/effective_go) and the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).
- Public packages live in `pkg/`; everything else in `internal/`. Default to `internal/`.
- Error wrapping: use `fmt.Errorf("doing X: %w", err)`. Never swallow errors silently.
- Logging: use `log/slog` from the stdlib. No third-party loggers.
- Avoid `panic` outside `main`. Return errors.
- Keep tool responses **terse by default**. Heavy payloads must require explicit opt-in.

## MCP tool conventions

Each new tool must:

- Have a clear name in `snake_case`, matching the catalog in the PRD.
- Declare every argument with a JSON Schema description an LLM can read.
- Include scale metadata in its description for any ambiguous numeric field (e.g. `feel` is 1-5, `sleepQuality` is 1-4).
- Render dates in the athlete's configured timezone and normalize athlete IDs to `i12345`.
- Have a terse default response under ~500 tokens and an `include_full: bool` opt-in.

## Security issues

Do **not** open a public issue for vulnerabilities. See [SECURITY.md](SECURITY.md).

## Licensing of contributions

By contributing, you agree that your contributions will be licensed under the [MIT license](LICENSE) covering the project.

Do not paste code from GPL-licensed projects (including `mvilanova/intervals-mcp-server`) into PRs. icuvisor is a clean-room implementation against intervals.icu's public API.

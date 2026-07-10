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
- Render dates in the athlete's configured timezone and preserve athlete-ID shapes: trim whitespace, lowercase only an optional leading `I`, validate digits, and never add or remove the `i` prefix.
- Have a terse default response under ~500 tokens and an `include_full: bool` opt-in.

### Tool schema snapshots

Stable MCP tools follow an additive-only argument schema rule because clients may cache schemas for an entire conversation. Snapshot files live in `internal/tools/schema_snapshot/<tool_name>.json` and are generated from the live registry:

```bash
go run ./scripts/snapshot_tool_schemas.go
```

Snapshots use canonical JSON: two-space indentation, sorted object keys from Go's JSON encoder, and one trailing newline. When a PR adds a new optional argument to an existing stable tool, regenerate and commit the updated snapshot. Do not remove or rename an existing stable-tool argument in place; ship a new tool name instead so cached clients do not call a mismatched schema. New tools add new snapshot files and may evolve until they are declared stable.

### Analyzer formula drift

The `icuvisor://analysis-formulas` resource and analyzer `_meta.formula_ref` outputs are public contracts. Changing a canonical formula ref, formula text, or pinned analyzer output is a breaking definition-drift event that needs an explicit product decision and an update to the golden fixtures in `internal/resources/testdata/analysis_formulas.md` and `testdata/analysis/`.

### Tool-name confusability

Tool descriptions in the same prefix/domain cluster must have distinguishing first sentences. CI compares normalized first sentences within each cluster using token Jaccard similarity and fails pairs at or above `0.58`. If the check fails, rewrite one first sentence to make the access pattern and payload shape obvious to an LLM reading only that sentence; do not rename the tool.

### Tool routing smoke eval

Run `make eval-tool-routing` to validate the local first-tool-call routing fixtures. By default the command validates fixtures and catalog wiring only; it does not call a model provider unless explicitly configured.

Provider-backed runs currently support Anthropic-compatible chat completions via environment variables:

```bash
ICUVISOR_ROUTING_EVAL_PROVIDER=anthropic \
ANTHROPIC_API_KEY=sk-ant-... \
make eval-tool-routing
```

Optional overrides are `ICUVISOR_ROUTING_EVAL_MODEL` for the model name and `ICUVISOR_ROUTING_EVAL_ANTHROPIC_URL` for an Anthropic-compatible endpoint.

When the provider is unset, `make eval-tool-routing` exits zero after validating fixture structure, expected-tool availability, and catalog construction, with each live model case reported as skipped. When a provider is configured, provider call errors or expected-versus-actual routing mismatches exit non-zero. The eval sends only registered tool names, descriptions, and JSON schemas to the provider; it does not execute icuvisor tool handlers and therefore does not call intervals.icu.

## Security issues

Do **not** open a public issue for vulnerabilities. See [SECURITY.md](SECURITY.md).

## Licensing of contributions

By contributing, you agree that your contributions will be licensed under the [MIT license](LICENSE) covering the project.

Do not paste code from GPL-licensed projects into PRs. icuvisor is a clean-room implementation against intervals.icu's public API.

# CLAUDE.md

Guidance for Claude Code (and any other AI assistant) working in this repository.

## What this project is

icuvisor is an open-source MCP server that connects [intervals.icu](https://intervals.icu) training data to AI assistants, shipped as a single signed Go binary. Read the [PRD](docs/prd/PRD-icuvisor.md) and [ROADMAP.md](ROADMAP.md) before making non-trivial changes — they own scope and phasing.

## Authoritative documents

- **What & why:** [`docs/prd/PRD-icuvisor.md`](docs/prd/PRD-icuvisor.md)
- **When / phasing:** [`ROADMAP.md`](ROADMAP.md)
- **Contributor rules:** [`CONTRIBUTING.md`](CONTRIBUTING.md)
- **Security policy:** [`SECURITY.md`](SECURITY.md)

If these conflict, the PRD wins for product behaviour, the roadmap wins for phasing, and CONTRIBUTING wins for process.

## Hard rules

1. **Clean-room.** This project is MIT-licensed and is built from intervals.icu's public API docs + black-box testing. **Never paste, paraphrase, or transliterate code from `mvilanova/intervals-mcp-server` (GPLv3) or any other copyleft source.** Inspiration from public API behaviour is fine; copying source is not.
2. **No GPL dependencies.** Check the license of every new module before adding it. Prefer stdlib first; permissive licenses (MIT, BSD, Apache-2.0, MPL-2.0) second.
3. **Default to `internal/`.** Only put a package in `pkg/` if external consumers genuinely need it.
4. **No `panic` outside `main`.** Return errors. Wrap with `%w`.
5. **Tools must be terse-by-default.** Heavy payloads (streams, raw samples) require an explicit `include_full: true` argument. Token budget is a product feature, not an afterthought.
6. **API keys live in the OS keychain.** Never log them. Never write them to disk in plain text. Never accept them as tool parameters from the LLM — they come from server config, not the conversation.
7. **Don't expose the HTTP transport beyond `127.0.0.1` by default.** A LAN bind is opt-in and must be documented.

## Project layout

```
cmd/icuvisor/        Binary entrypoint (main package only — keep it thin)
internal/intervals/  intervals.icu HTTP client (typed, tested against fixtures)
internal/mcp/        MCP server wiring + transports (stdio, Streamable HTTP)
internal/tools/      One file per tool, each implementing a small interface
internal/config/     Config loading, keychain access, athlete-ID/timezone normalization
docs/                PRD, roadmap-adjacent design docs, client setup guides
```

Add new tools as `internal/tools/<tool_name>.go` with a matching `_test.go`. Register them from a single `Register()` call so the catalog is greppable.

## Go conventions

- **Format:** `gofmt` + `goimports` with `-local github.com/ricardocabral/icuvisor`. CI fails on dirty diffs.
- **Lint:** `golangci-lint run ./...` must pass. Config: `.golangci.yml`. Don't disable a linter without a comment explaining why.
- **Errors:** sentinel errors for stable contract points (`var ErrNotFound = errors.New("not found")`), wrapped errors everywhere else (`fmt.Errorf("getting activity %s: %w", id, err)`). Use `errors.Is` / `errors.As` at call sites; never `err.Error() == "..."`.
- **Logging:** `log/slog` with structured fields. Use `slog.Default()` in libraries; do not import a global logger. Never log API keys, tokens, or raw athlete identifiers in a way that's hard to scrub later.
- **Context:** every function that does I/O or blocks takes `ctx context.Context` as the first argument. Honour cancellation. No `context.TODO()` in shipped code.
- **HTTP:** use a single shared `*http.Client` with a timeout. Set a `User-Agent: icuvisor/<version>`. Use `httpretry` (or equivalent) for 429/5xx with exponential backoff and jitter. Always close response bodies.
- **JSON:** prefer typed structs over `map[string]any`. Use `json.Decoder` for streams; `json.Unmarshal` for small bodies.
- **Time:** all times are `time.Time` in UTC at the boundary; render in the athlete's configured timezone only at the presentation layer.
- **Concurrency:** `errgroup.Group` for fan-out; `context` for cancellation; mutexes only as a last resort. Run `go test -race` locally before pushing.
- **Tests:** table-driven. Use `t.Run(tc.name, ...)`. Use `testdata/` fixtures for API responses. Never hit the network from tests — wrap intervals.icu calls behind an interface and stub it.
- **Generics:** fine when they remove duplication; not for cleverness.
- **Comments:** only when the *why* is non-obvious. Don't restate the code. Exported identifiers need a doc comment starting with the identifier name.

## MCP-server conventions

icuvisor uses `github.com/modelcontextprotocol/go-sdk`. Read its docs before changing transport code. General rules:

- **Tool names** are `snake_case`, match the catalog in PRD §7.2.C, and are stable — renames are breaking changes.
- **Schemas matter.** Every argument needs a JSON Schema description an LLM can read. Include units and scale ranges (`"feel: athlete-reported feel, scale 1-5"`, `"sleepQuality: 1-4"`). The LLM uses these descriptions to decide what to send.
- **Return shapes are part of the API.** Document them. Add a `_meta` field for things like `total_count`, `next_page`, and scale legends — clients can use it, LLMs can ignore it.
- **Pagination:** server-side. Default page size must fit comfortably inside a free-tier Claude context window. Expose a `next_page_token` (opaque string).
- **Destructive ops are registration-time gated.** `delete_event`, `delete_events_by_date_range`, and other write/delete tools declare their required capability and are registered only when `internal/safety` allows it for `ICUVISOR_DELETE_MODE`; never add a model-controlled `confirm: true` override.
- **Idempotency:** writes that can be safely retried should be idempotent. Document the ones that can't.
- **Errors back to the LLM** must be short, actionable, and free of internal stack traces. Log the detail; return the summary.
- **Athlete ID normalization:** accept `i12345` or `12345`; emit `i12345`. Centralize in `internal/config`.
- **Strava-imported activities:** detect via the upstream marker and label them in responses so the LLM doesn't hallucinate over `N/A` fields.
- **Coach mode:** the coach-scoped API key never leaves the server. The `athlete_id` argument selects which athlete the call targets — it is *not* a credential.

## Adding a tool — checklist

1. Confirm it's in the PRD §7.2.C catalog, or open an issue first.
2. Add a typed request/response struct in `internal/tools/<name>.go`.
3. Wire it into the registry; add it to the README catalog.
4. Add table-driven tests, including: terse default, `include_full` opt-in, scale-metadata in response, athlete-ID normalization, pagination if applicable.
5. Update `CHANGELOG.md` under `[Unreleased]`.
6. Manual smoke test against at least one MCP client.

## Build, test, release

- `make build` — local build into `./bin/icuvisor`.
- `make test` / `make test-race` — unit tests.
- `make lint` — golangci-lint.
- `make snapshot` — local GoReleaser dry-run.
- Tag `vX.Y.Z` on `main` to trigger the release workflow.
- Tags are immutable. If a release is broken, ship a new patch — never retag.

## What I want from Claude in this repo

- **Read the PRD section before answering scope questions.** Don't invent product behaviour from your training data.
- **Prefer editing existing files** to creating new ones; this repo is small enough that sprawl is expensive.
- **Don't write code commentary** ("// This function returns the activity"). Use a doc comment if the identifier is exported, otherwise leave it.
- **No emojis** in code, commits, or PR descriptions unless explicitly asked.
- **Conventional Commits** for every commit message. Lowercase subjects; imperative mood.
- **Ask before doing wide refactors.** A focused PR beats a sweeping one every time.
- **If you change the PRD or the roadmap, update both this file's pointers and `CHANGELOG.md` if user-visible.**

## What is *not* in scope

- A multi-tenant SaaS.
- Replacing intervals.icu's own UI.
- Hosting athlete data on our infrastructure (the future opt-in relay is the only exception).
- Adding direct Strava/TrainingPeaks ingestion to the icuvisor binary (those are future companion MCP servers).

When in doubt: smaller change, clearer commit, ask the human.

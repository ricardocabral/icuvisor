# R004 Plan Review — Step 3: Tests

Verdict: **APPROVE**, with test-shape guardrails.

The Step 3 plan matches the prompt: add focused `wireServer` coverage using the new package-private `deps` seams, and keep the existing `app_test.go` dispatch/startup coverage passing unchanged. I also ran `go test ./internal/app`; it currently passes.

## Guardrails for the new `wireServer` test

1. Prefer a new `internal/app/wire_test.go` test that calls `wireServer` directly with fake `deps`. Do not start transports or exercise `defaultStartServer` in this new coverage.
2. Assert configuration through captured constructor inputs, not by peeking into `mcpserver.Server` internals. In practice, fake `NewServer(ctx, mcpserver.Options)` should capture the options and return `&mcpserver.Server{}`.
3. Swap at least two collaborators, but the most valuable seams to fake are:
   - `NewClient`, to avoid real config validation/network-adjacent setup and capture `intervals.Options` including trimmed/defaulted version.
   - `NewResourceRegistry` and `NewToolRegistry`, to capture `ResourceOptions` / `RegistryOptions` and verify debug metadata, timezone fallback, parsed delete mode/toolset, coach-mode flags, and coach config propagation.
   - `NewSelectionStore` and `RecentToolCallRecorder`, to verify default coach athlete ID and recorder propagation into `mcpserver.Options`.
4. Use a custom logger from `deps.Logger` rather than changing `slog.Default()` in this test. That avoids global logger races and lets the test run safely with other package tests.
5. Keep fake registries minimal implementations of the existing registry interfaces; do not change production interfaces or broaden deps types just to make the test easier.
6. Cover the current version behavior explicitly: whitespace-only version becomes `dev`, and non-empty versions are trimmed before being passed to client/server/registry options.

## Suggested assertions

A single focused test can validate most behavior by capturing inputs from the fake constructors:

- `NewClient` receives the expected `config.Config` and normalized `Version`.
- `NewSelectionStore` receives `info.Config.Coach.DefaultAthleteID`.
- `RecentToolCallRecorder` result is the same recorder passed to `mcpserver.Options`.
- `mcpserver.Options` contains the expected config, version, logger, capability fallback, parsed toolset, selection store, prompt/resource/tool registries.
- Resource options preserve `Version`, `TimezoneFallback`, `DebugMetadata`, `DeleteMode`, `Toolset`, and `DisableAthleteProfile`.
- Tool registry options preserve `Version`, `TimezoneFallback`, `DebugMetadata`, `Capability`, `Toolset`, `CoachModeEnabled`, and `CoachConfig`.

Optional but useful: add a small error-path subtest for `NewClient` or `NewServer` returning an error, ensuring `wireServer` returns the error and a non-nil cleanup function without continuing construction.

With those constraints, Step 3 should give useful regression coverage for the extracted wiring without expanding scope or coupling tests to MCP server implementation details.

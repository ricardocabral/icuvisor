# Plan Review: Step 1 — Inventory the wiring graph

## Verdict

Approved, with one scope note to carry into the inventory. The step is appropriately small: read the existing `defaultStartServer`, record the construction order in `STATUS.md`, and make the `Deps` decision before moving code.

## Required inventory coverage

When completing Step 1, make sure the inventory includes both constructed collaborators and the non-construction side effects/derived inputs that must remain behaviorally unchanged:

1. `slog.Default()` logger selection.
2. Version trim/defaulting to `dev`, mutation back into `info.Version`, and startup log.
3. HTTP non-loopback warning path, without logging secrets or athlete IDs.
4. Capability fallback from `info.Capability` or `safety.NewCapability(info.DeleteMode)`.
5. `deleteMode := safety.ParseMode(capability.Mode())` and `toolset := safety.ParseToolset(info.Toolset.String())`, plus the resolved-mode/toolset logs.
6. `intervals.NewClient(intervals.Options{Config: info.Config, Version: info.Version})`.
7. `coach.NewSelectionStore(info.Config.Coach.DefaultAthleteID)`.
8. `defaultRecentToolCallRecorder()`, including the current warn-and-continue behavior on recorder setup failure.
9. `prompts.NewRegistry()`.
10. `resources.NewRegistryWithOptions(client, resources.ResourceOptions{...})`, including `DisableAthleteProfile: info.Config.CoachModeEnabled()`.
11. `tools.NewRegistryWithOptions(client, tools.RegistryOptions{...})`, including capability, toolset, coach mode, and coach config propagation.
12. `mcpserver.NewServer(ctx, mcpserver.Options{...})`.
13. The final transport dispatch: HTTP uses `RunStreamableHTTP(ctx, info.Config.HTTPBindAddress)`, otherwise stdio uses `Run(ctx)`. Even if `wireServer` only builds and returns the server, this run behavior needs to be called out as adjacent behavior that must stay unchanged.

## `Deps` recommendation

A package-private deps struct is worth introducing because there are more than four construction seams. Prefer a small `wireDeps`/`Deps` with factory functions and zero-value/default filling, rather than pre-built instances, so production behavior remains identical while tests can replace specific constructors. In particular, the useful seams are the intervals client constructor, selection-store constructor, recent-call recorder factory, prompt/resource/tool registry constructors, MCP server constructor, and optionally logger selection.

Also record that `wireServer` needs the full `ServerInfo`-derived inputs, not just `config.Config`: version, debug metadata, capability/delete mode, and toolset are all currently inputs to the wiring. If the final signature uses `config.Config` only, it must have another way to preserve those values; otherwise direct `defaultStartServer` behavior and existing tests can regress.

## Non-blocking notes

- Keep `wireServer` and the deps type package-private, as required.
- Do not let Step 1 become a refactor; `STATUS.md` should receive the inventory and the `Deps` decision only.
- Existing `app_test.go` has direct coverage of startup logging and HTTP transport dispatch; the inventory should mention these so Step 2 can preserve them or move them cleanly with `defaultStartServer`.

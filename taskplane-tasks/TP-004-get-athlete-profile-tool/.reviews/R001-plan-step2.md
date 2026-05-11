# Plan Review: TP-004 Step 2 — implement the typed tool

## Verdict

**Needs changes before coding Step 2.** The Step 1 contract is now solid, but the Step 2 plan in `STATUS.md` is still only the prompt checklist. It does not yet spell out the implementation shape needed to connect the existing SDK-free registry, intervals client, request decoding, and version metadata safely.

## Findings

### 1. Missing dependency/wiring plan for the concrete tool registry

The repository currently has only the generic `tools.Registry` / `Registrar` interfaces (`internal/tools/registry.go`) and the MCP server accepts a registry via `mcp.Options.Registry`. There is no concrete default tools registry yet, and `internal/app/defaultStartServer` currently starts MCP without registering any tools.

Step 2 says “Register exactly `get_athlete_profile`”, but the plan does not say where that registration lives or how dependencies are injected.

**Required adjustment:** Add an explicit design before implementation, for example:

- `internal/tools/get_athlete_profile.go` defines the tool and handler.
- `internal/tools/registry.go` gains a concrete registry/constructor such as `NewRegistry(profileClient, version)` or equivalent.
- The registry takes an interface dependency, e.g. `GetAthleteProfile(ctx context.Context) (intervals.AthleteWithSportSettings, error)`, so tests can use a fake client and the handler can call the real intervals client later.
- App/MCP end-to-end wiring can be completed in Step 5 if desired, but Step 2 should leave a clear hook for it rather than a tool that cannot be registered outside tests.

### 2. No plan for `_meta.server_version` data flow

The accepted contract requires `_meta.server_version`, and typed response structs will likely be introduced in this step. The current code passes `Version` to the MCP server, but no version is available to `internal/tools` unless the registry/tool constructor receives it.

**Required adjustment:** Include version in the tool dependency plan now. If the registry constructor receives an empty version, normalize it to `dev`, matching `mcp.NewServer` behavior. This prevents Step 3 from having to refactor the tool API just to add required metadata.

### 3. Argument decoding should be strict, not just schema-described

The Step 1 contract allows only optional `include_full` and explicitly excludes credentials and v0.1 `athlete_id`. The Step 2 checklist mentions JSON Schema descriptions, but not runtime decoding behavior. Plain `json.Unmarshal` into a struct would silently ignore unknown fields such as `api_key` or `athlete_id`, which makes it harder to detect accidental credential passing and weakens the “minimal arguments” contract.

**Required adjustment:** Plan to:

- Publish an input schema with `type: object`, a described optional `include_full` boolean defaulting to false, and `additionalProperties: false`.
- Decode `req.Arguments` with `json.Decoder.DisallowUnknownFields()` or equivalent.
- Return a short `tools.NewUserError` for invalid arguments, e.g. `invalid get_athlete_profile arguments; only include_full is supported`.

### 4. Error handling path needs to use the existing public-error mechanism

The MCP adapter sanitizes handler errors and exposes only `tools.UserError` messages. The Step 2 plan should be explicit that upstream intervals errors are wrapped in `tools.NewUserError` with the public message from the Step 1 contract, not returned directly.

**Required adjustment:** Add to the Step 2 plan:

- On client failure, return `tools.NewUserError("could not fetch athlete profile; check intervals.icu credentials and athlete ID", err)`.
- Do not include upstream response bodies, API key/config values, request URLs, or raw athlete identifiers in the public message.
- Let the existing MCP adapter log the handler error; do not add ad-hoc logging that could leak secrets.

### 5. Response-shaping boundary between Step 2 and Step 3 is unclear

Step 2 says “Add typed request and response structs,” while Step 3 says “Shape the response.” Without a boundary, Step 2 may either under-implement unusable structs or accidentally complete Step 3 without tests.

**Required adjustment:** State the intended split. A workable split is:

- Step 2 creates the typed request/response structs, schema, registry entry, handler skeleton, context-aware client call, strict argument decode, and sanitized error path.
- Step 3 fills in/finishes the mapping from `intervals.AthleteWithSportSettings` to the exact contracted terse/full response fields, including units, pace key selection, normalized IDs, and `_meta`.

If Step 2 will already implement the mapper, say so and include the Step 1 field contract as the source of truth.

## Suggested Step 2 plan additions

Before starting code, update `STATUS.md` with a short design section covering:

1. Concrete files/functions/constructors to add or modify.
2. The profile-client interface and fakeability for tests.
3. How `server_version` reaches the tool response.
4. Strict JSON Schema and runtime argument validation.
5. Exact public error mapping via `tools.NewUserError`.
6. The Step 2/Step 3 implementation boundary.


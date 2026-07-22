# Public core library contract

`github.com/ricardocabral/icuvisor/pkg/icuvisor` is the public reuse boundary for hosts that need icuvisor's MCP core without copying packages from this repository's `internal/` tree.

The package is additive. The local CLI and existing internal packages remain the source implementation; the public facade converts small public types into internal types at the boundary and keeps implementation details private.

## Public import path

```go
import core "github.com/ricardocabral/icuvisor/pkg/icuvisor"
```

`icuvisor-host` and any future host integration should import this package instead of copying core packages or importing `github.com/ricardocabral/icuvisor/internal/...`.

## Exported API surface

The facade intentionally exposes a narrow host-safe surface:

- Intervals clients:
  - `NewAPIKeyClient(APIKeyClientOptions)` for local API-key Basic Auth.
  - `NewBearerClient(BearerClientOptions)` for hosted OAuth Bearer Auth.
  - `RetryConfig` for retry behavior.
- Runtime configuration and policy:
  - `Config` for API base URL, athlete ID, timezone, HTTP timeout, debug metadata, delete mode, and toolset.
  - `DeleteModeSafe`, `DeleteModeFull`, `DeleteModeNone`.
  - `ToolsetCore`, `ToolsetFull`.
- MCP construction:
  - `NewCoreRegistry(*Client, RegistryOptions)`.
  - `NewResourceRegistry(*Client, ResourceRegistryOptions)`.
  - `NewPromptRegistry()`.
  - `NewServer(context.Context, ServerOptions)`.
  - `Server.CatalogHash()`, `Run`, `ServeStreamableHTTP`, and `RunStreamableHTTP`.
  - `NewInvoker(context.Context, InvokerOptions)`, `Invoker.Tools()`, and `Invoker.Invoke()` for a non-MCP local view over the same safety-gated core catalog.
  - `RecentToolCallRecorder` for host or CLI diagnostics that record tool names without exposing arguments or payloads.
- Catalog helpers:
  - `CollectToolCatalog(context.Context, CatalogOptions)`.
  - `ComputeToolCatalogHash(context.Context, CatalogOptions)`.
- Host extension points:
  - `ToolFilter func(ToolInfo) bool` for OAuth-scope or tenant policy filtering.
  - `ExtraTools []Tool` for host-owned diagnostics such as setup status.
  - Public `Tool`, `ToolInfo`, `ToolRequest`, `ToolResult`, `Content`, and `TextResult` types.
- Streamable HTTP factory helpers:
  - `NewStreamableHTTPHandler(func(*http.Request) (*Server, error), StreamableHTTPHandlerOptions)`.
  - `StreamableHTTPPath` for the default `/mcp` endpoint.

The public types are not aliases of `internal/...` types. Callers should treat them as the compatibility contract and avoid depending on internal package names or representations.

## Public CLI adoption

The local `icuvisor` MCP binary and `icuvisor-cli` standalone direct-tool binary both construct the Intervals API-key client and core tool registry through `pkg/icuvisor`. The standalone view invokes registry handlers through `Invoker` rather than serializing an MCP request, while preserving the configured toolset and registration-time delete/write policy. This keeps the public facade exercised by both production views instead of only by host-side tests.

Coach mode remains on the existing internal wiring path for now because it still depends on internal selection-store and coach ACL machinery that is not part of the public host-safe contract. The public facade is covered by parity tests against the internal registries for catalog hashes/tool names, resources, prompts, delete-mode/toolset combinations, and API-key Basic Auth behavior.

## Hosted-safe semantics

The facade owns several safety behaviors required by stateless hosted MCP servers:

- Bearer auth uses an internal auth strategy that sends `Authorization: Bearer ...`, preserves User-Agent, retry, timeout, and base URL behavior, and does not require or store an API key.
- Athlete IDs are normalized at the public server/catalog boundary before MCP tool routing uses them.
- Public delete-mode/toolset values and extra-tool requirement/toolset metadata are validated. Invalid values fail closed instead of silently becoming write-enabled or core-visible defaults.
- `SkipRuntimeCatalogMetadata` is self-contained for facade-built core registries: `NewServer` computes the effective catalog hash and reinjects it into tool/resource shaping when the caller did not pre-supply a catalog hash.
- Tool responses, `icuvisor_check_server_version`, and the athlete-profile resource can use request-scoped catalog hashes without mutating process-global response metadata.
- `NewStreamableHTTPHandler` supports request-scoped server factories and stateless mode. Factory errors default to a short public message (`MCP authorization failed`) unless `FactoryErrorMessage` supplies a different safe message; detailed errors must stay in server logs.

## Ownership matrix

| Area | Public `icuvisor` core owns | Private `icuvisor-host` owns |
| --- | --- | --- |
| MCP protocol | Server construction, stdio/Streamable HTTP helpers, tool/resource/prompt registration | Request authentication, per-request principal resolution, hosted router wiring |
| Tools and resources | Core tool catalog, safety gates, response shaping, resources, prompts, catalog hash and stale-catalog diagnostics | OAuth-scope policy decisions, hosted-only setup/status diagnostics, per-tenant policy inputs |
| Intervals access | API-key and OAuth Bearer Intervals client transport hooks, retries, User-Agent, typed API client | OAuth session lifecycle, token refresh/decryption, user/grant storage |
| Data/security | No hosted secrets, no OAuth sessions, no Firestore/KMS/web settings | OAuth clients, sessions, Firestore/Secret Manager/KMS, hosted web settings and deployment |

## Host integration rules

- `icuvisor-host` should depend on `github.com/ricardocabral/icuvisor/pkg/icuvisor` for reusable core behavior.
- `icuvisor-host` must not copy core packages such as `internal/tools`, `internal/mcp`, `internal/intervals`, `internal/response`, `internal/resources`, or `internal/prompts`.
- `icuvisor-host` must not import `github.com/ricardocabral/icuvisor/internal/...`; Go's `internal` visibility rules make that invalid across modules.
- Hosted-only OAuth, session, store, Firestore, KMS, Secret Manager, web settings, and deploy code stays private to `icuvisor-host`.
- Do not expose API keys, OAuth tokens, cookies, raw hosted requests, or raw training payloads in docs, tests, logs, commits, or model-visible tool parameters.
- Do not copy real athlete identifiers into docs, tests, fixtures, logs, or commits. Model-visible `athlete_id` tool arguments are allowed when the public core schema requires athlete routing and the host has already authenticated the caller and enforced its ACLs; `athlete_id` is a routing selector, not a credential.

## Compatibility expectations

`pkg/icuvisor` is a public compatibility boundary. Additive changes are preferred. Breaking changes to exported names, option semantics, safety defaults, or response/catalog metadata behavior should be treated as compatibility-impacting and coordinated with `icuvisor-host` migrations.

Targeted verification for changes to this boundary:

```sh
go test ./internal/athleteprofile ./internal/resources ./internal/intervals ./internal/tools ./internal/toolexec ./internal/cli ./internal/mcp ./pkg/icuvisor
```

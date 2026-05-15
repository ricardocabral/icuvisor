# Post-update schema-change notification

MCP clients commonly cache a server's tool catalog for the lifetime of a conversation. If icuvisor is upgraded while a chat is still open, the running client may keep using the stale tool schema until the user starts a new conversation.

Every icuvisor tool response includes `_meta.server_version` and `_meta.catalog_hash`:

```json
{
  "_meta": {
    "server_version": "v0.5.0",
    "catalog_hash": "ab12cd..."
  }
}
```

`catalog_hash` is a deterministic SHA-256 over the exposed MCP tool catalog: tool names, tool descriptions, input schemas, and advertised output schemas after toolset/delete-mode registration filtering. Description-only changes are included because descriptions are part of the LLM-facing contract.

When icuvisor can tell that the catalog hash differs from the hash first seen by the current process/session, the response also includes:

```json
{
  "_meta": {
    "schema_changed": true,
    "schema_change_message": "icuvisor was upgraded from v0.4.1 to v0.5.0 since this conversation started; tool schemas may have changed. Open a new conversation to use the latest tools.",
    "previous_version": "v0.4.1",
    "current_version": "v0.5.0",
    "previous_catalog_hash": "9f3e22...",
    "catalog_hash": "ab12cd..."
  }
}
```

## Recommended user action

If an AI client or assistant reports `_meta.schema_changed: true`, open a new conversation in the MCP client. A new chat forces the client to fetch the latest tool catalog and use the current argument schemas.

## Limits

- Some MCP clients do not surface `_meta` back to the LLM or user. In those clients, icuvisor still sends the metadata, but the assistant may not be able to relay the warning.
- The current Go SDK response-shaping path does not expose a stable client session handle, so icuvisor tracks the first-seen catalog hash with a per-process fallback. Normal binary restarts reset process memory, so `schema_changed` is primarily a protocol guarantee and an integration-test seam until client/session resumption plumbing is available.
- The notification does not perform auto-update checks or release-channel polling. It only describes the schema state of responses produced by the running binary.

Check [`CHANGELOG.md`](../CHANGELOG.md) for release notes describing user-visible tool schema changes.

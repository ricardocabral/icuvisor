---
title: "After upgrading icuvisor"
description: "What to do when a client has a stale MCP tool schema after an icuvisor upgrade."
---

MCP clients commonly cache a server's tool catalog for the lifetime of a conversation. If icuvisor is upgraded while a chat is still open, the client may keep using the old tool schema until you start a new conversation.

## What icuvisor sends

Every icuvisor tool response includes response metadata like this:

```json
{
  "_meta": {
    "server_version": "v0.5.0",
    "catalog_hash": "ab12cd..."
  }
}
```

`catalog_hash` is a deterministic SHA-256 over the exposed MCP tool catalog: tool names, descriptions, input schemas, and advertised output schemas after toolset/delete-mode registration filtering.

When icuvisor can tell that the catalog hash differs from the hash first seen by the current process or session, the response also includes:

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

## What you should do

If an AI client or assistant reports `_meta.schema_changed: true`, open a new conversation in the MCP client. A new chat forces the client to fetch the latest tool catalog and use the current argument schemas. Reusing the old conversation can keep sending stale arguments even though the upgraded binary is running.

If your client hides `_meta`, reconnect or reload MCP tools, then ask the assistant to call `icuvisor_check_server_version` with no arguments. Compare the visible fields in that tool's description with the response:

- visible `description_server_version` → response `server_version`
- visible `description_catalog_fingerprint` → response `description_catalog_fingerprint`
- visible `description_toolset` → response `toolset`
- visible `description_delete_mode` → response `delete_mode`

If any of those pairs differ, the client is still showing stale tool descriptions: reconnect or reload MCP tools again, fully restart the client if needed, and start a new conversation if the current chat keeps using old schemas or assumptions. Do not compare `description_catalog_fingerprint` with the live `catalog_hash`; they are different diagnostic values.

If the client still shows old tools, stale timezone/date behavior, outdated zones, or write failures after the new chat, continue with the [stale conversation troubleshooting checklist]({{< relref "troubleshooting#stale-conversations-and-cached-tool-catalogs" >}}).

## Limits

- Some MCP clients do not surface `_meta` back to the assistant or user. icuvisor still sends the metadata, but the assistant may not relay the warning.
- Normal binary restarts reset process memory. The schema-change fields are primarily a protocol guarantee and integration-test seam until client/session resumption plumbing is available.
- This notification does not perform auto-update checks or release-channel polling. It only describes the schema state of responses produced by the running binary.
- `icuvisor_check_server_version` is local, read-only, and has no arguments. It does not contact intervals.icu, upload credentials, inspect files, or send telemetry; it returns catalog/version/toolset/delete-mode metadata from the running server.

Check the top-level [CHANGELOG](https://github.com/ricardocabral/icuvisor/blob/main/CHANGELOG.md) for user-visible tool schema changes.

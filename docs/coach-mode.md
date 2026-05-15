# Coach mode

Coach mode lets a locally held coach-scoped intervals.icu API key target multiple configured athletes without ever passing the credential through an MCP tool argument.

## Catalog-cache caveat

MCP clients may cache the tool catalog for the current conversation. `select_athlete` changes server-side routing immediately, and per-call `athlete_id` overrides are enforced immediately, but the model or client may not see a refreshed `tools/list` until a new conversation or reconnect.

When `select_athlete` returns `_meta.requires_new_conversation: true`, start a new conversation or reconnect the MCP client before relying on newly visible/hidden tools. This avoids stale catalog entries after switching between athletes with different ACLs. TP-040 is expected to add notifications so clients can refresh catalogs more gracefully.

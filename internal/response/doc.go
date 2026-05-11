// Package response owns shared MCP response-boundary shaping: JSON-tagged DTOs are
// converted to JSON values, null-valued object keys are stripped in terse mode,
// and row metadata is attached before values are serialized for clients. Tool
// responses must be object wrappers so response-level _meta.server_version has a
// single stable location; row collections inside wrappers are named through
// Options.RowCollections. Date/time presentation is rendered at this boundary in
// the athlete's configured IANA timezone using RenderTimeInTimezone.
//
// IncludeFull preserves JSON null values that reach this package. A struct field
// whose null must be visible in full mode must not use omitempty, or the caller
// must pass an explicit map value containing that nil; fields omitted by
// encoding/json cannot be reconstructed by the shaper.
package response

// Package safety owns the process-wide write/delete capability gate.
//
// The gate is resolved from ICUVISOR_DELETE_MODE once at startup and is used
// when registering tools. It deliberately exposes no per-call confirmation
// helper or confirm-style override because model-controlled arguments are not a
// credible safety boundary.
package safety

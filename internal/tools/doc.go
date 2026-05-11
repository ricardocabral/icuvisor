// Package tools implements icuvisor MCP tools.
//
// Activity splits use manual analyzed intervals when intervals.icu returns
// distance and duration for those rows. When no manual laps are available,
// get_activity_splits requests distance and time streams and linearly
// interpolates the elapsed time at each preferred split boundary: 1000 meters
// for metric athletes or 1609.344 meters for imperial athletes. If moving or
// paused-state samples are absent, virtual splits use elapsed stream time and do
// not attempt to remove stopped time. Missing, non-monotonic, or mismatched
// distance/time streams return an empty split list rather than inventing data.
package tools

# Code Review — Step 4: Tests

**Verdict: approve.**

The Step 4 change adds a registry-level regression test that registers the full production catalog with a no-network real `*intervals.Client`, asserts the exact tool-name set, and fails on duplicates. This covers the intended post-refactor risk: a future constructor/wiring change silently dropping or duplicating a tool when the registry uses the concrete intervals client path.

No blocking findings.

## Notes

- The test uses `newNoNetworkIntervalsClient(t)`, so registration remains guarded against accidental HTTP during catalog construction.
- The expected list includes the full 38-tool production surface, including the eight tools intentionally omitted from schema snapshots.
- The exact sorted slice comparison catches both name drift and count drift; the `seen` map catches duplicate registrations with a clearer failure.

## Verification

Commands run:

- `git diff 5071eb544d935089814f3ae49f85b350518e2bd6..HEAD --name-only`
- `git diff 5071eb544d935089814f3ae49f85b350518e2bd6..HEAD`
- `go test ./internal/tools` — passes
- `make test` — passes
- `make test-race` — passes
- `go run ./scripts/check_schema_stability.go` — passes (`snapshot freshness: PASS`)

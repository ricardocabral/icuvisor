# Plan Review — Step 1

## Verdict: REVISE

The plan does not cover all existing public guidance required by its completion criterion. In addition to the specified config reference, `web/content/guides/coach-mode.md:46` currently says that `12345` and `i12345` normalize to `i12345`. This is the same routing-sensitive false claim and would remain publicly published after the planned changes.

Resolve the scope conflict before implementation: either add that guide (and its assertion to the documentation contract) to Step 1's artifacts/file scope, or amend the completion criterion to explicitly limit the guarantee to the named pages. `CONTRIBUTING.md:73` also contains the obsolete canonicalization guidance and should be assessed or explicitly excluded.

Once that is resolved, the targeted contract should assert the exact behavior represented by `internal/config/athlete.go`: surrounding whitespace is trimmed; only an optional leading `I` is canonicalized to `i`; digits are validated; and neither accepted ID shape gains or loses its prefix. It should also require a hosted-connector link/URL and retain the explicit prohibition on generic public tunnels.

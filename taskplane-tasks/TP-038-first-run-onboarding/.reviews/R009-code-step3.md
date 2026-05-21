# Review R009 — Code Review for Step 3: Autodetect + verify

Verdict: **REVISE**

## Findings

1. **Detected timezone defaults to UTC on normal local-time configurations.**  
   `internal/app/setup.go:312-316` treats `time.Local.String() == "Local"` as an unusable value and immediately falls back to `config.DefaultTimezone` (`UTC`). On Go, `time.Local.String()` is commonly exactly `"Local"` even when the host is configured for a real zone (for example, this macOS worktree reports `date +%Z` as `-03` while a small Go program prints `time.Local.String() == "Local"`). That means `icuvisor setup` will prompt `Detected timezone: UTC. Use this? [Y/n]` for many first-run athletes, and the default answer writes the wrong timezone. Step 3 requires detecting the athlete's OS timezone and prompting to confirm an IANA zone. Please replace this fallback with a real IANA-zone detection path (for example, validate `TZ` when set and/or resolve platform localtime symlinks such as `/etc/localtime` under `zoneinfo`), and only use UTC as an explicit last-resort with clear user override behavior. Add a test that would fail for the current `"Local"` fallback.

## Notes

- The `/athlete/0/profile` client path and the `errors.Is(err, intervals.ErrUnauthorized)` mapping are in the right direction.
- `go test ./...` passes locally.

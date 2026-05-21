# R006 Code Review — Step 2: Backends

## Verdict: approved

No blocking findings.

The Step 2 backend implementation matches the approved plan: `OSKeychain()` wraps `the selected OS keyring module` behind an injectable adapter, uses the project-local service name, preserves `ErrNotFound` as the read fall-through sentinel, keeps unexpected backend errors wrapped with `%w`, and avoids logging secret values. Linux-specific unavailable-keychain classification is isolated behind build tags and only maps unavailable/headless lookup failures to `ErrNotFound` on `Get`; write/delete paths continue to surface backend failures.

The tests cover the important backend contract points for this step: success paths, upstream not-found mapping, unexpected error wrapping, context cancellation before backend calls, Linux headless degradation, non-Linux classifier behavior, and slog redaction for `Set`.

## Validation

- `git diff c53fe4c..HEAD --name-only`
- `git diff c53fe4c..HEAD`
- `go test ./...`
- `golangci-lint run ./internal/credstore`
- `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go test -c ./internal/credstore -o /tmp/credstore-linux.test`
- `GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go test -c ./internal/credstore -o /tmp/credstore-darwin.test`
- `GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go test -c ./internal/credstore -o /tmp/credstore-windows.test.exe`
- `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o /tmp/icuvisor-linux ./cmd/icuvisor`
- `GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o /tmp/icuvisor-darwin ./cmd/icuvisor`
- `GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o /tmp/icuvisor-windows.exe ./cmd/icuvisor`

## Non-blocking notes

- `go mod tidy` would reclassify `the selected OS keyring module` and `the D-Bus module` as direct dependencies and add checksum entries for the new module graph's test dependencies. This is not a Step 2 blocker, but it is worth tidying before merge if the project/CI enforces module-file cleanliness.
- Step 3 should keep using `errors.Is(err, credstore.ErrNotFound)` for fall-through behavior so wrapped backend failures do not get mistaken for missing credentials.

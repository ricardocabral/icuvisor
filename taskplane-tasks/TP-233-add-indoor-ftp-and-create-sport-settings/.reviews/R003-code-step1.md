# Code Review — TP-233 Step 1

## Verdict: APPROVE

The client adds `IndoorFTP` as a sparse update field and a separate, create-only parameter boundary. Creation uses the required no-query `POST` with `types:[sport]`; it cannot include IDs, recalculation, or zone replacement fields, and the shared POST transport retains its no-retry behavior. Threshold validation rejects invalid values before transport without inventing an indoor/outdoor FTP ordering rule, while pace values and supplied metadata remain canonical m/s/pass-through.

The local-server tests cover the exact update/create method, path, query, sparse bodies, response echo, canonical pace metadata, and zero-request validation failures.

Verified: `go test ./internal/intervals -run 'SportSettings' -count=1`, `go test ./internal/intervals -count=1`, `go test -race ./internal/intervals -run 'SportSettings' -count=1`, `go vet ./internal/intervals`, and `git diff --check 69c463cb9d4694007ab99168b625fdd5fe4f3f41..HEAD`.

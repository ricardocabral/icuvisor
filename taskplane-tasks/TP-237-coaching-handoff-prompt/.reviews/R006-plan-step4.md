# Plan Review — TP-237 Step 4

## Verdict: REVISE

The planned Go tests, eval validation, lint, binary build, and whitespace check cover the core implementation, but the verification plan omits two checks required by this change set:

1. Run `make test-race` (or `go test -race -count=1 ./...`) after `make test`. This task changes prompt-handler code, and the project guidance requires a local race run before delivery.
2. Run `make web-build`. New cookbook/front-matter content and `relref` links are not exercised by `make test`, `make eval-validate`, `make lint`, or `make build`; a Hugo build is needed to validate the published documentation surface.

Retain the existing required commands and `git diff --check`; add these two checks to Step 4 and fix any resulting failures.

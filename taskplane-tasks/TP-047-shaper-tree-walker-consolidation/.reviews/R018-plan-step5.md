# Plan Review — Step 5: Verify byte-identical output

**Decision: APPROVE.**

The Step 5 plan now includes the two guardrails needed for a byte-identical verification gate:

- Run only the golden snapshot comparison with update mode explicitly disabled:

  ```sh
  env -u UPDATE_RESPONSE_GOLDENS go test ./internal/response -run '^TestShapeGoldenSnapshots$' -count=1
  ```

- Require the committed baseline fixtures to remain unchanged:

  ```sh
  git diff --exit-code -- internal/response/testdata
  ```

This directly addresses the prior risk that `UPDATE_RESPONSE_GOLDENS=1` could mask output drift. The existing five fixtures cover the representative shapes required by the task, and keeping the broader build/test/race/lint work in Step 6 is appropriate.

Proceed with Step 5. If either command fails or any fixture changes, stop and resolve the drift before moving to Step 6.

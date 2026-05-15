# Plan Review — Step 5: Verify byte-identical output

**Decision: REVISE.**

Step 5 is intentionally small, but the current `STATUS.md` plan only repeats the high-level checkbox. Because the snapshot test has an update mode, the plan should spell out the exact verification command and guard against accidentally rewriting the baseline goldens.

## Blocking plan gap

1. **Pin the golden comparison to non-update mode.**  
   Add a Step 5 note/checklist item that runs the snapshot test with `UPDATE_RESPONSE_GOLDENS` explicitly unset, for example:

   ```sh
   env -u UPDATE_RESPONSE_GOLDENS go test ./internal/response -run '^TestShapeGoldenSnapshots$' -count=1
   ```

   This matters because `TestShapeGoldenSnapshots` writes fixtures when `UPDATE_RESPONSE_GOLDENS=1`; using update mode during Step 5 would mask exactly the byte drift this step is meant to catch.

2. **Require a clean testdata diff before proceeding.**  
   Add an explicit post-check such as:

   ```sh
   git diff --exit-code -- internal/response/testdata
   ```

   If either command fails or any golden fixture changes, stop and resolve the semantic drift before Step 6. Do not regenerate or accept new fixture bytes as part of Step 5 unless the investigation proves the Step 1 baseline itself was wrong and that rationale is documented.

## Non-blocking reminders

- The five fixtures already cover the required representative shapes: terse/full activities, fitness rows, wrapper rows, and provenance metadata.
- Step 5 does not need `make test`/race/lint; those remain Step 6. Running the focused response-package golden test is enough for the byte-identical gate.

Once the exact non-update command and clean-diff check are recorded in `STATUS.md`, the Step 5 plan should be safe to proceed.

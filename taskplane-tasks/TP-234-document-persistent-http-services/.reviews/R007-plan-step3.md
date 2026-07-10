# Plan Review — TP-234 Step 3

## Verdict: APPROVE

The R006 plan is concrete and addresses the Step 3 contract requirements. It defines a root-relative static test over fenced executable samples in both affected guides, including the plist's split XML arguments, while keeping explanatory warning prose outside the negative checks. It specifies the required three-platform, loopback, credential-store, lifecycle/recovery/removal, hosted HTTPS/OAuth, no-tunnel, and connector-key assertions.

The planned negative checks reject credential assignments/sources and plaintext keys as well as wildcard and any non-loopback HTTP bind in executable samples. Adding the test to `docs-guidance-test` is the appropriate CI integration; the existing Ubuntu CI job already invokes that target. The stated direct-script and Make-target verification completes the step plan.

# Plan review — TP-053 Step 2

Decision: **changes requested** before drafting the page.

The revised Step 2 plan fixes the two blockers from R003: it adds the missing Tutorials section/nav work and chooses a deterministic command path for the local-stdio ChatGPT JSON. One remaining issue is that the new build-from-source path still is not fully planned as something a fresh macOS reader can execute end-to-end.

## Findings

1. **Blocking — the build-from-source tutorial path still hides required toolchain prerequisites.**  
   The plan now installs by cloning to `/Users/Shared/icuvisor-src`, running `make build`, and copying the binary to `/Users/Shared/icuvisor/bin/icuvisor`. That path requires `git`, `make`/Xcode Command Line Tools, and a compatible Go toolchain; `go.mod` currently declares `go 1.25.10`. A fresh macOS machine does not have all of that by default, while the acceptance criteria say every command should work for a new user in about 10 minutes.

   Before drafting, update the plan to handle this explicitly. Either:

   - add the temporary build-from-source prerequisites to **What you'll need** and include exact preflight/install guidance that keeps the tutorial linear, or
   - mark the missing signed installer / build toolchain requirement as a dependency blocker instead of presenting the fallback as fresh-machine-ready.

   Do not leave this as an implicit failure the reader discovers at `git clone` or `make build`.

2. **Major — the newly chosen `/Users/Shared` install path was not the path validated in Step 1.**  
   `STATUS.md` records successful `make build` and Codex local-stdio validation using the worktree binary, then Step 2 introduces a new stable copy location under `/Users/Shared`. That is a good documentation choice, but the exact commands that create and run `/Users/Shared/icuvisor/bin/icuvisor` should be smoke-tested before they become the tutorial's copy-paste path.

   Add a Step 2 checklist item to run and record the exact sequence the page will show: clone into `/Users/Shared/icuvisor-src`, build, create `/Users/Shared/icuvisor/bin`, copy the binary, run `/Users/Shared/icuvisor/bin/icuvisor version`, and use that same absolute path in the setup command and ChatGPT JSON.

3. **Minor — make the website-nav change concrete.**  
   The plan says to add the Tutorials section/index/nav entry, which is directionally correct. Because the current site nav is hard-coded in the Hugo layouts, spell out that the implementation should add `web/content/tutorials/_index.md` and update every relevant header template, not just the new page, so the Step 4 nav check cannot pass on one page and fail on another.

4. **Minor — preserve the required code-fence detail.**  
   The prompt specifically asks for the ChatGPT JSON in a fenced `text` block, not a `json` block and not a Hugo shortcode. Add that to the drafting checklist so the page follows the tutorial instructions exactly.

Once these points are reflected in `STATUS.md`, the rest of the Step 2 plan is ready: it preserves the Diataxis tutorial shape, keeps the local-stdio honesty constraints from Step 1, avoids leaking secrets, and includes the required changelog update.

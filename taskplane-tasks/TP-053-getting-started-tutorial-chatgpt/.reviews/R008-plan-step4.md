# Plan review — TP-053 Step 4

Decision: **revise** before build verification.

The Step 4 plan covers the main requested checks: Hugo build, Tutorials navigation/list presence, Pagefind/search verification or limitation recording, and phone-width reading. However, it is missing one verification that is now essential for this task: a final privacy/content audit of the tutorial text itself, not just the screenshots.

## Blocking issue

- Add an explicit Markdown/generated-HTML privacy audit before marking Step 4 complete. The Step 1/3 records say real validation output must not leak into docs or screenshots, but the current tutorial text still contains the exact aggregate values recorded from the private Codex validation run (`11 activities`, training load `790`, `15 hours 24 minutes`, `312.19 km`) in `web/content/tutorials/getting-started-chatgpt.md`. Those are private training details from validation and should be replaced with clearly synthetic illustrative values before the build is accepted.

  The Step 4 plan should require a concrete check such as:

  ```bash
  grep -R "790\|312\.19\|15 hours 24\|11 activities" web/content web/static/img/tutorials/chatgpt web/public
  ```

  and should also state that any representative answer values in the page must be synthetic, not copied from validation logs.

## Tighten the verification plan

Please make these checks explicit rather than relying on the broad “documentation sanity checks” item:

- After `cd web && hugo --minify --gc`, verify the generated tutorial exists, for example `test -f public/tutorials/getting-started-chatgpt/index.html`.
- Verify all six referenced screenshots resolve in the built site, for example by grepping the generated HTML and checking the corresponding files under `web/public/img/tutorials/chatgpt/`.
- For Pagefind, either run the actual indexer if available, or record the missing Pagefind tooling limitation and do the accepted equivalent generated-site token check for this scaffold. The STATUS entry should name the exact commands and evidence.
- For phone-width review, record the viewport used and the result, not just that it was read.

Once the privacy audit is added and the generic sanity-check item is expanded into concrete evidence, the Step 4 plan will be ready to execute.
